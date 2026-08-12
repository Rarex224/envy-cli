package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const contentWidth = 66

var (
	accent     = lipgloss.Color("212") // UI accent (selection)
	deep       = lipgloss.Color("61")  // header bar
	brandGreen = lipgloss.Color("78")  // logo green (intro splash)
	fgBright   = lipgloss.Color("231")
	fgDim      = lipgloss.Color("244")

	pillText   = lipgloss.Color("235")
	pillGreen  = lipgloss.Color("78")
	pillRed    = lipgloss.Color("203")
	pillYellow = lipgloss.Color("214")
	pillGray   = lipgloss.Color("245")
)

var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	headerBar = lipgloss.NewStyle().Bold(true).Foreground(fgBright).Background(deep)

	subtitleStyle = lipgloss.NewStyle().Foreground(fgDim)
	taglineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Italic(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	dimStyle      = lipgloss.NewStyle().Foreground(fgDim)
	ruleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	valueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	statusStyle = lipgloss.NewStyle().
			Foreground(fgBright).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	barStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)

	nameText    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	nameTextSel = lipgloss.NewStyle().Foreground(accent).Bold(true)

	nameColStyle = lipgloss.NewStyle().Width(30).MaxWidth(30)
	keyColStyle  = lipgloss.NewStyle().Width(28).MaxWidth(28)
	valColStyle  = lipgloss.NewStyle().Width(24).MaxWidth(24)
)

// pill renders a rounded, colored badge.
func pill(bg lipgloss.Color, text string) string {
	return lipgloss.NewStyle().
		Foreground(pillText).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// header renders a full-width title bar with a left title and right meta.
func header(left, right string) string {
	gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return headerBar.Render(" " + left + strings.Repeat(" ", gap) + right + " ")
}

// rule renders a full-width horizontal divider.
func rule() string {
	return ruleStyle.Render(strings.Repeat("─", contentWidth))
}
