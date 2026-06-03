package database

type AdminMode string

const (
	AdminModeAdminsOnly AdminMode = "admin"
	AdminModeAdminAuth  AdminMode = "adminauth"
	AdminModeEveryone   AdminMode = "everyone"
)

func GetAdminMode(chatID int64) (AdminMode, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return "", err
	}
	if settings.AdminMode == "" {
		return AdminModeAdminAuth, nil
	}
	return settings.AdminMode, nil
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
