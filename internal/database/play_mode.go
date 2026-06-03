package database

func PlayModeAdminsOnly(chatID int64) (bool, error) {
	return getChatField(chatID, func(s *ChatSettings) bool { return s.PlayModeAdminsOnly })
}

func SetPlayModeAdminsOnly(chatID int64, adminsOnly bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.PlayModeAdminsOnly == adminsOnly {
			return false
		}
		s.PlayModeAdminsOnly = adminsOnly
		return true
	})
}
