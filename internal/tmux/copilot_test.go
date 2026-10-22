package tmux

import (
	"testing"
	"time"
)

func TestIsCopilotCommand(t *testing.T) {
	for _, tt := range []struct {
		command string
		want    bool
	}{
		{"copilot", true},
		{" copilot --resume=some-id ", true},
		{"/usr/local/bin/copilot --model claude-opus", true},
		{"claude", false},
		{"echo copilot", false},
		{"copilot-helper", false},
		{"", false},
	} {
		if got := IsCopilotCommand(tt.command); got != tt.want {
			t.Errorf("IsCopilotCommand(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}

func TestCopilotDetectionUsesCLIInsteadOfModel(t *testing.T) {
	s := NewSession("copilot-detection", t.TempDir())
	for _, command := range []string{"copilot", "copilot --model claude-opus", "copilot --resume=test"} {
		s.Command = command
		if got := s.ForceDetectTool(); got != "copilot" {
			t.Fatalf("%q detected as %q", command, got)
		}
	}
}

func TestCopilotPinnedDetectionOverridesStaleModelDetection(t *testing.T) {
	s := NewSession("copilot-wrapper", t.TempDir())
	s.Command = "env MODEL=claude copilot --resume=test"
	s.detectedTool = "claude"
	s.toolDetectedAt = time.Now()
	s.SetDetectPatterns("copilot", nil)
	if got := s.DetectTool(); got != "copilot" {
		t.Fatalf("model detection overrode the pinned CLI: %q", got)
	}
}
