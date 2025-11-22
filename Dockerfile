# Dockerfile
FROM golang:1.25.4-alpine AS builder

WORKDIR /build

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gitlab-downloader ./cmd/gitlab-downloader

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/gitlab-downloader .

# Default environment variables (können überschrieben werden)
ENV GITLAB_URL="https://gitlab.com"
ENV GITLAB_TOKEN=""

ENTRYPOINT ["/app/gitlab-downloader"]
