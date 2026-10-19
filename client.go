package bibleit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ServerEndpoint is the project-managed browser authentication origin. Applications
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
	execute  func(context.Context, Command) (Response, error)
}

// NewClient creates a credential client for browser-issued CLI token revocation.
// It cannot execute commands. Use package rest for HTTP resource operations.
// The default HTTP timeout is 15 seconds. Operations are never retried automatically.
func NewClient(apiToken string) (*Client, error) {
	return NewClientWithHTTPClient(apiToken, nil)
}

// NewClientWithHTTPClient configures browser-credential revocation transport/timeouts.
// The origin stays fixed; the supplied client is copied and redirects are blocked.
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

// NewSSHClient binds command helpers to an explicit caller-owned SSH executor.
// Bearer credential clients cannot execute commands; use package rest for HTTP.
func NewSSHClient(execute func(context.Context, Command) (Response, error)) (*Client, error) {
	if execute == nil {
		return nil, errors.New("an SSH command executor is required")
	}
	return &Client{execute: execute}, nil
}

func (c *Client) Execute(ctx context.Context, command Command) (Response, error) {
	if c.execute == nil {
		return Response{}, errors.New("command HTTP transport has been removed; use the REST client or an explicit SSH executor")
	}
	if command.wire == "" {
		return Response{}, errors.New("invalid zero-value command")
	}
	if command.streaming {
		return Response{}, errors.New("streaming commands require a persistent SSH subscription")
	}
	if len(command.wire) > MaxCommandBytes {
		return Response{}, &RequestSizeError{Part: "command", Bytes: len(command.wire), Limit: MaxCommandBytes}
	}
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	return c.execute(ctx, command)
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
