package database

import (
	"context"
	"time"

	"github.com/Laky-64/gologging"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"main/internal/utils"
)

var (
	client           *mongo.Client
	database         *mongo.Database
	settingsColl     *mongo.Collection
	chatSettingsColl *mongo.Collection

	logger = gologging.GetLogger("Database")

	chatCache  = utils.NewCache[int64, *ChatSettings](60 * time.Minute)
	stateCache = utils.NewCache[string, *BotState](60 * time.Minute)

	// Inicializados em Init() (precisam das collections já resolvidas).
	botStore  *docStore[string, BotState]
	chatStore *docStore[int64, ChatSettings]
)

func Init(mongoURL string) (func(), error) {
	var err error
	logger.Debug("Initializing MongoDB...")
	client, err = mongo.Connect(options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}

	logger.Debug("Successfully connected to MongoDB.")

	database = client.Database("WinxMusic")
	settingsColl = database.Collection("bot_settings")
	chatSettingsColl = database.Collection("chat_settings")

	botStore = &docStore[string, BotState]{
		coll:        settingsColl,
		cache:       stateCache,
		idOf:        func(string) any { return "global" }, // _id do singleton
		makeDefault: func(string) *BotState { return newDefaultBotState() },
		afterLoad:   buildIndexes,
	}
	chatStore = &docStore[int64, ChatSettings]{
		coll:        chatSettingsColl,
		cache:       chatCache,
		idOf:        func(id int64) any { return id },
		makeDefault: func(id int64) *ChatSettings { return defaultChatSettings(id) },
	}

	migrateData()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Error while disconnecting MongoDB: %v", err)
		} else {
			logger.Info("MongoDB disconnected successfully")
		}
	}, nil
}
