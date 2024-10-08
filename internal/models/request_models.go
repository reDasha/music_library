package models

type UpdateSongRequest struct {
	Group       string `json:"group,omitempty" example:""`
	Song        string `json:"song,omitempty" example:""`
	ReleaseDate string `json:"releaseDate,omitempty" example:""`
	Text        string `json:"text,omitempty" example:""`
	Link        string `json:"link,omitempty" example:""`
}

type CreateSongRequest struct {
	Group string `json:"group" binding:"required"`
	Song  string `json:"song" binding:"required"`
}

// ErrorResponse описывает структуру ошибки для Swagger.
// @Description Ошибка API
type ErrorResponse struct {
	Message string `json:"message"`
}

// MessageResponse описывает успешное сообщение для Swagger.
type MessageResponse struct {
	Message string `json:"message"`
}
