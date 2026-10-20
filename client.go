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
	"sort"
	"strings"
	"time"
)

// ServerEndpoint is the project-managed browser authentication origin. Applications
// cannot override it through configuration or environment variables. HTTPS
// certificate verification is enabled by the default transport.
const ServerEndpoint = "https://bibleit.app"

// Endpoint is the fixed resource API origin.
const Endpoint = "https://api.bibleit.app/v1"
const maxJSONBytes = 4 << 20

// Client provides resource operations, pagination and credential revocation.
type Client struct {
	token string
	http  *http.Client
}

// NewClient creates a resource client. It accepts personal bt_, delegated oa_, or organization svc_ credentials.
// The server enforces each credential's resource boundaries and scopes. An empty
// token is permitted for public discovery. Pass an optional HTTP client to configure
// transport and timeouts. No operation is automatically retried.
func NewClient(token string, transports ...*http.Client) (*Client, error) {
	if strings.ContainsAny(token, "\r\n\x00") || strings.TrimSpace(token) != token {
		return nil, errors.New("invalid bearer token")
	}
	if len(transports) > 1 {
		return nil, errors.New("at most one HTTP client is supported")
	}
	var transport *http.Client
	if len(transports) == 1 {
		transport = transports[0]
	}
	h := &http.Client{Timeout: 40 * time.Second}
	if transport != nil {
		*h = *transport
	}
	h.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	h.Jar = nil // Browser sessions do not authenticate REST resources.
	return &Client{token: token, http: h}, nil
}

// NewClientWithHTTPClient creates a resource client with a custom HTTP transport.
func NewClientWithHTTPClient(token string, transport *http.Client) (*Client, error) {
	return NewClient(token, transport)
}

type Operation struct{ Method, Path, Summary string }

// Operations returns a copy of the pinned contract's REST operation directory.
func Operations() map[string]Operation {
	out := make(map[string]Operation, len(operations))
	for k, v := range operations {
		out[k] = v
	}
	return out
}
func OperationNames() []string {
	names := make([]string, 0, len(operations))
	for k := range operations {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

type Options struct {
	Query url.Values
	JSON  any
	// Body streams raw uploads/imports. It is mutually exclusive with JSON.
	Body        io.Reader
	ContentType string
	// ContentLength is required for raw uploads; use the exact byte length.
	ContentLength                                    int64
	IfMatch, IdempotencyKey, DestinationToken, Range string
}

// Response retains ETag, Location, Retry-After and request IDs in Header.
// The caller must close Body, including after Decode. Binary downloads are
// streamed directly from Body, and are never retried or buffered by the SDK.
type Response struct{ *http.Response }

func (r *Response) Decode(target any) error {
	b, err := io.ReadAll(io.LimitReader(r.Body, maxJSONBytes+1))
	if err != nil {
		return err
	}
	if len(b) > maxJSONBytes {
		return errors.New("REST JSON response exceeds 4 MiB")
	}
	return json.Unmarshal(b, target)
}

type ProblemError struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	// Detail is server-supplied guidance. Error() intentionally omits it.
	Detail    string      `json:"detail,omitempty"`
	Status    int         `json:"status"`
	Code      string      `json:"code"`
	RequestID string      `json:"request_id"`
	Header    http.Header `json:"-"`
}

func (e *ProblemError) Error() string {
	return fmt.Sprintf("Bibleit REST HTTP %d: %s", e.Status, e.Code)
}
func (e *ProblemError) RetryAfter() time.Duration {
	v := e.Header.Get("Retry-After")
	if n, err := time.ParseDuration(v + "s"); err == nil && n >= 0 {
		return n
	}
	return 0
}

// Call invokes one named operation from the checked-in OpenAPI contract.
// Paths cannot select another host or bypass the pinned /v1 API. Conditional
// writes use the caller's exact ETag; 412/428 are returned without retry.
func (c *Client) Call(ctx context.Context, operation string, params map[string]string, options *Options) (*Response, error) {
	op, ok := operations[operation]
	if !ok {
		return nil, fmt.Errorf("unknown REST operation %q", operation)
	}
	path := op.Path
	for key, value := range params {
		placeholder := "{" + key + "}"
		if !strings.Contains(path, placeholder) || value == "" || value == "." || value == ".." || strings.ContainsAny(value, "/\\\x00\r\n") {
			return nil, fmt.Errorf("invalid path parameter %q", key)
		}
		path = strings.ReplaceAll(path, placeholder, url.PathEscape(value))
	}
	if strings.ContainsAny(path, "{}") {
		return nil, errors.New("missing REST path parameter")
	}
	o := Options{}
	if options != nil {
		o = *options
	}
	if o.JSON != nil && o.Body != nil {
		return nil, errors.New("JSON and raw Body are mutually exclusive")
	}
	body := o.Body
	contentType := o.ContentType
	if o.JSON != nil {
		b, err := json.Marshal(o.JSON)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
		if contentType == "" {
			contentType = "application/json"
		}
	}
	endpoint := Endpoint + path
	if q := o.Query.Encode(); q != "" {
		endpoint += "?" + q
	}
	req, err := http.NewRequestWithContext(ctx, op.Method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if o.Body != nil && o.ContentLength >= 0 {
		req.ContentLength = o.ContentLength
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range map[string]string{"Content-Type": contentType, "If-Match": o.IfMatch, "Idempotency-Key": o.IdempotencyKey, "X-Destination-Service-Token": o.DestinationToken, "Range": o.Range} {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		b, err := io.ReadAll(io.LimitReader(res.Body, maxJSONBytes+1))
		if err != nil {
			return nil, err
		}
		problem := &ProblemError{}
		_ = json.Unmarshal(b, problem)
		problem.Status = res.StatusCode
		problem.Header = res.Header.Clone()
		if problem.RequestID == "" {
			problem.RequestID = res.Header.Get("X-Request-Id")
		}
		return nil, problem
	}
	return &Response{res}, nil
}
func destinationAuthorization(token string) string {
	if token == "" {
		return ""
	}
	return "Bearer " + token
}
