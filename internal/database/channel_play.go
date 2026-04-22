package database

func LinkedChannel(chatID int64) (int64, error) {
	settings, err := getChatSettings(chatID)
	if err != nil {
		return 0, err
	}
	return settings.ChannelPlayID, nil
}

func LinkChannel(chatID, channelID int64) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.ChannelPlayID == channelID {
			return false
		}
		s.ChannelPlayID = channelID
		return true
	})
}
