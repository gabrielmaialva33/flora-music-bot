package core

import (
	"os"
	"path/filepath"

	"github.com/Laky-64/gologging"

	state "main/internal/core/models"
)

// check if a track is used in any room (other than the given room).
// Deve ser chamado sob roomsMu (R)Lock e SEM segurar nenhum room.mu — usa os
// getters travados Track()/Queue() de cada room, então segurar um room.mu aqui
// inverteria a ordem de lock e poderia travar.
func isTrackUsed(trackID string, skipChatID int64) bool {
	for chatID, room := range rooms {
		if room == nil || chatID == skipChatID {
			continue
		}

		if t := room.Track(); t != nil && t.ID == trackID {
			return true
		}

		if isTrackInQueue(trackID, room.Queue()) {
			return true
		}
	}
	return false
}

// checks first N (2) queued tracks
func isTrackInQueue(trackID string, queue []*state.Track) bool {
	limit := 2
	if len(queue) < limit {
		limit = len(queue)
	}

	for _, q := range queue[:limit] {
		if q != nil && q.ID == trackID {
			return true
		}
	}
	return false
}

// releaseTrackFile remove o arquivo de track se ele não estiver em uso em nenhuma
// outra room. DEVE ser chamado sem segurar nenhum room.mu (ver isTrackUsed).
func releaseTrackFile(track *state.Track, skipChatID int64) {
	if track == nil {
		return
	}

	roomsMu.RLock()
	used := isTrackUsed(track.ID, skipChatID)
	roomsMu.RUnlock()

	if used {
		gologging.DebugF(
			"file still in use, skipped remove: %s:%s",
			string(track.Source),
			track.ID,
		)
		return
	}

	findAndRemove(track)
}

// cleanup current + queued track files if unused.
// Chamado por DeleteRoom, já fora de roomsMu e de r.mu — usa os getters travados.
func (r *RoomState) cleanupFile() {
	if r == nil {
		return
	}

	chatID := r.ChatID()

	// collect current + next tracks (max 2)
	tracks := []*state.Track{}
	if t := r.Track(); t != nil {
		tracks = append(tracks, t)
	}
	tracks = append(tracks, r.Queue()...)
	if len(tracks) > 2 {
		tracks = tracks[:2]
	}

	for _, t := range tracks {
		if t == nil || t.ID == "" {
			continue
		}

		roomsMu.RLock()
		used := isTrackUsed(t.ID, chatID)
		roomsMu.RUnlock()

		if used {
			gologging.DebugF(
				"track still in use, skip delete: %s:%s",
				string(t.Source),
				t.ID,
			)
			continue
		}

		findAndRemove(t)
	}
}

func findAndRemove(track *state.Track) {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+state.SafeFileID(track.ID)+"*"))
	if err != nil {
		return
	}

	for _, f := range files {
		os.Remove(f)
		gologging.DebugF(
			"removed unused file: %s",
			f,
		)
	}
}
