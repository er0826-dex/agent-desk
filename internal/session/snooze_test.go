package session

import (
	"testing"
	"time"
)

// Saturday 2026-10-03 16:30 local.
var snoozeNow = time.Date(2026, 10, 3, 16, 30, 0, 0, time.Local)

func TestParseSnooze(t *testing.T) {
	cases := []struct {
		spec string
		want time.Time
	}{
		{"30m", snoozeNow.Add(30 * time.Minute)},
		{"2h", snoozeNow.Add(2 * time.Hour)},
		{"1d", snoozeNow.AddDate(0, 0, 1)},
		{"1w", snoozeNow.AddDate(0, 0, 7)},
		{"7pm", time.Date(2026, 10, 3, 19, 0, 0, 0, time.Local)},
		{"7am", time.Date(2026, 10, 4, 7, 0, 0, 0, time.Local)}, // already past today
		{"7:30pm", time.Date(2026, 10, 3, 19, 30, 0, 0, time.Local)},
		{"19:00", time.Date(2026, 10, 3, 19, 0, 0, 0, time.Local)},
		{"12am", time.Date(2026, 10, 4, 0, 0, 0, 0, time.Local)},
		{"12pm", time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)},
		{"tomorrow", time.Date(2026, 10, 4, 9, 0, 0, 0, time.Local)},
		{"tomorrow 9am", time.Date(2026, 10, 4, 9, 0, 0, 0, time.Local)},
		{"mon", time.Date(2026, 10, 5, 9, 0, 0, 0, time.Local)},
		{"Monday 11:30", time.Date(2026, 10, 5, 11, 30, 0, 0, time.Local)},
		{"sat", time.Date(2026, 10, 10, 9, 0, 0, 0, time.Local)}, // today is Sat: next Sat
		{"10/12", time.Date(2026, 10, 12, 9, 0, 0, 0, time.Local)},
		{"10/12 11am", time.Date(2026, 10, 12, 11, 0, 0, 0, time.Local)},
		{"1/4", time.Date(2027, 1, 4, 9, 0, 0, 0, time.Local)},        // past this year → next
		{"10/3 5pm", time.Date(2026, 10, 3, 17, 0, 0, 0, time.Local)}, // later today
		{"10/3/27", time.Date(2027, 10, 3, 9, 0, 0, 0, time.Local)},
		{"  Tomorrow   7AM ", time.Date(2026, 10, 4, 7, 0, 0, 0, time.Local)},
	}
	for _, c := range cases {
		got, err := ParseSnooze(c.spec, snoozeNow)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.spec, err)
			continue
		}
		if !got.Equal(c.want) {
			t.Errorf("%q: got %v, want %v", c.spec, got, c.want)
		}
	}
}

func TestParseSnoozeRejects(t *testing.T) {
	for _, spec := range []string{"", "0m", "yesterday", "25:00", "13pm", "10/3 9am", "13/1", "soon", "7 am tomorrow"} {
		if got, err := ParseSnooze(spec, snoozeNow); err == nil {
			t.Errorf("%q: expected error, got %v", spec, got)
		}
	}
}

