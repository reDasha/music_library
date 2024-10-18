package models

type UpdateSongRequest struct {
	Group       *string `json:"group,omitempty"`
	Song        *string `json:"song,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty" example:"0001-01-01"`
	Text        *string `json:"text,omitempty"`
	Link        *string `json:"link,omitempty"`
}

type CreateSongRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}
