FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine3.23 AS build

ARG TARGETOS=linux
ARG TARGETARCH
ARG BUILDPLATFORM

WORKDIR /go/gitlab-downloader

COPY . .
RUN apk update --no-cache \
    && apk add --no-cache make zip \
    && go mod tidy \
    && CGO_ENABLED=0 \
       GOOS=${TARGETOS} \
       GOARCH=${TARGETARCH} \
       go build -ldflags="-w -s" -o gitlab-downloader ./cmd/...

FROM --platform=$BUILDPLATFORM alpine:3.23 AS prod
LABEL org.opencontainers.image.source=https://gitlab.hufschlaeger.net/
LABEL org.opencontainers.image.description="gitlab.downloader"
LABEL org.opencontainers.image.licenses=MIT

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=build /go/gitlab-downloader/gitlab-downloader /app/

ENTRYPOINT ["/app/gitlab-downloader"]