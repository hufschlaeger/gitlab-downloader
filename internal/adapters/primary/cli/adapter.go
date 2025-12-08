package cli

import (
	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
	"hufschlaeger.net/gitlab-downloader/internal/core/ports"
)

type Adapter struct {
	service ports.ReleaseDownloadPort
}

func NewAdapter(service ports.ReleaseDownloadPort) *Adapter {
	return &Adapter{service: service}
}

func (a *Adapter) DownloadRelease(config *Config) error {
	req := domain.DownloadRequest{
		ProjectName: config.Project,
		ReleaseTag:  config.Release,
		OutputPath:  config.Output,
		ExtIndex:    config.ExtIndex,
		Token:       config.Token,
	}

	return a.service.DownloadRelease(req)
}
