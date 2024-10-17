package api

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"music_storage/internal/db"
	"music_storage/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
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
// @Param song query string false "Фильтр по названию песни"
// @Param id query int false "Фильтр по id"
// @Param group query string false "Фильтр по названию группы"
// @Param text query string false "Фильтр по фрагменту текста песни"
// @Param link query string false "Фильтр по ссылке"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество записей на странице" default(10)
// @Success 200 {object} models.SongResponse "Список песен с фильтрацией и пагинацией"
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
	query := db.DB.Model(&models.Song{}).Joins("Group")
	if group != "" {
		query = query.Joins("JOIN groups ON groups.id = songs.group_id").Where("groups.name = ?", group)
	}
	if song != "" {
		query = query.Where("songs.song = ?", song)
	}
	if releaseDate != "" {
		parsedDate, err := time.Parse("2006-02-01", releaseDate)
		if err == nil {
			query = query.Where("songs.release_date = ?", parsedDate)
		} else {
			logrus.Warnf("Некорректный формат даты: %s", releaseDate)
		}
	}
	if text != "" {
		query = query.Where("songs.text LIKE ?", "%"+text+"%")
	}
	if link != "" {
		query = query.Where("songs.link = ?", link)
	}
	if id != "" {
		query = query.Where("songs.id = ?", id)
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

	if len(songs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode([]models.SongResponse{})
		if err != nil {
			logrus.Errorf("Ошибка при кодировании ответа: %v", err)
			return
		}
		logrus.Info("Ответ успешно отправлен: пустой массив")
		return
	}

	var responses []models.SongResponse
	for _, song := range songs {
		responses = append(responses, models.SongResponse{
			Song:        song.Song,
			ID:          song.ID,
			Group:       song.Group.Name,
			Link:        song.Link,
			ReleaseDate: song.ReleaseDate.Format("2006-01-02"),
			Text:        song.Text,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responses)
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
		groupName := updateData.Group
		var group models.Group
		if err := db.DB.Where("name = ?", groupName).First(&group).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				group = models.Group{Name: groupName}
				if err := db.DB.Create(&group).Error; err != nil {
					logrus.Errorf("Ошибка при создании группы: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					err := json.NewEncoder(w).Encode(models.ErrorResponse{
						Message: "Внутренняя ошибка сервера",
					})
					if err != nil {
						return
					}
					return
				}
			} else {
				logrus.Errorf("Ошибка при выполнении запроса к базе данных: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				err := json.NewEncoder(w).Encode(models.ErrorResponse{
					Message: "Внутренняя ошибка сервера",
				})
				if err != nil {
					return
				}
				return
			}
		}
		song.GroupID = group.ID
	}
	if updateData.Song != "" {
		song.Song = updateData.Song
	}
	if updateData.ReleaseDate != "" {
		parsedDate, err := time.Parse("2006-01-02", updateData.ReleaseDate)
		if err != nil {
			logrus.Errorf("Некорректный формат даты: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(models.ErrorResponse{
				Message: "Некорректный формат даты, ожидается формат YYYY-MM-DD",
			})
			if err != nil {
				return
			}
			return
		}
		song.ReleaseDate = parsedDate
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

	response := models.SongResponse{
		ID:          song.ID,
		Song:        song.Song,
		Group:       song.Group.Name,
		Link:        song.Link,
		ReleaseDate: song.ReleaseDate.Format("2006-01-02"),
		Text:        song.Text,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
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
// @Success 200 {object} models.SongResponse "Успешное добавление песни с внешними данными"
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

	var group models.Group
	result := db.DB.Where("name = ?", newSong.Group).First(&group)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			group = models.Group{Name: newSong.Group}
			result = db.DB.Create(&group)
			if result.Error != nil {
				logrus.Errorf("Ошибка при создании группы: %v", result.Error)
				w.WriteHeader(http.StatusInternalServerError)
				err := json.NewEncoder(w).Encode(models.ErrorResponse{
					Message: "Ошибка при создании группы",
				})
				if err != nil {
					return
				}
				return
			}
			logrus.Info("Группа успешно создана")
		} else {
			logrus.Errorf("Ошибка при поиске группы: %v", result.Error)
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Message: "Внутренняя ошибка сервера",
			})
			if err != nil {
				return
			}
			return
		}
	}

	song := models.Song{
		GroupID: group.ID,
		Song:    newSong.Song,
	}

	externalData := FetchExternalSongData(song.Group.Name, song.Song)
	if externalData != nil {
		logrus.Info("Получены данные из внешнего API")
		if releaseDate, err := time.Parse("2006-01-02", externalData.ReleaseDate); err == nil {
			song.ReleaseDate = releaseDate
		}
		if err != nil {
			logrus.Errorf("Ошибка при парсинге даты: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(models.ErrorResponse{
				Message: "Внутренняя ошибка сервера",
			})
			if err != nil {
				return
			}
			return
		}
		song.Text = externalData.Text
		song.Link = externalData.Link
	} else {
		logrus.Warn("Данные из внешнего API не получены")
	}

	result = db.DB.Create(&song)
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

	response := models.SongResponse{
		ID:          song.ID,
		Song:        song.Song,
		Group:       newSong.Group,
		Link:        song.Link,
		ReleaseDate: song.ReleaseDate.Format("2006-01-02"),
		Text:        song.Text,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logrus.Errorf("Ошибка при кодировании ответа: %v", err)
		return
	}
	logrus.Info("Ответ успешно отправлен")
}
