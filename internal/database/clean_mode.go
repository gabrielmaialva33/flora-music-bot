package database

func CleanMode(chatID int64) (bool, error) {
	s, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return s.CleanMode, nil
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
