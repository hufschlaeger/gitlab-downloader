FROM docker.io/golang:1.26.0-alpine3.23 AS build
LABEL org.opencontainers.image.source=https://gitlab.hufschlaeger.net/
LABEL org.opencontainers.image.description="gitlab.downlaoder"
LABEL org.opencontainers.image.licenses=MIT
WORKDIR /go/gitlab-downloader
COPY . .
RUN apk update --no-cache \
    && apk add --no-cache make zip golangci-lint \
    && go mod tidy \
    && make build-linux

FROM alpine:3.23 AS prod

COPY --from=build /go/gitlab-downloader/bin/gitlab-downloader_unix /go/bin/gitlab-downloader

# Default environment variables (können überschrieben werden)
ENV GITLAB_URL="https://gitlab.com"
ENV GITLAB_TOKEN=""

ENTRYPOINT ["/go/bin/markscribe"]