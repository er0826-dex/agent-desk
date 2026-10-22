package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/millwright-software/agent-desk/internal/session"
)

func TestCopilotRecoveryPickerFilteringAndCancel(t *testing.T) {
	inst := &session.Instance{ID: "entry", Title: "Work", ProjectPath: "/project"}
	conversations := []session.CopilotConversation{
		{ID: "one", Name: "Dashboard", Cwd: "/project"},
		{ID: "two", Name: "Expenses", Cwd: "/other"},
	}
	d := NewCopilotRecoveryDialog(inst, conversations, nil)
	if len(d.filtered()) != 1 || d.selected().ID != "one" {
		t.Fatal("picker should initially show this folder")
	}
	d.update(tea.KeyMsg{Type: tea.KeyTab})
	if len(d.filtered()) != 2 {
		t.Fatal("other folders must remain discoverable")
	}
	d.input.SetValue("Expenses")
	if d.selected().ID != "two" {
		t.Fatal("search did not find conversation")
	}
	if !strings.Contains(d.View(100, 30), "Expenses") {
		t.Fatal("selection is not visible")
	}
	h := &Home{copilotRecoveryDialog: d}
	h.handleCopilotRecoveryKey(tea.KeyMsg{Type: tea.KeyEsc})
	if h.copilotRecoveryDialog != nil || inst.CopilotSessionID != "" {
		t.Fatal("cancel changed the conversation")
	}
}

func TestCopilotRestartRequestsPickerForLegacyEntry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	inst := &session.Instance{ID: "entry", Tool: "copilot"}
	h := &Home{instanceByID: map[string]*session.Instance{inst.ID: inst}}
	msg := h.restartSession(inst)()
	if _, ok := msg.(copilotRecoveryLoadedMsg); !ok {
		t.Fatalf("wanted recovery, got %T", msg)
	}
	if inst.CopilotSessionID != "" {
		t.Fatal("legacy ID was guessed")
	}
}

func TestCopilotRecoverySavesBeforeRestart(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "b6a5a130-786e-4f89-af81-5b9ceff1f262"
	dir := filepath.Join(home, ".copilot", "session-state", id)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}

	metadata := fmt.Sprintf("id: %s\nname: Dashboard\ncwd: /project\n", id)
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}
	storage, err := session.NewStorageWithProfile("_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = storage.GetDB().Close() })
	inst := session.NewInstanceWithTool("entry", "/project", "copilot")
	inst.Command = "copilot"
	if err := storage.SaveWithGroups([]*session.Instance{inst}, nil); err != nil {
		t.Fatal(err)
	}
	conversations, err := session.ListCopilotConversations()
	if err != nil {
		t.Fatal(err)
	}
	h := &Home{
		storage: storage, instances: []*session.Instance{inst},
		instanceByID:          map[string]*session.Instance{inst.ID: inst},
		resumingSessions:      make(map[string]time.Time),
		copilotRecoveryDialog: NewCopilotRecoveryDialog(inst, conversations, nil),
	}
	_, cmd := h.handleCopilotRecoveryKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || h.copilotRecoveryDialog != nil {
		t.Fatal("successful recovery should queue restart and close picker")
	}
	loaded, _, err := storage.LoadLite()
	if err != nil || len(loaded) != 1 || loaded[0].CopilotSessionID != id {
		t.Fatalf("identity not saved before executing restart: %v, %v", loaded, err)
	}

	inst.CopilotSessionID = ""
	h.copilotRecoveryDialog = NewCopilotRecoveryDialog(inst, conversations, nil)
	if err := storage.GetDB().Close(); err != nil {
		t.Fatal(err)
	}
	_, cmd = h.handleCopilotRecoveryKey(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || h.copilotRecoveryDialog == nil || inst.CopilotSessionID != "" {
		t.Fatal("failed save must not queue restart or retain an unpersisted binding")
	}
}

func TestCopilotPrelaunchReloadDoesNotDuplicateEntry(t *testing.T) {
	old := &session.Instance{ID: "same-id"}
	current := &session.Instance{ID: old.ID, CopilotSessionID: "saved-conversation"}
	h := &Home{
		instances:    []*session.Instance{old},
		instanceByID: map[string]*session.Instance{old.ID: old},
	}
	h.registerCreatedInstance(current)
	if len(h.instances) != 1 || h.instances[0] != current || h.instanceByID[old.ID] != current {
		t.Fatal("prelaunch storage reload duplicated or lost the new entry")
	}
}
