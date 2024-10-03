package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"music_storage/internal/db"
	"music_storage/internal/models"
	"net/http"
	"strconv"
	"strings"
)

// @title Songs API
// @version 1.0
// @description API для работы с библиотекой песен.
// @host localhost:8080
// @BasePath /

// GetFilteredSongs возвращает список песен с фильтрацией по полям и пагинацией.
// @Summary Получить список песен с фильтрацией
// @Description Возвращает список песен с поддержкой фильтрации по полям и пагинации.
// @Tags Песни
// @Produce  json
// @Param group query string false "Фильтр по названию группы"
// @Param song query string false "Фильтр по названию песни"
// @Param id query int false "Фильтр по id"
// @Param text query string false "Фильтр по фрагменту текста песни"
// @Param link query string false "Фильтр по ссылке"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество записей на странице" default(10)
// @Success 200 {object} []models.Song "Список песен с фильтрацией и пагинацией"
// @Failure 400 {object} models.ErrorResponse "Некорректный ID"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /songs [get]
func GetFilteredSongs(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Начало обработки запроса на получение отфильтрованного списка песен")
	id := r.URL.Query().Get("id")
	group := r.URL.Query().Get("group")
	song := r.URL.Query().Get("song")
	releaseDate := r.URL.Query().Get("releaseDate")
	text := r.URL.Query().Get("text")
	link := r.URL.Query().Get("link")

	logrus.Debugf("Параметры запроса - id: %s, group: %s, song: %s, releaseDate: %s, text: %s, link: %s", id, group, song, releaseDate, text, link)

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	logrus.Debugf("Параметры пагинации - page: %d, limit: %d", page, limit)

	var songs []models.Song
	query := db.DB.Model(&models.Song{})
	if group != "" {
		query = query.Where(`"group" = ?`, group)
	}
	if song != "" {
		query = query.Where("song = ?", song)
	}
	if releaseDate != "" {
		query = query.Where("releaseDate = ?", releaseDate)
	}
	if text != "" {
		query = query.Where("text LIKE ?", "%"+text+"%")
	}
	if link != "" {
		query = query.Where(" link = ?", link)
	}
	if id != "" {
		query = query.Where(" id = ?", id)
	}

	result := query.Limit(limit).Offset(offset).Find(&songs)
	if result.Error != nil {
		logrus.Errorf("Ошибка при выполнении запроса к базе данных: %v", result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Внутренняя ошибка сервера",
		})
		if err != nil {
			return
		}
		return
	}
	logrus.Info("Запрос к базе данных успешно выполнен")

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(songs)
	if err != nil {
		logrus.Errorf("Ошибка при кодировании ответа: %v", err)
		return
	}
	logrus.Info("Ответ успешно отправлен")
}

// GetSongText возвращает текст песни или конкретный куплет.
// @Summary Получить текст песни
// @Description Возвращает текст песни с возможностью выбора конкретного куплета или всего текста.
// @Tags Песни
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param verse query int false "Номер куплета"
// @Success 200 {string} string "Текст песни или куплет"
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /songs/{id}/text [get]
func GetSongText(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Начало обработки запроса на получение текста песни")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil || id < 1 {
		logrus.Errorf("Некорректный ID: %s", vars["id"])
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Некорректный ID"})
		if err != nil {
			return
		}
		return
	}
	logrus.Debugf("Получение песни с ID: %d", id)
	var song models.Song
	result := db.DB.First(&song, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logrus.Error("Песня не найдена")
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Песня не найдена"})
			if err != nil {
				return
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Внутренняя ошибка сервера при поиске песни"})
		if err != nil {
			return
		}
		return
	}

	if song.Text == "" {
		logrus.Error("Текст песни не найден")
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Текст песни не найден"})
		if err != nil {
			return
		}
		return
	}

	verses := strings.Split(song.Text, "\n\n")
	verseStr := r.URL.Query().Get("verse")
	if verseStr != "" {
		verse, err := strconv.Atoi(verseStr)
		if err != nil || verse < 1 || verse > len(verses) {
			logrus.Errorf("Некорректный номер куплета: %s", verseStr)
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Некорректный номер куплета"})
			if err != nil {
				return
			}
			return
		}

		logrus.Infof("Отправка куплета номер %d", verse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]string{"text": verses[verse-1]})
		if err != nil {
			logrus.Errorf("Ошибка при кодировании куплета: %v", err)
			return
		}
		return
	}

	logrus.Info("Отправка полного текста песни")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"text": song.Text})
	if err != nil {
		logrus.Errorf("Ошибка при кодировании текста: %v", err)
		return
	}
}

