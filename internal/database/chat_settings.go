package database

type RTMPConfig struct {
	URL string `bson:"rtmp_url"`
	Key string `bson:"rtmp_key"`
}

type ChatSettings struct {
	ChatID                int64      `bson:"_id"`
	ChannelPlayID         int64      `bson:"cplay_id"`
	AuthUsers             []int64    `bson:"auth_users"`
	AdminMode             AdminMode  `bson:"admin_mode,omitempty"`
	Language              string     `bson:"language"`
	RTMP                  RTMPConfig `bson:"rtmp_config"`
	AssistantIndex        int        `bson:"ass_index,omitempty"`
	ThumbnailsDisabled    bool       `bson:"no_thumb"`
	PlayModeAdminsOnly    bool       `bson:"play_mode"`
	CommandDelete         bool       `bson:"cmd_delete"`
	CleanMode             bool       `bson:"clean_mode"`
	CleanModeDurationMins int        `bson:"clean_mode_duration_mins"`
}

func defaultChatSettings(chatID int64) *ChatSettings {
	return &ChatSettings{
		ChatID:                chatID,
		AuthUsers:             []int64{},
		CleanModeDurationMins: 15,
	}
}

// GetChatSettings returns the (cached) settings for a chat, creating defaults if absent.
func GetChatSettings(chatID int64) (*ChatSettings, error) {
	return getChatSettings(chatID)
}

// UpdateChatSettings persists the given chat settings.
func UpdateChatSettings(settings *ChatSettings) error {
	return updateChatSettings(settings)
}

func getChatSettings(chatID int64) (*ChatSettings, error) {
	return chatStore.get(chatID)
}

func updateChatSettings(settings *ChatSettings) error {
	return chatStore.update(settings.ChatID, settings)
}

func modifyChatSettings(chatID int64, fn func(*ChatSettings) bool) error {
	return chatStore.modify(chatID, fn)
}
