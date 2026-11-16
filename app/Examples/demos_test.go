package examples

import (
	"testing"
	"time"
)

func TestJobAt(t *testing.T) {
	start := time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		after time.Duration
		pct   int
		done  bool
	}{
		{0, 0, false},
		{700 * time.Millisecond, 10, false},
		{3 * time.Second, 50, false},
		{5900 * time.Millisecond, 90, false},
		{jobLength, 100, true},
		{time.Hour, 100, true},
		{-time.Second, 0, false}, // a clock that went back is not a negative bar
	}
	for _, c := range cases {
		got := jobAt(start, start.Add(c.after))
		if got.Percent != c.pct || got.Done != c.done {
			t.Errorf("after %v: %d%% done=%v, want %d%% done=%v", c.after, got.Percent, got.Done, c.pct, c.done)
		}
	}
}

func TestStateIsSignedAndBounded(t *testing.T) {
	s := newStore([]byte("0123456789abcdef0123456789abcdef"), true)
	other := newStore([]byte("fedcba9876543210fedcba9876543210"), true)
	rec := recorder()
	if !s.write(rec, "x", map[string]string{"a": "b"}) {
		t.Fatal("a small state was refused")
	}
	req := requestWith(rec)
	var got map[string]string
	if !s.read(req, "x", &got) || got["a"] != "b" {
		t.Errorf("the state did not come back: %v", got)
	}
	if other.read(req, "x", &got) {
		t.Errorf("a cookie signed with another key was accepted")
	}
	if s.read(req, "y", &got) {
		t.Errorf("a cookie for one demo was accepted by another")
	}
	if s.write(recorder(), "x", map[string]string{"a": string(make([]byte, 4000))}) {
		t.Errorf("a state over the limit was kept")
	}
}
