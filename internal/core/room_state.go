package core

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
)

var (
	rooms   = make(map[int64]*RoomState)
	roomsMu sync.RWMutex

	ErrRoomDestroyed = errors.New("room destroyed")
)

// RoomState aggregates everything we track for a single chat room.
//
// All sub-structs below are protected by the single mutex mu; they carry no
// locks of their own. Only chatID (immutable) and destroyed (atomic) are
// accessed outside mu.
type RoomState struct {
	mu sync.RWMutex

	// Identity (immutable after creation)
	chatID int64

	// Grouped state, all guarded by mu
	pb   playbackState
	q    queueState
	meta roomMeta

	// Automation (embedded so its exported Remaining*Duration methods and its
	// promoted fields stay accessible exactly as before)
	*scheduledTimers

	// Arbitrary per-room data (exported)
	Data map[string]any

	// Internal Components
	Assistant *Assistant
	destroyed atomic.Bool
}

// playbackState holds everything about the currently loaded/playing track.
type playbackState struct {
	filePath  string
	track     *state.Track
	playing   bool
	paused    bool
	muted     bool
	speed     float64
	position  int
	updatedAt int64
	loop      int
}

// queueState holds the pending track queue and its ordering mode.
type queueState struct {
	queue   []*state.Track
	shuffle bool
}

// roomMeta holds UI/routing metadata for the room.
type roomMeta struct {
	statusMsg     *telegram.NewMessage
	channelPlayID int64
}

type scheduledTimers struct {
	scheduledUnmuteTimer *time.Timer
	scheduledResumeTimer *time.Timer
	scheduledSpeedTimer  *time.Timer

	scheduledUnmuteUntil time.Time
	scheduledResumeUntil time.Time
	scheduledSpeedUntil  time.Time
}

// Room management functions

func DeleteRoom(chatID int64) bool {
	_, file, line, _ := runtime.Caller(1)
	gologging.DebugF("DeleteRoom called from %s:%d", file, line)

	roomsMu.Lock()
	room, ok := rooms[chatID]
	if !ok || room == nil || room.destroyed.Load() {
		roomsMu.Unlock()
		return false
	}

	delete(rooms, chatID)
	roomsMu.Unlock()

	room.cleanupFile()
	room.Stop()
	room.destroyed.Store(true)
	return true
}

// GetRoom retrieves an existing room or creates a new one if requested.
func GetRoom(chatID int64, ass *Assistant, create bool) (*RoomState, bool) {
	roomsMu.RLock()
	room, exists := rooms[chatID]
	roomsMu.RUnlock()

	if exists {
		return room, true
	}

	if create {
		return createNewRoom(chatID, ass)
	}

	return nil, false
}

func createNewRoom(chatID int64, ass *Assistant) (*RoomState, bool) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	room, exists := rooms[chatID]
	if !exists {
		room = &RoomState{
			chatID:    chatID,
			q:         queueState{queue: []*state.Track{}},
			pb:        playbackState{speed: 1.0},
			Assistant: ass,
			Data:      make(map[string]any),
		}
		room.destroyed.Store(false)
		rooms[chatID] = room
	}

	return room, true
}

func GetAllRooms() map[int64]*RoomState {
	roomsMu.RLock()

	out := make(map[int64]*RoomState, len(rooms))
	var dead []int64

	for chatID, room := range rooms {
		if room == nil || room.destroyed.Load() {
			dead = append(dead, chatID)
			continue
		}
		out[chatID] = room
	}

	roomsMu.RUnlock()

	if len(dead) > 0 {
		roomsMu.Lock()
		for _, chatID := range dead {
			if room := rooms[chatID]; room == nil || room.destroyed.Load() {
				delete(rooms, chatID)
			}
		}
		roomsMu.Unlock()
	}

	return out
}

// Helpers

func (r *RoomState) IsDestroyed() bool {
	return r.destroyed.Load()
}

// roomGet runs read under the room's read lock, returning zero if the room is
// already destroyed.
func roomGet[T any](r *RoomState, zero T, read func() T) T {
	if r.IsDestroyed() {
		return zero
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return read()
}

// roomSet runs write under the room's write lock, no-op if the room is
// already destroyed.
func roomSet(r *RoomState, write func()) {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	write()
}

func (r *RoomState) updatePosition() {
	if r == nil || r.pb.track == nil || r.pb.updatedAt == 0 {
		return
	}

	current := time.Now().Unix()
	elapsed := float64(current - r.pb.updatedAt)

	if r.pb.playing && !r.pb.paused {
		r.pb.position += int(elapsed * r.pb.speed)
		if r.pb.position >= r.pb.track.Duration {
			r.pb.position = r.pb.track.Duration
			r.pb.playing = false
		}
	}
	r.pb.updatedAt = current
}

func (st *scheduledTimers) RemainingUnmuteDuration() time.Duration {
	if st == nil || st.scheduledUnmuteUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledUnmuteUntil)
}

