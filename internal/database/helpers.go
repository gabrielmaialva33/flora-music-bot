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
