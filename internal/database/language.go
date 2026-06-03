package database

import "main/internal/config"

func Language(chatID int64) (string, error) {
	lang, err := getChatField(chatID, func(s *ChatSettings) string { return s.Language })
	if err != nil || lang == "" {
		return config.DefaultLang, err
	}
	return lang, nil
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
