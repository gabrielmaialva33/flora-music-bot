package platforms

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/cookies"
	state "main/internal/core/models"
)

const PlatformYtDlp state.PlatformName = "YtDlp"

type YtdlpPlatform struct {
	name state.PlatformName
}

type ytdlpInfo struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Duration    float64     `json:"duration"`
	Thumbnail   string      `json:"thumbnail"`
	URL         string      `json:"webpage_url"`
	OriginalURL string      `json:"original_url"`
	Uploader    string      `json:"uploader"`
	Description string      `json:"description"`
	IsLive      bool        `json:"is_live"`
	Extractor   string      `json:"extractor"`
	Entries     []ytdlpInfo `json:"entries"`
}

var bannedExtractors = map[string]bool{
	"alphaporno":     true,
	"beeg":           true,
	"behindkink":     true,
	"bongacams":      true,
	"cam4":           true,
	"cammodels":      true,
	"camsoda":        true,
	"chaturbate":     true,
	"drtuber":        true,
	"eporner":        true,
	"erocast":        true,
	"eroprofile":     true,
	"fourtube":       true,
	"goshgay":        true,
	"hellporno":      true,
	"iwara":          true,
	"lovehomeporn":   true,
	"manyvids":       true,
	"motherless":     true,
	"murrtube":       true,
	"nonktube":       true,
	"noodlemagazine": true,
	"nubilesporn":    true,
	"nuvid":          true,
	"oftv":           true,
	"peekvids":       true,
	"pornbox":        true,
	"pornflip":       true,
	"pornhub":        true,
	"pornotube":      true,
	"pornovoisines":  true,
	"pornoxo":        true,
	"redgifs":        true,
	"redtube":        true,
	"rule34video":    true,
	"sauceplus":      true,
	"sexu":           true,
	"slutload":       true,
	"spankbang":      true,
	"stripchat":      true,
	"sunporno":       true,
	"thisvid":        true,
	"tnaflix":        true,
	"toypics":        true,
	"txxx":           true,
	"xhamster":       true,
	"xnxx":           true,
	"xvideos":        true,
	"xxxymovies":     true,
	"youjizz":        true,
	"youporn":        true,
	"zenporn":        true,
}

