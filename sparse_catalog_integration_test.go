package bibleit

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"
)

func TestServerSparseCatalogIntegration(t *testing.T) {
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the isolated --catalog-only server harness")
	}
	if os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("fixed endpoint requires isolated TLS fixture")
	}
	addresses, err := net.LookupIP("bibleit.app")
	if err != nil || len(addresses) == 0 {
		t.Fatal("fixture resolution failed")
	}
	for _, address := range addresses {
		if !address.IsLoopback() {
			t.Fatal("fixture resolves outside loopback")
		}
	}
	c, err := NewClient(os.Getenv("BIBLEIT_INTEGRATION_TOKEN"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	t.Run("present verse keeps source offset", func(t *testing.T) {
		value, err := c.ReadVerses(ctx, Reference{Translation: "BPML", Book: "42", Chapter: 17, Verse: 37})
		if err != nil || value.Verse == nil || *value.Verse != 37 || len(value.Verses) != 1 || value.Verses[0].Text != "Luke 17:37 after-gap" {
			t.Fatalf("present verse: %+v %v", value, err)
		}
		for _, named := range []struct {
			slug, book, text string
			id               int64
		}{
			{"BPML", "Gênesis", "Book 1 1:1 fixture", 1},
			{"LUT1912", "1. Könige", "Book 11 1:1 fixture", 11},
		} {
			value, err := c.ReadVerses(ctx, Reference{Translation: named.slug, Book: named.book, Chapter: 1, Verse: 1})
			if err != nil || value.Book == nil || *value.Book != named.id || len(value.Verses) != 1 || value.Verses[0].Text != named.text {
				t.Errorf("localized named read %s: %+v %v", named.book, value, err)
			}
		}
	})
	t.Run("omitted verse stays not found", func(t *testing.T) {
		_, err := c.ReadVerses(ctx, Reference{Translation: "BPML", Book: "42", Chapter: 17, Verse: 36})
		var serverErr *ServerError
		if !errors.As(err, &serverErr) || serverErr.Code != "verse_not_found" {
			t.Fatalf("omitted verse: %v", err)
		}
	})
	t.Run("chapter preserves sparse numbering", func(t *testing.T) {
		value, err := c.ReadVerses(ctx, Reference{Translation: "BPML", Book: "42", Chapter: 17})
		if err != nil || len(value.Verses) != 2 || value.Verses[0].Text != "Luke 17:35 before-gap" || value.Verses[1].Text != "Luke 17:37 after-gap" {
			t.Fatalf("sparse chapter: %+v %v", value, err)
		}
	})
	t.Run("catalogue preserves book and max verse IDs", func(t *testing.T) {
		for _, language := range []struct{ slug, first, eleventh, luke string }{
			{"BPML", "Gênesis", "1 Reis", "Lucas"},
			{"LUT1912", "1. Mose", "1. Könige", "Lukas"},
		} {
			value, err := c.GetTranslationCatalog(ctx, language.slug)
			if err != nil || len(value.Books) != 66 {
				t.Errorf("%s catalogue: books=%d %v", language.slug, len(value.Books), err)
				continue
			}
			book := value.Books[41]
			if value.Books[0].Name != language.first || value.Books[10].Name != language.eleventh || book.ID != 42 || book.Name != language.luke || len(book.Chapters) != 2 || book.Chapters[1].Number != 17 || book.Chapters[1].Verses != 37 {
				t.Errorf("%s localized sparse catalogue mismatch: %+v", language.slug, book)
			}
		}
	})
	t.Run("missing translation preserves error", func(t *testing.T) {
		_, err := c.GetTranslationCatalog(ctx, "not-a-translation")
		var serverErr *ServerError
		if !errors.As(err, &serverErr) || serverErr.Code != "translation_not_found" {
			t.Fatalf("missing translation: %v", err)
		}
	})
}
