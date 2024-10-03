# Music Library API

Проект предоставляет REST API для хранения и выдачи по запросу информации о музыкальных произведениях. API поддерживает создание, чтение, обновление и удаление (CRUD) записей о песнях, а также интеграцию с внешним источником данных для получения дополнительной информации.

## Технологии

- **Golang** — язык программирования.
- **Gorilla Mux** — библиотека для маршрутизации HTTP-запросов.
- **GORM** — ORM-библиотека для работы с базами данных.
- **PostgreSQL** — реляционная база данных.
- **Logrus** — библиотека ддля логирования.

## Запуск проекта

1. Клонируйте репозиторий:

    ```bash
    git clone https://github.com/reDasha/music_library.git
    ```

2. Запустите PostgreSQL и создайте новую базу данных:

 ```sql
 CREATE DATABASE music_library;
 ```

3. Настройте переменные окружения в файле `.env`. Например, так:

    ```
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=yourpassword
    DB_NAME=music_library
    ```

2. Запустите сервер:

    ```bash
    go run main.go
    ```

## Использование

После запуска приложения, API будет доступен по адресу `http://localhost:8080/`.

### API эндпоинты

- GET /songs — получение списка песен с фильтрацией.
- POST /songs — добавление новой песни.
- PUT /songs/{id} — обновление информации о песне.
- DELETE /songs/{id} — удаление песни по ID.

### Пример запроса для добавления песни:

```
bash
curl -X POST http://localhost:8080/songs \
  -H "Content-Type: application/json" \
  -d '{
        "group": "Queen",
        "song": "Bohemian Rhapsody"
      }'
```
Swagger доступен по адресу `http://localhost:8080/swagger/index.html#`.