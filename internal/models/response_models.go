package models

type SongResponse struct {
	ID          int    `json:"id"`
	Song        string `json:"song"`
	Group       string `json:"group"`
	Link        string `json:"link"`
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
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
