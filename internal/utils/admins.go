package utils

import (
	"slices"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

var adminCache = NewCache[int64, []int64](30 * time.Minute)

// Checks if a user is an admin in a chat

func IsChatAdmin(c *telegram.Client, chatID, userID int64) (bool, error) {
	if chatID == userID { // chat anon admin or pvt chat
		return true, nil
	}

	ids, ok := adminCache.Get(chatID)
	if ok {
		return slices.Contains(ids, userID), nil
	}

	ids, err := ReloadChatAdmin(c, chatID)
	if err != nil {
		return false, err
	}

	return slices.Contains(ids, userID), nil
}

// Reloads the chat admins from Telegram and updates the cache

func ReloadChatAdmin(c *telegram.Client, chatID int64) ([]int64, error) {
	ids, err := fetchAdmins(c, chatID)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		adminCache.Delete(chatID)
	} else {
		adminCache.Set(chatID, ids)
	}

	return ids, nil
}

// Adds a user to the cached admin list, auto-reloading if cache is missing

func AddChatAdmin(c *telegram.Client, chatID, userID int64) error {
	ids, ok := adminCache.Get(chatID)
	if !ok || len(ids) == 0 {
		var err error
		ids, err = ReloadChatAdmin(c, chatID)
		if err != nil {
			return err
		}
	}

	if !slices.Contains(ids, userID) {
		ids = append(ids, userID)
		adminCache.Set(chatID, ids)
	}

	return nil
}

// Removes a user from the cached admin list, auto-reloading if cache is missing

func RemoveChatAdmin(c *telegram.Client, chatID, userID int64) error {
	ids, ok := adminCache.Get(chatID)
	if !ok || len(ids) == 0 {
		var err error
		ids, err = ReloadChatAdmin(c, chatID)
		if err != nil {
			return err
		}
	}

	newIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id != userID {
			newIDs = append(newIDs, id)
		}
	}

	if len(newIDs) == 0 {
		adminCache.Delete(chatID)
	} else {
		adminCache.Set(chatID, newIDs)
	}

	return nil
}

// Fetches admins from Telegram
func fetchAdmins(c *telegram.Client, chatID int64) ([]int64, error) {
	admins, _, err := c.GetChatMembers(chatID, &telegram.ParticipantOptions{
		Filter:           &telegram.ChannelParticipantsAdmins{},
		SleepThresholdMs: 3000,
		Limit:            -1,
	})
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, p := range admins {
		if p.User.Bot || p.User.Deleted {
			continue
		}
		ids = append(ids, p.User.ID)
	}
	return ids, nil
}

var ownerCache = NewCache[int64, int64](30 * time.Minute)

// GetChatOwner returns the creator's user ID for a chat, using a short-lived cache.
// Returns (0, nil) when no creator could be determined.
func GetChatOwner(c *telegram.Client, chatID int64) (int64, error) {
	if ownerID, ok := ownerCache.Get(chatID); ok && ownerID != 0 {
		return ownerID, nil
	}

	admins, _, err := c.GetChatMembers(chatID, &telegram.ParticipantOptions{
		Filter:           &telegram.ChannelParticipantsAdmins{},
		SleepThresholdMs: 3000,
		Limit:            -1,
	})
	if err != nil {
		return 0, err
	}

	for _, p := range admins {
		if p.Status == telegram.Creator && p.User != nil {
			ownerCache.Set(chatID, p.User.ID)
			return p.User.ID, nil
		}
	}
	return 0, nil
}
