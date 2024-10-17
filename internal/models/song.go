package models

import (
	"time"
)

type Group struct {
	ID   int    `gorm:"primaryKey"`
	Name string `json:"name"`
}

type Song struct {
	ID          int       `gorm:"primaryKey"`
	GroupID     int       `json:"groupID"`
	Group       Group     `gorm:"foreignKey:GroupID"`
	Song        string    `json:"song"`
	ReleaseDate time.Time `json:"releaseDate" gorm:"type:date"`
	Text        string    `json:"text"`
	Link        string    `json:"link"`
}
