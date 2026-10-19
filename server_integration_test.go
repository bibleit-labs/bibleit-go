package bibleit

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bibleit-labs/bibleit-go/rest"
)

func integrationREST(t *testing.T) (context.Context, *rest.Client, rest.PersonalIdentity) {
	t.Helper()
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the isolated server harness")
	}
	if os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("REST integration requires the isolated TLS fixture")
	}
	addresses, err := net.LookupIP("api.bibleit.app")
	if err != nil || len(addresses) == 0 {
		t.Fatal("fixture resolution failed", err)
	}
	for _, address := range addresses {
		if !address.IsLoopback() {
			t.Fatal("fixture resolved outside loopback")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	c, err := rest.NewClient(os.Getenv("BIBLEIT_INTEGRATION_TOKEN"), nil)
	if err != nil {
		t.Fatal(err)
	}
	var identity rest.PersonalIdentity
	decodeREST(t, c.GetPersonalIdentity, ctx, &identity)
	if identity.ID == "" || identity.PersonalAccountID == "" {
		t.Fatal("missing personal identity")
	}
	return ctx, c, identity
}
func decodeREST(t *testing.T, call func(context.Context, *rest.Options) (*rest.Response, error), ctx context.Context, target any) {
	t.Helper()
	r, err := call(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	if err = r.Decode(target); err != nil {
		t.Fatal(err)
	}
}
func restStatus(t *testing.T, err error, status int) {
	t.Helper()
	var p *rest.ProblemError
	if !errors.As(err, &p) || p.Status != status {
		t.Fatalf("expected HTTP %d: %v", status, err)
	}
}

// Exercises current REST resources in the disposable real-server fixture.
func TestServerHTTPIntegration(t *testing.T) {
	ctx, c, identity := integrationREST(t)
	t.Run("reading search catalogue and rights", func(t *testing.T) {
		r, err := c.ReadPersonalVerse(ctx, identity.PersonalAccountID, "web", "19", "23", "1", nil)
		if err != nil {
			t.Fatal(err)
		}
		var verse rest.ScriptureRead
		err = r.Decode(&verse)
		r.Body.Close()
		if err != nil || len(verse.Verses) != 1 || verse.Verses[0].BookID != 19 || !strings.Contains(verse.Verses[0].Text, "shepherd") {
			t.Fatal(verse, err)
		}
		r, err = c.ReadPersonalChapter(ctx, identity.PersonalAccountID, "web", "19", "23", nil)
		if err != nil {
			t.Fatal(err)
		}
		var chapter rest.ScriptureRead
		err = r.Decode(&chapter)
		r.Body.Close()
		if err != nil || len(chapter.Verses) != 2 {
			t.Fatal(chapter, err)
		}
		r, err = c.SearchPersonalScripture(ctx, identity.PersonalAccountID, "web", &rest.Options{Query: url.Values{"q": {"shepherd"}, "limit": {"100"}}})
		if err != nil {
			t.Fatal(err)
		}
		var search rest.ScriptureSearch
		err = r.Decode(&search)
		r.Body.Close()
		if err != nil || len(search.Verses) == 0 {
			t.Fatal(search, err)
		}
		r, err = c.GetTranslation(ctx, "web", nil)
		if err != nil {
			t.Fatal(err)
		}
		var translation rest.CatalogueTranslation
		err = r.Decode(&translation)
		r.Body.Close()
		if err != nil || translation.Slug != "web" {
			t.Fatal(translation, err)
		}
		r, err = c.ListPersonalTranslations(ctx, identity.PersonalAccountID, nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
	})
	t.Run("personal quotas and safe credential metadata", func(t *testing.T) {
		r, err := c.GetPersonalQuotas(ctx, identity.PersonalAccountID, nil)
		if err != nil {
			t.Fatal(err)
		}
		var quotas rest.AccountQuotaCollection
		err = r.Decode(&quotas)
		r.Body.Close()
		if err != nil || len(quotas.Data) == 0 {
			t.Fatal("quota resources missing", err)
		}
		r, err = c.ListPersonalTokens(ctx, identity.PersonalAccountID, nil)
		if err != nil {
			t.Fatal(err)
		}
		var tokens rest.PersonalTokenCollection
		err = r.Decode(&tokens)
		r.Body.Close()
		if err != nil || len(tokens.Data) == 0 {
			t.Fatal("credential metadata missing", err)
		}
		for _, token := range tokens.Data {
			if token.ID == "" || token.IssuedAt <= 0 {
				t.Fatal("invalid credential metadata")
			}
		}
	})
	t.Run("book and native catalogue resources", func(t *testing.T) {
		r, err := c.ReadPersonalBook(ctx, identity.PersonalAccountID, "web", "19", nil)
		if err != nil {
			t.Fatal(err)
		}
		var book rest.ScriptureBookRead
		err = r.Decode(&book)
		r.Body.Close()
		if err != nil || len(book.Verses) != 100 || book.NextCursor == nil || book.Verses[0].BookID != 19 {
			t.Fatal("book contents differ", err)
		}
		r, err = c.GetTranslationCatalogueDetails(ctx, "web", nil)
		if err != nil {
			t.Fatal(err)
		}
		var catalogue rest.TranslationCatalogueDetails
		err = r.Decode(&catalogue)
		r.Body.Close()
		if err != nil || len(catalogue.Data) == 0 {
			t.Fatal("native catalogue missing", err)
		}
	})
	t.Run("scope and expiry enforcement", func(t *testing.T) {
		scoped, _ := rest.NewClient(os.Getenv("BIBLEIT_SCOPED_TOKEN"), nil)
		_, err := scoped.ListPersonalLives(ctx, identity.PersonalAccountID, nil)
		restStatus(t, err, 403)
		denied, _ := rest.NewClient(os.Getenv("BIBLEIT_DENIED_TOKEN"), nil)
		_, err = denied.ReadPersonalVerse(ctx, identity.PersonalAccountID, "web", "19", "23", "1", nil)
		restStatus(t, err, 403)
		expired, _ := rest.NewClient(os.Getenv("BIBLEIT_EXPIRED_TOKEN"), nil)
		_, err = expired.GetPersonalIdentity(ctx, nil)
		restStatus(t, err, 401)
	})
	t.Run("browser exchange replay and revocation", func(t *testing.T) {
		auth, _ := NewAuthClient()
		request := CodeExchange{Code: os.Getenv("BIBLEIT_AUTH_CODE"), Verifier: os.Getenv("BIBLEIT_AUTH_VERIFIER"), DeviceName: "REST integration"}
		issued, err := auth.ExchangeCode(ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		credential, err := NewClient(issued.AccessToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = credential.RevokeCredential(context.Background()) })
		personal, _ := rest.NewClient(issued.AccessToken, nil)
		r, err := personal.GetPersonalIdentity(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		_, err = auth.ExchangeCode(ctx, request)
		var authError *HTTPError
		if !errors.As(err, &authError) || authError.Code != "invalid_authorization_code" {
			t.Fatal(err)
		}
		invalid := CodeExchange{Code: os.Getenv("BIBLEIT_AUTH_BAD_CODE"), Verifier: "wrong-verifier"}
		_, err = auth.ExchangeCode(ctx, invalid)
		if !errors.As(err, &authError) || authError.Code != "invalid_code_verifier" {
			t.Fatal(err)
		}
		invalid.Verifier = request.Verifier
		_, err = auth.ExchangeCode(ctx, invalid)
		if !errors.As(err, &authError) || authError.Code != "invalid_authorization_code" {
			t.Fatal(err)
		}
		if err = credential.RevokeCredential(ctx); err != nil {
			t.Fatal(err)
		}
		_, err = personal.GetPersonalIdentity(ctx, nil)
		restStatus(t, err, 401)
	})
	t.Run("Live conditional lifecycle and cleanup", func(t *testing.T) {
		r, err := c.CreatePersonalLive(ctx, identity.PersonalAccountID, &rest.Options{JSON: rest.CreateLive{Name: "REST integration"}, IdempotencyKey: "integration-live"})
		if err != nil {
			t.Fatal(err)
		}
		var live rest.PersonalLiveSummary
		err = r.Decode(&live)
		r.Body.Close()
		if err != nil || live.ID == "" {
			t.Fatal(live, err)
		}
		t.Cleanup(func() {
			r, err := c.GetLive(context.Background(), live.ID, nil)
			if err != nil {
				return
			}
			etag := r.Header.Get("ETag")
			r.Body.Close()
			r, err = c.DeleteLive(context.Background(), live.ID, &rest.Options{IfMatch: etag})
			if err != nil {
				t.Error(err)
			} else {
				r.Body.Close()
			}
		})
		r, err = c.GetLive(ctx, live.ID, nil)
		if err != nil {
			t.Fatal(err)
		}
		etag := r.Header.Get("ETag")
		r.Body.Close()
		r, err = c.UpdateLiveMetadata(ctx, live.ID, &rest.Options{IfMatch: etag, JSON: rest.LiveMetadataPatch{Name: "Renamed"}})
		if err != nil {
			t.Fatal(err)
		}
		etag = r.Header.Get("ETag")
		r.Body.Close()
		r, err = c.UpdateLivePlayback(ctx, live.ID, &rest.Options{IfMatch: etag, JSON: map[string]any{"status": "stopped"}})
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
		_, err = c.UpdateLivePlayback(ctx, live.ID, &rest.Options{IfMatch: etag, JSON: map[string]any{"status": "running"}})
		restStatus(t, err, 412)
		r, err = c.ListPersonalLives(ctx, identity.PersonalAccountID, nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Body.Close()
	})
}
