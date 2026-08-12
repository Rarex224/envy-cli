package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Rarex224/envy-cli/internal/config"
	"github.com/Rarex224/envy-cli/internal/scanner"
	"github.com/Rarex224/envy-cli/internal/ui"
)

// run is the testable entry point. args is os.Args; cwd is the working dir.
func run(args []string, cwd string) error {
	var path string
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "-") {
			return fmt.Errorf("unknown flag: %s", a)
		}
		path = a
	}

	home, _ := os.UserHomeDir()
	cfg, err := config.Resolve(path, cwd, home)
	if err != nil {
		return err
	}

	projects := scanner.Scan(cfg)
	if len(projects) == 0 {
		fmt.Printf("envy: no projects with a .env found under %s\n", strings.Join(cfg.Roots, ", "))
		return nil
	}

	_, err = tea.NewProgram(ui.New(projects), tea.WithAltScreen()).Run()
	return err
}

func main() {
	cwd, _ := os.Getwd()
	if err := run(os.Args, cwd); err != nil {
		fmt.Fprintln(os.Stderr, "envy:", err)
		os.Exit(1)
	}
}
