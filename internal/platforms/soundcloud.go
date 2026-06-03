package platforms

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
	"main/internal/utils"
)

type SoundCloudPlatform struct {
	name state.PlatformName
}

var (
	soundcloudLinkRegex = regexp.MustCompile(
		`(?i)^(https?://)?(www\.)?(soundcloud\.com|snd\.sc)/`,
	)
	soundcloudCache = utils.NewCache[string, []*state.Track](1 * time.Hour)
)

const PlatformSoundCloud state.PlatformName = "SoundCloud"

func init() {
	Register(85, &SoundCloudPlatform{
		name: PlatformSoundCloud,
	})
}

func (s *SoundCloudPlatform) Name() state.PlatformName {
	return s.name
}

func (s *SoundCloudPlatform) CanGetTracks(query string) bool {
	return soundcloudLinkRegex.MatchString(strings.TrimSpace(query))
}

func (s *SoundCloudPlatform) GetTracks(
	ctx context.Context,
	query string,
	_ bool,
) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	cacheKey := "soundcloud:" + strings.ToLower(query)
	if cached, ok := soundcloudCache.Get(cacheKey); ok {
		gologging.Debug("SoundCloud: Using cached tracks")
		return cached, nil
	}

	gologging.InfoF("SoundCloud: Fetching metadata for %s", query)

	info, err := s.extractMetadata(ctx, query)
	if err != nil {
		gologging.ErrorF("SoundCloud: Failed to extract metadata: %v", err)
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	var tracks []*state.Track

	if len(info.Entries) > 0 {
		gologging.InfoF(
			"SoundCloud: Found playlist with %d tracks",
			len(info.Entries),
		)
		for _, entry := range info.Entries {
			track := s.infoToTrack(&entry)
			tracks = append(tracks, track)
		}
	} else {
		track := s.infoToTrack(info)
		tracks = []*state.Track{track}
	}

	if len(tracks) > 0 {
		soundcloudCache.Set(cacheKey, tracks)
		gologging.InfoF(
			"SoundCloud: Successfully extracted %d track(s)",
			len(tracks),
		)
	}

	return tracks, nil
}

func (s *SoundCloudPlatform) CanDownload(
	source state.PlatformName,
) bool {
	return source == PlatformSoundCloud
}

func (s *SoundCloudPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	track.Video = false

	gologging.InfoF("SoundCloud: Downloading %s", track.Title)

	formatArgs := []string{
		"-f", "ba[abr>=128]/ba",
		"-x",
		"--concurrent-fragments", "4",
		"--no-overwrites",
	}

	path, err := runYtdlpDownload(ctx, track, formatArgs)
	if err != nil {
		return "", err
	}

	gologging.InfoF("SoundCloud: Successfully downloaded %s", track.Title)
	return path, nil
}

func (s *SoundCloudPlatform) extractMetadata(
	ctx context.Context,
	url string,
) (*ytdlpInfo, error) {
	return runYtdlpJSON(ctx, url)
}

func (s *SoundCloudPlatform) infoToTrack(info *ytdlpInfo) *state.Track {
	title := info.Title
	duration := int(info.Duration)

	track := &state.Track{
		ID:       info.ID,
		Title:    title,
		Duration: duration,
		Artwork:  info.Thumbnail,
		URL:      info.URL,
		Source:   PlatformSoundCloud,
		Video:    false,
	}

	return track
}
