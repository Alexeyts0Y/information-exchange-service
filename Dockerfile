# syntax=docker/dockerfile:1.7

# Версию Go можно переопределить: docker build --build-arg GO_VERSION=1.23 ...
ARG GO_VERSION=1.26

# ---------- builder ----------
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src

# Кеш зависимостей: пока go.mod/go.sum не менялись — слой переиспользуется.
# go.sum* — чтобы сборка не падала, если go.sum ещё не сгенерирован.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Статический бинарь: CGO выключен, пути вырезаны, символы удалены.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---------- runtime: app ----------
FROM gcr.io/distroless/static-debian12:nonroot AS app
WORKDIR /
COPY --from=builder /out/app /app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app"]

# ---------- runtime: migrate ----------
FROM gcr.io/distroless/static-debian12:nonroot AS migrate
WORKDIR /
COPY --from=builder /out/migrate /migrate
USER nonroot:nonroot
ENTRYPOINT ["/migrate"]