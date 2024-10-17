package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/http-swagger"
	"music_storage/internal/api"
	"music_storage/internal/db"
	"net/http"
	"os"

	_ "music_storage/docs"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Info("Инициализация сервера")
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Ошибка загрузки файла .env: ", err)
	}

	logrus.Info("Подключение к базе данных...")
	db.Connect()
	logrus.Info("Подключение к базе данных установлено")
	logrus.Info("Настройка маршрутов API")
	r := mux.NewRouter()
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	r.HandleFunc("/songs", api.GetFilteredSongs).Methods("GET")
	r.HandleFunc("/songs/{id}/text", api.GetSongText).Methods("GET")
	r.HandleFunc("/songs/{id}", api.DeleteSong).Methods("DELETE")
	r.HandleFunc("/songs/{id}", api.UpdateSong).Methods("PUT")
	r.HandleFunc("/songs", api.CreateSong).Methods("POST")

	logrus.Info("Маршруты API настроены")

	logrus.Info("Запуск HTTP-сервера")
	if err := http.ListenAndServe(os.Getenv("SERVICE_ADDRESS"), r); err != nil {
		logrus.Fatal("Ошибка запуска HTTP-сервера:", err)
	}
}
