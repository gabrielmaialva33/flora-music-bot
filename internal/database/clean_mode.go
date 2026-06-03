package database

func CleanMode(chatID int64) (bool, error) {
	return getChatField(chatID, func(s *ChatSettings) bool { return s.CleanMode })
}

func SetCleanMode(chatID int64, enabled bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.CleanMode == enabled {
			return false
		}
		s.CleanMode = enabled
		return true
	})
}
