package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunUnknownFlag(t *testing.T) {
	if err := run([]string{"envy", "--nope"}, t.TempDir()); err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestRunEmptyScanSucceeds(t *testing.T) {
	// A dir with no .env anywhere: run should print a message and return nil
	// without launching the interactive TUI.
	dir := t.TempDir()
	if err := run([]string{"envy", dir}, dir); err != nil {
		t.Fatalf("expected nil for empty scan, got %v", err)
	}
}

func TestRunMalformedConfigErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "envy.config.json"), []byte("{bad"), 0o644)
	if err := run([]string{"envy"}, dir); err == nil {
		t.Fatal("expected error for malformed config")
	}
}
