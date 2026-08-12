package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncMissingAppends(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env")
	original := "A=1\nB=2\n"
	os.WriteFile(env, []byte(original), 0o644)

	if err := SyncMissing(env, []string{"REDIS_URL", "STRIPE_KEY"}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(env)
	got := string(data)
	if !strings.HasPrefix(got, original) {
		t.Fatalf("existing content changed:\n%q", got)
	}
	if !strings.Contains(got, "# from .env.example") {
		t.Fatal("missing provenance comment")
	}
	if !strings.Contains(got, "\nREDIS_URL=\n") || !strings.Contains(got, "\nSTRIPE_KEY=\n") {
		t.Fatalf("missing appended keys:\n%q", got)
	}
}

func TestSyncMissingNoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env")
	os.WriteFile(env, []byte("A=1"), 0o644) // no trailing newline

	if err := SyncMissing(env, []string{"B"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(env)
	if !strings.Contains(string(data), "A=1\n") {
		t.Fatalf("expected newline inserted before appended block:\n%q", string(data))
	}
}

func TestSyncMissingCreatesFile(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env") // does not exist yet
	if err := SyncMissing(env, []string{"A", "B"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(env)
	if err != nil {
		t.Fatalf(".env was not created: %v", err)
	}
	if !strings.Contains(string(data), "A=\n") || !strings.Contains(string(data), "B=\n") {
		t.Fatalf("created file missing keys:\n%q", string(data))
	}
}

func TestSyncMissingEmptyIsNoop(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env")
	os.WriteFile(env, []byte("A=1\n"), 0o644)
	if err := SyncMissing(env, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(env)
	if string(data) != "A=1\n" {
		t.Fatalf("no-op changed file: %q", string(data))
	}
}
