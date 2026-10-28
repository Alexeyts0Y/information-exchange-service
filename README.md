# Information Exchange Service

> REST API для обмена информацией об угрозах (Threat Intelligence) с оптимистичной блокировкой, версионированием записей и полной поддержкой JSONB-полей.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat-square&logo=postgresql)](https://www.postgresql.org)
[![Gin](https://img.shields.io/badge/Gin-v1.9-00ACD7?style=flat-square)](https://gin-gonic.com)

---

## 📋 Содержание

- [О проекте](#-о-проекте)
- [Соответствие требованиям](#-соответствие-требованиям)
- [Возможности](#-возможности)
- [Архитектура](#-архитектура)
- [Быстрый старт](#-быстрый-старт)
- [Конфигурация](#️-конфигурация)
- [API](#-api)
- [Примеры](#-примеры)
- [Разработка](#️-разработка)
- [Тестирование](#-тестирование)
- [Структура проекта](#-структура-проекта)
- [Коды ошибок](#-коды-ошибок)
- [Что можно улучшить](#-что-можно-улучшить)
- [Лицензия](#-лицензия)

---

## 🎯 О проекте

**Information Exchange Service** — сервис для централизованного хранения и обмена информацией об угрозах информационной безопасности. Построен на принципах чистой архитектуры с чётким разделением слоёв: `handler → service → repository`.

Каждая запись об угрозе идентифицируется по **MD5-хешу** и может содержать десятки атрибутов: вторичные хеши (SHA1, SHA256), поведенческие признаки, IOC (сетевые и файловые), детекты антивирусов, источник, критичность и многое другое.

---

## ✅ Соответствие требованиям

Ниже — сводка по обязательным и бонусным пунктам задания.

### Обязательное

| Требование | Статус | Комментарий |
|------------|:------:|-------------|
| Реализация сервиса по OpenAPI-спеке | ✅ | `POST /info`, `GET /info/{md5}`, `PUT /info/{md5}` |
| Использование СУБД | ✅ | PostgreSQL 16 + GORM |
| Кодогенерация по спеке | ✅ | `oapi-codegen` → `internal/gen/threat` (версия закреплена в `Makefile`) |

### Бонусные

| Бонус | Статус | Где посмотреть |
|-------|:------:|----------------|
| Юнит- и интеграционные тесты | ✅ | `*_test.go` в каждом пакете + `//go:build integration` для testcontainers |
| Осмысленное дополнение полей и валидация | ✅ | `internal/validation`, `internal/entity`, миграции с CHECK-констрейнтами |
| Docker для развертывания | ✅ | `Dockerfile` + `docker-compose.yml` + цели `make up/run/down` |
| Библиотеки кодогенерации | ✅ | `oapi-codegen` (из спеки), `golang-migrate` (embed-миграции) |
| Документация кода | ✅ | GoDoc-комментарии для пакетов и экспортируемых сущностей |
| Дополнительная защита (TLS, rate-limit, ACL) | ⚠️ | **Осознанно вынесено за MVP**, план — в разделе [«Что можно улучшить»](#-что-можно-улучшить) |

---

## ✨ Возможности

| Категория | Возможности |
|-----------|-------------|
| 🔐 **API** | REST, JSON, генерируется из OpenAPI 3.1 через `oapi-codegen` |
| 🛡 **Валидация** | Хеши (MD5/SHA1/SHA256), enum-значения, URI, диапазоны чисел |
| 🔄 **Консистентность** | Оптимистичная блокировка по полю `version` (409 Conflict при рассинхроне) |
| 🗄 **База данных** | PostgreSQL 16, миграции через `golang-migrate`, JSONB для гибких полей |
| 🚀 **Производительность** | GIN-индексы для JSONB, partial unique-индексы, connection pool |
| 🩺 **Наблюдаемость** | Structured-логирование, X-Request-ID, health-проверка `/health` |
| 🛑 **Graceful shutdown** | Корректное завершение с таймаутом |
| 🧪 **Тесты** | Unit + integration (testcontainers) + combined coverage через `make test-cover-all` |
| 📚 **Документация** | Swagger UI из OpenAPI-спеки + GoDoc |

---

## 🏗 Архитектура

```
┌──────────────┐    ┌──────────────┐    ┌────────────────┐    ┌───────────┐
│   HTTP       │───▶│   Handler    │───▶│    Service     │───▶│ Repository│───▶ PostgreSQL
│  (gin)       │    │ (validation) │    │ (business log) │    │  (GORM)   │
└──────────────┘    └──────────────┘    └────────────────┘    └───────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  gen/threat  │  ← oapi-codegen (ServerInterface)
                    └──────────────┘
```

**Слои и их ответственность:**

- **`handler`** — HTTP-слой. Парсит JSON, вызывает валидаторы, транслирует ошибки в HTTP-статусы.
- **`service`** — бизнес-логика. Маппинг между API-DTO и доменной сущностью, версионирование.
- **`repository`** — доступ к данным. Транслирует ошибки GORM в доменные (`ErrNotFound`, `ErrAlreadyExists`, `ErrVersionConflict`).
- **`entity`** — доменная модель (`Threat`).
- **`validation`** — чистые функции-валидаторы.
- **`config`** — загрузка конфигурации из окружения + `.env`.

---

## 🚀 Быстрый старт

### Через Docker Compose

```bash
git clone https://github.com/AlexeytsOY/information-exchange-service.git
cd information-exchange-service

cp .env.example .env
# Отредактируйте .env — как минимум DB_USER, DB_PASSWORD, DB_NAME

# Пересобрать образы и запустить в foreground (со стримом логов)
make run

# Или в фоне с автоматическим стримом логов приложения
make run-detached
```

Сервис будет доступен по адресу: **http://localhost:8080**

Полезные команды Docker:

```bash
make up      # запустить стек без пересборки
make build   # только собрать образы
make stop    # остановить контейнеры без удаления
make down    # остановить и удалить контейнеры
make remove  # удалить контейнеры и сети (включая orphans)
```

### Локально

```bash
# 1. Установите зависимости
go mod download

# 2. Поднимите PostgreSQL (или используйте уже существующий)
docker run -d --name pg-threats \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=threats \
  -p 5432:5432 \
  postgres:16-alpine

# 3. Настройте окружение (.env или export)
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=threats

# 4. Соберите и запустите (бинарь в ./bin/information-exchange-service)
make go
```

`make go` сначала выполняет `make compile` (собирает `./bin/information-exchange-service`), затем запускает бинарь. При старте сервис **автоматически применяет миграции** и проверяет соединение с БД.

Проверка готовности:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Swagger UI: **http://localhost:8080/docs**

---

## ⚙️ Конфигурация

Все параметры задаются через переменные окружения (или `.env`). Значения по умолчанию указаны ниже.

### Приложение

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `APP_ENV` | `development` | Окружение (`development`, `prod`) |
| `APP_PORT` | `8080` | Порт HTTP-сервера |
| `APP_READ_TIMEOUT` | `5s` | Таймаут чтения |
| `APP_WRITE_TIMEOUT` | `10s` | Таймаут записи |
| `APP_IDLE_TIMEOUT` | `60s` | Idle-таймаут |
| `APP_SHUTDOWN_TIMEOUT` | `5s` | Таймаут graceful shutdown |
| `APP_HEALTH_TIMEOUT` | `2s` | Таймаут ping к БД в `/health` |

### База данных

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `5432` | Порт |
| `DB_USER` | — (**required**) | Пользователь |
| `DB_PASSWORD` | — | Пароль |
| `DB_NAME` | — (**required**) | Имя БД |
| `DB_SSLMODE` | `disable` | SSL-режим (`disable`, `require`, `verify-full`) |
| `DB_MAX_OPEN_CONNS` | `25` | Максимум открытых соединений |
| `DB_MAX_IDLE_CONNS` | `5` | Максимум idle-соединений |
| `DB_CONN_MAX_LIFETIME` | `30m` | Время жизни соединения |
| `DB_CONN_MAX_IDLE_TIME` | `5m` | Время жизни idle-соединения |
| `DB_PING_TIMEOUT` | `5s` | Таймаут ping при старте |
| `DB_LOG_LEVEL` | `warn` | Уровень логов GORM (`silent`, `error`, `warn`, `info`) |

---

## 📡 API

### Endpoints

| Метод | Путь | Описание | Успех |
|-------|------|----------|-------|
| `POST` | `/info` | Создать запись об угрозе | `201 Created` |
| `GET` | `/info/{md5}` | Получить запись по MD5 | `200 OK` |
| `PUT` | `/info/{md5}` | Обновить запись (full replace) | `200 OK` |
| `GET` | `/health` | Проверка готовности | `200 OK` / `503` |
| `GET` | `/docs` | Swagger UI | `200` |
| `GET` | `/openapi.yaml` | OpenAPI-спецификация | `200` |

### Модель `Threat`

Полная спецификация доступна в [`api/openapi.yaml`](api/openapi.yaml). Ключевые поля:

```jsonc
{
  "md5": "d41d8cd98f00b204e9800998ecf8427e",   // PK, 32 hex
  "sha1": "da39a3ee...",                        // опционально
  "sha256": "e3b0c442...",                      // опционально
  "name": "Trojan.Win32.Generic",               // required
  "aliases": ["Generic Trojan"],
  "platform": ["Windows"],
  "tags": ["executable", "c2"],
  "threat_type": "trojan",
  "severity": "high",                           // low|medium|high|critical
  "status": "active",                           // active|resolved|false_positive|expired
  "confidence": 85,                             // 0..100
  "source": { "name": "Sandbox", "type": "sandbox", "url": "https://..." },
  "network_iocs": [{ "type": "domain", "value": "evil.example.com" }],
  "file_iocs":    [{ "type": "path",   "value": "C:\\Temp\\update.exe" }],
  "detections":   [{ "engine": "Kaspersky", "signature": "Trojan.Win32" }],
  "version": 1
}
```

> 💡 **Оптимистичная блокировка:** при `PUT` передавайте `version` из последнего `GET`. Если запись была изменена — вернётся `409 Conflict`.

---

## 💡 Примеры

### Создание записи

```bash
curl -X POST http://localhost:8080/info \
  -H 'Content-Type: application/json' \
  -d '{
    "md5": "d41d8cd98f00b204e9800998ecf8427e",
    "name": "Trojan.Win32.Generic",
    "severity": "high",
    "status": "active",
    "threat_type": "trojan",
    "platform": ["Windows"],
    "tags": ["executable", "persistence"],
    "network_iocs": [
      { "type": "domain", "value": "evil.example.com", "description": "C2" }
    ]
  }'
```

**Ответ** `201 Created`:

```json
{
  "md5": "d41d8cd98f00b204e9800998ecf8427e",
  "name": "Trojan.Win32.Generic",
  "severity": "high",
  "status": "active",
  "version": 1,
  "created_at": "2026-10-28T12:00:00Z",
  "updated_at": "2026-10-28T12:00:00Z",
  "...": "..."
}
```

### Получение

```bash
curl http://localhost:8080/info/d41d8cd98f00b204e9800998ecf8427e
```

### Обновление

```bash
curl -X PUT http://localhost:8080/info/d41d8cd98f00b204e9800998ecf8427e \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Trojan.Win32.Generic.Updated",
    "severity": "critical",
    "version": 1
  }'
```

Если запись уже изменил кто-то другой:

```json
HTTP/1.1 409 Conflict
{
  "message": "version conflict: threat was modified by another request; reload and retry"
}
```

### Ошибки валидации

```bash
curl -X POST http://localhost:8080/info \
  -H 'Content-Type: application/json' \
  -d '{"md5":"short","name":""}'
```

```json
HTTP/1.1 400 Bad Request
{
  "message": "validation failed",
  "details": "md5: md5 must match ^[a-fA-F0-9]{32}$"
}
```

---

## 🛠 Разработка

Все команды доступны через `Makefile`. Полный список:

```bash
make help
```

### Требования

- **Go** 1.22+
- **PostgreSQL** 16+ (или Docker)
- **Docker** (для `make run`, `make up`, integration-тестов)
- **golangci-lint** (для `make lint`)
- **goimports** (для `make fmt`)
- `oapi-codegen` и `gocovmerge` подтягиваются автоматически через `go run`, отдельная установка не нужна

### Генерация кода из OpenAPI

Версия `oapi-codegen` закреплена в `Makefile` (`OAPI_VERSION := v2.8.1`), конфиг — `oapi-codegen.yaml`:

```bash
make gen
```

### Сборка

```bash
make compile          # ./bin/information-exchange-service (из ./cmd/app)
make compile-migrate  # ./bin/migrate (из ./cmd/migrate)
make go               # compile + запуск бинаря приложения
```

### Миграции

Миграции встроены в бинарник через `//go:embed` и **применяются автоматически при старте приложения**. Для ручного применения есть отдельный бинарь `cmd/migrate`:

```bash
make migrate       # go run ./cmd/migrate
make migrate-bin   # собрать ./bin/migrate и запустить
```

Файлы миграций лежат в `internal/migrations/`.

### Форматирование и линтинг

```bash
make lint       # golangci-lint run ./...
make lint-fix   # golangci-lint run ./... --fix
make fmt        # gofmt -s -w . && goimports -w .
make fmt-check  # проверка форматирования (для CI)
```

---

## 🧪 Тестирование

### Юнит-тесты

Быстрые, без Docker:

```bash
make test
```

Под капотом: `go test -race -count=1 ./internal/...`

Покрывают:

- Валидацию (все правила, включая граничные значения)
- Сервисный слой (маппинг, версионирование, обработку ошибок)
- HTTP-хендлеры (все статусы: 200/201/400/404/409/500)
- Маппинг между DTO и сущностью
- Middleware (X-Request-ID, логирование)
- Конфигурацию (дефолты, override, валидация)

### Интеграционные и e2e-тесты

Поднимают настоящий PostgreSQL через `testcontainers-go` (**нужен Docker**):

```bash
make test-integration
```

Под капотом: `go test -race -count=1 -tags=integration ./internal/repository/... ./test/...`

Проверяют:

- Создание и чтение (round-trip) через реальный PG
- Дубликаты (`ErrAlreadyExists`)
- Оптимистичную блокировку (`ErrVersionConflict`)
- Миграции
- e2e-сценарии (`./test/...`)

### Все тесты

```bash
make test-all   # test + test-integration
```

### Покрытие

```bash
make test-cover              # unit coverage → .coverage/unit.out
make test-cover-integration  # integration coverage → .coverage/integration.out
make test-cover-all          # merge unit+integration → .coverage/all.out + .coverage/coverage.html
make cover-func              # покрытие по функциям (combined)
make cover-html              # открыть HTML-отчёт в браузере
make clean-cov               # удалить артефакты покрытия
```

`test-cover-all` использует `gocovmerge` (подтягивается через `go run`) и кладёт итоговый HTML в `.coverage/coverage.html`.

---

## 📁 Структура проекта

```
.
├── api/
│   └── openapi.yaml                 # Спецификация OpenAPI 3.1
├── cmd/
│   ├── app/
│   │   └── main.go                  # Точка входа сервиса
│   └── migrate/
│       └── main.go                  # Отдельный бинарь миграций
├── internal/
│   ├── app/                         # Сборка приложения (DI, роутер, shutdown)
│   ├── config/                      # Загрузка конфигурации
│   ├── database/                    # Подключение к PostgreSQL + миграции
│   ├── entity/                      # Доменные сущности (Threat)
│   ├── gen/threat/                  # Сгенерированный код из OpenAPI
│   ├── handler/                     # HTTP-слой
│   ├── middleware/                  # Логирование, request ID
│   ├── migrations/                  # SQL-миграции (embed)
│   ├── repository/                  # Доступ к данным (GORM)
│   ├── service/                     # Бизнес-логика
│   ├── testutil/                    # Хелперы для integration-тестов
│   └── validation/                  # Валидаторы
├── test/                            # e2e-тесты (integration tag)
├── .coverage/                       # Артефакты покрытия (в .gitignore)
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── oapi-codegen.yaml                # Конфиг кодогенерации
├── go.mod
└── README.md
```

---

## 🩹 Коды ошибок

| Код | Сообщение | Когда |
|-----|-----------|-------|
| `400` | `invalid request body` | Битый JSON |
| `400` | `validation failed` | Не прошла валидация (см. `details`) |
| `400` | `invalid md5` | MD5 в пути не соответствует `^[a-fA-F0-9]{32}$` |
| `404` | `threat not found` | Запись отсутствует |
| `409` | `threat with this md5 already exists` | Дубликат при создании |
| `409` | `version conflict: threat was modified...` | Устаревшая `version` при `PUT` |
| `500` | `internal server error` | Внутренняя ошибка (см. логи) |
| `503` | `unhealthy` | `/health`: БД недоступна |

Все ошибки имеют единый формат:

```json
{
  "message": "human-readable description",
  "details": "optional technical details"
}
```

---

## 🔍 Наблюдаемость

- **X-Request-ID** — генерируется автоматически или пробрасывается из входящего заголовка; возвращается в ответе и пишется в логи.
- **Логи** — `start`/`end` запросов с методом, путём, статусом, размером ответа и латентностью.
- **`/health`** — ping к БД с настраиваемым таймаутом; `200` при успехе, `503` при ошибке.

Пример:

```text
http: start request_id=9f2a... method=POST path="/info" remote_addr=127.0.0.1
http: end   request_id=9f2a... method=POST path="/info" status=201 size=412 latency=3.2ms errors=""
```

---

## 🚧 Что можно улучшить

Раздел намеренно оставлен как честный backlog — часть бонусных требований сознательно вынесена за пределы MVP, чтобы не раздувать скоуп. Ниже — план расширения по направлениям.

### 🔒 Дополнительная защита сервиса

| Направление | Идея реализации |
|-------------|-----------------|
| **TLS / mTLS** | Нативный `http.Server` с `TLSConfig`, автоподхват сертификатов из файла или `autocert` (Let's Encrypt). Для межсервисного взаимодействия — mTLS с валидацией клиентских сертификатов. |
| **Rate limiting** | `golang.org/x/time/rate` в middleware; лимит per-IP / per-token, с конфигурацией через env. Для распределённого сценария — Redis-счётчики. |
| **Аутентификация / ACL** | API-ключи (`X-API-Key`) или JWT (OIDC-провайдер). Роли `reader` / `writer` / `admin`; проверка на уровне middleware. |
| **Аудит-лог** | Отдельная таблица `audit_log` (кто, что, когда изменил) с записью в рамках той же транзакции, что и `UPDATE threats`. |
| **CORS и CSP** | Строгий список `Allow-Origin`, `Content-Security-Policy` для Swagger UI. |
| **Ограничение размера тела** | `http.MaxBytesReader` в middleware — защита от payload-бомб. |
| **Секреты** | Вынос секретов из env в Vault / AWS Secrets Manager, ротация без рестарта. |

### 🧱 Инфраструктура

- Helm-чарт и манифесты Kubernetes (readiness/liveness probes уже совместимы с `/health`).
- CI/CD: `golangci-lint`, `go test -race`, публикация образа, `trivy`-скан.
- OpenTelemetry: трейсы через `otelgin`, метрики в Prometheus.

### 📊 Данные и API

- Пагинация и фильтрация для `GET /info` (сейчас endpoint отсутствует — можно добавить `?severity=&status=&tags=`).
- Полнотекстовый поиск по `description` / `name` (`tsvector` + GIN-индекс).
- Мягкое удаление (`deleted_at`) и TTL-очистка по полю `ttl` (cron-задача).
- Webhooks / события при создании/обновлении — интеграция с SIEM.