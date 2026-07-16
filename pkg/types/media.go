package types

type MediaType string

const (
	Movie  MediaType = "movie"
	Series MediaType = "series"
)

type MediaRef struct {
	ProviderName string    `json:"provider_name"`
	ProviderID   string    `json:"provider_id"`
	Title        string    `json:"title"`
	MediaType    MediaType `json:"media_type"`
	Year         string    `json:"year"`
	Poster       string    `json:"poster"`
	URL          string    `json:"url"`
}

type Stream struct {
	URL       string     `json:"url"`
	Referer   string     `json:"referer"`
	UserAgent string     `json:"user_agent"`
	Subtitles []Subtitle `json:"subtitles"`
	Quality   string     `json:"quality"`
	Provider  string     `json:"provider"`
	IsM3U8    bool       `json:"is_m3u8"`
}

type Subtitle struct {
	URL  string `json:"url"`
	Lang string `json:"lang"`
}

type PlayOptions struct {
	StreamURL string
	Referer   string
	UserAgent string
	Subtitles []Subtitle
	Title     string
}
