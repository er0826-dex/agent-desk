package session

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/millwright-software/agent-desk/internal/tmux"
)

const copilotTestID = "b6a5a130-786e-4f89-af81-5b9ceff1f262"

func writeCopilotMetadata(t *testing.T, home, id, content string) {
	t.Helper()
	dir := filepath.Join(home, ".copilot", "session-state", id)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCopilotIdentitySurvivesStorageAndReboot(t *testing.T) {
	s := newTestStorage(t)
	var instances []*Instance
	for _, name := range []string{"dashboard", "expenses"} {
		inst := NewInstanceWithTool(name, "/same/project", "copilot")
		inst.Command = "copilot"
		if err := inst.PrepareCopilotSession(); err != nil {
			t.Fatal(err)
		}
		if !validCopilotID(inst.CopilotSessionID) {
			t.Fatalf("not a UUID: %q", inst.CopilotSessionID)
		}
		if got := inst.buildCopilotCommand(inst.Command); got != "copilot --session-id="+inst.CopilotSessionID {
			t.Fatalf("unexpected first launch command: %q", got)
		}
		instances = append(instances, inst)
	}
	if instances[0].CopilotSessionID == instances[1].CopilotSessionID {
		t.Fatal("same-directory sessions must have different conversation IDs")
	}
	if err := s.SaveWithGroups(instances, nil); err != nil {
		t.Fatal(err)
	}
	lite, _, err := s.LoadLite()
	if err != nil || len(lite) != 2 {
		t.Fatalf("LoadLite: %v, %d rows", err, len(lite))
	}
	full, _, err := s.LoadWithGroups()
	if err != nil || len(full) != 2 {
		t.Fatalf("LoadWithGroups: %v, %d rows", err, len(full))
	}
	expected := map[string]string{}
	for _, inst := range instances {
		expected[inst.ID] = inst.CopilotSessionID
	}
	for _, inst := range lite {
		if inst.CopilotSessionID != expected[inst.ID] || !inst.CopilotSessionPending {
			t.Fatalf("lite load lost identity: %+v", inst)
		}
	}
	for _, inst := range full {
		if !inst.CopilotSessionPending {
			t.Fatal("pre-launch state must survive a reload")
		}
		inst.CopilotSessionPending = false
		inst.tmuxSession = nil // A reboot removes processes, not persisted identity.
		cmd, err := inst.copilotResumeCommand()
		if err != nil || cmd != "copilot --resume="+expected[inst.ID] {
			t.Fatalf("wrong recovery command: %q, %v", cmd, err)
		}
		if !inst.CanRestart() {
			t.Fatal("persisted Copilot entry should be recoverable")
		}
	}
	if err := s.SaveWithGroups(full, nil); err != nil {
		t.Fatal(err)
	}
	reloaded, _, err := s.LoadLite()
	if err != nil {
		t.Fatal(err)
	}
	for _, inst := range reloaded {
		if inst.CopilotSessionPending {
			t.Fatal("launched conversation must not become a new conversation after reload")
		}
	}
}

func TestCopilotLegacyRestartDoesNotStartAnotherConversation(t *testing.T) {
	inst := NewInstanceWithTool("legacy", "/same/project", "copilot")
	inst.Command = "copilot"
	if !errors.Is(inst.Restart(), ErrCopilotRecoveryRequired) {
		t.Fatal("legacy restart must request recovery, not run --continue")
	}
	if !errors.Is(inst.Start(), ErrCopilotRecoveryRequired) {
		t.Fatal("start must not silently discard legacy history either")
	}
	if !errors.Is(inst.StartWithMessage("hello"), ErrCopilotRecoveryRequired) {
		t.Fatal("message launch must also require identity")
	}
}

func TestCopilotMetadataAndBinding(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeCopilotMetadata(t, home, copilotTestID, "id: "+copilotTestID+"\nname: Dashboard\ncwd: /project\nupdated_at: 2026-10-22T12:00:00Z\n")
	badID := "b6a5a130-786e-4f89-af81-5b9ceff1f263"
	writeCopilotMetadata(t, home, badID, "id: wrong\n")
	conversations, err := ListCopilotConversations()
	if len(conversations) != 1 || err == nil {
		t.Fatalf("expected one usable entry and a visible metadata error: %v, %v", conversations, err)
	}
	if conversations[0].Name != "Dashboard" || conversations[0].Cwd != "/project" {
		t.Fatalf("wrong metadata: %+v", conversations[0])
	}
	inst := &Instance{Tool: "copilot"}
	if err := inst.BindCopilotConversation(copilotTestID); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"../other", badID, "not-a-uuid"} {
		if err := inst.BindCopilotConversation(invalid); err == nil {
			t.Errorf("accepted invalid conversation %q", invalid)
		}
		if inst.CopilotSessionID != copilotTestID {
			t.Fatal("failed recovery changed existing binding")
		}
	}
}

