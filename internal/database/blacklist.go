package database

func BlacklistedUsers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), state.Blacklisted.Users...), nil
}

func BlacklistedChats() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), state.Blacklisted.Chats...), nil
}

func IsBlacklistedUser(userID int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Blacklisted.Users, userID), nil
}

func IsBlacklistedChat(chatID int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Blacklisted.Chats, chatID), nil
}

func AddBlacklistedUser(userID int64) error {
	return addToBotList(func(s *BotState) *[]int64 { return &s.Blacklisted.Users }, userID)
}

func RemoveBlacklistedUser(userID int64) error {
	return removeFromBotList(func(s *BotState) *[]int64 { return &s.Blacklisted.Users }, userID)
}

func AddBlacklistedChat(chatID int64) error {
	return addToBotList(func(s *BotState) *[]int64 { return &s.Blacklisted.Chats }, chatID)
}

func RemoveBlacklistedChat(chatID int64) error {
	return removeFromBotList(func(s *BotState) *[]int64 { return &s.Blacklisted.Chats }, chatID)
}
