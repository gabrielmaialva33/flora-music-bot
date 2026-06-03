package database

func IsAuthorized(chatID, userID int64) (bool, error) {
	return getChatField(chatID, func(s *ChatSettings) bool { return contains(s.AuthUsers, userID) })
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
	// defensive copy so callers can't mutate the cached slice in place.
	return getChatField(chatID, func(s *ChatSettings) []int64 {
		return append([]int64(nil), s.AuthUsers...)
	})
}
