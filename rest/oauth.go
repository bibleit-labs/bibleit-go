package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const OAuthEndpoint = "https://bibleit.app"

// OAuthClient implements the public-client protocol for registered clients.
// It never sends a bearer token, opens a browser, or persists issued secrets.
// Callers validate callback state and own secure refresh-token storage.
type OAuthClient struct{ http *http.Client }

func NewOAuthClient(transport *http.Client) (*OAuthClient, error) {
	c, err := NewClient("", transport)
	if err != nil {
		return nil, err
	}
	return &OAuthClient{http: c.http}, nil
}

type OAuthProtocolError struct {
	Code        string      `json:"error"`
	Description string      `json:"error_description"`
	Status      int         `json:"-"`
	Header      http.Header `json:"-"`
}

func (e *OAuthProtocolError) Error() string { return "Bibleit OAuth: " + e.Code }

// AuthorizationURL builds the browser consent URL with S256 PKCE. The callback
// and scopes must match the server-side client registration. Keep the verifier
// private; only its SHA-256 base64url challenge belongs in this URL.
func AuthorizationURL(clientID, redirectURI, state, challenge string, scopes []string) (string, error) {
	if clientID == "" || redirectURI == "" || state == "" || len(challenge) != 43 || len(scopes) == 0 {
		return "", errors.New("client ID, callback, state, S256 challenge and explicit scopes are required")
	}
	for _, r := range challenge {
		if !strings.ContainsRune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_", r) {
			return "", errors.New("invalid S256 challenge")
		}
	}
	q := url.Values{"response_type": {"code"}, "client_id": {clientID}, "redirect_uri": {redirectURI}, "state": {state}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}, "scope": {strings.Join(scopes, " ")}}
	return OAuthEndpoint + "/oauth/authorize?" + q.Encode(), nil
}
func (c *OAuthClient) Metadata(ctx context.Context) (OAuthMetadata, error) {
	res, err := c.request(ctx, "GET", "/.well-known/oauth-authorization-server", nil)
	if err != nil {
		return OAuthMetadata{}, err
	}
	defer res.Body.Close()
	var v OAuthMetadata
	err = res.Decode(&v)
	return v, err
}

// ExchangeCode is a single attempt. Replaying a consumed code can revoke the
// entire grant; this client never retries, including after transport failures.
func (c *OAuthClient) ExchangeCode(ctx context.Context, clientID, code, redirectURI, verifier string) (OAuthToken, error) {
	if clientID == "" || code == "" || redirectURI == "" || len(verifier) < 43 || len(verifier) > 128 {
		return OAuthToken{}, errors.New("client ID, code, callback and PKCE verifier are required")
	}
	return c.token(ctx, url.Values{"grant_type": {"authorization_code"}, "client_id": {clientID}, "code": {code}, "redirect_uri": {redirectURI}, "code_verifier": {verifier}})
}

// Refresh rotates a single-use refresh token. Replace stored credentials
// atomically after success. Reusing a consumed refresh token revokes the grant.
func (c *OAuthClient) Refresh(ctx context.Context, clientID, refreshToken string, scopes []string) (OAuthToken, error) {
	if clientID == "" || refreshToken == "" {
		return OAuthToken{}, errors.New("client ID and refresh token are required")
	}
	f := url.Values{"grant_type": {"refresh_token"}, "client_id": {clientID}, "refresh_token": {refreshToken}}
	if len(scopes) > 0 {
		f.Set("scope", strings.Join(scopes, " "))
	}
	return c.token(ctx, f)
}
func (c *OAuthClient) Revoke(ctx context.Context, clientID, token string) error {
	if clientID == "" || token == "" {
		return errors.New("client ID and token are required")
	}
	res, err := c.request(ctx, "POST", "/oauth/revoke", url.Values{"client_id": {clientID}, "token": {token}})
	if err != nil {
		return err
	}
	return res.Body.Close()
}
func (c *OAuthClient) token(ctx context.Context, form url.Values) (OAuthToken, error) {
	res, err := c.request(ctx, "POST", "/oauth/token", form)
	if err != nil {
		return OAuthToken{}, err
	}
	defer res.Body.Close()
	var v OAuthToken
	err = res.Decode(&v)
	return v, err
}
func (c *OAuthClient) request(ctx context.Context, method, path string, form url.Values) (*Response, error) {
	encoded := form.Encode()
	if len(encoded) > 16384 {
		return nil, errors.New("OAuth form exceeds 16 KiB")
	}
	var body io.Reader
	if method == "POST" {
		body = strings.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, OAuthEndpoint+path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		b, err := io.ReadAll(io.LimitReader(res.Body, 16385))
		if err != nil {
			return nil, err
		}
		p := &OAuthProtocolError{}
		_ = json.Unmarshal(b, p)
		p.Status = res.StatusCode
		p.Header = res.Header.Clone()
		return nil, p
	}
	return &Response{res}, nil
}
