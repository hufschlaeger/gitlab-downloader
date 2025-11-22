package http

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNewInsecureClient_ConfiguresTransportAndTimeout(t *testing.T) {
	client := NewInsecureClient("")
	if client.Timeout != 30*time.Minute {
		t.Fatalf("expected 30m timeout, got %v", client.Timeout)
	}

	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify=true")
	}
	if tr.Proxy != nil {
		t.Fatalf("expected no proxy when proxyURL empty")
	}
}

func TestNewInsecureClient_WithProxy(t *testing.T) {
	proxy := &url.URL{Scheme: "http", Host: "proxy.local:8080"}
	client := NewInsecureClient(proxy.String())
	tr := client.Transport.(*http.Transport)
	if tr.Proxy == nil {
		t.Fatalf("expected proxy function to be set")
	}
	// The proxy func returns the configured URL regardless of request
	u, err := tr.Proxy(&http.Request{})
	if err != nil || u == nil || u.String() != proxy.String() {
		t.Fatalf("expected proxy URL %q, got %v (err=%v)", proxy.String(), u, err)
	}
}
