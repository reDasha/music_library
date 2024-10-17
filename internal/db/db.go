package db

import (
	"fmt"
	"music_storage/internal/models"
	"os"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	logrus.Info("Успешное подключение к базе данных")

	err = DB.AutoMigrate(&models.Group{}, &models.Song{})
	if err != nil {
		logrus.Fatalf("Ошибка автоматической миграции: %v", err)
	}
	logrus.Info("Автоматическая миграция завершена успешно")
}
