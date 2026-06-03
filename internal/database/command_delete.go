package database

func CommandDelete(chatID int64) (bool, error) {
	return getChatField(chatID, func(s *ChatSettings) bool { return s.CommandDelete })
}

func SetCommandDelete(chatID int64, enabled bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.CommandDelete == enabled {
			return false
		}
		s.CommandDelete = enabled
		return true
	})
}
