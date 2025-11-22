package ports

import (
	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
	"io"
)

// GitLabPort - Secondary Port (Driven)
type GitLabPort interface {
	GetProject(name string) (*domain.Project, error)
	GetRelease(projectID int, tag string) (*domain.Release, error)
}

// DownloadPort - Secondary Port (Driven)
type DownloadPort interface {
	DownloadFromURL(url string, writer io.Writer) error
}

// FileSystemPort - Secondary Port (Driven)
type FileSystemPort interface {
	CreateFile(path string) (io.WriteCloser, error)
}
