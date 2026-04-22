package database

func SetMaintenance(enabled bool, reason ...string) error {
	return modifyBotState(func(s *BotState) bool {
		changed := false
		if s.Maintenance.Enabled != enabled {
			s.Maintenance.Enabled = enabled
			changed = true
		}

		newReason := ""
		if enabled && len(reason) > 0 {
			newReason = reason[0]
		}

		if s.Maintenance.Reason != newReason {
			s.Maintenance.Reason = newReason
			changed = true
		}
		return changed
	})
}

func MaintenanceReason() (string, error) {
	state, err := getBotState()
	if err != nil {
		return "", err
	}
	return state.Maintenance.Reason, nil
}

func IsMaintenanceEnabled() (bool, error) {
	state, err := getBotState()
	if err != nil {
		return false, err
	}
	return state.Maintenance.Enabled, nil
}
