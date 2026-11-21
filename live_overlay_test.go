package bibleit

import (
	"context"
	"strings"
	"testing"
)

func TestLiveOverlayCommands(t *testing.T) {
	cases := []struct {
		command Command
		err     error
		want    string
	}{}
	c, e := GetLiveOverlayCommand("L1")
	cases = append(cases, struct {
		command Command
		err     error
		want    string
	}{c, e, "LIVE L1 OVERLAY GET"})
	c, e = SetLiveOverlayCommand("L1", "builtin-youtube-stream-link", 7)
	cases = append(cases, struct {
		command Command
		err     error
		want    string
	}{c, e, "LIVE L1 OVERLAY SET builtin-youtube-stream-link 7"})
	c, e = ClearLiveOverlayCommand("L1", 7)
	cases = append(cases, struct {
		command Command
		err     error
		want    string
	}{c, e, "LIVE L1 OVERLAY CLEAR 7"})
	for _, v := range cases {
		if v.err != nil || v.command.String() != v.want || v.command.Streaming() {
			t.Fatalf("%q %v", v.command.String(), v.err)
		}
	}
	if _, e := SetLiveOverlayCommand("L1", "bad\ncommand", 0); e == nil {
		t.Fatal("accepted command injection")
	}
	if _, e := SetLiveOverlayCommand("L1", strings.Repeat("x", 65), 0); e == nil {
		t.Fatal("accepted oversized design")
	}
	if _, e := ClearLiveOverlayCommand("L1", -1); e == nil {
		t.Fatal("accepted negative sequence")
	}
}
func TestTypedLiveOverlaySelection(t *testing.T) {
	c := discoveryClient(t, map[string]string{"LIVE L1 OVERLAY GET": "OK live_id=L1 sequence=7 active=true design_id=builtin-youtube-stream-link design_revision=1 name=\"YouTube link\" future=kept\n", "LIVE L1 OVERLAY CLEAR 7": "OK live_id=L1 sequence=8 active=false\n"})
	got, e := c.GetLiveOverlay(context.Background(), "L1")
	if e != nil || got.Sequence != 7 || got.Overlay == nil || got.Overlay.YouTubeURL != nil || got.Fields["future"] != "kept" {
		t.Fatalf("%+v %v", got, e)
	}
	got, e = c.ClearLiveOverlay(context.Background(), "L1", 7)
	if e != nil || got.Sequence != 8 || got.Overlay != nil {
		t.Fatalf("%+v %v", got, e)
	}
}
func TestLiveOverlayRejectsMalformedResponse(t *testing.T) {
	for _, wire := range []string{"OK live_id=L1 sequence=7 active=yes\n", "OK live_id=L1 sequence=7 active=true\n", "OK live_id=L1 sequence=-1 active=false\n"} {
		c := discoveryClient(t, map[string]string{"LIVE L1 OVERLAY GET": wire})
		if _, e := c.GetLiveOverlay(context.Background(), "L1"); e == nil {
			t.Fatal("accepted malformed response")
		}
	}
}
