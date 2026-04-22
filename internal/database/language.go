package database

import "main/internal/config"

func Language(chatID int64) (string, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return config.DefaultLang, err
	}
	if settings.Language == "" {
		return config.DefaultLang, nil
	}
	return settings.Language, nil
}

func SetLanguage(chatID int64, lang string) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.Language == lang {
			return false
		}
		s.Language = lang
		return true
	})
}
