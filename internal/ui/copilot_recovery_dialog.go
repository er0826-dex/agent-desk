package ui

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/millwright-software/agent-desk/internal/session"
)

type copilotRecoveryLoadedMsg struct {
	instanceID    string
	conversations []session.CopilotConversation
	err           error
}

type CopilotRecoveryDialog struct {
	instanceID    string
	title         string
	projectPath   string
	conversations []session.CopilotConversation
	input         textinput.Model
	cursor        int
	allFolders    bool
	warning       string
}

func NewCopilotRecoveryDialog(inst *session.Instance, conversations []session.CopilotConversation, err error) *CopilotRecoveryDialog {
	input := textinput.New()
	input.Placeholder = "Filter by name, UUID, or folder..."
	input.Focus()
	d := &CopilotRecoveryDialog{
		instanceID: inst.ID, title: inst.Title, projectPath: inst.ProjectPath,
		conversations: conversations, input: input,
	}
	if err != nil {
		d.warning = "Some saved metadata could not be read. See Agent Desk logs."
		uiLog.Warn("copilot_recovery_metadata", slog.String("error", err.Error()))
	}
	if len(d.filtered()) == 0 {
		d.allFolders = true
	}
	return d
}

func (d *CopilotRecoveryDialog) filtered() []session.CopilotConversation {
	query := strings.ToLower(strings.TrimSpace(d.input.Value()))
	var result []session.CopilotConversation
	for _, c := range d.conversations {
		if !d.allFolders && filepath.Clean(c.Cwd) != filepath.Clean(d.projectPath) {
			continue
		}
		if query == "" || strings.Contains(strings.ToLower(c.Name+" "+c.ID+" "+c.Cwd), query) {
			result = append(result, c)
		}
	}
	return result
}

func (d *CopilotRecoveryDialog) selected() *session.CopilotConversation {
	rows := d.filtered()
	if d.cursor < 0 || d.cursor >= len(rows) {
		return nil
	}
	return &rows[d.cursor]
}

func (d *CopilotRecoveryDialog) update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "up", "ctrl+p":
		if d.cursor > 0 {
			d.cursor--
		}
	case "down", "ctrl+n":
		if d.cursor+1 < len(d.filtered()) {
			d.cursor++
		}
	case "tab":
		d.allFolders = !d.allFolders
		d.cursor = 0
	default:
		var cmd tea.Cmd
		d.input, cmd = d.input.Update(msg)
		d.cursor = 0
		return cmd
	}
	return nil
}

func (d *CopilotRecoveryDialog) View(width, height int) string {
	innerWidth := max(24, min(90, width-8))
	d.input.Width = innerWidth - 4
	scope := "this folder"
	if d.allFolders {
		scope = "all folders"
	}
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render("Recover Copilot conversation"),
		"Agent Desk entry: " + d.title,
		"Choose its saved conversation. Enter confirms the link.",
		d.input.View(),
		fmt.Sprintf("Showing %s (Tab toggles)", scope),
	}
	rows := d.filtered()
	if len(rows) == 0 {
		lines = append(lines, "No matching conversations. Try Tab or clear the filter.")
	} else {
		pageSize := max(1, min(7, (height-14)/2))
		start := max(0, d.cursor-pageSize+1)
		end := min(len(rows), start+pageSize)
		for idx := start; idx < end; idx++ {
			c := rows[idx]
			label := c.Name
			if idx == d.cursor {
				label = "> " + label
				label = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(label)
			} else {
				label = "  " + label
			}
			lines = append(lines, label,
				fmt.Sprintf("  %s | %s", c.UpdatedAt.Local().Format("Jan 02 15:04"), c.ID))
		}
		if selected := d.selected(); selected != nil {
			lines = append(lines, "Folder: "+session.SanitizeDisplayName(selected.Cwd))
		}
		lines = append(lines, fmt.Sprintf("%d of %d", d.cursor+1, len(rows)))
	}
	if d.warning != "" {
		lines = append(lines, d.warning)
	}
	lines = append(lines, "Enter link + resume | Esc cancel | Up/Down navigate")
	box := DialogBoxStyle.Width(innerWidth).Render(strings.Join(lines, "\n"))
	return centerInScreen(box, width, height)
}

func (h *Home) handleCopilotRecoveryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	d := h.copilotRecoveryDialog
	switch msg.String() {
	case "esc":
		h.copilotRecoveryDialog = nil
		return h, nil
	case "enter":
		chosen := d.selected()
		if chosen == nil {
			return h, nil
		}
		inst := h.getInstanceByID(d.instanceID)
		if inst == nil {
			d.warning = "This Agent Desk entry no longer exists. Press Esc."
			return h, nil
		}
		// A storage reload may have linked this entry while the picker was open.
		if inst.CopilotSessionID != "" {
			d.warning = "This entry was linked elsewhere. Press Esc and retry."
			return h, nil
		}
		for _, other := range h.instances {
			if other.ID != inst.ID && other.CopilotSessionID == chosen.ID {
				d.warning = "Already linked to: " + other.Title
				return h, nil
			}
		}
		if h.storage == nil {
			d.warning = "Storage unavailable; conversation was not changed."
			return h, nil
		}
		if err := inst.BindCopilotConversation(chosen.ID); err != nil {
			d.warning = err.Error()
			return h, nil
		}
		if h.storageWatcher != nil {
			h.storageWatcher.NotifySave()
		}
		// Save just this entry. Failure must leave the picker open, not launch
		// an unpersisted binding that will be lost at the next reboot.
		if err := h.storage.SaveWithGroups([]*session.Instance{inst}, nil); err != nil {
			inst.CopilotSessionID = ""
			d.warning = fmt.Sprintf("Could not save conversation link: %v", err)
			return h, nil
		}
		h.copilotRecoveryDialog = nil
		h.resumingSessions[inst.ID] = time.Now()
		return h, h.restartSession(inst)
	default:
		return h, d.update(msg)
	}
}