func TestCopilotCapabilityAndCustomCommands(t *testing.T) {
	bin := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	fake := filepath.Join(bin, "copilot")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho '--resume'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	inst := &Instance{Tool: "copilot", Command: "copilot", CopilotSessionID: copilotTestID, CopilotSessionPending: true}
	if err := inst.validateCopilotLaunch(); err == nil || !strings.Contains(err.Error(), "--session-id") {
		t.Fatalf("unsupported CLI must report an actionable error: %v", err)
	}
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho '--session-id <id>'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := inst.validateCopilotLaunch(); err != nil {
		t.Fatal(err)
	}
	inst.CopilotSessionID = "bad; command"
	if err := inst.validateCopilotLaunch(); err == nil {
		t.Fatal("invalid ID accepted")
	}
	custom := &Instance{Tool: "copilot", Command: "copilot --model custom"}
	if err := custom.PrepareCopilotSession(); err != nil || custom.CopilotSessionID != "" {
		t.Fatal("custom command must not claim an ID it does not pass to Copilot")
	}
	custom.CopilotSessionID = copilotTestID
	if _, err := custom.copilotResumeCommand(); err == nil {
		t.Fatal("custom flags must not silently be discarded")
	}
}

func TestCopilotRestartRecreatesThenRespawnsSameIdentity(t *testing.T) {
	bin := t.TempDir()
	log := filepath.Join(bin, "commands")
	t.Setenv("COPILOT_TEST_LOG", log)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	// A fake tmux keeps this regression independent of the user's real server.
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$COPILOT_TEST_LOG\"\ncase \"$1\" in\nhas-session) exit 1;;\nlist-windows) exit 0;;\nesac\n"
	if err := os.WriteFile(filepath.Join(bin, "tmux"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	tmux.RefreshSessionCache()
	t.Cleanup(tmux.RefreshSessionCache)
	inst := NewInstanceWithTool("copilot-recovery-test", bin, "copilot")
	inst.Command = "copilot"
	inst.CopilotSessionID = copilotTestID
	inst.Wrapper = "env RECOVERY_TEST=1 {command}"
	if err := inst.Restart(); err != nil {
		t.Fatal(err)
	}
	name := inst.tmuxSession.Name
	if err := inst.Restart(); err != nil {
		t.Fatal(err)
	}
	if name != inst.tmuxSession.Name {
		t.Fatal("live restart unnecessarily recreated tmux")
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)
	if !strings.Contains(output, "new-session") || !strings.Contains(output, "respawn-pane") {
		t.Fatalf("did not exercise both reboot and live restart: %s", output)
	}
	if strings.Count(output, "copilot --resume="+copilotTestID) < 2 || strings.Contains(output, "--continue") {
		t.Fatalf("wrong conversation selected: %s", output)
	}
	if !strings.Contains(output, "RECOVERY_TEST=1") {
		t.Fatal("wrapper lost on restart")
	}
	if err := os.WriteFile(filepath.Join(bin, "copilot"), []byte("#!/bin/sh\necho '--session-id <id>'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	pending := NewInstanceWithTool("copilot-not-yet-launched", bin, "copilot")
	if err := pending.PrepareCopilotSession(); err != nil {
		t.Fatal(err)
	}
	if err := pending.Restart(); err != nil {
		t.Fatal(err)
	}
	if pending.CopilotSessionPending {
		t.Fatal("successful first launch did not clear pending state")
	}
	if err := pending.Restart(); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "copilot --session-id="+pending.CopilotSessionID) ||
		!strings.Contains(string(data), "copilot --resume="+pending.CopilotSessionID) {
		t.Fatal("pre-launch recovery must allocate once, then resume")
	}
}

func TestCopilotPreparationIsStable(t *testing.T) {
	inst := &Instance{Tool: "copilot", CreatedAt: time.Now()}
	if err := inst.PrepareCopilotSession(); err != nil {
		t.Fatal(err)
	}
	id := inst.CopilotSessionID
	if err := inst.PrepareCopilotSession(); err != nil || inst.CopilotSessionID != id {
		t.Fatal("preparation changed an already assigned identity")
	}
}
