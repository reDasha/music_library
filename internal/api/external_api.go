package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type ExternalSongData struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func FetchExternalSongData(group, song string) *ExternalSongData {
	apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", os.Getenv("API_BASE_URL"), group, song)

	logrus.Infof("Запрос данных через внешний API: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		logrus.Errorf("Ошибка при запросе к внешнему API: %v", err)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("Неверный статус ответа от внешнего API: %d", resp.StatusCode)
		return nil
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logrus.Errorf("Ошибка при закрытии тела ответа: %v", err)
		}
	}(resp.Body)

	var data ExternalSongData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		logrus.Errorf("Ошибка при декодировании ответа: %v", err)
		return nil
	}
	logrus.Infof("Данные успешно получены: %+v", data)
	return &data
}
