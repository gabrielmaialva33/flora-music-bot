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

	if r.pb.track != nil && r.pb.loop > 0 {
		r.pb.position = 0
		r.pb.playing = false
		r.pb.paused = false
		r.pb.muted = false
		r.pb.loop--
		r.pb.updatedAt = time.Now().Unix()
		t := r.pb.track
		r.mu.Unlock()
		return t
	}

	// Captura o track atual pra liberar o arquivo FORA do lock. releaseTrackFile
	// inspeciona as outras rooms (roomsMu + room.mu de cada uma); fazer isso sob
	// r.mu daria deadlock de ordem inversa entre duas rooms trocando de faixa ao
	// mesmo tempo (A segura rA.mu e quer rB.mu; B segura rB.mu e quer rA.mu).
	oldTrack := r.pb.track
	oldChatID := r.chatID

	if len(r.q.queue) == 0 {
		r.mu.Unlock()
		releaseTrackFile(oldTrack, oldChatID)
		return nil
	}

	index := 0
	if r.q.shuffle {
		index = rand.Intn(len(r.q.queue))
	}

	next := r.q.queue[index]
	r.q.queue = append(r.q.queue[:index], r.q.queue[index+1:]...)

	r.pb.track = next
	r.pb.position = 0
	r.pb.playing = false
	r.pb.paused = false
	r.pb.muted = false
	r.pb.updatedAt = time.Now().Unix()

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
		r.q.queue = []*state.Track{}
		return
	}

	if index >= 0 && index < len(r.q.queue) {
		r.q.queue = append(r.q.queue[:index], r.q.queue[index+1:]...)
	}
}

// MoveInQueue moves a track from one position to another
func (r *RoomState) MoveInQueue(from, to int) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if from < 0 || from >= len(r.q.queue) ||
		to < 0 || to >= len(r.q.queue) ||
		from == to {
		return
	}

	item := r.q.queue[from]
	r.q.queue = append(r.q.queue[:from], r.q.queue[from+1:]...)

	if to >= len(r.q.queue) {
		r.q.queue = append(r.q.queue, item)
	} else {
		r.q.queue = append(r.q.queue[:to], append([]*state.Track{item}, r.q.queue[to:]...)...)
	}
}

// AddTracksToQueue appends multiple tracks to the queue
func (r *RoomState) AddTracksToQueue(tracks []*state.Track) {
	if r.IsDestroyed() {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.q.queue = append(r.q.queue, tracks...)
}
