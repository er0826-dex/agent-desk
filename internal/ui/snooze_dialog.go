package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/millwright-software/agent-desk/internal/session"
)

// snoozeSetMsg is sent when the user picks a wake time. pick is the
// suggestion label, or "typed" for a hand-entered time, for the tally.
type snoozeSetMsg struct {
	instanceID string
	until      time.Time
	pick       string
}

// SnoozeDialog is the s-key picker. The text field comes first and has focus,
// so typing starts immediately ("2h", "7am", "mon 9am", "10/12 11:00") and
// the resolved time shows underneath as you type. Below it, a short list of
// suggestions built for the time of day (session.SnoozeSuggestions), which
// narrows to whatever matches the typed text; ↓ moves onto the list. Enter
// takes the first matching suggestion while any match, else the typed time.
type SnoozeDialog struct {
	visible     bool
	width       int
	height      int
	instanceID  string
	title       string
	input       textinput.Model
	cursor      int // -1 = the text field; 0.. = a row of filtered()
	suggestions []session.SnoozePreset
	err         string
}

// NewSnoozeDialog creates the dialog.
func NewSnoozeDialog() *SnoozeDialog {
	ti := textinput.New()
	ti.Placeholder = "2h · 7am · tomorrow · fri 9am · 10/12 11:00"
	ti.CharLimit = 32
	ti.Width = 30
	ti.Prompt = ""
	return &SnoozeDialog{input: ti, cursor: -1}
}

// Show opens the dialog for a session with the text field focused.
func (d *SnoozeDialog) Show(instanceID, title string) {
	d.visible = true
	d.instanceID = instanceID
	d.title = title
	d.err = ""
	d.suggestions = session.SnoozeSuggestions(time.Now())
	d.input.SetValue("")
	d.focusInput()
}

// Hide closes the dialog.
func (d *SnoozeDialog) Hide() {
	d.visible = false
	d.input.Blur()
}

// IsVisible reports whether the dialog is shown.
func (d *SnoozeDialog) IsVisible() bool { return d.visible }

// SetSize updates dialog dimensions.
func (d *SnoozeDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

func (d *SnoozeDialog) focusInput() {
	d.cursor = -1
	d.input.Focus()
}

// filtered returns the suggestions matching the typed text (all when empty).
func (d *SnoozeDialog) filtered() []session.SnoozePreset {
	q := strings.ToLower(strings.TrimSpace(d.input.Value()))
	if q == "" {
		return d.suggestions
	}
	var out []session.SnoozePreset
	for _, p := range d.suggestions {
		if strings.Contains(strings.ToLower(p.Label), q) {
			out = append(out, p)
		}
	}
	return out
}

// resolve returns what Enter would pick right now: the highlighted row; else
// the first suggestion still matching the typed text (what is on screen
// wins: "week" means the Later this week row, not a parse of the word);
// else the typed text as a time. The second return is the tally label.
func (d *SnoozeDialog) resolve(now time.Time) (time.Time, string, error) {
	rows := d.filtered()
	if d.cursor >= 0 && d.cursor < len(rows) {
		return rows[d.cursor].At, rows[d.cursor].Label, nil
	}
	if len(rows) > 0 {
		return rows[0].At, rows[0].Label, nil
	}
	t, err := session.ParseSnooze(d.input.Value(), now)
	return t, "typed", err
}

// Update handles key input. Returns a snoozeSetMsg command on Enter.
func (d *SnoozeDialog) Update(msg tea.KeyMsg) (*SnoozeDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}
	switch msg.String() {
	case "esc":
		d.Hide()
		return d, nil
	case "down", "ctrl+n", "tab":
		if d.cursor < len(d.filtered())-1 {
			d.cursor++
			d.input.Blur()
		}
		return d, nil
	case "up", "ctrl+p", "shift+tab":
		if d.cursor > 0 {
			d.cursor--
		} else {
			d.focusInput()
		}
		return d, nil
	case "enter":
		// Rebuild the rows for this instant so a dialog left open does not
		// hand back a landmark that has since passed.
		d.suggestions = session.SnoozeSuggestions(time.Now())
		until, pick, err := d.resolve(time.Now())
		if err != nil {
			d.err = err.Error()
			return d, nil
		}
		instanceID := d.instanceID
		d.Hide()
		return d, func() tea.Msg {
			return snoozeSetMsg{instanceID: instanceID, until: until, pick: pick}
		}
	}
	// Anything else types into the field, pulling focus back from the list.
	if d.cursor >= 0 {
		d.focusInput()
	}
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	d.err = ""
	if d.cursor >= len(d.filtered()) {
		d.cursor = -1
	}
	return d, cmd
}

