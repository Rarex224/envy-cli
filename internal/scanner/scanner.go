// Package scanner discovers projects (dirs with a .env file) under scan roots.
package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Rarex224/envy-cli/internal/config"
)

type Project struct {
	Name       string
	Dir        string
	EnvPath    string // sync target: an existing .env, else <Dir>/.env
	EnvFiles   []string
	SchemaPath string
	Err        error
}

var schemaNames = map[string]bool{
	".env.example":  true,
	".env.sample":   true,
	".env.template": true,
}

var dirsToSkip = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".cache":       true,
}

func Scan(cfg config.Config) []Project {
	seen := map[string]bool{}
	var out []Project
	for _, root := range cfg.Roots {
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				return nil
			}
			if path != root && skipDir(d.Name(), cfg) {
				return filepath.SkipDir
			}
			if p, ok := candidate(path, root); ok && !seen[p.Dir] {
				seen[p.Dir] = true
				out = append(out, p)
			}
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func skipDir(name string, cfg config.Config) bool {
	if dirsToSkip[name] || strings.HasPrefix(name, ".") {
		return true
	}
	return cfg.Ignored(name)
}

func candidate(dir, root string) (Project, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Project{}, false
	}
	var envFiles []string
	var schema string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		switch {
		case schemaNames[name]:
			if schema == "" {
				schema = filepath.Join(dir, name)
			}
		case isEnvFile(name):
			envFiles = append(envFiles, filepath.Join(dir, name))
		}
	}
	if len(envFiles) == 0 {
		return Project{}, false
	}
	sort.Strings(envFiles)

	return Project{
		Name:       relName(root, dir),
		Dir:        dir,
		EnvFiles:   envFiles,
		SchemaPath: schema,
		EnvPath:    filepath.Join(dir, ".env"),
	}, true
}

func isEnvFile(name string) bool {
	if schemaNames[name] || strings.HasPrefix(name, ".env.tmp-") {
		return false
	}
	return name == ".env" || strings.HasPrefix(name, ".env.")
}

func relName(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return filepath.Base(dir)
	}
	return rel
}
