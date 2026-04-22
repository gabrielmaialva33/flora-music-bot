package database

func ServedChats() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Chats, nil
}

func ServedUsers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Served.Users, nil
}

func IsServedChat(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedChatsMap[id]
	return ok, nil
}

func IsServedUser(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	_, ok := state.servedUsersMap[id]
	return ok, nil
}

func AddServedChat(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Served.Chats, added = addUnique(s.Served.Chats, id)
		if added {
			s.servedChatsMap[id] = struct{}{}
		}
		return added
	})
}

func AddServedUser(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Served.Users, added = addUnique(s.Served.Users, id)
		if added {
			s.servedUsersMap[id] = struct{}{}
		}
		return added
	})
}

func RemoveServedChat(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Served.Chats, removed = removeElement(s.Served.Chats, id)
		if removed {
			delete(s.servedChatsMap, id)
		}
		return removed
	})
}

func RemoveServedUser(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Served.Users, removed = removeElement(s.Served.Users, id)
		if removed {
			delete(s.servedUsersMap, id)
		}
		return removed
	})
}
