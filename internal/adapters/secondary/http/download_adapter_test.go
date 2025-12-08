package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errWriter struct{ wrote int }

func (w *errWriter) Write(p []byte) (int, error) {
	w.wrote += len(p)
	return w.wrote, errors.New("sink error")
}

func TestDownloadAdapter_Success(t *testing.T) {
	// Serve some bytes with content-length
	body := strings.Repeat("x", 1024)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		_, _ = io.WriteString(w, body)
	}))
	defer ts.Close()

	client := &http.Client{}
	var Token string

	a := NewDownloadAdapter(client)

	var buf strings.Builder
	if err := a.DownloadFromURL(ts.URL, Token, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.String() != body {
		t.Fatalf("unexpected body size: got %d want %d", len(buf.String()), len(body))
	}
}

func TestDownloadAdapter_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("not found"))
	}))
	defer ts.Close()

	client := &http.Client{}
	a := NewDownloadAdapter(client)
	var buf strings.Builder
	var Token string
	err := a.DownloadFromURL(ts.URL, Token, &buf)
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected HTTP 404 error, got %v", err)
	}
}

func TestDownloadAdapter_RequestBuildError(t *testing.T) {
	client := &http.Client{}
	a := NewDownloadAdapter(client)
	var buf strings.Builder
	var Token string
	err := a.DownloadFromURL(":", Token, &buf) // invalid URL triggers http.NewRequest error
	if err == nil || !strings.Contains(err.Error(), "failed to create request") {
		t.Fatalf("expected request creation error, got %v", err)
	}
}

func TestDownloadAdapter_WriterError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		_, _ = io.WriteString(w, "0123456789")
	}))
	defer ts.Close()

	client := &http.Client{}
	a := NewDownloadAdapter(client)
	ew := &errWriter{}
	var token string
	err := a.DownloadFromURL(ts.URL, token, ew)
	if err == nil || !strings.Contains(err.Error(), "download failed") {
		t.Fatalf("expected download failed due to writer error, got %v", err)
	}
}
