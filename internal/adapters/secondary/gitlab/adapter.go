package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"hufschlaeger.net/gitlab-downloader/internal/core/domain"
)

type Adapter struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewAdapter(baseURL, token string, httpClient *http.Client) *Adapter {
	return &Adapter{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpClient,
	}
}

func (a *Adapter) GetProject(name string) (*domain.Project, error) {
	encodedName := url.PathEscape(name)
	url := fmt.Sprintf("%s/api/v4/projects/%s", a.baseURL, encodedName)

	var response projectResponse
	if err := a.doRequest(url, &response); err != nil {
		return nil, err
	}

	return &domain.Project{
		ID:   response.ID,
		Name: response.Name,
	}, nil
}

func (a *Adapter) GetRelease(projectID int, tag string) (*domain.Release, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/releases/%s", a.baseURL, projectID, tag)

	var response releaseResponse
	if err := a.doRequest(url, &response); err != nil {
		return nil, err
	}

	return a.mapToRelease(projectID, &response), nil
}

func (a *Adapter) doRequest(url string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", a.token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (a *Adapter) mapToRelease(projectID int, response *releaseResponse) *domain.Release {
	release := &domain.Release{
		ProjectID: projectID,
		Tag:       response.TagName,
	}

	for _, link := range response.Assets.Links {
		release.Assets.Links = append(release.Assets.Links, domain.Link{
			Name: link.Name,
			URL:  link.URL,
		})
	}

	for _, source := range response.Assets.Sources {
		release.Assets.Sources = append(release.Assets.Sources, domain.Source{
			Format: source.Format,
			URL:    source.URL,
		})
	}

	return release
}

// DTOs
type projectResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
	Assets  struct {
		Links []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"links"`
		Sources []struct {
			Format string `json:"format"`
			URL    string `json:"url"`
		} `json:"sources"`
	} `json:"assets"`
}
