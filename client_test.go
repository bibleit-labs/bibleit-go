package bibleit

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestClientsExposeSeparateOperations(t *testing.T) {
	api, err := NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := any(api).(interface {
		Ping(context.Context) (CommandResponse, error)
	}); ok {
		t.Fatal("resource client exposes SSH commands")
	}
	ssh, _ := NewSSHClient(func(context.Context, Command) (CommandResponse, error) { return CommandResponse{}, nil })
	if _, ok := any(ssh).(interface{ RevokeCredential(context.Context) error }); ok {
		t.Fatal("SSH client exposes credential revocation")
	}
}
func TestSSHClientRequiresExplicitExecutor(t *testing.T) {
	if _, err := NewSSHClient(nil); err == nil {
		t.Fatal("accepted missing executor")
	}
	calls := 0
	c, _ := NewSSHClient(func(ctx context.Context, cmd Command) (CommandResponse, error) {
		calls++
		if cmd.String() != "PING" {
			t.Fatal(cmd)
		}
		return DecodeCommandResponse(cmd, strings.NewReader("OK pong=true\n"))
	})
	res, err := c.Ping(context.Background())
	if err != nil || res.Records[0].Fields["pong"] != "true" || calls != 1 {
		t.Fatal(res, err, calls)
	}
	if _, err = c.Execute(context.Background(), Command{}); err == nil {
		t.Fatal("accepted zero command")
	}
	if _, err = c.Execute(context.Background(), Command{wire: strings.Repeat("x", MaxCommandBytes+1)}); err == nil {
		t.Fatal("accepted oversized command")
	}
	sub, _ := LiveCommand("L1", LiveSubscribe)
	if _, err = c.Execute(context.Background(), sub); err == nil {
		t.Fatal("finite executor accepted subscription")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.Ping(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("invalid operation reached executor")
	}
}
func TestCredentialTransportAndValidation(t *testing.T) {
	for _, token := range []string{" ", "bad\ntoken", "bad\r", "bad\x00"} {
		if _, err := NewClient(token); err == nil {
			t.Fatal("invalid token accepted")
		}
	}
	for _, endpoint := range []string{"http://remote.example", "https://user:secret@example.com", "https://example.com?secret=x", "https://example.com#fragment", "invalid"} {
		if _, _, err := newHTTPTransport(endpoint, nil); err == nil {
			t.Fatal(endpoint)
		}
	}
	supplied := &http.Client{}
	c, err := NewClientWithHTTPClient("bt_test", supplied)
	if err != nil || c.http == supplied || supplied.CheckRedirect != nil || c.http.CheckRedirect(nil, nil) != http.ErrUseLastResponse {
		t.Fatal("credential transport must copy client and block redirects", err)
	}
}
