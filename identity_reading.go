package bibleit

import (
	"context"
	"fmt"
	"strings"
)

type IdentityInfo struct {
	Actor, DisplayName, IdentityName, AuthProvider, KeyFingerprint string
	Authenticated                                                  bool
	Roles, Permissions                                             []string
	Raw                                                            Response
}
type VerseText struct {
	// Text retains the historical wire text, including any reference prefix.
	Text string
	// Structured source metadata is optional on older SSH servers.
	Content                          *string
	Translation, BookName, Reference string
	BookID, Chapter, Verse           *int64
	Fields                           map[string]string
}
type Reading struct {
	Translation string
	// Header coordinates describe the requested range; per-verse source metadata
	// is available in Verses when supplied by the server.
	Book, Chapter, Verse *int64
	Verses               []VerseText
	Raw                  Response
}

func boolean(fields map[string]string, key string) (bool, error) {
	switch fields[key] {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	return false, fmt.Errorf("invalid response field %q: expected true or false", key)
}
func commaList(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}
func optionalNumber(fields map[string]string, key string) (*int64, error) {
	if _, ok := fields[key]; !ok {
		return nil, nil
	}
	value, err := number(fields, key)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func optionalBoolean(fields map[string]string, key string) (*bool, error) {
	if _, ok := fields[key]; !ok {
		return nil, nil
	}
	value, err := boolean(fields, key)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
func optionalString(fields map[string]string, key string) *string {
	value, ok := fields[key]
	if !ok {
		return nil
	}
	return &value
}
func present(fields map[string]string, key string) (string, error) {
	value, ok := fields[key]
	if !ok {
		return "", fmt.Errorf("missing response field %q", key)
	}
	return value, nil
}
func (c *Client) GetIdentity(ctx context.Context) (IdentityInfo, error) {
	result, err := c.Identity(ctx)
	if err != nil {
		return IdentityInfo{}, err
	}
	fields, err := single(result)
	if err != nil {
		return IdentityInfo{}, err
	}
	value := IdentityInfo{Raw: result, DisplayName: fields["display_name"], IdentityName: fields["identity_name"], AuthProvider: fields["auth_provider"], KeyFingerprint: fields["key_fingerprint"], Roles: commaList(fields["roles"]), Permissions: commaList(fields["permissions"])}
	value.Actor, err = required(fields, "actor")
	if err != nil {
		return IdentityInfo{}, err
	}
	value.Authenticated, err = boolean(fields, "auth")
	if err != nil {
		return IdentityInfo{}, err
	}
	return value, nil
}
func (c *Client) ReadVerses(ctx context.Context, ref Reference) (Reading, error) {
	result, err := c.Read(ctx, ref)
	if err != nil {
		return Reading{}, err
	}
	return reading(result, false)
}
func (c *Client) SearchVerses(ctx context.Context, translation, query string) (Reading, error) {
	result, err := c.Search(ctx, translation, query)
	if err != nil {
		return Reading{}, err
	}
	return reading(result, true)
}
func reading(result Response, search bool) (Reading, error) {
	if len(result.Records) == 0 || result.Records[0].Type != "ok" {
		return Reading{}, fmt.Errorf("expected reading OK header")
	}
	fields := result.Records[0].Fields
	value := Reading{Raw: result, Verses: []VerseText{}}
	var err error
	value.Translation, err = required(fields, "translation")
	if err != nil {
		return Reading{}, err
	}
	for key, target := range map[string]**int64{"book": &value.Book, "chapter": &value.Chapter, "verse": &value.Verse} {
		if _, ok := fields[key]; ok {
			n, err := positive(fields, key)
			if err != nil {
				return Reading{}, err
			}
			*target = &n
		}
	}
	if !search && value.Book == nil {
		return Reading{}, fmt.Errorf("missing response field book")
	}
	if !search && value.Verse != nil {
		if value.Chapter == nil {
			return Reading{}, fmt.Errorf("verse reply requires chapter")
		}
		fields, err = single(result)
		if err != nil {
			return Reading{}, err
		}
		verse, err := verseText(fields)
		if err != nil {
			return Reading{}, err
		}
		value.Verses = append(value.Verses, verse)
		return value, nil
	}
	countKey := "verses"
	if search {
		countKey = "results"
	}
	_, err = counted(result, countKey, "verse")
	if err != nil {
		return Reading{}, err
	}
	for _, record := range result.Records[1:] {
		verse, err := verseText(record.Fields)
		if err != nil {
			return Reading{}, err
		}
		value.Verses = append(value.Verses, verse)
	}
	return value, nil
}

// verseText decodes only explicit metadata. Historical display text cannot be
// parsed reliably into verse coordinates across translations and locales.
func verseText(fields map[string]string) (VerseText, error) {
	text, err := present(fields, "text")
	if err != nil {
		return VerseText{}, err
	}
	value := VerseText{Text: text, Fields: fields, Content: optionalString(fields, "content"), Translation: fields["translation"], BookName: fields["book_name"], Reference: fields["reference"]}
	coordinates := 0
	for key, target := range map[string]**int64{"book": &value.BookID, "chapter": &value.Chapter, "verse": &value.Verse} {
		if _, ok := fields[key]; ok {
			n, err := positive(fields, key)
			if err != nil {
				return VerseText{}, err
			}
			*target = &n
			coordinates++
		}
	}
	if coordinates != 0 && coordinates != 3 {
		return VerseText{}, fmt.Errorf("incomplete verse coordinates")
	}
	return value, nil
}
