package models

type UpdateSongRequest struct {
	Group       string `json:"group,omitempty" example:""`
	Song        string `json:"song,omitempty" example:""`
	ReleaseDate string `json:"releaseDate,omitempty" example:""`
	Text        string `json:"text,omitempty" example:""`
	Link        string `json:"link,omitempty" example:""`
}

type CreateSongRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}
