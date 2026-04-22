package platforms

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
)

func getPath(track *state.Track, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	mediaType := "audio"
	if track.Video {
		mediaType = "video"
	}

	filename := mediaType + "_" + track.ID + ext

	return filepath.Join("downloads", filename)
}

func fileExists(path string) bool {
	i, err := os.Stat(path)
	if err != nil {
		gologging.ErrorF("os.Stat: %v", err)
		return false
	}

	return i.Size() > 0
}

func findFile(track *state.Track) string {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		gologging.ErrorF("filepath.Glob: %v", err)
		return ""
	}

	for _, f := range files {
		if i, err := os.Stat(f); err == nil && i.Size() > 0 {
			return f
		}
	}

	return ""
}

func findAndRemove(track *state.Track) {
	t := "audio"
	if track.Video {
		t = "video"
	}

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+track.ID+"*"))
	if err != nil {
		return
	}

	for _, f := range files {
		os.Remove(f)
	}
}

func sanitizeAPIError(err error, apiKey string) error {
	if err == nil || apiKey == "" {
		return err
	}
	masked := strings.ReplaceAll(err.Error(), apiKey, "***REDACTED***")
	return errors.New(masked)
}

func playableMedia(m *telegram.NewMessage) (bool, bool) {
	if m == nil {
		return false, false
	}

	check := func(msg *telegram.NewMessage) (bool, bool) {
		switch {
		case msg.Audio() != nil, msg.Voice() != nil:
			return false, true

		case msg.Video() != nil:
			return true, false

		case msg.Document() != nil:
			mimeType := strings.ToLower(msg.Document().MimeType)

			if mimeType == "" {
				return false, false
			}

			switch {
			case strings.HasPrefix(mimeType, "audio/"):
				return false, true
			case strings.HasPrefix(mimeType, "video/"):
				return true, false
			}
		}

		return false, false
	}

	if m.IsReply() {
		rmsg, err := m.GetReplyMessage()
		if err != nil {
			return false, false
		}
		return check(rmsg)
	}

	return check(m)
}
