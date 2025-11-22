# Variables
BINARY_NAME=gitlab-downloader
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin
BINARY_ARM64=$(BINARY_NAME)_arm64

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build configuration
MAIN_FILE=cmd/gitlab-downloader/main.go
BUILD_FLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
BUILD_DIR=./bin

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

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin build-arm64

.PHONY: build-linux
build-linux:
	@echo "🔨 Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) $(MAIN_FILE)

.PHONY: build-windows
build-windows:
	@echo "🔨 Building for Windows amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) $(MAIN_FILE)

.PHONY: build-darwin
build-darwin:
	@echo "🔨 Building for macOS amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) $(MAIN_FILE)

.PHONY: build-arm64
build-arm64:
	@echo "🔨 Building for macOS arm64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_ARM64) $(MAIN_FILE)


# Create release archives
.PHONY: release
release: clean build-all
	@echo "📦 Creating release archives v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)/releases
	@mkdir -p $(BUILD_DIR)/temp

	# Linux
	@cp $(BUILD_DIR)/$(BINARY_UNIX) $(BUILD_DIR)/temp/$(BINARY_NAME)
	@tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR)/temp $(BINARY_NAME)

	# Darwin amd64
	@cp $(BUILD_DIR)/$(BINARY_DARWIN) $(BUILD_DIR)/temp/$(BINARY_NAME)
	@tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR)/temp $(BINARY_NAME)

	# Darwin arm64
	@cp $(BUILD_DIR)/$(BINARY_ARM64) $(BUILD_DIR)/temp/$(BINARY_NAME)
	@tar -czf $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(BUILD_DIR)/temp $(BINARY_NAME)

	# Windows
	@cd $(BUILD_DIR) && zip -j releases/$(BINARY_NAME)-windows-amd64.zip $(BINARY_WINDOWS)

	@echo "📦 Release archives created in $(BUILD_DIR)/releases/"

# Run tests
test:
	go test -v ./...

# Clean
clean:
	rm -rf bin/ downloads/
	docker rmi gitlab-downloader:latest gitlab-downloader:minimal 2>/dev/null || true
