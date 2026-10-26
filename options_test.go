package bibleit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestTypedQueriesAndDefaultOptions(t *testing.T) {
	after := "opaque / + &"
	calls := 0
	client, _ := NewClient("", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 && (r.URL.Query().Get("limit") != "7" || r.URL.Query().Get("after") != after) {
			t.Fatal(r.URL)
		}
		if calls == 2 && r.URL.RawQuery != "" {
			t.Fatal(r.URL)
		}
		return response(200, `{}`), nil
	})})
	params := ListBooksParams{}.WithLimit(7).WithAfter(after)
	res, err := client.ListBooks(context.Background(), &Options{Params: params})
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	res, err = client.ListBooks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if _, err := client.ListBooks(context.Background(), nil, nil); err == nil {
		t.Fatal("accepted multiple options")
	}
	if _, err := client.ListBooks(context.Background(), &Options{Params: params, Query: url.Values{}}); err == nil {
		t.Fatal("accepted conflicting typed and raw queries")
	}
	if calls != 2 {
		t.Fatal("invalid options caused a request", calls)
	}
}

func TestTypedBookParamsSnapshotBeforeIteration(t *testing.T) {
	limit := int64(1)
	params := &ReadPersonalBookParams{Limit: &limit}
	client, _ := NewClient("", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("limit") != "1" {
			t.Fatal(r.URL)
		}
		return response(200, bookPageFixture(1, "null")), nil
	})})
	pages := client.ReadPersonalBookPages(context.Background(), "account", "web", "19", &Options{Params: params})
	limit = 999
	count := 0
	for _, err := range pages {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 1 {
		t.Fatal(count)
	}
}

func TestEventPollDefaultsAndExplicitZeroWait(t *testing.T) {
	calls := 0
	client, _ := NewClient("", &http.Client{Transport: trip(func(r *http.Request) (*http.Response, error) {
		calls++
		want := "25"
		if calls == 2 {
			want = "0"
		}
		if r.URL.Query().Get("wait") != want || r.URL.Query().Has("after") {
			t.Fatal(r.URL)
		}
		return response(200, `{"schema_version":1,"type":"stream.heartbeat","resource_type":"live","resource_id":"L1","sequence":1,"cursor":"opaque","data":null}`), nil
	})})
	if _, err := client.PollLiveEventsWithOptions(context.Background(), "L1"); err != nil {
		t.Fatal(err)
	}
	params := GetLiveEventsParams{}.WithWait(0)
	if _, err := client.PollLiveEventsWithOptions(context.Background(), "L1", &params); err != nil {
		t.Fatal(err)
	}
	if params.After != nil || *params.Wait != 0 {
		t.Fatal("poll changed caller parameters")
	}
	invalid := int64(26)
	if _, err := client.PollLiveEventsWithOptions(context.Background(), "L1", &GetLiveEventsParams{Wait: &invalid}); err == nil {
		t.Fatal("accepted invalid wait")
	}
	if calls != 2 {
		t.Fatal(calls)
	}
}

func TestQueryWithMethodsPreserveOriginalAndExplicitZero(t *testing.T) {
	original := ReadPersonalBookParams{}.WithLimit(100).WithAfter("original")
	updated := original.WithLimit(0).WithAfter("")
	if *original.Limit != 100 || *original.After != "original" {
		t.Fatal("With methods changed the original parameters", original)
	}
	values := updated.QueryValues()
	if values.Get("limit") != "0" || !values.Has("after") || values.Get("after") != "" {
		t.Fatal("explicit zero values were omitted", values)
	}
	if len((ReadPersonalBookParams{}).QueryValues()) != 0 {
		t.Fatal("unset query fields were encoded")
	}
	search := SearchScriptureParams{}.WithQ("love & peace").WithLimit(10)
	if search.QueryValues().Get("q") != "love & peace" || *search.Limit != 10 {
		t.Fatal(search)
	}
	asset := UploadPresentationAssetParams{}.WithKind(UploadPresentationAssetParamsKindImage)
	if asset.QueryValues().Get("kind") != "image" {
		t.Fatal(asset)
	}
}

func TestConcreteSchemaFieldsRejectWrongWireTypes(t *testing.T) {
	var playback Playback
	if err := json.Unmarshal([]byte(`{"status":42}`), &playback); err == nil {
		t.Fatal("status accepted a number")
	}
	var display LiveDisplay
	if err := json.Unmarshal([]byte(`{"snapshot":{"id":"snapshot"},"status":"running","type":"presentation"}`), &display); err != nil || display.Snapshot == nil {
		t.Fatal(display, err)
	}
	if err := json.Unmarshal([]byte(`{"snapshot":null}`), &display); err != nil || display.Snapshot != nil {
		t.Fatal(display, err)
	}
	data, err := json.Marshal(EmptyObject{})
	if err != nil || string(data) != "{}" {
		t.Fatal(string(data), err)
	}
}
