// Package tomato is a thin HTTP client for the TomatoAnimes (betomato) API.
//
// Credentials and base URL come from config (BETOMATO_TOKEN / BETOMATO_BASE_URL).
// The client is safe for concurrent use and returns typed DTOs that the
// /anime module renders in Telegram.
package tomato

import (
	"errors"
	"fmt"
	"strconv"
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
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("feed: %s", resp.Status())
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
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("search: %s", resp.Status())
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
		Get(fmt.Sprintf("/v2/anime/%d", animeID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("anime: %s", resp.Status())
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
		Post(fmt.Sprintf("/season/%d/episodes", seasonID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("season: %s", resp.Status())
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
		Get(fmt.Sprintf("/v2/anime/episode/%d/stream", episodeID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("stream: %s", resp.Status())
	}
	return out, nil
}

// ErrNotConfigured is returned when BETOMATO_TOKEN is empty.
var ErrNotConfigured = errors.New(
	"tomato: BETOMATO_TOKEN is not configured",
)
