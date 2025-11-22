package cli

import (
	"errors"
	"testing"

	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
	"hufschlaeger.net/gitlab-downloader/internal/core/ports"
)

type mockService struct {
	received domain.DownloadRequest
	retErr   error
}

func (m *mockService) DownloadRelease(req domain.DownloadRequest) error {
	m.received = req
	return m.retErr
}

func TestAdapter_DownloadRelease_PassesThroughConfig(t *testing.T) {
	ms := &mockService{}
	a := NewAdapter(ms)
	cfg := &Config{
		Project:  "group/proj",
		Release:  "v1.2.3",
		Output:   "out.zip",
		ExtIndex: 1,
	}

	if err := a.DownloadRelease(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := domain.DownloadRequest{ProjectName: "group/proj", ReleaseTag: "v1.2.3", OutputPath: "out.zip", ExtIndex: 1}
	if ms.received != want {
		t.Fatalf("unexpected request: %+v", ms.received)
	}
}

func TestAdapter_DownloadRelease_PropagatesError(t *testing.T) {
	ms := &mockService{retErr: errors.New("boom")}
	a := NewAdapter(ms)
	cfg := &Config{}
	if err := a.DownloadRelease(cfg); err == nil {
		t.Fatalf("expected error")
	}
}

// ensure mockService implements the interface
var _ ports.ReleaseDownloadPort = (*mockService)(nil)
