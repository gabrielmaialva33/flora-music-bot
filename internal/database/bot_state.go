package database

type UsersChats struct {
	Users []int64 `bson:"users"`
	Chats []int64 `bson:"chats"`
}

type Maintenance struct {
	Enabled bool   `bson:"enabled,omitempty"`
	Reason  string `bson:"reason,omitempty"`
}

type BotState struct {
	ID            string      `bson:"_id"`
	Served        UsersChats  `bson:"served"`
	Sudoers       []int64     `bson:"sudoers"`
	Blacklisted   UsersChats  `bson:"blacklisted"`
	AutoLeave     bool        `bson:"autoleave"`
	LoggerEnabled bool        `bson:"logger"`
	Maintenance   Maintenance `bson:"maint,omitempty"`

	// runtime indexes for fast lookup
	servedUsersMap map[int64]struct{} `bson:"-"`
	servedChatsMap map[int64]struct{} `bson:"-"`
}

const botStateCacheKey = "bot_state"

func newDefaultBotState() *BotState {
	s := &BotState{
		ID: "global",
		Served: UsersChats{
			Users: []int64{},
			Chats: []int64{},
		},
		Blacklisted: UsersChats{
			Users: []int64{},
			Chats: []int64{},
		},
		Sudoers:       []int64{},
		LoggerEnabled: true,
	}

	buildIndexes(s)
	return s
}

// BotState é singleton: cache key "bot_state", _id "global" no Mongo (ver botStore).
func getBotState() (*BotState, error) { return botStore.get(botStateCacheKey) }

func updateBotState(newState *BotState) error {
	return botStore.update(botStateCacheKey, newState)
}

func modifyBotState(fn func(*BotState) bool) error {
	return botStore.modify(botStateCacheKey, fn)
}

func buildIndexes(s *BotState) {
	s.servedUsersMap = make(map[int64]struct{}, len(s.Served.Users))
	for _, u := range s.Served.Users {
		s.servedUsersMap[u] = struct{}{}
	}

	s.servedChatsMap = make(map[int64]struct{}, len(s.Served.Chats))
	for _, c := range s.Served.Chats {
		s.servedChatsMap[c] = struct{}{}
	}
}
