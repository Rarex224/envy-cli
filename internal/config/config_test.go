package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	if got := ExpandTilde("~/code", "/home/x"); got != "/home/x/code" {
		t.Fatalf("ExpandTilde = %q", got)
	}
	if got := ExpandTilde("/abs", "/home/x"); got != "/abs" {
		t.Fatalf("ExpandTilde abs = %q", got)
	}
}

func TestIgnored(t *testing.T) {
	c := Config{Ignore: []string{"legacy", "*-archive"}}
	if !c.Ignored("legacy") || !c.Ignored("old-archive") {
		t.Fatal("expected matches to be ignored")
	}
	if c.Ignored("myapp") {
		t.Fatal("myapp should not be ignored")
	}
}

func TestResolveArgWins(t *testing.T) {
	c, err := Resolve("~/proj", "/cwd", "/home/x")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Roots) != 1 || c.Roots[0] != "/home/x/proj" {
		t.Fatalf("Roots = %v", c.Roots)
	}
}

func TestResolveConfigFile(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "envy.config.json"),
		[]byte(`{"roots":["~/code"],"ignore":["legacy"]}`), 0o644)
	c, err := Resolve("", cwd, "/home/x")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Roots) != 1 || c.Roots[0] != "/home/x/code" {
		t.Fatalf("Roots = %v", c.Roots)
	}
	if !c.Ignored("legacy") {
		t.Fatal("ignore not loaded")
	}
}

func TestResolveDefaultCwd(t *testing.T) {
	cwd := t.TempDir()
	c, err := Resolve("", cwd, "/home/x")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Roots) != 1 || c.Roots[0] != cwd {
		t.Fatalf("Roots = %v, want [%s]", c.Roots, cwd)
	}
}

func TestResolveMalformed(t *testing.T) {
	cwd := t.TempDir()
	os.WriteFile(filepath.Join(cwd, "envy.config.json"), []byte(`{bad`), 0o644)
	if _, err := Resolve("", cwd, "/home/x"); err == nil {
		t.Fatal("expected parse error")
	}
}
