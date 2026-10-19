package rest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func bookPageFixture(verse int, cursor string) string {
	return fmt.Sprintf(`{"translation":"web","rights":{"edition":"fixture"},"verses":[{"book_id":19,"book":"Psalms","chapter":1,"verse":%d,"text":"fixture"}],"next_cursor":%s}`, verse, cursor)
}

type closeBookBody struct {
	io.Reader
	closed *int
}

func (b closeBookBody) Close() error { *b.closed++; return nil }

func TestBookPagesLazyTraversalAndOptions(t *testing.T) {
	for _, personal := range []bool{true, false} {
		t.Run(fmt.Sprint(personal), func(t *testing.T) {
			calls, closed := 0, 0
			opts := &Options{Query: url.Values{"limit": {"1"}}}
			client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
				calls++
				scope := "/organizations/org/"
				if personal {
					scope = "/accounts/account/"
				}
				if r.URL.Host != "api.bibleit.app" || r.URL.Path != "/v1"+scope+"translations/web/books/19" || r.URL.Query().Get("limit") != "1" || r.Header.Get("Authorization") != "Bearer bt_fixture" {
					t.Fatal(r.URL)
				}
				cursor := "\"opaque / +\""
				if calls == 2 {
					if r.URL.Query().Get("after") != "opaque / +" {
						t.Fatal(r.URL)
					}
					cursor = "null"
				}
				res := response(200, "")
				res.Body = closeBookBody{strings.NewReader(bookPageFixture(calls, cursor)), &closed}
				return res, nil
			})})
			pages := client.ReadBookPages(context.Background(), "org", "web", "19", opts)
			if personal {
				pages = client.ReadPersonalBookPages(context.Background(), "account", "web", "19", opts)
			}
			if calls != 0 {
				t.Fatal("iterator was eager")
			}
			opts.Query.Set("limit", "999") // Snapshot must already belong to the iterator.
			count := 0
			for page, err := range pages {
				if err != nil || len(page.Verses) != 1 || closed != calls {
					t.Fatal(page, err, closed, calls)
				}
				count++
			}
			if calls != 2 || count != 2 || opts.Query.Get("after") != "" || opts.Query.Get("limit") != "999" {
				t.Fatal(calls, count, opts.Query)
			}
		})
	}
}

func TestBookPagesStopCancelAndNoRetry(t *testing.T) {
	t.Run("early stop", func(t *testing.T) {
		calls := 0
		client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Query().Get("limit") != "50" {
				t.Fatal(r.URL)
			}
			return response(200, bookPageFixture(1, `"next"`)), nil
		})})
		for _, err := range client.ReadPersonalBookPages(context.Background(), "account", "web", "19", nil) {
			if err != nil {
				t.Fatal(err)
			}
			break
		}
		if calls != 1 {
			t.Fatal(calls)
		}
	})
	t.Run("cancel between pages", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
			calls++
			return response(200, bookPageFixture(1, `"next"`)), nil
		})})
		var final error
		for _, err := range client.ReadPersonalBookPages(ctx, "account", "web", "19", nil) {
			if err != nil {
				final = err
				continue
			}
			cancel()
		}
		if !errors.Is(final, context.Canceled) || calls != 1 {
			t.Fatal(final, calls)
		}
	})
	t.Run("stale cursor", func(t *testing.T) {
		calls := 0
		client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return response(200, bookPageFixture(1, `"next"`)), nil
			}
			return response(409, `{"status":409,"code":"stale_cursor","detail":"Restart this book read."}`), nil
		})})
		count := 0
		var final error
		for _, err := range client.ReadPersonalBookPages(context.Background(), "account", "web", "19", nil) {
			if err != nil {
				final = err
			} else {
				count++
			}
		}
		var problem *ProblemError
		if calls != 2 || count != 1 || !errors.As(final, &problem) || problem.Code != "stale_cursor" || problem.Detail != "Restart this book read." || strings.Contains(problem.Error(), "Restart") {
			t.Fatal(calls, count, final)
		}
	})
}

func TestBookPagesRejectUnsafeContinuation(t *testing.T) {
	first := bookPageFixture(1, `"next"`)
	for name, second := range map[string]string{
		"repeated cursor":        bookPageFixture(2, `"next"`),
		"duplicate verse":        bookPageFixture(1, `null`),
		"changed edition":        strings.Replace(bookPageFixture(2, `null`), `"fixture"}`, `"changed"}`, 1),
		"changed translation":    strings.Replace(bookPageFixture(2, `null`), `"web"`, `"bpml"`, 1),
		"missing cursor":         strings.Replace(bookPageFixture(2, `null`), `,"next_cursor":null`, "", 1),
		"empty cursor":           bookPageFixture(2, `""`),
		"invalid cursor type":    bookPageFixture(2, `7`),
		"missing rights":         `{"translation":"web","verses":[],"next_cursor":null}`,
		"null rights":            `{"translation":"web","rights":null,"verses":[],"next_cursor":null}`,
		"empty nonterminal page": `{"translation":"web","rights":{},"verses":[],"next_cursor":"new"}`,
		"null verses":            `{"translation":"web","rights":{},"verses":null,"next_cursor":null}`,
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(*http.Request) (*http.Response, error) {
				calls++
				body := first
				if calls > 1 {
					body = second
				}
				return response(200, body), nil
			})})
			var final error
			count := 0
			for _, err := range client.ReadPersonalBookPages(context.Background(), "account", "web", "19", nil) {
				if err != nil {
					final = err
				} else {
					count++
				}
			}
			if !errors.Is(final, ErrInvalidBookPage) || count != 1 || calls != 2 {
				t.Fatal(final, count, calls)
			}
		})
	}
}

func TestBookPagesValidateOptionsBeforeNetwork(t *testing.T) {
	client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid options contacted server")
		return nil, nil
	})})
	for _, query := range []url.Values{{"limit": {"0"}}, {"limit": {"501"}}, {"limit": {"bad"}}, {"limit": {"1", "2"}}, {"after": {""}}, {"after": {"one", "two"}}} {
		errorsSeen := 0
		for _, err := range client.ReadPersonalBookPages(context.Background(), "account", "web", "19", &Options{Query: query}) {
			if err == nil {
				t.Fatal(query)
			}
			errorsSeen++
		}
		if errorsSeen != 1 {
			t.Fatal(query, errorsSeen)
		}
	}
}

func TestSingleBookMethodsDoNotAutoPaginate(t *testing.T) {
	client, _ := NewClient("bt_fixture", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		if len(r.URL.Query()) != 0 {
			t.Fatal(r.URL)
		}
		return response(200, bookPageFixture(1, `"next"`)), nil
	})})
	personal, err := client.ReadPersonalBook(context.Background(), "account", "web", "19", nil)
	if err != nil {
		t.Fatal(err)
	}
	personal.Body.Close()
	organization, err := client.ReadBook(context.Background(), "org", "web", "19", nil)
	if err != nil {
		t.Fatal(err)
	}
	organization.Body.Close()
}