func (st *scheduledTimers) RemainingResumeDuration() time.Duration {
	if st == nil || st.scheduledResumeUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledResumeUntil)
}

func (st *scheduledTimers) RemainingSpeedDuration() time.Duration {
	if st == nil || st.scheduledSpeedUntil.IsZero() {
		return 0
	}
	return time.Until(st.scheduledSpeedUntil)
}

func (st *scheduledTimers) cancelScheduledUnmute() {
	if st != nil && st.scheduledUnmuteTimer != nil {
		st.scheduledUnmuteTimer.Stop()
		st.scheduledUnmuteTimer = nil
		st.scheduledUnmuteUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledResume() {
	if st != nil && st.scheduledResumeTimer != nil {
		st.scheduledResumeTimer.Stop()
		st.scheduledResumeTimer = nil
		st.scheduledResumeUntil = time.Time{}
	}
}

func (st *scheduledTimers) cancelScheduledSpeed() {
	if st != nil && st.scheduledSpeedTimer != nil {
		st.scheduledSpeedTimer.Stop()
		st.scheduledSpeedTimer = nil
		st.scheduledSpeedUntil = time.Time{}
	}
}

// Getters

func (r *RoomState) EffectiveChatID() int64 {
	return roomGet(r, 0, func() int64 {
		if r.meta.channelPlayID != 0 {
			return r.meta.channelPlayID
		}
		return r.chatID
	})
}

func (r *RoomState) ChannelPlayID() int64 {
	return roomGet(r, 0, func() int64 { return r.meta.channelPlayID })
}

func (r *RoomState) ChatID() int64 {
	if r.IsDestroyed() {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.chatID
}

func (r *RoomState) FilePath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pb.filePath
}

func (r *RoomState) Loop() int {
	return roomGet(r, 0, func() int { return r.pb.loop })
}

func (r *RoomState) Position() int {
	return roomGet(r, 0, func() int { return r.pb.position })
}

func (r *RoomState) Queue() []*state.Track {
	if r.IsDestroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	q := make([]*state.Track, len(r.q.queue))
	copy(q, r.q.queue)
	return q
}

func (r *RoomState) Shuffle() bool {
	return roomGet(r, false, func() bool { return r.q.shuffle })
}

func (r *RoomState) Speed() float64 {
	return roomGet(r, 0, func() float64 { return r.pb.speed })
}

func (r *RoomState) Track() *state.Track {
	if r.IsDestroyed() {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pb.track
}

func (r *RoomState) StatusMsg() *telegram.NewMessage {
	return roomGet(r, nil, func() *telegram.NewMessage { return r.meta.statusMsg })
}

func (r *RoomState) GetData(k string) (bool, any) {
	type result struct {
		ok bool
		v  any
	}
	res := roomGet(r, result{}, func() result {
		v, ok := r.Data[k]
		return result{ok, v}
	})
	return res.ok, res.v
}

// Setters

func (r *RoomState) SetLoop(loop int) {
	roomSet(r, func() { r.pb.loop = loop })
}

func (r *RoomState) SetChannelPlayID(chatID int64) {
	roomSet(r, func() { r.meta.channelPlayID = chatID })
}

func (r *RoomState) SetData(k string, v any) {
	roomSet(r, func() {
		if r.Data == nil {
			r.Data = make(map[string]any)
		}
		r.Data[k] = v
	})
}

func (r *RoomState) DeleteData(k string) {
	roomSet(r, func() { delete(r.Data, k) })
}

func (r *RoomState) SetShuffle(enabled bool) {
	roomSet(r, func() { r.q.shuffle = enabled })
}

func (r *RoomState) SetStatusMsg(m *telegram.NewMessage) {
	roomSet(r, func() { r.meta.statusMsg = m })
}

// State checks

func (r *RoomState) IsActiveChat() bool {
	if r.IsDestroyed() {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
	return r.pb.track != nil && r.pb.playing
}

func (r *RoomState) IsPaused() bool {
	return roomGet(r, false, func() bool {
		return r.pb.paused && r.pb.track != nil && r.pb.playing
	})
}

func (r *RoomState) IsMuted() bool {
	return roomGet(r, false, func() bool {
		return r.pb.muted && r.pb.track != nil && r.pb.playing
	})
}

func (r *RoomState) Parse() {
	if r.IsDestroyed() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.updatePosition()
}
