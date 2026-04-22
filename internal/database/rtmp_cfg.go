package database

func RTMP(chatID int64) (string, string, error) {
	s, err := getChatSettings(chatID)
	if err != nil {
		return "", "", err
	}
	return s.RTMP.URL, s.RTMP.Key, nil
}

func SetRTMP(chatID int64, url, key string) error {
	return modifyChatSettings(chatID, func(s *ChatSettings) bool {
		if s.RTMP.URL == url && s.RTMP.Key == key {
			return false
		}
		s.RTMP.URL = url
		s.RTMP.Key = key
		return true
	})
}
