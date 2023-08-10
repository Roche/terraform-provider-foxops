package helpers

import (
	"fmt"
	"net/http"
	"runtime"
)

type transport struct {
	version    string
	HTTPClient http.RoundTripper
}

func NewTransport(version string, httpClient http.RoundTripper) http.RoundTripper {
	return &transport{version, httpClient}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", fmt.Sprintf("Foxops Terraform Provider/%s (%s; %s)", t.version, runtime.GOOS, runtime.GOARCH))
	return t.HTTPClient.RoundTrip(req)
}
