package database

// ThumbnailsDisabled returns whether thumbnails are disabled for the chat.
// Returns false by default (thumbnails enabled).
func ThumbnailsDisabled(chatID int64) (bool, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return false, err
	}
	return settings.ThumbnailsDisabled, nil
}

// SetThumbnailsDisabled sets whether thumbnails should be disabled for the chat.
func SetThumbnailsDisabled(chatID int64, disabled bool) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.ThumbnailsDisabled == disabled {
			return false
		}
		s.ThumbnailsDisabled = disabled
		return true
	})
}
