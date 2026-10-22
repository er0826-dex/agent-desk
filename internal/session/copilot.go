package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

var ErrCopilotRecoveryRequired = errors.New("no Copilot conversation is linked; press R in Agent Desk to choose its saved conversation, or use session set <id> copilot-session-id <uuid>")

type CopilotConversation struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	Cwd       string    `yaml:"cwd"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

func copilotStateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".copilot", "session-state"), nil
}

func validCopilotID(id string) bool {
	parsed, err := uuid.Parse(id)
	return err == nil && parsed.String() == id
}

func readCopilotConversation(root, id string) (CopilotConversation, error) {
	var conversation CopilotConversation
	if !validCopilotID(id) {
		return conversation, fmt.Errorf("invalid Copilot conversation UUID: %q", id)
	}
	path := filepath.Join(root, id, "workspace.yaml")
	info, err := os.Stat(path)
	if err != nil {
		return conversation, err
	}
	if info.Size() > 64*1024 {
		return conversation, fmt.Errorf("saved Copilot metadata too large: %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return conversation, err
	}
	if err := yaml.Unmarshal(data, &conversation); err != nil {
		return conversation, fmt.Errorf("read Copilot metadata %s: %w", id, err)
	}
	if conversation.ID != id {
		return conversation, fmt.Errorf("saved Copilot metadata ID does not match directory %s", id)
	}
	conversation.Name = SanitizeDisplayName(conversation.Name)
	if conversation.Name == "" {
		conversation.Name = id
	}
	return conversation, nil
}

// ListCopilotConversations reads metadata only, never conversation contents.
// Partial failures are returned alongside usable entries so recovery stays available.
func ListCopilotConversations() ([]CopilotConversation, error) {
	root, err := copilotStateDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("list saved Copilot conversations: %w", err)
	}
	var conversations []CopilotConversation
	var failures []error
	for _, entry := range entries {
		if !entry.IsDir() || !validCopilotID(entry.Name()) {
			continue
		}
		conversation, err := readCopilotConversation(root, entry.Name())
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", entry.Name(), err))
			continue
		}
		conversations = append(conversations, conversation)
	}
	sort.Slice(conversations, func(a, b int) bool {
		if conversations[a].UpdatedAt.Equal(conversations[b].UpdatedAt) {
			return conversations[a].ID < conversations[b].ID
		}
		return conversations[a].UpdatedAt.After(conversations[b].UpdatedAt)
	})
	return conversations, errors.Join(failures...)
}

func (i *Instance) BindCopilotConversation(id string) error {
	if i.Tool != "copilot" {
		return fmt.Errorf("session is not a Copilot session")
	}
	root, err := copilotStateDir()
	if err != nil {
		return err
	}
	if _, err := readCopilotConversation(root, id); err != nil {
		return fmt.Errorf("cannot recover Copilot conversation: %w", err)
	}
	i.CopilotSessionID = id
	i.CopilotSessionPending = false
	return nil
}

// PrepareCopilotSession is for NEW entries only. Existing unlinked entries must
// use recovery rather than silently receiving a new conversation.
func (i *Instance) PrepareCopilotSession() error {
	if i.Tool != "copilot" || i.CopilotSessionID != "" || (i.Command != "" && i.Command != "copilot") {
		return nil
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("allocate Copilot conversation ID: %w", err)
	}
	i.CopilotSessionID = id.String()
	i.CopilotSessionPending = true
	return nil
}

func (i *Instance) validateCopilotLaunch() error {
	if i.Tool != "copilot" || (i.Command != "" && i.Command != "copilot") {
		return nil
	}
	if i.CopilotSessionID == "" {
		return ErrCopilotRecoveryRequired
	}
	if !validCopilotID(i.CopilotSessionID) {
		return fmt.Errorf("invalid Copilot conversation UUID: %q", i.CopilotSessionID)
	}
	if err := i.validateCopilotWrapper(); err != nil {
		return err
	}
	if !i.CopilotSessionPending {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	help, err := exec.CommandContext(ctx, "copilot", "--help").Output()
	if err != nil {
		return fmt.Errorf("check Copilot CLI capabilities: %w", err)
	}
	if !strings.Contains(string(help), "--session-id") {
		return fmt.Errorf("this Copilot CLI does not support --session-id; update Copilot before launching a managed session")
	}
	return nil
}

func (i *Instance) copilotResumeCommand() (string, error) {
	if i.CopilotSessionID == "" {
		return "", ErrCopilotRecoveryRequired
	}
	if !validCopilotID(i.CopilotSessionID) {
		return "", fmt.Errorf("invalid Copilot conversation UUID: %q", i.CopilotSessionID)
	}
	if i.Command != "" && i.Command != "copilot" {
		return "", fmt.Errorf("cannot safely rewrite custom Copilot command; set the command to copilot before recovering, or resume the conversation explicitly in your custom command")
	}
	if err := i.validateCopilotWrapper(); err != nil {
		return "", err
	}
	if i.CopilotSessionPending {
		if err := i.validateCopilotLaunch(); err != nil {
			return "", err
		}
		return "copilot --session-id=" + i.CopilotSessionID, nil
	}
	// Unlike --session-id, --resume must not create a new conversation when
	// history has been removed. Let the CLI report missing or locked history.
	return "copilot --resume=" + i.CopilotSessionID, nil
}

func (i *Instance) validateCopilotWrapper() error {
	wrapper := i.Wrapper
	if wrapper == "" {
		if def := GetToolDef(i.Tool); def != nil {
			wrapper = def.Wrapper
		}
	}
	if wrapper != "" && !strings.Contains(wrapper, wrapperPlaceholder) {
		return fmt.Errorf("the Copilot wrapper must include {command} to preserve the conversation ID")
	}
	return nil
}
