package bibleit

import (
	"github.com/bibleit-labs/bibleit-go/rest"
	"testing"
)

func TestServerSparseCatalogIntegration(t *testing.T) {
	ctx, c, identity := integrationREST(t)
	t.Run("present and missing source verse IDs", func(t *testing.T) {
		r, err := c.ReadPersonalVerse(ctx, identity.PersonalAccountID, "bpml", "42", "17", "37", nil)
		if err != nil {
			t.Fatal(err)
		}
		var reading rest.ScriptureRead
		err = r.Decode(&reading)
		r.Body.Close()
		if err != nil || len(reading.Verses) != 1 || reading.Verses[0].Verse != 37 || reading.Verses[0].BookID != 42 || reading.Verses[0].Text != "after-gap" {
			t.Fatal(reading, err)
		}
		_, err = c.ReadPersonalVerse(ctx, identity.PersonalAccountID, "bpml", "42", "17", "36", nil)
		restStatus(t, err, 404)
	})
	t.Run("chapter retains sparse numbering", func(t *testing.T) {
		r, err := c.ReadPersonalChapter(ctx, identity.PersonalAccountID, "bpml", "42", "17", nil)
		if err != nil {
			t.Fatal(err)
		}
		var reading rest.ScriptureRead
		err = r.Decode(&reading)
		r.Body.Close()
		if err != nil || len(reading.Verses) != 2 || reading.Verses[0].Verse != 35 || reading.Verses[1].Verse != 37 {
			t.Fatal(reading, err)
		}
	})
	t.Run("reviewed book catalogues use canonical IDs", func(t *testing.T) {
		for _, slug := range []string{"bpml", "lut1912"} {
			r, err := c.ListTranslationBooks(ctx, slug, nil)
			if err != nil {
				t.Fatal(err)
			}
			var books rest.BookCollection
			err = r.Decode(&books)
			r.Body.Close()
			if err != nil || len(books.Data) == 0 || books.Data[0].ID != 1 {
				t.Fatal(books, err)
			}
		}
		_, err := c.GetTranslation(ctx, "not-a-translation", nil)
		restStatus(t, err, 404)
	})
}
