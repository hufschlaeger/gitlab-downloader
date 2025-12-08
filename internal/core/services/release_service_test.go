package services

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
	"hufschlaeger.net/gitlab-downloader/internal/core/ports"
)

type mockGitLab struct {
	project *domain.Project
	release *domain.Release
	projErr error
	relErr  error
}

func (m *mockGitLab) GetProject(name string) (*domain.Project, error) {
	if m.projErr != nil {
		return nil, m.projErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return &domain.Project{ID: 1, Name: name}, nil
}

func (m *mockGitLab) GetRelease(projectID int, tag string) (*domain.Release, error) {
	if m.relErr != nil {
		return nil, m.relErr
	}
	if m.release != nil {
		return m.release, nil
	}
	return &domain.Release{ProjectID: projectID, Tag: tag}, nil
}

type writeCatcher struct {
	bytes.Buffer
}

func (w *writeCatcher) Close() error { return nil }

type mockFS struct {
	createErr error
	lastPath  string
	wc        *writeCatcher
}

func (m *mockFS) CreateFile(path string) (io.WriteCloser, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.lastPath = path
	m.wc = &writeCatcher{}
	return m.wc, nil
}

type mockDownloader struct {
	lastURL     string
	downloadErr error
}

func (m *mockDownloader) DownloadFromURL(url string, token string, writer io.Writer) error {
	if m.downloadErr != nil {
		return m.downloadErr
	}
	m.lastURL = url
	_, err := writer.Write([]byte("DATA"))
	return err
}

// Helper to build a service with pluggable parts
func newTestService(gl ports.GitLabPort, dl ports.DownloadPort, fs ports.FileSystemPort) *ReleaseService {
	return NewReleaseService(gl, dl, fs)
}

func TestDetermineDownloadURL_Ingest(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{ProjectID: 123, Tag: "v1.2.3"}
	url := service.determineDownloadURL("DiMAG/Ingest/IngestProzessModul", release, 0)
	expected := "https://gitlab.la-bw.de/api/v4/projects/123/packages/generic/releases/v1.2.3/ipm.v1.2.3.zip"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}

func TestDetermineDownloadURL_Access(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{
		ProjectID: 456,
		Tag:       "v2.0.0",
		Assets: domain.Assets{Links: []domain.Link{
			{ // Simulates an access file link inside repo browser URL
				Name: "ACCESS-Installer",
				URL:  "https://gitlab.la-bw.de/DiMAG/Access/AccessModul/-/blob/v2.0.0/path/to/access-installer.exe",
			},
		}},
	}
	url := service.buildAccessURL("https://gitlab.la-bw.de", "DiMAG/Access/AccessModul", release)
	expected := "https://gitlab.la-bw.de/api/v4/projects/456/repository/files/path/to/access-installer.exe/raw?ref=v2.0.0"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}

func TestDetermineDownloadURL_Generic_Artifacts(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{
		ProjectID: 42,
		Tag:       "v0.1.0",
		Assets: domain.Assets{Links: []domain.Link{
			{URL: "https://gitlab.la-bw.de/group/proj/-/jobs/123456/artifacts/download"},
		}},
	}
	url := service.buildGenericURL("https://gitlab.la-bw.de", "group/proj", release, 0)
	expected := "https://gitlab.la-bw.de/api/v4/projects/42/jobs/123456/artifacts"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}

func TestDetermineDownloadURL_Generic_Uploads_UsesSourceByIndex(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{
		ProjectID: 1,
		Tag:       "v0.0.1",
		Assets: domain.Assets{
			Links:   []domain.Link{{URL: "https://gitlab.la-bw.de/group/proj/-/uploads/abcd/file.zip"}},
			Sources: []domain.Source{{URL: "src-0"}, {URL: "src-1"}},
		},
	}
	url := service.buildGenericURL("https://gitlab.la-bw.de", "group/proj", release, 1)
	expected := "src-1"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}

func TestDetermineDownloadURL_Generic_DefaultsToFirstLink(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{
		ProjectID: 1,
		Tag:       "v0.0.1",
		Assets:    domain.Assets{Links: []domain.Link{{URL: "https://example.com/file.tgz"}}},
	}
	url := service.buildGenericURL("https://gitlab.la-bw.de", "group/proj", release, 0)
	expected := "https://example.com/file.tgz"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}

func TestDetermineDownloadURL_NoLinksOrSources_ReturnsEmpty(t *testing.T) {
	service := newTestService(nil, nil, nil)
	release := &domain.Release{ProjectID: 99, Tag: "v9.9.9"}
	url := service.buildGenericURL("https://gitlab.la-bw.de", "group/proj", release, 0)
	if url != "" {
		t.Fatalf("expected empty url, got %q", url)
	}
}

func TestDownloadRelease_Success(t *testing.T) {
	gl := &mockGitLab{
		project: &domain.Project{ID: 77, Name: "group/proj"},
		release: &domain.Release{
			ProjectID: 77,
			Tag:       "v1.0.0",
			Assets:    domain.Assets{Links: []domain.Link{{URL: "https://example.com/app.zip"}}},
		},
	}
	dl := &mockDownloader{}
	fs := &mockFS{}
	service := newTestService(gl, dl, fs)

	req := domain.DownloadRequest{ProjectName: "group/proj", ReleaseTag: "v1.0.0", OutputPath: "out.zip", ExtIndex: 0}
	if err := service.DownloadRelease(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dl.lastURL != "https://example.com/app.zip" {
		t.Fatalf("expected download url to be propagated, got %q", dl.lastURL)
	}
	if fs.wc == nil || fs.wc.Len() == 0 {
		t.Fatalf("expected data written to file, buffer len=%d", fs.wc.Len())
	}
}

func TestDownloadRelease_Errors(t *testing.T) {
	cases := []struct {
		name      string
		gitlab    *mockGitLab
		fs        *mockFS
		dl        *mockDownloader
		req       domain.DownloadRequest
		expectErr string
	}{
		{
			name:      "GetProject fails",
			gitlab:    &mockGitLab{projErr: errors.New("boom")},
			fs:        &mockFS{},
			dl:        &mockDownloader{},
			req:       domain.DownloadRequest{ProjectName: "p", ReleaseTag: "t", OutputPath: "out"},
			expectErr: "failed to get project",
		},
		{
			name:      "GetRelease fails",
			gitlab:    &mockGitLab{project: &domain.Project{ID: 1}, relErr: errors.New("boom")},
			fs:        &mockFS{},
			dl:        &mockDownloader{},
			req:       domain.DownloadRequest{ProjectName: "p", ReleaseTag: "t", OutputPath: "out"},
			expectErr: "failed to get release",
		},
		{
			name:   "No URL found",
			gitlab: &mockGitLab{project: &domain.Project{ID: 1}, release: &domain.Release{ProjectID: 1, Tag: "t"}},
			fs:     &mockFS{}, dl: &mockDownloader{},
			req:       domain.DownloadRequest{ProjectName: "p", ReleaseTag: "t", OutputPath: "out"},
			expectErr: "no download URL",
		},
		{
			name:      "CreateFile fails",
			gitlab:    &mockGitLab{project: &domain.Project{ID: 1}, release: &domain.Release{ProjectID: 1, Tag: "t", Assets: domain.Assets{Links: []domain.Link{{URL: "https://example.com"}}}}},
			fs:        &mockFS{createErr: errors.New("disk full")},
			dl:        &mockDownloader{},
			req:       domain.DownloadRequest{ProjectName: "p", ReleaseTag: "t", OutputPath: "out"},
			expectErr: "failed to create output file",
		},
		{
			name:      "Download fails",
			gitlab:    &mockGitLab{project: &domain.Project{ID: 1}, release: &domain.Release{ProjectID: 1, Tag: "t", Assets: domain.Assets{Links: []domain.Link{{URL: "https://example.com"}}}}},
			fs:        &mockFS{},
			dl:        &mockDownloader{downloadErr: errors.New("net")},
			req:       domain.DownloadRequest{ProjectName: "p", ReleaseTag: "t", OutputPath: "out"},
			expectErr: "download failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newTestService(tc.gitlab, tc.dl, tc.fs)
			err := service.DownloadRelease(tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.expectErr) {
				t.Fatalf("expected error containing %q, got %v", tc.expectErr, err)
			}
		})
	}
}
