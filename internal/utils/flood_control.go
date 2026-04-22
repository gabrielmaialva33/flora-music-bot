package utils

import (
	"sync"
	"time"
)

var (
	floodMap = make(map[string]time.Time)
	floodMu  sync.RWMutex
)

// GetFlood returns remaining cooldown for a key.
func GetFlood(key string) time.Duration {
	floodMu.RLock()
	t, ok := floodMap[key]
	floodMu.RUnlock()

	if !ok {
		return 0
	}

	remaining := time.Until(t)

	if remaining <= 0 {
		floodMu.Lock()
		delete(floodMap, key)
		floodMu.Unlock()
	}

	return remaining
}

// SetFlood sets cooldown duration for a key.
func SetFlood(key string, duration time.Duration) {
	floodMu.Lock()
	floodMap[key] = time.Now().Add(duration)
	floodMu.Unlock()
}
