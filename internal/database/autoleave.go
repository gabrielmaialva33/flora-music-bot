package database

func AutoLeave() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.AutoLeave, nil
}

func SetAutoLeave(value bool) error {
	return modifyBotState(func(s *BotState) bool {
		if s.AutoLeave == value {
			return false
		}
		s.AutoLeave = value
		return true
	})
}