// DeleteSong удаляет песню по её ID.
// @Summary Удалить песню
// @Description Удаляет песню по её ID.
// @Tags Песни
// @Param id path int true "ID песни"
// @Success 200 {object} models.MessageResponse "Успешное удаление песни"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} models.ErrorResponse "Песня не найдена"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /songs/{id} [delete]
func DeleteSong(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Начало обработки запроса на удаление песни")
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("Некорректный ID: %s", idStr)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Некорректный ID",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Debugf("ID песни для удаления: %d", id)

	result := db.DB.Delete(&models.Song{}, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logrus.Warnf("Песня с ID %d не найдена", id)
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Message: "Песня не найдена",
			})
			if err != nil {
				return
			}
			return
		}
		logrus.Errorf("Ошибка при удалении песни из базы данных: %v", result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Внутренняя ошибка сервера",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Infof("Песня с ID %d успешно удалена", id)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(models.MessageResponse{
		Message: "Песня успешно удалена",
	})
	if err != nil {
		logrus.Errorf("Ошибка при кодировании ответа: %v", err)
		return
	}
	logrus.Info("Ответ успешно отправлен")
}

// UpdateSong обновляет данные песни по её ID.
// @Summary      Изменить данные песни
// @Description  Обновляет информацию о песне по её ID. Поля, которые не переданы, остаются без изменений.
// @Tags         Песни
// @Accept       json
// @Produce      json
// @Param        id      path      int     true   "ID песни"
// @Param        song    body      models.UpdateSongRequest true "Обновленные данные песни (поля, которые могут быть изменены)"
// @Success      200     {object}  models.Song "Успешное обновление песни"
// @Failure      400     {object}  models.ErrorResponse "Некорректные данные запроса"
// @Failure      404     {object}  models.ErrorResponse "Песня не найдена"
// @Failure      500     {object}  models.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /songs/{id} [put]
func UpdateSong(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Начало обработки запроса на изменение данных песни")
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("Некорректный ID: %s", idStr)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Некорректный ID",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Debugf("ID песни для обновления: %d", id)

	var updateData models.UpdateSongRequest
	err = json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		logrus.Errorf("Ошибка при декодировании запроса: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Некорректные данные запроса",
		})
		if err != nil {
			return
		}
		return
	}

	var song models.Song
	result := db.DB.First(&song, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logrus.Warnf("Песня с ID %d не найдена", id)
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Message: "Песня не найдена",
			})
			if err != nil {
				return
			}
			return
		}
		logrus.Errorf("Ошибка при выполнении запроса к базе данных: %v", result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Внутренняя ошибка сервера",
		})
		if err != nil {
			return
		}
		return
	}

	if updateData.Group != "" {
		song.Group = updateData.Group
	}
	if updateData.Song != "" {
		song.Song = updateData.Song
	}
	if updateData.ReleaseDate != "" {
		song.ReleaseDate = updateData.ReleaseDate
	}
	if updateData.Text != "" {
		song.Text = updateData.Text
	}
	if updateData.Link != "" {
		song.Link = updateData.Link
	}

	result = db.DB.Save(&song)
	if result.Error != nil {
		logrus.Errorf("Ошибка при обновлении песни в базе данных: %v", result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Внутренняя ошибка сервера",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Infof("Данные песни с ID %d успешно обновлены", id)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(song)
	if err != nil {
		logrus.Errorf("Ошибка при кодировании ответа: %v", err)
		return
	}
	logrus.Info("Ответ успешно отправлен")
}

// CreateSong добавляет новую песню в библиотеку.
// @Summary Добавить новую песню
// @Description Добавление новой песни в базу данных. Данные о песне обогащаются информацией с внешнего API.
// @Tags Песни
// @Accept  json
// @Produce  json
// @Param song body models.CreateSongRequest true "Данные песни (группа, название)"
// @Success 200 {object} models.Song "Успешное добавление песни с внешними данными"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера при сохранении песни"
// @Router /songs [post]
func CreateSong(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Начало обработки запроса на создание новой песни")
	var newSong models.CreateSongRequest

	err := json.NewDecoder(r.Body).Decode(&newSong)
	if err != nil {
		logrus.Errorf("Ошибка при декодировании запроса: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Некорректные данные запроса",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Debugf("Данные новой песни - group: %s, song: %s", newSong.Group, newSong.Song)

	song := models.Song{
		Group: newSong.Group,
		Song:  newSong.Song,
	}

	externalData := FetchExternalSongData(song.Group, song.Song)
	if externalData != nil {
		logrus.Info("Получены данные из внешнего API")
		song.ReleaseDate = externalData.ReleaseDate
		song.Text = externalData.Text
		song.Link = externalData.Link
	} else {
		logrus.Warn("Данные из внешнего API не получены")
	}

	result := db.DB.Create(&song)
	if result.Error != nil {
		logrus.Errorf("Ошибка при сохранении песни в базу данных: %v", result.Error)
		w.WriteHeader(http.StatusInternalServerError)
		err := json.NewEncoder(w).Encode(models.ErrorResponse{
			Message: "Внутренняя ошибка сервера",
		})
		if err != nil {
			return
		}
		return
	}

	logrus.Info("Песня успешно сохранена в базе данных")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(song)
	if err != nil {
		logrus.Errorf("Ошибка при кодировании ответа: %v", err)
		return
	}
	logrus.Info("Ответ успешно отправлен")
}
