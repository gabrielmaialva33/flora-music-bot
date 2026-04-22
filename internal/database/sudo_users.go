package database

func Sudoers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return state.Sudoers, nil
}

func IsSudoWithoutError(id int64) bool {
	is, _ := IsSudo(id)
	return is
}

func IsSudo(id int64) (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return contains(state.Sudoers, id), nil
}

func AddSudo(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var added bool
		s.Sudoers, added = addUnique(s.Sudoers, id)
		return added
	})
}

func RemoveSudo(id int64) error {
	return modifyBotState(func(s *BotState) bool {
		var removed bool
		s.Sudoers, removed = removeElement(s.Sudoers, id)
		return removed
	})
}
