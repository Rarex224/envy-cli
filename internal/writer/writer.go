// Package writer performs atomic, append-only edits to .env files.
package writer

import (
	"os"
	"path/filepath"
	"strings"
)

// SyncMissing appends missingKeys (empty-valued) to the .env at envPath,
// preserving existing bytes. It creates the file if absent; empty input is a no-op.
func SyncMissing(envPath string, missingKeys []string) error {
	if len(missingKeys) == 0 {
		return nil
	}
	existing, err := os.ReadFile(envPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		existing = nil
	}

	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("\n# from .env.example\n")
	for _, k := range missingKeys {
		b.WriteString(k)
		b.WriteString("=\n")
	}

	return atomicWrite(envPath, []byte(b.String()))
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".env.tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
