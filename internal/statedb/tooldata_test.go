package statedb

import (
	"encoding/json"
	"testing"
	"time"
)

func TestToolDataRoundtripFlag(t *testing.T) {
	detectedAt := time.Unix(1750000000, 0)
	blob := MarshalToolData(
		"claude-sid", detectedAt, "claude-fable-5",
		"", time.Time{},
		nil, "",
		"", time.Time{},
		"", time.Time{},
		[]string{"mcp-a"},
		nil, "ocean", 3, 1750009999, // FlagParkedBlue, lastLogActivity
		1750020000, "Work", // snoozeUntil, snoozeFromGroup
		"copilot-sid",
		true,
	)

	claudeSID, claudeAt, claudeModel,
		_, _,
		_, _,
		_, _,
		_, _,
		mcps,
		_, colorScheme, flag, lastLog, snooze, snoozeFrom, copilotSID, copilotPending := UnmarshalToolData(blob)

	if copilotSID != "copilot-sid" || !copilotPending {
		t.Fatalf("copilotSID = %q, pending = %v", copilotSID, copilotPending)
	}
	if claudeSID != "claude-sid" {
		t.Errorf("claudeSID = %q, want claude-sid", claudeSID)
	}
	if !claudeAt.Equal(detectedAt) {
		t.Errorf("claudeAt = %v, want %v", claudeAt, detectedAt)
	}
	if claudeModel != "claude-fable-5" {
		t.Errorf("claudeModel = %q, want claude-fable-5", claudeModel)
	}
	if len(mcps) != 1 || mcps[0] != "mcp-a" {
		t.Errorf("mcps = %v, want [mcp-a]", mcps)
	}
	if colorScheme != "ocean" {
		t.Errorf("colorScheme = %q, want ocean", colorScheme)
	}
	if flag != 3 {
		t.Errorf("flag = %d, want 3", flag)
	}
	if lastLog != 1750009999 {
		t.Errorf("lastLogActivity = %d, want 1750009999", lastLog)
	}
	if snooze != 1750020000 {
		t.Errorf("snoozeUntil = %d, want 1750020000", snooze)
	}
	if snoozeFrom != "Work" {
		t.Errorf("snoozeFromGroup = %q, want Work", snoozeFrom)
	}

	// Clean session stays clean
	blob = MarshalToolData(
		"", time.Time{}, "", "", time.Time{}, nil, "",
		"", time.Time{}, "", time.Time{},
		nil, nil, "", 0, 0, 0, "", "", false,
	)
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, flag, lastLog, snooze, _, _, _ = UnmarshalToolData(blob)
	if flag != 0 || lastLog != 0 || snooze != 0 {
		t.Errorf("expected clean flag/lastLog/snooze, got %d/%d/%d", flag, lastLog, snooze)
	}
}

// TestToolDataLegacyParkedMigration verifies an old row that stored the
// parked bool (before the Flag enum existed) maps to FlagParkedRed (2).
func TestToolDataLegacyParkedMigration(t *testing.T) {
	legacy, _ := json.Marshal(map[string]any{"parked": true})
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, flag, _, _, _, _, _ := UnmarshalToolData(legacy)
	if flag != 2 {
		t.Errorf("legacy parked should migrate to flag 2 (FlagParkedRed), got %d", flag)
	}
}
