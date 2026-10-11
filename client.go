package bibleit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ServerEndpoint is the project-managed Bibleit service address. Applications
// cannot override it through configuration or environment variables. HTTPS
// certificate verification is enabled by the default transport.
const ServerEndpoint = "https://bibleit.app"

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

type Client struct {
	endpoint string
	apiToken string
	http     *http.Client
}

// NewClient creates a client using an API token and the project-managed service.
// The default HTTP timeout is 15 seconds. Operations are never retried automatically.
func NewClient(apiToken string) (*Client, error) {
	return NewClientWithHTTPClient(apiToken, nil)
}

// NewClientWithHTTPClient configures HTTP transport/timeouts while retaining the
// fixed service endpoint. The supplied HTTP client is copied and redirects blocked.
func NewClientWithHTTPClient(apiToken string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(apiToken) == "" || strings.ContainsAny(apiToken, "\r\n\x00") {
		return nil, errors.New("a valid API token is required")
	}
	endpoint, h, err := newHTTPTransport(ServerEndpoint, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{endpoint: endpoint, apiToken: apiToken, http: h}, nil
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

func (c *Client) Execute(ctx context.Context, command Command) (Response, error) {
	if command.wire == "" {
		return Response{}, errors.New("invalid zero-value command")
	}
	if command.streaming {
		return Response{}, &ServerError{Code: "streaming_requires_ssh"}
	}
	if len(command.wire) > MaxCommandBytes {
		return Response{}, &RequestSizeError{Part: "command", Bytes: len(command.wire), Limit: MaxCommandBytes}
	}
	body, err := json.Marshal(map[string]string{"command": command.wire})
	if err != nil {
		return Response{}, err
	}
	if len(body) > MaxRequestBytes {
		return Response{}, &RequestSizeError{Part: "JSON body", Bytes: len(body), Limit: MaxRequestBytes}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/cli/command", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer res.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if readErr != nil {
		return Response{}, readErr
	}
	if len(responseBody) > maxResponseBytes {
		return Response{}, errors.New("protocol response exceeds 2 MiB")
	}
	response, decodeErr := DecodeCommandResponse(command, bytes.NewReader(responseBody))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var serverErr *ServerError
		if errors.As(decodeErr, &serverErr) {
			return Response{}, serverErr
		}
		var fields map[string]any
		if json.Unmarshal(responseBody, &fields) == nil {
			if code, ok := fields["error"].(string); ok {
				return Response{}, &HTTPError{StatusCode: res.StatusCode, Code: code, Fields: fields}
			}
		}
	}
	if res.StatusCode == http.StatusUnauthorized {
		return Response{}, &ServerError{Code: "unauthorized"}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Response{}, &HTTPError{StatusCode: res.StatusCode}
	}
	return response, decodeErr
}

func (c *Client) Read(ctx context.Context, ref Reference) (Response, error) {
	command, err := ReadCommand(ref)
	if err != nil {
		return Response{}, err
	}
	return c.Execute(ctx, command)
}
func (c *Client) Search(ctx context.Context, translation, query string) (Response, error) {
	command, err := SearchCommand(translation, query)
	if err != nil {
		return Response{}, err
	}
	return c.Execute(ctx, command)
}
func (c *Client) Ping(ctx context.Context) (Response, error) { return c.Execute(ctx, PingCommand()) }
func (c *Client) Identity(ctx context.Context) (Response, error) {
	return c.Execute(ctx, IdentityCommand())
}
func (c *Client) ListTranslations(ctx context.Context) (Response, error) {
	return c.Execute(ctx, TranslationListCommand())
}
