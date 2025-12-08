# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM docker.io/golang:1.25.5-alpine3.23 AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

LABEL org.opencontainers.image.source=https://gitlab.hufschlaeger.net/
LABEL org.opencontainers.image.description="gitlab.downloader"
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /go/gitlab-downloader

COPY . .

RUN apk update --no-cache \
    && apk add --no-cache make zip \
    && go mod tidy

# Cross-compilation Build
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -ldflags="-w -s" -o bin/gitlab-downloader ./cmd/...

FROM --platform=$TARGETPLATFORM alpine:3.23 AS prod

RUN apk --no-cache add ca-certificates

COPY --from=build /go/gitlab-downloader/bin/gitlab-downloader /go/bin/gitlab-downloader

# Default environment variables (können überschrieben werden)
ENV GITLAB_URL="https://gitlab.com"
ENV GITLAB_TOKEN=""

ENTRYPOINT ["/go/bin/gitlab-downloader"]
