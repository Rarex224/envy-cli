package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Rarex224/envy-cli/internal/scanner"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAnalyzeMissingAndExtra(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), "A=1\nZ=9\n")
	writeFile(t, filepath.Join(dir, ".env.example"), "A=\nB=\nC=\n")
	p := scanner.Project{
		Name: "myapp", Dir: dir,
		EnvPath:    filepath.Join(dir, ".env"),
		EnvFiles:   []string{filepath.Join(dir, ".env")},
		SchemaPath: filepath.Join(dir, ".env.example"),
	}

	s := Analyze(p)
	if s.Missing != 2 || s.Extra != 1 {
		t.Fatalf("summary = missing %d extra %d, want 2/1", s.Missing, s.Extra)
	}
	if got := MissingKeys(s); len(got) != 2 || got[0] != "B" || got[1] != "C" {
		t.Fatalf("MissingKeys = %v, want [B C]", got)
	}
}

func TestAnalyzeMergesEnvFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), "A=1\n")
	writeFile(t, filepath.Join(dir, ".env.local"), "B=2\n") // B defined only here
	writeFile(t, filepath.Join(dir, ".env.example"), "A=\nB=\nC=\n")
	p := scanner.Project{
		Name: "app", Dir: dir,
		EnvPath:    filepath.Join(dir, ".env"),
		EnvFiles:   []string{filepath.Join(dir, ".env"), filepath.Join(dir, ".env.local")},
		SchemaPath: filepath.Join(dir, ".env.example"),
	}
	s := Analyze(p)
	// Only C is missing; B is satisfied by .env.local.
	if got := MissingKeys(s); len(got) != 1 || got[0] != "C" {
		t.Fatalf("MissingKeys = %v, want [C]", got)
	}
	if s.Values["B"] != "2" {
		t.Fatalf("Values[B] = %q, want 2", s.Values["B"])
	}
}

func TestAnalyzeNoSchema(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".env"), "A=1\n")
	p := scanner.Project{
		Name: "api", Dir: dir,
		EnvPath:  filepath.Join(dir, ".env"),
		EnvFiles: []string{filepath.Join(dir, ".env")},
	}
	s := Analyze(p)
	if !s.NoSchema {
		t.Fatal("expected NoSchema = true")
	}
}