// previewTime renders a resolved wake time for the dialog.
func previewTime(t, now time.Time) string {
	if t.Sub(now) < 6*24*time.Hour {
		return t.Format("Mon 3:04pm")
	}
	return t.Format("Mon 1/2 3:04pm")
}

// View renders the field, the live preview, and the suggestions.
func (d *SnoozeDialog) View() string {
	if !d.visible {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorCyan)
	dimStyle := lipgloss.NewStyle().Foreground(ColorComment)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
	timeStyle := lipgloss.NewStyle().Foreground(ColorTextDim)
	previewStyle := lipgloss.NewStyle().Foreground(ColorGreen)
	errStyle := lipgloss.NewStyle().Foreground(ColorRed)

	dialogWidth := 50
	if d.width > 0 && d.width < dialogWidth+10 {
		dialogWidth = d.width - 10
		if dialogWidth < 36 {
			dialogWidth = 36
		}
	}

	now := time.Now()
	var content strings.Builder
	content.WriteString(titleStyle.Render("Snooze session"))
	content.WriteString(dimStyle.Render("            [Esc] Cancel"))
	content.WriteString("\n")
	content.WriteString(strings.Repeat("-", dialogWidth-4))
	content.WriteString("\n")
	name := d.title
	if len(name) > dialogWidth-8 {
		name = name[:dialogWidth-11] + "..."
	}
	content.WriteString(dimStyle.Render(name))
	content.WriteString("\n\n")

	// The field, with what Enter would do underneath.
	if d.cursor < 0 {
		content.WriteString("> " + selectedStyle.Render("When:") + " " + d.input.View())
	} else {
		content.WriteString("  When: " + d.input.View())
	}
	content.WriteString("\n")
	val := strings.TrimSpace(d.input.Value())
	switch {
	case d.err != "":
		content.WriteString("  " + errStyle.Render(d.err))
	case d.cursor < 0 && val == "":
		content.WriteString("  " + dimStyle.Render("type a time, or ↓ to pick one"))
	case d.cursor < 0:
		if t, _, err := d.resolve(now); err == nil {
			content.WriteString("  " + previewStyle.Render("→ "+previewTime(t, now)))
		} else {
			content.WriteString("  " + dimStyle.Render("?"))
		}
	default:
		content.WriteString("  " + dimStyle.Render("↑ to edit"))
	}
	content.WriteString("\n\n")

	rows := d.filtered()
	for i, p := range rows {
		label := p.Label
		pad := 22 - len(label)
		if pad < 1 {
			pad = 1
		}
		line := "  " + label
		if i == d.cursor {
			line = "> " + selectedStyle.Render(label)
		}
		line += strings.Repeat(" ", pad) + timeStyle.Render(previewTime(p.At, now))
		content.WriteString(line)
		content.WriteString("\n")
	}
	if len(rows) == 0 && val != "" {
		content.WriteString(dimStyle.Render("  (no suggestions match; Enter uses the typed time)"))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(dimStyle.Render("↑/↓ Pick  Enter Snooze  Esc Cancel"))

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyan).
		Background(ColorBg).
		Padding(1, 2).
		Width(dialogWidth)

	return lipgloss.Place(
		d.width, d.height,
		lipgloss.Center, lipgloss.Center,
		dialogStyle.Render(content.String()),
	)
}
