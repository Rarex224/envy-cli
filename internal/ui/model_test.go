package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Rarex224/envy-cli/internal/scanner"
)

func makeProject(t *testing.T) scanner.Project {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), "A=1\n")
	writeFile(t, filepath.Join(dir, ".env.example"), "A=\nB=\n")
	return scanner.Project{
		Name: "myapp", Dir: dir,
		EnvPath:    filepath.Join(dir, ".env"),
		EnvFiles:   []string{filepath.Join(dir, ".env")},
		SchemaPath: filepath.Join(dir, ".env.example"),
	}
}

// ready builds a model past the intro splash.
func ready(projects []scanner.Project) Model {
	m := New(projects)
	done, _ := m.Update(bootDoneMsg{})
	return done.(Model)
}

func TestEnterOpensFileView(t *testing.T) {
	m := ready([]scanner.Project{makeProject(t)})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if updated.(Model).screen != screenFile {
		t.Fatal("enter did not open file view")
	}
}

func TestEscReturnsToDashboard(t *testing.T) {
	m := ready([]scanner.Project{makeProject(t)})
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	back, _ := opened.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	if back.(Model).screen != screenDashboard {
		t.Fatal("esc did not return to dashboard")
	}
}

func TestSyncFlipsMissingToOk(t *testing.T) {
	p := makeProject(t)
	m := ready([]scanner.Project{p})
	opened, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	synced, _ := opened.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	sm := synced.(Model)
	if sm.summaries[sm.cursor].Missing != 0 {
		t.Fatalf("after sync, missing = %d, want 0", sm.summaries[sm.cursor].Missing)
	}
	data, _ := os.ReadFile(p.EnvPath)
	if !strings.Contains(string(data), "\nB=\n") {
		t.Fatalf(".env not updated: %q", string(data))
	}
}

func TestViewRendersBadges(t *testing.T) {
	m := ready([]scanner.Project{makeProject(t)})
	out := m.View()
	if out == "" {
		t.Fatal("empty view")
	}
}
