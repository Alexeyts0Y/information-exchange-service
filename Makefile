PROJECT_NAME = information-exchange-service

# Golang
compile :
	CGO_ENABLED=0 go build -o ./bin/${PROJECT_NAME} ./cmd/app

go:	compile
	./bin/${PROJECT_NAME}

lint:
	golangci-lint run ./... --fix

fmt:
	gofmt -s -w .
	goimports -l -w .

test:
	go test -v ./internal/...

# Docker
down:
	docker-compose down

build:
	docker-compose build

up:
	docker-compose up

run: down build up

stop:
	docker compose stop

remove:
	docker compose down --remove-orphans

gen:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml api/openapi.yaml
