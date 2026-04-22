package database

func IsAuthorized(chatID, userID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return contains(settings.AuthUsers, userID), nil
}

func Authorize(chatID, userID int64) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		var added bool
		s.AuthUsers, added = addUnique(s.AuthUsers, userID)
		return added
	})
}

func Unauthorize(chatID, userID int64) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		var removed bool
		s.AuthUsers, removed = removeElement(s.AuthUsers, userID)
		return removed
	})
}

func AuthorizedUsers(chatID int64) ([]int64, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return nil, err
	}
	return settings.AuthUsers, nil
}
