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

func (a *DownloadAdapter) DownloadFromURL(url string, token string, writer io.Writer) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// GitLab Authentication
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Bei 401/403 hilfreiche Fehlermeldung
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("authentication failed (HTTP %d): please provide a valid GitLab token", resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var bar *progressbar.ProgressBar

	if resp.ContentLength > 0 {
		bar = progressbar.DefaultBytes(
			resp.ContentLength,
			"downloading",
		)
	} else {
		bar = progressbar.NewOptions64(
			-1,
			progressbar.OptionSetDescription("downloading"),
			progressbar.OptionShowBytes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(15),
			progressbar.OptionThrottle(100),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		)
	}

	written, err := io.Copy(io.MultiWriter(writer, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	fmt.Printf("\nDownload completed successfully: %d bytes\n", written)

	return nil
}
