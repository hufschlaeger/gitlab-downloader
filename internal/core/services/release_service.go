package services

import (
	"fmt"
	"strings"

	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
	"hufschlaeger.net/gitlab-downloader/internal/core/ports"
)

type ReleaseService struct {
	gitlab     ports.GitLabPort
	downloader ports.DownloadPort
	filesystem ports.FileSystemPort
}

func NewReleaseService(
	gitlab ports.GitLabPort,
	downloader ports.DownloadPort,
	filesystem ports.FileSystemPort,
) *ReleaseService {
	return &ReleaseService{
		gitlab:     gitlab,
		downloader: downloader,
		filesystem: filesystem,
	}
}

func (s *ReleaseService) DownloadRelease(req domain.DownloadRequest) error {
	// Get project
	project, err := s.gitlab.GetProject(req.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Get release
	release, err := s.gitlab.GetRelease(project.ID, req.ReleaseTag)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	// Determine download URL
	url := s.determineDownloadURL(req.ProjectName, release, req.ExtIndex)
	if url == "" {
		return fmt.Errorf("no download URL found")
	}

	// Create output file
	file, err := s.filesystem.CreateFile(req.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Download
	if err := s.downloader.DownloadFromURL(url, file); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

func (s *ReleaseService) determineDownloadURL(projectName string, release *domain.Release, extIndex int) string {
	projectLower := strings.ToLower(projectName)
	host := "https://gitlab.la-bw.de"

	switch projectLower {
	case "dimag/ingest/ingestprozessmodul":
		return s.buildIngestURL(host, release)
	case "dimag/access/accessmodul":
		return s.buildAccessURL(host, projectName, release)
	default:
		return s.buildGenericURL(host, projectName, release, extIndex)
	}
}

func (s *ReleaseService) buildIngestURL(host string, release *domain.Release) string {
	return fmt.Sprintf("%s/api/v4/projects/%d/packages/generic/releases/%s/ipm.%s.zip",
		host, release.ProjectID, release.Tag, release.Tag)
}

func (s *ReleaseService) buildAccessURL(host, projectName string, release *domain.Release) string {
	for _, link := range release.Assets.Links {
		if strings.Contains(strings.ToLower(link.Name), "access") {
			urlPath := strings.Replace(
				link.URL,
				fmt.Sprintf("%s/%s/-/blob/", host, projectName),
				"",
				1,
			)
			parts := strings.SplitN(urlPath, "/", 2)
			if len(parts) == 2 {
				return fmt.Sprintf("%s/api/v4/projects/%d/repository/files/%s/raw?ref=%s",
					host, release.ProjectID, parts[1], parts[0])
			}
		}
	}
	return ""
}

func (s *ReleaseService) buildGenericURL(host, projectName string, release *domain.Release, extIndex int) string {
	if len(release.Assets.Links) > 0 {
		url := release.Assets.Links[0].URL

		if strings.Contains(url, "artifacts") {
			url = strings.Replace(
				url,
				fmt.Sprintf("%s/%s/-/", host, projectName),
				fmt.Sprintf("%s/api/v4/projects/%d/", host, release.ProjectID),
				1,
			)
			return strings.TrimSuffix(url, "/download")
		}

		if strings.Contains(url, "uploads") && extIndex < len(release.Assets.Sources) {
			return release.Assets.Sources[extIndex].URL
		}

		return url
	}

	if extIndex < len(release.Assets.Sources) {
		return release.Assets.Sources[extIndex].URL
	}

	return ""
}