// URLs that are likely handled by YouTube
var youtubePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(youtube\.com|youtu\.be|music\.youtube\.com)`),
}

func init() {
	Register(60, &YtdlpPlatform{
		name: PlatformYtDlp,
	})
}

func (y *YtdlpPlatform) Name() state.PlatformName {
	return y.name
}

// CanGetTracks checks if this is a valid URL that yt-dlp might handle
func (y *YtdlpPlatform) CanGetTracks(query string) bool {
	query = strings.TrimSpace(query)

	if _, err := sanitizeMediaURL(query); err != nil {
		return false
	}

	// Must be a URL
	parsedURL, err := url.Parse(query)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	host := strings.ToLower(parsedURL.Host)

	// Ignore Telegram URLs ( already handled by TeleramPlatform)
	if host == "t.me" ||
		host == "telegram.me" ||
		host == "telegram.dog" ||
		strings.HasSuffix(host, ".t.me") {
		return false
	}

	return true
}

// GetTracks extracts metadata using yt-dlp
func (y *YtdlpPlatform) GetTracks(
	ctx context.Context,
	query string,
	video bool,
) ([]*state.Track, error) {
	query = strings.TrimSpace(query)

	safeURL, err := sanitizeMediaURL(query)
	if err != nil {
		gologging.InfoF("YtDlp: Rejected unsafe URL: %s", query)
		return nil, err
	}
	query = safeURL

	gologging.InfoF("YtDlp: Extracting metadata for %s", query)

	info, err := y.extractMetadata(ctx, query)
	if err != nil {
		gologging.ErrorF("YtDlp: Failed to extract metadata: %v", err)
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Check if it's a live stream
	if info.IsLive {
		gologging.Info("YtDlp: Detected live stream, returning error")
		return nil, errors.New(
			"live streams are not supported by yt-dlp downloader",
		)
	}

	// Check for banned extractor
	if bannedExtractors[strings.ToLower(info.Extractor)] {
		gologging.InfoF("YtDlp: Blocked adult content from extractor: %s", info.Extractor)
		return nil, errors.New("adult content is not allowed")
	}

	var tracks []*state.Track

	// Handle playlists
	if len(info.Entries) > 0 {
		gologging.InfoF(
			"YtDlp: Found playlist with %d entries",
			len(info.Entries),
		)
		for _, entry := range info.Entries {
			if entry.IsLive {
				continue // Skip live entries
			}
			// Check entry extractor if present (sometimes entries have their own extractor info)
			if entry.Extractor != "" &&
				bannedExtractors[strings.ToLower(entry.Extractor)] {
				gologging.InfoF(
					"YtDlp: Skipping banned entry from extractor: %s",
					entry.Extractor,
				)
				continue
			}
			track := y.infoToTrack(&entry, video)
			tracks = append(tracks, track)
		}
	} else {
		track := y.infoToTrack(info, video)
		tracks = []*state.Track{track}
	}

	if len(tracks) > 0 {
		gologging.InfoF(
			"YtDlp: Successfully extracted %d track(s)",
			len(tracks),
		)
	}

	return tracks, nil
}

func (y *YtdlpPlatform) CanDownload(source state.PlatformName) bool {
	// YtDlp can download from itself (when it extracts info)
	// and from YouTube platform
	return source == y.name || source == PlatformYouTube
}

func (y *YtdlpPlatform) Download(
	ctx context.Context,
	track *state.Track,
	_ *telegram.NewMessage,
) (string, error) {
	gologging.InfoF("YtDlp: Downloading %s", track.Title)

	formatArgs := []string{"--geo-bypass"}

	// Format selection
	if track.Video {
		formatArgs = append(
			formatArgs,
			"-f",
			"(b[height>=360][height<=1080]/bv*[height>=360][height<=1080]/bv*)+(ba[abr>=180][abr<=360]/ba)/b",
		)
	} else {
		formatArgs = append(
			formatArgs,
			"-f", "ba[abr>=180][abr<=360]/ba",
			"-x",
			"--concurrent-fragments", "4",
		)
	}

	// Cookies (YouTube only)
	if y.isYouTubeURL(track.URL) {
		if cookieFile, err := cookies.GetRandomCookieFile(); err == nil &&
			cookieFile != "" {
			formatArgs = append(formatArgs, "--cookies", cookieFile)
		}
	}

	path, err := runYtdlpDownload(ctx, track, formatArgs)
	if err != nil {
		return "", err
	}

	gologging.InfoF("YtDlp: Successfully downloaded %s", path)
	return path, nil
}

// extractMetadata uses yt-dlp to extract video/audio metadata
func (y *YtdlpPlatform) extractMetadata(
	ctx context.Context,
	urlStr string,
) (*ytdlpInfo, error) {
	var extraArgs []string

	// Add cookies only for YouTube
	if y.isYouTubeURL(urlStr) {
		cookieFile, err := cookies.GetRandomCookieFile()
		if err == nil && cookieFile != "" {
			extraArgs = append(extraArgs, "--cookies", cookieFile)
		}
	}

	return runYtdlpJSON(ctx, urlStr, extraArgs...)
}

// infoToTrack converts yt-dlp info to Track
func (y *YtdlpPlatform) infoToTrack(
	info *ytdlpInfo,
	video bool,
) *state.Track {
	duration := int(info.Duration)

	// Use original_url if available, otherwise webpage_url
	trackURL := info.URL
	if info.OriginalURL != "" {
		trackURL = info.OriginalURL
	}

	return &state.Track{
		ID:       info.ID,
		Title:    info.Title,
		Duration: duration,
		Artwork:  info.Thumbnail,
		URL:      trackURL,
		Source:   PlatformYtDlp,
		Video:    video,
	}
}

// isYouTubeURL checks if the URL is from YouTube
func (y *YtdlpPlatform) isYouTubeURL(urlStr string) bool {
	for _, pattern := range youtubePatterns {
		if pattern.MatchString(urlStr) {
			return true
		}
	}
	return false
}
