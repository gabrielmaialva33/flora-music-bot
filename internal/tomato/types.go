package tomato

// --- /v2/animes/feed ---

type FeedResponse struct {
	Status         bool           `json:"status"`
	StatusCode     int            `json:"status_code"`
	RemoteSettings RemoteSettings `json:"remote_settings"`
	Data           []FeedSection  `json:"data"`
}

type RemoteSettings struct {
	BlockScreenshots bool `json:"block_screenshots"`
}

// FeedSection is a single row on the home screen.
// Known types observed so far:
//
//	3: "Em alta" — banners (uses Thumbnails)
//	5: regular rail — cards (uses Thumbnails)
//	6: featured hero block (single item with banner + text)
//	7: new episodes rail (uses Episodes)
//	8: history row (often empty)
type FeedSection struct {
	Type       int            `json:"type"`
	Title      string         `json:"title,omitempty"`
	HyperTitle string         `json:"hyper_title,omitempty"`
	Banner     string         `json:"banner,omitempty"`
	Text       string         `json:"text,omitempty"`
	Tags       string         `json:"tags,omitempty"`
	Year       string         `json:"year,omitempty"`
	Rating     string         `json:"rating,omitempty"`
	AnimeID    int            `json:"anime_id,omitempty"`
	EpisodeID  int            `json:"episode_id,omitempty"`
	Data       []FeedDataItem `json:"data,omitempty"`
}

// FeedDataItem is a polymorphic item inside a FeedSection.Data slice.
// For rails of animes only AnimeID/Thumbnail are populated;
// for "novos episódios" the Ep* fields carry the payload.
type FeedDataItem struct {
	// anime rail
	AnimeID   int    `json:"anime_id,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
	TagID     int    `json:"tag_id,omitempty"`

	// episode rail
	EpID      int    `json:"ep_id,omitempty"`
	EpAnimeID int    `json:"ep_anime_id,omitempty"`
	AnimeName string `json:"anime_name,omitempty"`
	EpName    string `json:"ep_name,omitempty"`
	Dubbed    bool   `json:"dubbed,omitempty"`
}

// --- /v2/content/search ---

type SearchResponse struct {
	Status     bool           `json:"status"`
	Message    string         `json:"message"`
	StatusCode int            `json:"status_code"`
	Result     []SearchResult `json:"result"`
}

type SearchResult struct {
	ID       int    `json:"id"`
	Type     string `json:"type"` // "anime" | "manga"
	Name     string `json:"name"`
	Episodes int    `json:"episodes"`
	Date     string `json:"date"`
	Image    string `json:"image"`
	Priority int    `json:"priority"`
	Tags     string `json:"tags"`
}

// --- /v2/anime/:id ---

type AnimeResponse struct {
	AnimeDetails  AnimeDetails  `json:"anime_details"`
	Liked         bool          `json:"liked"`
	Favorited     bool          `json:"favorited"`
	Notify        bool          `json:"notify"`
	NotifyCapable bool          `json:"notify_capable"`
	CommentsCount int           `json:"comments_count"`
	AnimeSeasons  []AnimeSeason `json:"anime_seasons"`
}

type AnimeDetails struct {
	AnimeID             int    `json:"anime_id"`
	AnimeName           string `json:"anime_name"`
	AnimeDescription    string `json:"anime_description"`
	AnimeParentalRating string `json:"anime_parental_rating"`
	ReleaseDay          string `json:"release_day"`
	AnimeEpisodes       int    `json:"anime_episodes"`
	AnimeDate           string `json:"anime_date"`
	AnimeCoverURL       string `json:"anime_cover_url"`
	AnimeCapeURL        string `json:"anime_cape_url"`
	AnimeBannerURL      string `json:"anime_banner_url"`
	AnimeGenre          string `json:"anime_genre"`
	DubAvailable        bool   `json:"dub_available"`
}

type AnimeSeason struct {
	SeasonID     int    `json:"season_id"`
	SeasonName   string `json:"season_name"`
	SeasonNumber int    `json:"season_number"`
	SeasonDubbed int    `json:"season_dubbed"` // 0 | 1
}

// --- /season/:id/episodes ---

type SeasonEpisodesResponse struct {
	Status   bool      `json:"status"`
	Episodes int       `json:"episodes"`
	Data     []Episode `json:"data"`
}

type Episode struct {
	EpID            int    `json:"ep_id"`
	EpName          string `json:"ep_name"`
	EpNumber        int    `json:"ep_number"`
	EpAnimeID       int    `json:"ep_anime_id"`
	EpSeasonID      int    `json:"ep_season_id"`
	EpThumbnail     string `json:"ep_thumbnail"`
	EpDescription   string `json:"ep_description"`
	EpLenghtMinutes int    `json:"ep_lenght_minutes"`
}

// --- /v2/anime/episode/:id/stream ---

type EpisodeStreamResponse struct {
	Streams          StreamQualities `json:"streams"`
	EpisodeType      int             `json:"episodeType"`
	EpisodeAnimeID   int             `json:"episodeAnimeID"`
	EpisodeNumber    int             `json:"episodeNumber"`
	EpisodeName      string          `json:"episodeName"`
	EpisodeHasNext   bool            `json:"episodeHasNext"`
	NextEpisodeID    int             `json:"nextEpisodeID"`
	NextEpisodeTitle string          `json:"nextEpisodeTitle"`
	NextEpisodeData  *NextEpisode    `json:"nextEpisodeData,omitempty"`
}

type StreamQualities struct {
	SHD string `json:"shd"` // 480p
	MHD string `json:"mhd"` // 720p (often null)
	FHD string `json:"fhd"` // 1080p
}

// Best returns the best-quality URL available (fhd > mhd > shd).
func (s StreamQualities) Best() string {
	switch {
	case s.FHD != "":
		return s.FHD
	case s.MHD != "":
		return s.MHD
	default:
		return s.SHD
	}
}

type NextEpisode struct {
	EpisodeID        int    `json:"episodeID"`
	EpisodeName      string `json:"episodeName"`
	EpisodeThumbnail string `json:"episodeThumbnail"`
}