func TestSnoozeSuggestions(t *testing.T) {
	day := func(y, mo, d, h, mi int) time.Time { return time.Date(y, time.Month(mo), d, h, mi, 0, 0, time.Local) }
	type row struct {
		label string
		at    time.Time
	}
	cases := []struct {
		name string
		now  time.Time
		want []row
	}{
		{"Monday 1am", day(2026, 10, 12, 1, 10), []row{
			{"Tonight", day(2026, 10, 12, 20, 0)},
			{"Later this week", day(2026, 10, 15, 9, 0)},
			{"This weekend", day(2026, 10, 17, 9, 0)},
			{"Next Monday", day(2026, 10, 19, 9, 0)},
			{"In 2 weeks", day(2026, 10, 26, 9, 0)},
			{"In a month", day(2026, 11, 12, 9, 0)},
		}},
		{"Wednesday 2pm", day(2026, 10, 14, 14, 0), []row{
			{"Tonight", day(2026, 10, 14, 20, 0)},
			{"Later this week", day(2026, 10, 16, 9, 0)},
			{"This weekend", day(2026, 10, 17, 9, 0)},
			{"Next Monday", day(2026, 10, 19, 9, 0)},
			{"In 2 weeks", day(2026, 10, 28, 9, 0)},
			{"In a month", day(2026, 11, 14, 9, 0)},
		}},
		{"Friday 1pm: no later-this-week", day(2026, 10, 16, 13, 0), []row{
			{"Tonight", day(2026, 10, 16, 20, 0)},
			{"This weekend", day(2026, 10, 17, 9, 0)},
			{"Next Monday", day(2026, 10, 19, 9, 0)},
			{"In 2 weeks", day(2026, 10, 30, 9, 0)},
			{"In a month", day(2026, 11, 16, 9, 0)},
		}},
		{"Saturday 4:30pm: weekend means Sunday", snoozeNow, []row{
			{"Tonight", day(2026, 10, 3, 20, 0)},
			{"This weekend", day(2026, 10, 4, 9, 0)},
			{"Next Monday", day(2026, 10, 5, 9, 0)},
			{"In 2 weeks", day(2026, 10, 17, 9, 0)},
			{"In a month", day(2026, 11, 3, 9, 0)},
		}},
		{"Saturday 7am: weekend is still today", day(2026, 10, 3, 7, 0), []row{
			{"Tonight", day(2026, 10, 3, 20, 0)},
			{"This weekend", day(2026, 10, 3, 9, 0)},
			{"Next Monday", day(2026, 10, 5, 9, 0)},
			{"In 2 weeks", day(2026, 10, 17, 9, 0)},
			{"In a month", day(2026, 11, 3, 9, 0)},
		}},
		{"Sunday 8:30pm: tonight gone, no weekend", day(2026, 10, 11, 20, 30), []row{
			{"Next Monday", day(2026, 10, 12, 9, 0)},
			{"In 2 weeks", day(2026, 10, 25, 9, 0)},
			{"In a month", day(2026, 11, 11, 9, 0)},
		}},
		{"Monday 7:45pm: tonight under the lead", day(2026, 10, 12, 19, 45), []row{
			{"Later this week", day(2026, 10, 15, 9, 0)},
			{"This weekend", day(2026, 10, 17, 9, 0)},
			{"Next Monday", day(2026, 10, 19, 9, 0)},
			{"In 2 weeks", day(2026, 10, 26, 9, 0)},
			{"In a month", day(2026, 11, 12, 9, 0)},
		}},
	}
	for _, c := range cases {
		got := SnoozeSuggestions(c.now)
		if len(got) != len(c.want) {
			t.Errorf("%s: want %d rows, got %v", c.name, len(c.want), labels(got))
			continue
		}
		for i, w := range c.want {
			if got[i].Label != w.label || !got[i].At.Equal(w.at) {
				t.Errorf("%s row %d = %q %v, want %q %v", c.name, i, got[i].Label, got[i].At, w.label, w.at)
			}
		}
	}
}

func labels(ps []SnoozePreset) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Label
	}
	return out
}

func TestFormatSnoozeUntil(t *testing.T) {
	cases := []struct {
		t    time.Time
		want string
	}{
		{time.Date(2026, 10, 3, 19, 0, 0, 0, time.Local), "7p"},
		{time.Date(2026, 10, 3, 19, 15, 0, 0, time.Local), "7:15p"},
		{time.Date(2026, 10, 4, 7, 0, 0, 0, time.Local), "Sun 7a"},
		{time.Date(2026, 10, 5, 11, 30, 0, 0, time.Local), "Mon 11:30a"},
		{time.Date(2026, 10, 10, 7, 0, 0, 0, time.Local), "10/10 7a"},
		{time.Date(2027, 1, 4, 7, 0, 0, 0, time.Local), "1/4 7a"},
	}
	for _, c := range cases {
		if got := FormatSnoozeUntil(c.t, snoozeNow); got != c.want {
			t.Errorf("%v: got %q, want %q", c.t, got, c.want)
		}
	}
}

func TestSnoozeLifecycle(t *testing.T) {
	inst := &Instance{}
	until := snoozeNow.Add(time.Hour)
	inst.Snooze(until)
	if inst.Flag != FlagSnoozed || !inst.SnoozeUntil.Equal(until) {
		t.Fatalf("after Snooze: flag=%v until=%v", inst.Flag, inst.SnoozeUntil)
	}
	if inst.SnoozeDue(snoozeNow.Add(59 * time.Minute)) {
		t.Error("due a minute early")
	}
	if !inst.SnoozeDue(until) {
		t.Error("not due at the wake time")
	}
	inst.WakeFromSnooze()
	if inst.Flag != FlagUnread || !inst.SnoozeUntil.IsZero() {
		t.Errorf("after wake: flag=%v until=%v", inst.Flag, inst.SnoozeUntil)
	}
	if inst.SnoozeDue(until.Add(time.Hour)) {
		t.Error("woken session reported due again")
	}

	// u on a snoozed session leaves the cycle at none.
	inst.Snooze(until)
	inst.Flag = inst.Flag.Next()
	inst.SnoozeUntil = time.Time{}
	if inst.Flag != FlagNone {
		t.Errorf("Next from snoozed = %v, want none", inst.Flag)
	}

	// ClearSnooze leaves a non-snooze marker alone.
	inst.Flag = FlagParkedBlue
	inst.SnoozeUntil = until
	inst.ClearSnooze()
	if inst.Flag != FlagParkedBlue || !inst.SnoozeUntil.IsZero() {
		t.Errorf("ClearSnooze on blue: flag=%v until=%v", inst.Flag, inst.SnoozeUntil)
	}
	if !FlagSnoozed.IsParked() || !FlagSnoozed.IsSkipped() || FlagParkedBlue.IsSkipped() {
		t.Error("flag predicates wrong for snoozed/blue")
	}
}
