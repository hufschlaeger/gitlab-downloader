package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/schollz/progressbar/v3"
)

type DownloadAdapter struct {
	client *http.Client
}

func NewDownloadAdapter(client *http.Client) *DownloadAdapter {
	return &DownloadAdapter{client: client}
}

func (a *DownloadAdapter) DownloadFromURL(url string, writer io.Writer) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = fmt.Errorf("failed to close connection: %w", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	_, err = io.Copy(io.MultiWriter(writer, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}
