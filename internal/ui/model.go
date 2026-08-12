package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Rarex224/envy-cli/internal/drift"
	"github.com/Rarex224/envy-cli/internal/scanner"
	"github.com/Rarex224/envy-cli/internal/writer"
)

type screen int

const (
	screenDashboard screen = iota
	screenFile
)

// bootDoneMsg ends the intro splash.
type bootDoneMsg struct{}

func bootTimer() tea.Cmd {
	return tea.Tick(1300*time.Millisecond, func(time.Time) tea.Msg { return bootDoneMsg{} })
}

type Model struct {
	summaries     []Summary
	cursor        int
	row           int
	screen        screen
	revealed      map[int]bool
	status        string
	booting       bool
	width, height int
}

func New(projects []scanner.Project) Model {
	sums := make([]Summary, len(projects))
	for i, p := range projects {
		sums[i] = Analyze(p)
	}
	return Model{summaries: sums, revealed: map[int]bool{}, booting: true}
}

func (m Model) Init() tea.Cmd { return bootTimer() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case bootDoneMsg:
		m.booting = false
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
		if m.booting { // any key skips the intro
			m.booting = false
			return m, nil
		}
		if m.screen == screenDashboard {
			return m.updateDashboard(msg)
		}
		return m.updateFile(msg)
	}
	return m, nil
}

func (m Model) updateDashboard(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.summaries)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.summaries) > 0 {
			m.screen = screenFile
			m.row = 0
			m.revealed = map[int]bool{}
			m.status = ""
		}
	}
	return m, nil
}

func (m Model) updateFile(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := m.summaries[m.cursor]
	switch key.String() {
	case "esc":
		m.screen = screenDashboard
		m.status = ""
	case "up", "k":
		if m.row > 0 {
			m.row--
		}
	case "down", "j":
		if m.row < len(s.Changes)-1 {
			m.row++
		}
	case "m":
		m.revealed[m.row] = !m.revealed[m.row]
	case "c":
		m.status = m.copyValue(s)
	case "s":
		m.status = m.sync(s)
	}
	return m, nil
}

func (m Model) copyValue(s Summary) string {
	if m.row >= len(s.Changes) {
		return ""
	}
	key := s.Changes[m.row].Key
	val, ok := s.Values[key]
	if !ok {
		return key + " has no value to copy"
	}
	if err := clipboard.WriteAll(val); err != nil {
		return "clipboard unavailable"
	}
	return "copied " + key
}

func (m Model) sync(s Summary) string {
	missing := MissingKeys(s)
	if len(missing) == 0 {
		return "nothing to sync"
	}
	if err := writer.SyncMissing(s.Project.EnvPath, missing); err != nil {
		return "sync failed: " + err.Error()
	}
	m.summaries[m.cursor] = Analyze(s.Project)
	return fmt.Sprintf("synced %d key(s)", len(missing))
}

func (m Model) View() string {
	if m.booting {
		return m.viewSplash()
	}
	if m.screen == screenFile {
		return m.viewFile()
	}
	return m.viewDashboard()
}

// viewSplash is the branded intro shown briefly on launch.
func (m Model) viewSplash() string {
	green := lipgloss.NewStyle().Foreground(brandGreen).Bold(true)
	eye := green.Render("  ◜◉◝")
	smile := green.Render("  ╰‿‿╯")
	name := lipgloss.NewStyle().Foreground(fgBright).Bold(true).Render("e n v y")
	tag := taglineStyle.Render("find drift · sync keys · keep secrets masked")

	block := lipgloss.JoinVertical(lipgloss.Center, eye, smile, "", name, "", tag)
	card := cardStyle.Padding(2, 4).Render(block)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}
	return card
}

