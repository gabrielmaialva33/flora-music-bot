package database

func CommandDelete(chatID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return settings.CommandDelete, nil
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
