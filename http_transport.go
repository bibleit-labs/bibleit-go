package bibleit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBytes = 2 << 20

const MaxCommandBytes = 8192
const MaxRequestBytes = 16384

// RequestSizeError reports a request rejected locally before any network call.
type RequestSizeError struct {
	Part  string
	Bytes int
	Limit int
}

func (e *RequestSizeError) Error() string {
	return fmt.Sprintf("%s is %d bytes; maximum is %d", e.Part, e.Bytes, e.Limit)
}

func newHTTPTransport(endpoint string, supplied *http.Client) (string, *http.Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return "", nil, errors.New("invalid server endpoint")
	}
	if u.Scheme == "http" && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" && u.Hostname() != "::1" {
		return "", nil, errors.New("remote server endpoints require HTTPS")
	}
	h := &http.Client{Timeout: 15 * time.Second}
	if supplied != nil {
		*h = *supplied
	}
	// Requests must not be redirected or replayed at another endpoint.
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return strings.TrimRight(endpoint, "/"), h, nil
}

// HTTPError represents a failure without a usable protocol error envelope.
type HTTPError struct {
	StatusCode int
	Code       string
	Fields     map[string]any
}

func (e *HTTPError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Bibleit HTTP request failed with status %d: %s", e.StatusCode, e.Code)
	}
	return fmt.Sprintf("Bibleit HTTP request failed with status %d", e.StatusCode)
}
