package database

type AdminMode string

const (
	AdminModeAdminsOnly AdminMode = "admin"
	AdminModeAdminAuth  AdminMode = "adminauth"
	AdminModeEveryone   AdminMode = "everyone"
)

func GetAdminMode(chatID int64) (AdminMode, error) {
	mode, err := getChatField(chatID, func(s *ChatSettings) AdminMode { return s.AdminMode })
	if err != nil {
		return "", err
	}
	if mode == "" {
		return AdminModeAdminAuth, nil
	}
	return mode, nil
}

func SetAdminMode(chatID int64, mode AdminMode) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.AdminMode == mode {
			return false
		}
		s.AdminMode = mode
		return true
	})
}
