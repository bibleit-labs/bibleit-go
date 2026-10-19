package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"strconv"
)

// ErrInvalidBookPage means a paged response could not be traversed safely.
// Callers should discard partial results rather than treat them as a full book.
var ErrInvalidBookPage = errors.New("invalid REST book page")

// ReadPersonalBookPages lazily reads a personal book in bounded pages. It opts
// into cursor pagination with limit=100 unless Options.Query specifies a limit
// (1..500). Query.after can resume a previous traversal. Options are copied;
// caller query values are never changed. Each iteration starts a new traversal.
// Stop iteration or cancel ctx to stop fetching. Errors are yielded once, with
// a zero page, and are never retried. Every response body is closed before yield.
// ReadPersonalBook remains a single request and returns one page (100 verses
// by default). Use this iterator to read the complete book.
func (c *Client) ReadPersonalBookPages(ctx context.Context, accountID, slug, bookID string, options *Options) iter.Seq2[ScriptureBookRead, error] {
	return c.bookPages(ctx, "readPersonalBook", map[string]string{"account_id": accountID, "slug": slug, "book_id": bookID}, options)
}

// ReadBookPages is the organization-scoped counterpart of ReadPersonalBookPages.
// It has the same lazy iteration, cancellation and error behavior. ReadBook
// remains a single-request method returning one bounded page.
func (c *Client) ReadBookPages(ctx context.Context, organizationID, slug, bookID string, options *Options) iter.Seq2[ScriptureBookRead, error] {
	return c.bookPages(ctx, "readBook", map[string]string{"organization_id": organizationID, "slug": slug, "book_id": bookID}, options)
}

func (c *Client) bookPages(ctx context.Context, operation string, params map[string]string, options *Options) iter.Seq2[ScriptureBookRead, error] {
	snapshot := Options{}
	if options != nil {
		snapshot = *options
	}
	snapshot.Query = cloneBookQuery(snapshot.Query)
	if !snapshot.Query.Has("limit") {
		snapshot.Query.Set("limit", "100")
	}
	return func(yield func(ScriptureBookRead, error) bool) {
		fail := func(err error) { yield(ScriptureBookRead{}, err) }
		opts := snapshot
		opts.Query = cloneBookQuery(snapshot.Query)
		limit, err := strconv.Atoi(opts.Query.Get("limit"))
		if err != nil || limit < 1 || limit > 500 || len(opts.Query["limit"]) != 1 {
			fail(fmt.Errorf("book page limit must be an integer between 1 and 500"))
			return
		}
		if opts.Query.Has("after") && (opts.Query.Get("after") == "" || len(opts.Query["after"]) != 1) {
			fail(fmt.Errorf("book page after must be a single nonempty cursor"))
			return
		}
		seen := map[string]bool{}
		if after := opts.Query.Get("after"); after != "" {
			seen[after] = true
		}
		firstTranslation, firstRights := "", ""
		var previous Verse
		for {
			if err := ctx.Err(); err != nil {
				fail(err)
				return
			}
			response, err := c.Call(ctx, operation, params, &opts)
			if err != nil {
				fail(err)
				return
			}
			var raw json.RawMessage
			err = response.Decode(&raw)
			closeErr := response.Body.Close()
			if err == nil {
				err = closeErr
			}
			if err != nil {
				fail(err)
				return
			}
			var page ScriptureBookRead
			var fields map[string]json.RawMessage
			if json.Unmarshal(raw, &page) != nil || json.Unmarshal(raw, &fields) != nil || page.Verses == nil || page.Translation == "" || len(fields["rights"]) == 0 || string(fields["rights"]) == "null" {
				fail(fmt.Errorf("%w: missing or malformed Scripture fields", ErrInvalidBookPage))
				return
			}
			cursor, exists := fields["next_cursor"]
			if !exists || len(cursor) == 0 {
				fail(fmt.Errorf("%w: missing next_cursor", ErrInvalidBookPage))
				return
			}
			var next *string
			if json.Unmarshal(cursor, &next) != nil || next != nil && *next == "" {
				fail(fmt.Errorf("%w: next_cursor must be a nonempty string or null", ErrInvalidBookPage))
				return
			}
			if next != nil && len(page.Verses) == 0 {
				fail(fmt.Errorf("%w: nonterminal page made no progress", ErrInvalidBookPage))
				return
			}
			if len(page.Verses) > limit {
				fail(fmt.Errorf("%w: page exceeds requested limit", ErrInvalidBookPage))
				return
			}
			rights, err := json.Marshal(page.Rights)
			if err != nil {
				fail(err)
				return
			}
			if firstTranslation == "" {
				firstTranslation, firstRights = page.Translation, string(rights)
			} else if page.Translation != firstTranslation || string(rights) != firstRights {
				fail(fmt.Errorf("%w: translation or rights changed during traversal", ErrInvalidBookPage))
				return
			}
			for _, verse := range page.Verses {
				if verse.BookID < 1 || verse.BookID > 255 || verse.Chapter < 1 || verse.Chapter > 152 || verse.Verse < 1 || verse.Verse > 255 || previous.BookID != 0 && (verse.BookID != previous.BookID || verse.Chapter < previous.Chapter || verse.Chapter == previous.Chapter && verse.Verse <= previous.Verse) {
					fail(fmt.Errorf("%w: duplicate or unordered verse coordinates", ErrInvalidBookPage))
					return
				}
				previous = verse
			}
			if next != nil && seen[*next] {
				fail(fmt.Errorf("%w: repeated continuation cursor", ErrInvalidBookPage))
				return
			}
			if !yield(page, nil) || next == nil {
				return
			}
			seen[*next] = true
			opts.Query.Set("after", *next)
		}
	}
}

func cloneBookQuery(source url.Values) url.Values {
	clone := url.Values{}
	for key, values := range source {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}
