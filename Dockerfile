# Build auf der nativen Plattform des Runners (amd64)
FROM golang:1.25.5-alpine3.23 AS build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /go/gitlab-downloader

COPY . .
RUN apk update --no-cache \
    && apk add --no-cache make zip \
    && go mod tidy \
    && CGO_ENABLED=0 \
       GOOS=${TARGETOS} \
       GOARCH=${TARGETARCH} \
       go build -ldflags="-w -s" -o gitlab-downloader ./cmd/...

# Prod-Stage mit Zielplattform
FROM alpine:3.23 AS prod

LABEL org.opencontainers.image.source=https://gitlab.hufschlaeger.net/
LABEL org.opencontainers.image.description="gitlab.downloader"
LABEL org.opencontainers.image.licenses=MIT

# Nur das Binary kopieren (kein apk add nötig wenn QEMU fehlt)
WORKDIR /app
COPY --from=build /go/gitlab-downloader/gitlab-downloader /app/

ENTRYPOINT ["/app/gitlab-downloader"]
