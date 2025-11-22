package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetProject_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Ensure encoded name is used in path
		if !strings.Contains(r.URL.EscapedPath(), "/api/v4/projects/group%2Fproj") {
			t.Fatalf("unexpected path: %s", r.URL.EscapedPath())
		}
		_ = json.NewEncoder(w).Encode(projectResponse{ID: 123, Name: "group/proj"})
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	proj, err := a.GetProject("group/proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.ID != 123 || proj.Name != "group/proj" {
		t.Fatalf("unexpected project: %+v", proj)
	}
}

func TestGetProject_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("not found"))
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	_, err := a.GetProject("group/proj")
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected HTTP 404 error, got %v", err)
	}
}

func TestGetProject_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{"))
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	_, err := a.GetProject("group/proj")
	if err == nil || !strings.Contains(err.Error(), "failed to decode response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestGetRelease_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !strings.Contains(r.URL.Path, "/api/v4/projects/77/releases/v1.2.3") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		resp := releaseResponse{
			TagName: "v1.2.3",
		}
		resp.Assets.Links = append(resp.Assets.Links, struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}{Name: "bin", URL: "https://example.com/bin.zip"})
		resp.Assets.Sources = append(resp.Assets.Sources, struct {
			Format string `json:"format"`
			URL    string `json:"url"`
		}{Format: "zip", URL: "https://example.com/src.zip"})
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	rel, err := a.GetRelease(77, "v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.ProjectID != 77 || rel.Tag != "v1.2.3" {
		t.Fatalf("unexpected release: %+v", rel)
	}
	if len(rel.Assets.Links) != 1 || rel.Assets.Links[0].Name != "bin" || rel.Assets.Links[0].URL != "https://example.com/bin.zip" {
		t.Fatalf("unexpected links: %+v", rel.Assets.Links)
	}
	if len(rel.Assets.Sources) != 1 || rel.Assets.Sources[0].Format != "zip" || rel.Assets.Sources[0].URL != "https://example.com/src.zip" {
		t.Fatalf("unexpected sources: %+v", rel.Assets.Sources)
	}
}

func TestGetRelease_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("oops"))
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	_, err := a.GetRelease(1, "v")
	if err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("expected HTTP 500 error, got %v", err)
	}
}

func TestGetRelease_BadJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("{"))
	}))
	defer ts.Close()

	a := NewAdapter(ts.URL, "tok", ts.Client())
	_, err := a.GetRelease(1, "v")
	if err == nil || !strings.Contains(err.Error(), "failed to decode response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
