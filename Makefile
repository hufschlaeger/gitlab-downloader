# Makefile
.PHONY: build build-docker run test clean

# Build binary
build:
	go build -o bin/gitlab-downloader ./cmd/gitlab-downloader

# Build Docker image
build-docker:
	podman build -t gitlab-downloader:latest .

# Build minimal Docker image
build-docker-minimal:
	podman build -f Dockerfile.scratch -t gitlab-downloader:minimal .

# Run locally (CLI)
run:
	@export GITLAB_TOKEN=$${GITLAB_TOKEN} && \
	./bin/gitlab-downloader \
		-project "dimag/ingest/ingestprozessmodul" \
		-release "3.5.0" \
		-out "./downloads/release.zip"

# Run in Docker
run-docker:
	docker run --rm \
		-e GITLAB_TOKEN=$${GITLAB_TOKEN} \
		-e GITLAB_URL=https://gitlab.la-bw.de \
		-v $$(pwd)/downloads:/downloads \
		gitlab-downloader:latest \
		-project "dimag/ingest/ingestprozessmodul" \
		-release "3.5.0" \
		-out "/downloads/release.zip"

# Run tests
test:
	go test -v ./...

# Clean
clean:
	rm -rf bin/ downloads/
	docker rmi gitlab-downloader:latest gitlab-downloader:minimal 2>/dev/null || true