func (m Model) viewDashboard() string {
	var b strings.Builder
	b.WriteString(header("🔎 envy", "v0.1.0") + "\n")

	keys, miss, extra := aggregate(m.summaries)
	b.WriteString(subtitleStyle.Render(fmt.Sprintf(" %d projects · %d keys · %d missing · %d extra",
		len(m.summaries), keys, miss, extra)) + "\n")
	b.WriteString(rule() + "\n\n")

	if len(m.summaries) == 0 {
		b.WriteString(dimStyle.Render(" no projects with a .env found"))
		return cardStyle.Render(b.String())
	}

	for i, s := range m.summaries {
		mark := "  "
		name := nameText.Render(s.Project.Name)
		if i == m.cursor {
			mark = barStyle.Render("┃ ")
			name = nameTextSel.Render(s.Project.Name)
		}
		b.WriteString(mark + nameColStyle.Render(name) + " " + badge(s) + "\n")
	}

	b.WriteString("\n" + helpStyle.Render(" ↑↓ move · enter open · q quit"))
	return cardStyle.Render(b.String())
}

// aggregate totals keys/missing/extra across all analyzed projects.
func aggregate(ss []Summary) (keys, missing, extra int) {
	for _, s := range ss {
		keys += len(s.Changes)
		missing += s.Missing
		extra += s.Extra
	}
	return keys, missing, extra
}

func badge(s Summary) string {
	switch {
	case s.Err != nil:
		return pill(pillRed, "unreadable")
	case s.NoSchema:
		return pill(pillGray, "no schema")
	case s.Missing == 0 && s.Extra == 0:
		return pill(pillGreen, "ok ✓")
	default:
		var parts []string
		if s.Missing > 0 {
			parts = append(parts, pill(pillRed, fmt.Sprintf("%d missing", s.Missing)))
		}
		if s.Extra > 0 {
			parts = append(parts, pill(pillYellow, fmt.Sprintf("%d extra", s.Extra)))
		}
		return strings.Join(parts, " ")
	}
}

func (m Model) viewFile() string {
	s := m.summaries[m.cursor]
	var b strings.Builder
	b.WriteString(header("🔎 envy  ›  "+s.Project.Name, "v0.1.0") + "\n")

	b.WriteString(subtitleStyle.Render(fmt.Sprintf(" .env vs .env.example · %d keys · %d missing · %d extra",
		len(s.Changes), s.Missing, s.Extra)) + "\n")
	b.WriteString(rule() + "\n\n")

	for i, c := range s.Changes {
		mark := "  "
		keyStyle := nameText
		if i == m.row {
			mark = barStyle.Render("┃ ")
			keyStyle = nameTextSel
		}
		key := keyColStyle.Render(keyStyle.Render(c.Key))
		val := valColStyle.Render(renderValue(c, s.Values[c.Key], m.revealed[i]))
		b.WriteString(mark + key + val + " " + tag(c.Kind) + "\n")
	}

	if m.status != "" {
		b.WriteString("\n" + statusStyle.Render(m.status))
	}
	b.WriteString("\n\n" + helpStyle.Render(" esc back · m mask · c copy · s sync · q quit"))
	return cardStyle.Render(b.String())
}

// maxValLen keeps long secrets and multi-line keys from wrapping the tag column.
const maxValLen = 22

func renderValue(c drift.Change, value string, revealed bool) string {
	if c.Kind == drift.Missing {
		return dimStyle.Render("—")
	}
	if value == "" {
		return dimStyle.Render("(empty)")
	}
	if revealed {
		return valueStyle.Render(truncate(oneLine(value), maxValLen))
	}
	return dimStyle.Render(truncate(mask(value), maxValLen))
}

func mask(v string) string {
	v = oneLine(v)
	r := []rune(v)
	if len(r) <= 4 {
		return strings.Repeat("•", len(r))
	}
	return string(r[:4]) + strings.Repeat("•", len(r)-4)
}

func oneLine(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "⏎")
}

// truncate shortens s to at most n cells, adding an ellipsis when it cuts.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func tag(k drift.Kind) string {
	switch k {
	case drift.Missing:
		return pill(pillRed, "missing")
	case drift.Extra:
		return pill(pillYellow, "extra")
	default:
		return pill(pillGreen, "ok")
	}
}
