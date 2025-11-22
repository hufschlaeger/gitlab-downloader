package http

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

func NewInsecureClient(proxyURL string) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if proxyURL != "" {
		if proxy, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Minute,
	}
}
