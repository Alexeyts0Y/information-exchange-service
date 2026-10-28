PROJECT_NAME := information-exchange-service
MODULE       := github.com/AlexeytsOY/information-exchange-service
OAPI_VERSION := v2.8.1
COVERAGE_DIR := .coverage

.PHONY: help compile compile-migrate go lint lint-fix fmt fmt-check \
        test test-integration test-all \
        test-cover test-cover-integration test-cover-all \
        cover-func cover-html clean-cov \
        down build up run run-detached stop remove gen

help: ## Показать доступные команды
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-24s\033[0m %s\n", $$1, $$2}'

# ---------- Golang ----------

.PHONY: help compile compile-migrate go migrate migrate-bin \
        lint lint-fix fmt fmt-check \
        test test-integration test-all \
        test-cover test-cover-integration test-cover-all \
        cover-func cover-html clean-cov \
        down build up run run-detached stop remove gen

compile: ## Собрать основной бинарь в ./bin
	@mkdir -p ./bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
		-o ./bin/$(PROJECT_NAME) ./cmd/app

compile-migrate: ## Собрать бинарь миграций в ./bin
	@mkdir -p ./bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
		-o ./bin/migrate ./cmd/migrate

go: compile ## Собрать и запустить приложение локально
	./bin/$(PROJECT_NAME)

lint: ## Проверить код линтером
	golangci-lint run ./...

lint-fix: ## Проверить и автоматически исправить
	golangci-lint run ./... --fix

fmt: ## Форматировать код
	gofmt -s -w .
	goimports -w .

fmt-check: ## Проверить форматирование (для CI)
	@test -z "$$(gofmt -s -l .)" || (echo "gofmt:"; gofmt -s -l .; exit 1)
	@go run golang.org/x/tools/cmd/goimports -l . | grep . && (echo "goimports: исправьте импорты"; exit 1) || true

# ---------- Tests ----------

test: ## Юнит-тесты (без Docker, быстро)
	go test -race -count=1 ./internal/...

test-integration: ## Интеграционные + e2e (testcontainers, нужен Docker)
	go test -race -count=1 -tags=integration ./internal/repository/... ./test/...

test-all: test test-integration ## Все тесты (unit + integration)

test-cover: ## Юнит-тесты + coverage
	@mkdir -p $(COVERAGE_DIR)
	go test -race -count=1 \
		-coverprofile=$(COVERAGE_DIR)/unit.out \
		./internal/...
	@echo "=== unit coverage ==="
	@go tool cover -func=$(COVERAGE_DIR)/unit.out | tail -1

test-cover-integration: ## Интеграция + coverage по всем ./internal/...
	@mkdir -p $(COVERAGE_DIR)
	go test -race -count=1 -tags=integration \
		-coverpkg=$(MODULE)/internal/... \
		-coverprofile=$(COVERAGE_DIR)/integration.out \
		./internal/repository/... ./test/...
	@echo "=== integration coverage ==="
	@go tool cover -func=$(COVERAGE_DIR)/integration.out | tail -1

test-cover-all: test-cover test-cover-integration ## Merge unit+integration → combined coverage
	@command -v gocovmerge >/dev/null 2>&1 || \
		go install github.com/wadey/gocovmerge@latest
	go run github.com/wadey/gocovmerge@latest \
    	$(COVERAGE_DIR)/unit.out $(COVERAGE_DIR)/integration.out > $(COVERAGE_DIR)/all.out
	@echo "=== combined coverage ==="
	@go tool cover -func=$(COVERAGE_DIR)/all.out | tail -1
	@go tool cover -html=$(COVERAGE_DIR)/all.out -o $(COVERAGE_DIR)/coverage.html
	@echo "HTML: $(COVERAGE_DIR)/coverage.html"

cover-func: ## Показать покрытие по функциям (combined)
	@test -f $(COVERAGE_DIR)/all.out || (echo "запустите make test-cover-all"; exit 1)
	@go tool cover -func=$(COVERAGE_DIR)/all.out

cover-html: ## Открыть combined coverage в браузере
	@test -f $(COVERAGE_DIR)/all.out || (echo "запустите make test-cover-all"; exit 1)
	@go tool cover -html=$(COVERAGE_DIR)/all.out

clean-cov: ## Удалить coverage-артефакты
	rm -rf $(COVERAGE_DIR)
	rm -f coverage.out coverage.unit.out coverage.integration.out coverage.html

migrate: ## Применить миграции к БД (конфиг из .env / окружения)
	go run ./cmd/migrate

migrate-bin: compile-migrate ## Применить миграции через скомпилированный ./bin/migrate
	./bin/migrate

# ---------- Docker ----------

down: ## Остановить и удалить контейнеры
	docker compose down

build: ## Собрать Docker-образы
	docker compose build

up: ## Запустить стек (foreground)
	docker compose up

run: down ## Пересобрать и запустить (foreground)
	docker compose up --build

run-detached: down ## Пересобрать и запустить в фоне + стрим логов
	docker compose up -d --build
	docker compose logs -f app

stop: ## Остановить контейнеры без удаления
	docker compose stop

remove: ## Удалить контейнеры и сети (включая orphans)
	docker compose down --remove-orphans

# ---------- Codegen ----------

gen: ## Сгенерировать код по OpenAPI
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_VERSION) \
		-config oapi-codegen.yaml api/openapi.yaml