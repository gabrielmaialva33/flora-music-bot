package database

func LinkedChannel(chatID int64) (int64, error) {
	return getChatField(chatID, func(s *ChatSettings) int64 { return s.ChannelPlayID })
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
