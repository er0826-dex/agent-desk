package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SortKey reorders the top-level sessions inside each group, newest key
// first, keeps sub-sessions under their parent, and leaves persisted Order
// alone so switching back to manual restores the hand-built list.
func TestFlattenSortKey(t *testing.T) {
	a := &Instance{ID: "a", Title: "a", GroupPath: "Work", Order: 0}
	b := &Instance{ID: "b", Title: "b", GroupPath: "Work", Order: 1}
	c := &Instance{ID: "c", Title: "c", GroupPath: "Work", Order: 2}
	sub := &Instance{ID: "b-sub", Title: "b-sub", GroupPath: "Work", Order: 3, ParentSessionID: "b"}
	tree := NewGroupTree([]*Instance{a, b, c, sub})

	ids := func() []string {
		var out []string
		for _, it := range tree.Flatten() {
			if it.Type == ItemTypeSession {
				out = append(out, it.Session.ID)
			}
		}
		return out
	}
	if got := strings.Join(ids(), " "); got != "a b b-sub c" {
		t.Fatalf("manual order = %q", got)
	}

	keys := map[string]int64{"a": 10, "b": 30, "c": 20}
	tree.SortKey = func(i *Instance) int64 { return keys[i.ID] }
	if got := strings.Join(ids(), " "); got != "b b-sub c a" {
		t.Errorf("sorted order = %q, want b b-sub c a", got)
	}
	// Last-in-group flags follow the new order.
	items := tree.Flatten()
	last := items[len(items)-1]
	if last.Session.ID != "a" || !last.IsLastInGroup {
		t.Errorf("last item = %s lastInGroup=%v", last.Session.ID, last.IsLastInGroup)
	}
	if a.Order != 0 || b.Order != 1 || c.Order != 2 {
		t.Error("persisted Order must not change under a sort key")
	}

	// Ties keep manual order.
	tree.SortKey = func(*Instance) int64 { return 1 }
	if got := strings.Join(ids(), " "); got != "a b b-sub c" {
		t.Errorf("tied keys should keep manual order, got %q", got)
	}
	tree.SortKey = nil
	if got := strings.Join(ids(), " "); got != "a b b-sub c" {
		t.Errorf("nil key should restore manual order, got %q", got)
	}
}

func jsonlLine(typ string, at time.Time) string {
	return fmt.Sprintf(`{"type":%q,"timestamp":%q,"message":{"role":%q}}`, typ, at.Format(time.RFC3339Nano), typ)
}

// The 24h count walks newest-first, counts assistant entries only, and stops
// at the first entry older than the window instead of reading the whole file.
func TestCountRecentTurns(t *testing.T) {
	now := time.Date(2026, 10, 15, 5, 0, 0, 0, time.UTC)
	since := now.Add(-24 * time.Hour)
	lines := []string{
		jsonlLine("assistant", now.Add(-30*time.Hour)), // old, must not count
		jsonlLine("assistant", now.Add(-25*time.Hour)), // old
		jsonlLine("user", now.Add(-20*time.Hour)),
		jsonlLine("assistant", now.Add(-20*time.Hour)),
		"not json at all",
		jsonlLine("assistant", now.Add(-3*time.Hour)),
		jsonlLine("user", now.Add(-time.Hour)),
		jsonlLine("assistant", now.Add(-time.Hour)),
		`{"type":"summary"}`, // no timestamp: skipped, does not stop the scan
	}
	data := []byte(strings.Join(lines, "\n") + "\n")
	if got := countRecentTurnsIn(data, since); got != 3 {
		t.Errorf("in-memory count = %d, want 3", got)
	}

	path := filepath.Join(t.TempDir(), "s.jsonl")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := countRecentTurns(path, int64(len(data)), since); got != 3 {
		t.Errorf("file count = %d, want 3", got)
	}

	// The read is bounded: pad the front with old junk larger than the cap
	// and the recent tail still counts.
	pad := strings.Repeat(jsonlLine("assistant", now.Add(-40*time.Hour))+"\n", int(recentTurnsMaxBytes/80)+100)
	big := []byte(pad + string(data))
	if err := os.WriteFile(path, big, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := countRecentTurns(path, int64(len(big)), since); got != 3 {
		t.Errorf("bounded file count = %d, want 3", got)
	}

	// refreshRecentTurns caches on size and age.
	inst := &Instance{}
	inst.refreshRecentTurns(path, int64(len(big)), now)
	if inst.RecentTurns() != 3 {
		t.Fatalf("RecentTurns = %d", inst.RecentTurns())
	}
	inst.recentTurns = 99 // would be overwritten only on a recompute
	inst.refreshRecentTurns(path, int64(len(big)), now.Add(5*time.Minute))
	if inst.RecentTurns() != 99 {
		t.Error("same size within 10 min should not recompute")
	}
	inst.refreshRecentTurns(path, int64(len(big)), now.Add(11*time.Minute))
	if inst.RecentTurns() != 3 {
		t.Error("after 10 min it should recompute")
	}
}
