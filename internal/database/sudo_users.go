package database

func Sudoers() ([]int64, error) {
	state, err := getBotState()
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), state.Sudoers...), nil
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
	return addToBotList(func(s *BotState) *[]int64 { return &s.Sudoers }, id)
}

func RemoveSudo(id int64) error {
	return removeFromBotList(func(s *BotState) *[]int64 { return &s.Sudoers }, id)
}
