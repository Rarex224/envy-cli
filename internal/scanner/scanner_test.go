package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Rarex224/envy-cli/internal/config"
)

func mkproj(t *testing.T, dir string, files ...string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("A=1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestScanFindsProjectsAndSchema(t *testing.T) {
	root := t.TempDir()
	mkproj(t, filepath.Join(root, "myapp"), ".env", ".env.example")
	mkproj(t, filepath.Join(root, "api"), ".env") // no schema
	mkproj(t, filepath.Join(root, "docs"))        // no .env, not a project
	mkproj(t, filepath.Join(root, "legacy"), ".env", ".env.example")

	cfg := config.Config{Roots: []string{root}, Ignore: []string{"legacy"}}
	got := Scan(cfg)

	if len(got) != 2 {
		t.Fatalf("got %d projects, want 2: %+v", len(got), got)
	}
	// sorted by name: api, myapp
	if got[0].Name != "api" || got[1].Name != "myapp" {
		t.Fatalf("names = %s, %s", got[0].Name, got[1].Name)
	}
	if got[0].SchemaPath != "" {
		t.Fatalf("api should have no schema, got %q", got[0].SchemaPath)
	}
	if filepath.Base(got[1].SchemaPath) != ".env.example" {
		t.Fatalf("myapp schema = %q", got[1].SchemaPath)
	}
}

func TestScanRootItselfIsProject(t *testing.T) {
	root := t.TempDir()
	mkproj(t, root, ".env", ".env.example")
	got := Scan(config.Config{Roots: []string{root}})
	if len(got) != 1 || got[0].Name != filepath.Base(root) {
		t.Fatalf("expected root as project, got %+v", got)
	}
}

func TestScanRecursiveAndSkips(t *testing.T) {
	root := t.TempDir()
	mkproj(t, filepath.Join(root, "mono", "apps", "web"), ".env", ".env.example")
	mkproj(t, filepath.Join(root, "mono", "services", "api"), ".env")
	mkproj(t, filepath.Join(root, "mono", "node_modules", "pkg"), ".env") // must be skipped

	got := Scan(config.Config{Roots: []string{root}})
	names := map[string]bool{}
	for _, p := range got {
		names[p.Name] = true
	}
	if !names[filepath.Join("mono", "apps", "web")] {
		t.Fatalf("expected nested apps/web project, got names %v", names)
	}
	if !names[filepath.Join("mono", "services", "api")] {
		t.Fatalf("expected nested services/api project, got names %v", names)
	}
	for n := range names {
		if filepath.Base(filepath.Dir(n)) == "node_modules" || n == filepath.Join("mono", "node_modules", "pkg") {
			t.Fatalf("node_modules should be skipped, found %q", n)
		}
	}
}

func TestScanCollectsMultipleEnvFiles(t *testing.T) {
	root := t.TempDir()
	mkproj(t, filepath.Join(root, "app"), ".env", ".env.local", ".env.example")
	got := Scan(config.Config{Roots: []string{root}})
	if len(got) != 1 {
		t.Fatalf("want 1 project, got %d", len(got))
	}
	if len(got[0].EnvFiles) != 2 {
		t.Fatalf("want 2 env files (.env, .env.local), got %v", got[0].EnvFiles)
	}
	if filepath.Base(got[0].EnvPath) != ".env" {
		t.Fatalf("sync target should be .env, got %q", got[0].EnvPath)
	}
	if filepath.Base(got[0].SchemaPath) != ".env.example" {
		t.Fatalf("schema not detected: %q", got[0].SchemaPath)
	}
}
