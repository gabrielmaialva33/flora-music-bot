package core

import (
	"math/rand"
	"time"

	state "main/internal/core/models"
)

// NextTrack retrieves and prepares the next track in queue
func (r *RoomState) NextTrack() *state.Track {
	if r.IsDestroyed() {
		return nil
	}

	r.mu.Lock()

	if r.track != nil && r.loop > 0 {
		r.position = 0
		r.playing = false
		r.paused = false
		r.muted = false
		r.loop--
		r.updatedAt = time.Now().Unix()
		t := r.track
		r.mu.Unlock()
		return t
	}

	// Captura o track atual pra liberar o arquivo FORA do lock. releaseTrackFile
	// inspeciona as outras rooms (roomsMu + room.mu de cada uma); fazer isso sob
	// r.mu daria deadlock de ordem inversa entre duas rooms trocando de faixa ao
	// mesmo tempo (A segura rA.mu e quer rB.mu; B segura rB.mu e quer rA.mu).
	oldTrack := r.track
	oldChatID := r.chatID

	if len(r.queue) == 0 {
		r.mu.Unlock()
		releaseTrackFile(oldTrack, oldChatID)
		return nil
	}

	index := 0
	if r.shuffle {
		index = rand.Intn(len(r.queue))
	}

	next := r.queue[index]
	r.queue = append(r.queue[:index], r.queue[index+1:]...)

	r.track = next
	r.position = 0
	r.playing = false
	r.paused = false
	r.muted = false
	r.updatedAt = time.Now().Unix()

	r.mu.Unlock()
	releaseTrackFile(oldTrack, oldChatID)
	return next
}

// RemoveFromQueue removes track(s) from queue
func (r *RoomState) RemoveFromQueue(index int) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if index == -1 {
		r.queue = []*state.Track{}
		return
	}

	if index >= 0 && index < len(r.queue) {
		r.queue = append(r.queue[:index], r.queue[index+1:]...)
	}
}

// MoveInQueue moves a track from one position to another
func (r *RoomState) MoveInQueue(from, to int) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if from < 0 || from >= len(r.queue) ||
		to < 0 || to >= len(r.queue) ||
		from == to {
		return
	}

	item := r.queue[from]
	r.queue = append(r.queue[:from], r.queue[from+1:]...)

	if to >= len(r.queue) {
		r.queue = append(r.queue, item)
	} else {
		r.queue = append(r.queue[:to], append([]*state.Track{item}, r.queue[to:]...)...)
	}
}

// AddTracksToQueue appends multiple tracks to the queue
func (r *RoomState) AddTracksToQueue(tracks []*state.Track) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue = append(r.queue, tracks...)
}
