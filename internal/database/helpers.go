package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var upsertOpt = options.UpdateOne().SetUpsert(true)

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// getChatField fetches the chat settings and returns a single derived field,
// removing the get→return-field boilerplate from the per-field getters.
func getChatField[T any](chatID int64, pick func(*ChatSettings) T) (T, error) {
	s, err := getChatSettings(chatID)
	if err != nil {
		var z T
		return z, err
	}
	return pick(s), nil
}

// addToBotList adds an id to a bot-state list selected by pick, persisting only
// when the id was actually appended.
func addToBotList(pick func(*BotState) *[]int64, id int64) error {
	return modifyBotState(func(s *BotState) bool {
		list := pick(s)
		var added bool
		*list, added = addUnique(*list, id)
		return added
	})
}

// removeFromBotList removes an id from a bot-state list selected by pick,
// persisting only when the id was actually removed.
func removeFromBotList(pick func(*BotState) *[]int64, id int64) error {
	return modifyBotState(func(s *BotState) bool {
		list := pick(s)
		var removed bool
		*list, removed = removeElement(*list, id)
		return removed
	})
}

// addUnique adds an element to a slice if it's not already present.
// Returns the new slice and true if the element was added.
func addUnique[T comparable](slice []T, element T) ([]T, bool) {
	for _, v := range slice {
		if v == element {
			return slice, false
		}
	}
	return append(slice, element), true
}

// removeElement removes an element from a slice if it's present.
// Returns the new slice and true if the element was removed.
func removeElement[T comparable](slice []T, element T) ([]T, bool) {
	for i, v := range slice {
		if v == element {
			return append(slice[:i], slice[i+1:]...), true
		}
	}
	return slice, false
}

// contains checks if a slice contains an element.
func contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}
