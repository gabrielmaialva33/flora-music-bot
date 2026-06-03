package platforms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"
	"resty.dev/v3"

	state "main/internal/core/models"
)

// newHTTPClient cria um resty.Client isolado. No resty v3 SetTimeout/SetHeader
// mutam o client in-place, então cada consumidor com requisitos distintos
// (timeout/User-Agent) deve ter o seu — compartilhar e remutar geraria race e
// corromperia config alheia. Se userAgent != "", fixa o UA uma vez aqui em vez
// de repetir SetHeader em cada request no hot path.
func newHTTPClient(timeout time.Duration, userAgent string) *resty.Client {
	c := resty.New().SetTimeout(timeout)
	if userAgent != "" {
		c.SetHeader("User-Agent", userAgent)
	}
	return c
}

// runYtdlpJSON roda `yt-dlp -j --flat-playlist` na URL e faz o parsing
// linha-a-linha (multi-linha => playlist; linha única => track simples).
// O timeout de 60s herda do ctx passado, então a extração de metadata nunca
// pendura numa rede lenta. extraArgs entram antes da URL (ex.: --cookies).
func runYtdlpJSON(
	ctx context.Context,
	urlStr string,
	extraArgs ...string,
) (*ytdlpInfo, error) {
	args := []string{
		"-j",
		"--flat-playlist",
		"--no-warnings",
		"--no-check-certificate",
	}
	args = append(args, extraArgs...)
	args = append(args, urlStr)

	cmdCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		gologging.ErrorF(
			"yt-dlp: metadata extraction failed: %v\n%s",
			err,
			stderr.String(),
		)
		return nil, fmt.Errorf("metadata extraction failed: %w", err)
	}

	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Playlist: múltiplos objetos JSON, um por linha.
	if len(lines) > 1 {
		info := ytdlpInfo{Entries: make([]ytdlpInfo, 0, len(lines))}
		for _, line := range lines {
			var entry ytdlpInfo
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				gologging.ErrorF("yt-dlp: failed to parse entry JSON: %v", err)
				continue
			}
			info.Entries = append(info.Entries, entry)
		}

		if len(info.Entries) == 0 {
			return nil, errors.New("no valid entries found in playlist")
		}

		return &info, nil
	}

	// Single track.
	var info ytdlpInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		gologging.ErrorF("yt-dlp: failed to parse JSON: %v", err)
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &info, nil
}

// runYtdlpDownload roda o yt-dlp de download pra um track, reaproveitando o
// arquivo já em cache quando existir. formatArgs traz só os flags específicos
// (formato/cookies); os flags comuns são adicionados aqui. No erro: limpa
// arquivos parciais, propaga cancelamento/timeout do context cru e devolve um
// erro com stdout/stderr pra facilitar debug.
func runYtdlpDownload(
	ctx context.Context,
	track *state.Track,
	formatArgs []string,
) (string, error) {
	if f := findFile(track); f != "" {
		gologging.Debug("yt-dlp: Download -> Cached File -> " + f)
		return f, nil
	}

	args := []string{
		"--no-playlist",
		"--no-part",
		"--no-warnings",
		"--ignore-errors",
		"--no-check-certificate",
		"-q",
		"-o", getPath(track, ".%(ext)s"),
	}
	args = append(args, formatArgs...)
	args = append(args, track.URL)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		outStr := strings.TrimSpace(stdout.String())
		errStr := strings.TrimSpace(stderr.String())

		gologging.ErrorF(
			"yt-dlp: download failed for %s: %v\nSTDOUT:\n%s\nSTDERR:\n%s",
			track.URL, err, outStr, errStr,
		)
		findAndRemove(track)

		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		return "", fmt.Errorf(
			"yt-dlp error: %w\nstdout: %s\nstderr: %s",
			err,
			outStr,
			errStr,
		)
	}

	path := findFile(track)
	if path == "" {
		return "", errors.New("yt-dlp did not return output file path")
	}

	return path, nil
}

// withVideoFlag clona cada track setando .Video, pulando entradas nil.
// Semântica única (sem buracos no slice) usada por YouTube e Spotify.
func withVideoFlag(tracks []*state.Track, video bool) []*state.Track {
	out := make([]*state.Track, 0, len(tracks))
	for _, t := range tracks {
		if t == nil {
			continue
		}
		clone := *t
		clone.Video = video
		out = append(out, &clone)
	}
	return out
}

func getPath(track *state.Track, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	mediaType := "audio"
	if track.Video {
		mediaType = "video"
	}

	filename := mediaType + "_" + state.SafeFileID(track.ID) + ext

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

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+state.SafeFileID(track.ID)+"*"))
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

	files, err := filepath.Glob(filepath.Join("downloads", t+"_"+state.SafeFileID(track.ID)+"*"))
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
