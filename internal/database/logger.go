package database

func IsLoggerEnabled() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.LoggerEnabled, nil
}

func SetLoggerEnabled(enabled bool) error {
	return modifyBotState(func(s *BotState) bool {
		if s.LoggerEnabled == enabled {
			return false
		}
		s.LoggerEnabled = enabled
		return true
	})
}
