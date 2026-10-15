package ui

import (
	"testing"
)

func TestSortModeCycle(t *testing.T) {
	m := sortManual
	seq := []string{}
	for i := 0; i < 4; i++ {
		seq = append(seq, m.String())
		m = m.next()
	}
	if got := seq[0] + " " + seq[1] + " " + seq[2] + " " + seq[3]; got != "manual recent busy manual" {
		t.Errorf("cycle = %q", got)
	}
	for _, s := range []string{"manual", "recent", "busy", ""} {
		if parseSortMode(s).String() != s && !(s == "" && parseSortMode(s) == sortManual) {
			t.Errorf("parse %q round-trips to %q", s, parseSortMode(s).String())
		}
	}
	h := &Home{}
	if h.sortKey() != nil {
		t.Error("manual should have no key")
	}
	h.sortMode = sortRecent
	if h.sortKey() == nil {
		t.Error("recent should have a key")
	}
	h.sortMode = sortBusy
	if h.sortKey() == nil {
		t.Error("busy should have a key")
	}
}
