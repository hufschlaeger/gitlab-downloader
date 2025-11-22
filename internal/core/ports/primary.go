package ports

import "hufschlaeger.net/gitlab-downloader/internal/core/domain"

// ReleaseDownloadPort - Primary Port (Driver)
type ReleaseDownloadPort interface {
	DownloadRelease(req domain.DownloadRequest) error
}
