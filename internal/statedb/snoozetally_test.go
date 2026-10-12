package statedb

import "testing"

func TestSnoozeTally(t *testing.T) {
	db := newTestDB(t)

	m, err := db.SnoozeTally()
	if err != nil || len(m) != 0 {
		t.Fatalf("empty tally: %v %v", m, err)
	}
	for _, l := range []string{"Next Monday", "typed", "Next Monday", "This weekend"} {
		if err := db.BumpSnoozeTally(l); err != nil {
			t.Fatal(err)
		}
	}
	m, _ = db.SnoozeTally()
	if m["Next Monday"] != 2 || m["typed"] != 1 || m["This weekend"] != 1 {
		t.Errorf("tally = %v", m)
	}
	rows := SortedTally(m)
	if rows[0].Label != "Next Monday" || rows[1].Label != "This weekend" || rows[2].Label != "typed" {
		t.Errorf("sorted = %v", rows)
	}

	// A damaged blob starts over instead of blocking picks.
	_ = db.SetMeta(snoozeTallyKey, "{not json")
	if err := db.BumpSnoozeTally("Tonight"); err != nil {
		t.Fatal(err)
	}
	m, _ = db.SnoozeTally()
	if len(m) != 1 || m["Tonight"] != 1 {
		t.Errorf("after damage: %v", m)
	}
}
