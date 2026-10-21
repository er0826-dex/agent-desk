package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/millwright-software/agent-desk/internal/session"
)

// The right-edge columns are the same width on every row, whatever the
// model, the auto bolt, the turn count or a snooze, so they line up.
func TestSessionColumnsFixedWidth(t *testing.T) {
	h := &Home{}
	want := colModelW + 1 + colTurnsW + 1 + colAgeW
	mk := func(tool, model, opts string) *session.Instance {
		i := &session.Instance{ID: model, Title: model, Tool: tool, ClaudeModel: model}
		if opts != "" {
			i.ToolOptionsJSON = []byte(opts)
		}
		return i
	}
	rows := []*session.Instance{
		mk("claude", "claude-opus-5-5", ""),
		mk("claude", "claude-fable-5-1", `{"tool":"claude","options":{"skip_permissions":true}}`),
		mk("shell", "", ""),
		mk("claude", "some-very-long-unknown-model-id", ""),
	}
	snoozed := mk("claude", "claude-opus-5-5", "")
	snoozed.Snooze(time.Now().Add(72 * time.Hour))
	rows = append(rows, snoozed)
	for _, r := range rows {
		for _, sel := range []bool{false, true} {
			got := h.sessionColumns(r, sel)
			if w := lipgloss.Width(got); w != want {
				t.Errorf("%q selected=%v: width %d, want %d: %q", r.ID, sel, w, want, got)
			}
		}
	}
}

// Rows render to exactly the list width, with the title truncated to make room.
func TestSessionRowFitsWidth(t *testing.T) {
	h := &Home{listRowWidth: 40}
	long := &session.Instance{ID: "x", Title: strings.Repeat("Long title ", 8), Tool: "claude", ClaudeModel: "claude-opus-5-5"}
	short := &session.Instance{ID: "y", Title: "Phone", Tool: "claude", ClaudeModel: "claude-opus-5-5"}
	for _, inst := range []*session.Instance{long, short} {
		var b strings.Builder
		h.renderSessionItem(&b, session.Item{Type: session.ItemTypeSession, Session: inst, Level: 1}, false)
		line := strings.TrimSuffix(b.String(), "\n")
		if w := lipgloss.Width(line); w != 39 {
			t.Errorf("%s: row width %d, want 39: %q", inst.ID, w, line)
		}
	}
}

func TestColumnAge(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local)
	cases := map[time.Duration]string{
		10 * time.Second:    "1m",
		40 * time.Minute:    "40m",
		5 * time.Hour:       "5h",
		3 * 24 * time.Hour:  "3d",
		20 * 24 * time.Hour: "2w",
		90 * 24 * time.Hour: "3mo",
	}
	for d, want := range cases {
		if got := columnAge(now.Add(-d), now); got != want {
			t.Errorf("%v: got %q want %q", d, got, want)
		}
		if lipgloss.Width(columnAge(now.Add(-d), now)) > colAgeW {
			t.Errorf("%v overflows the age column", d)
		}
	}
}
