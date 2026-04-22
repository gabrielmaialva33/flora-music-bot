// Package tomato is a thin HTTP client for the TomatoAnimes (betomato) API.
//
// Credentials and base URL come from config (BETOMATO_TOKEN / BETOMATO_BASE_URL).
// The client is safe for concurrent use and returns typed DTOs that the
// /anime module renders in Telegram.
package tomato

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"resty.dev/v3"

	"main/internal/config"
)

const (
	userAgent = "okhttp/4.11.0"
	httpTmo   = 15 * time.Second
)

// Client talks to the betomato edge API.
type Client struct {
	http *resty.Client
}

var (
	defaultClient *Client
	once          sync.Once
)

// Default returns a lazily-built singleton client bound to the config.
func Default() *Client {
	once.Do(func() {
		defaultClient = New()
	})
	return defaultClient
}

// New builds a client bound to the current config values.
func New() *Client {
	c := resty.New().
		SetBaseURL(config.BetomatoBaseURL).
		SetTimeout(httpTmo).
		SetHeader("User-Agent", userAgent).
		SetHeader("Accept", "application/json, text/plain, */*").
		SetHeader("Accept-Encoding", "gzip, deflate")

	if config.BetomatoToken != "" {
		c.SetAuthToken(config.BetomatoToken)
	}
	return &Client{http: c}
}

// Configured reports whether the client has credentials to hit the API.
func (c *Client) Configured() bool {
	return config.BetomatoToken != ""
}

func (c *Client) reqTime() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// Feed returns the home feed sections.
func (c *Client) Feed() (*FeedResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	out := &FeedResponse{}
	resp, err := c.http.R().
		SetHeader("Request-Time", c.reqTime()).
		SetHeader("X-App", "1.4.3").
		SetResult(out).
		Get("/v2/animes/feed")
	if err != nil {
		return nil, wrapTransportErr(err)
	}
	if resp.IsError() {
		return nil, statusErr(resp.StatusCode())
	}
	return out, nil
}

// Search searches for animes and mangas. page is zero-indexed.
func (c *Client) Search(query string, page int) (*SearchResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	out := &SearchResponse{}
	resp, err := c.http.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Request-Time", c.reqTime()).
		SetBody(map[string]any{
			"token":        config.BetomatoToken,
			"search":       query,
			"content_type": "all",
			"page":         page,
		}).
		SetResult(out).
		Post("/v2/content/search")
	if err != nil {
		return nil, wrapTransportErr(err)
	}
	if resp.IsError() {
		return nil, statusErr(resp.StatusCode())
	}
	return out, nil
}

// Anime returns details for a specific anime by ID.
func (c *Client) Anime(animeID int) (*AnimeResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	out := &AnimeResponse{}
	resp, err := c.http.R().
		SetHeader("Request-Time", c.reqTime()).
		SetResult(out).
		Get("/v2/anime/" + strconv.Itoa(animeID))
	if err != nil {
		return nil, wrapTransportErr(err)
	}
	if resp.IsError() {
		return nil, statusErr(resp.StatusCode())
	}
	return out, nil
}

// SeasonEpisodes returns episodes for a season.
// order is "ASC" or "DESC". page is zero-indexed.
func (c *Client) SeasonEpisodes(
	seasonID, page int,
	order string,
) (*SeasonEpisodesResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	if order != "DESC" {
		order = "ASC"
	}
	out := &SeasonEpisodesResponse{}
	resp, err := c.http.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Request-Time", c.reqTime()).
		SetBody(map[string]any{
			"token": config.BetomatoToken,
			"page":  page,
			"order": order,
		}).
		SetResult(out).
		Post("/season/" + strconv.Itoa(seasonID) + "/episodes")
	if err != nil {
		return nil, wrapTransportErr(err)
	}
	if resp.IsError() {
		return nil, statusErr(resp.StatusCode())
	}
	return out, nil
}

// EpisodeStream returns playable stream URLs for an episode.
func (c *Client) EpisodeStream(episodeID int) (*EpisodeStreamResponse, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	out := &EpisodeStreamResponse{}
	resp, err := c.http.R().
		SetHeader("User-Agent", "tomato-android").
		SetHeader("Content-Type", "application/json").
		SetResult(out).
		Get("/v2/anime/episode/" + strconv.Itoa(episodeID) + "/stream")
	if err != nil {
		return nil, wrapTransportErr(err)
	}
	if resp.IsError() {
		return nil, statusErr(resp.StatusCode())
	}
	return out, nil
}

// Sentinel errors. Callers should use errors.Is to branch and
// render a user-facing message (see modules/anime.go:friendlyErr).
//
// All error strings intentionally avoid leaking the upstream URL,
// HTTP verb or raw response body.
var (
	ErrNotConfigured = errors.New("tomato: BETOMATO_TOKEN is not configured")
	ErrTimeout       = errors.New("tomato: request timed out")
	ErrTransport     = errors.New("tomato: network error")
	ErrAuth          = errors.New("tomato: authentication failed")
	ErrNotFound      = errors.New("tomato: resource not found")
	ErrUnavailable   = errors.New("tomato: service unavailable")
	ErrProtocol      = errors.New("tomato: unexpected response")
)

// wrapTransportErr converts a raw resty/net error into a sentinel.
// This hides the upstream URL from anything that logs err.Error().
func wrapTransportErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	var nerr net.Error
	if errors.As(err, &nerr) && nerr.Timeout() {
		return ErrTimeout
	}
	// Strip the URL wrapped by net/http / resty so the string stays generic
	var uerr *url.Error
	if errors.As(err, &uerr) {
		lowered := strings.ToLower(uerr.Err.Error())
		if strings.Contains(lowered, "deadline") ||
			strings.Contains(lowered, "timeout") {
			return ErrTimeout
		}
	}
	return ErrTransport
}

// statusErr maps an HTTP status code to a sentinel error.
func statusErr(code int) error {
	switch {
	case code == 401 || code == 403:
		return ErrAuth
	case code == 404:
		return ErrNotFound
	case code >= 500:
		return ErrUnavailable
	case code >= 400:
		return ErrProtocol
	default:
		return ErrProtocol
	}
}
