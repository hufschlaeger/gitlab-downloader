# GitLab Release Downloader

A tiny, batteries‑included CLI to fetch artifacts or source archives from GitLab releases and save them locally. It speaks directly to the GitLab API, supports proxies, shows a progress bar while downloading, and can be built into a minimal Docker image.


## ✨ Features
- Simple one‑binary CLI (no runtime deps)
- Works with any GitLab instance (self‑hosted or gitlab.com)
- Auth via `PRIVATE-TOKEN`
- Proxy support (`HTTPS_PROXY`/`HTTP_PROXY` or `-proxy`)
- Smart URL handling for common GitLab release layouts:
  - Project‑specific rules used in this repo (see How it chooses what to download)
  - GitLab CI job artifacts links → converted to API URLs
  - Upload links with fallback to release sources by extension index
- Progress bar during download


## 🚀 Quickstart
1) Build the binary
```bash
# From repository root
go build -o gitlab-downloader ./cmd/gitlab-downloader
```

2) Run it
```bash
./gitlab-downloader \
  -gitlab-url https://gitlab.com \
  -token "$GITLAB_TOKEN" \
  -project group/project \
  -release v1.2.3 \
  -out ./artifact.zip
```

If you don’t pass `-gitlab-url`, it defaults to `https://gitlab.com` or reads `GITLAB_URL` from the environment.


## 🔧 CLI
Flags are defined in `internal/adapters/primary/cli/config.go` and surfaced via `cmd/gitlab-downloader/main.go`.

```text
-gitlab-url string   GitLab instance URL (defaults to env GITLAB_URL or https://gitlab.com)
-token string        Your private GitLab token (required) (alias: -t)
-proxy string        Proxy URL (e.g. http://proxy.local:8080)
-ext int             Source extension index (0=zip, 1=tar.gz, 2=tar.bz2, 3=tar)
-out string          Path to store the release (required) (alias: -o)
-release string      Version string of release (required) (alias: -r)
-project string      Project name with namespace/group (required) (alias: -p)
```

Environment variables
- `GITLAB_URL` — GitLab base URL if `-gitlab-url` not provided
- `GITLAB_TOKEN` — token if `-token`/`-t` not provided
- `HTTPS_PROXY` / `HTTP_PROXY` — used if `-proxy` is not provided


## 🧠 How it chooses what to download
The core logic lives in `internal/core/services/release_service.go`.

- Special projects (opinionated rules used here):
  - `DiMAG/Ingest/IngestProzessModul`: pulls a generic package URL
  - `DiMAG/Access/AccessModul`: converts a repository “blob” link into the GitLab raw file API URL
- Generic behavior:
  - If the first asset link looks like a CI artifacts download (`/-/jobs/.../artifacts/download`), it is converted to the corresponding API endpoint (`/api/v4/projects/{id}/jobs/.../artifacts`).
  - If the first asset link is an uploads URL and there are release sources, it selects a source URL by `-ext` index.
  - Otherwise falls back to the first asset link.
  - If there are no links, it tries `Sources[extIndex]`.

Tip: Use `-ext` to switch between `sources` entries (e.g., zip vs tar.gz) when a release provides multiple source formats.


## 🧪 Tests
The project includes a comprehensive unit test suite that is fully hermetic (no network or real filesystem writes).

Run all tests:
```bash
go test ./...
```

The suite covers:
- CLI flag/env parsing and validation
- HTTP client behavior (timeout, TLS, proxy)
- Download adapter (success, non‑200 responses, writer errors)
- File adapter (creation, error wrapping)
- GitLab adapter (request headers, path encoding, JSON mapping, error handling)
- Release service (URL selection and end‑to‑end flow with mocks)


## 🐳 Docker
Two Dockerfiles are provided:

- `Dockerfile` — standard multi‑stage build (alpine base)
- `Dockerfile.scratch` — builds a minimal static binary and runs it in `scratch`

Example build with the scratch image:
```bash
docker build -f Dockerfile.scratch -t gitlab-downloader:latest .
```
Run:
```bash
docker run --rm \
  -e GITLAB_URL=https://gitlab.com \
  -e GITLAB_TOKEN=YOUR_TOKEN \
  -v "$PWD":/data \
  gitlab-downloader:latest \
  -project group/project -release v1.2.3 -out /data/artifact.zip
```


## 🔒 TLS and proxies
- The HTTP client is configured with `InsecureSkipVerify: true` to accommodate internal CAs. For production with properly trusted CAs, you may want to adjust this.
- Proxy can be specified either via `-proxy` or environment (`HTTPS_PROXY` or `HTTP_PROXY`).


## 🧩 Architecture (ports & adapters)
- Domain and ports: `internal/core/domain`, `internal/core/ports`
- Core service (business rules): `internal/core/services/release_service.go`
- Drivers (primary): `internal/adapters/primary/cli`
- Driven (secondary):
  - GitLab API adapter: `internal/adapters/secondary/gitlab`
  - HTTP download + file adapters: `internal/adapters/secondary/http`

Entry point: `cmd/gitlab-downloader/main.go`


## 🛠️ Build from source
```bash
go build -o gitlab-downloader ./cmd/gitlab-downloader
```
Cross‑compile examples:
```bash
GOOS=linux   GOARCH=amd64 go build -o gitlab-downloader-linux-amd64   ./cmd/gitlab-downloader
GOOS=darwin  GOARCH=arm64 go build -o gitlab-downloader-darwin-arm64  ./cmd/gitlab-downloader
GOOS=windows GOARCH=amd64 go build -o gitlab-downloader-windows-amd64.exe ./cmd/gitlab-downloader
```

Go version: `go 1.25.4`


## 📎 Example commands
- Download first asset link:
```bash
./gitlab-downloader -t "$GITLAB_TOKEN" -p group/proj -r v1.0.0 -o out.zip
```

- Prefer release sources (e.g., tar.gz):
```bash
./gitlab-downloader -t "$GITLAB_TOKEN" -p group/proj -r v1.0.0 -ext 1 -o src.tar.gz
```

- Via proxy:
```bash
HTTPS_PROXY=http://proxy.local:8080 \
./gitlab-downloader -t "$GITLAB_TOKEN" -p group/proj -r v1.0.0 -o out.zip
```


## 🧰 CI
A GitLab CI pipeline is included with stages:
- Lint (golangci-lint)
- Test (gotestsum + JUnit report)

You can extend it to build and publish release artifacts by enabling the commented sections in `.gitlab-ci.yml`.


## 📄 License
Add your preferred license here (e.g., MIT). If you already have a LICENSE file, mention it and its terms.
