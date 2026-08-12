// Package config resolves scan roots and ignore globs.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	Roots  []string `json:"roots"`
	Ignore []string `json:"ignore"`
}

const fileName = "envy.config.json"

func Parse(data []byte) (Config, error) {
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", fileName, err)
	}
	return c, nil
}

func ExpandTilde(p, home string) string {
	if p == "~" {
		return home
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}

func (c Config) Ignored(name string) bool {
	for _, g := range c.Ignore {
		if ok, _ := path.Match(g, name); ok {
			return true
		}
	}
	return false
}

// Resolve order: arg > envy.config.json in cwd > cwd.
func Resolve(arg, cwd, home string) (Config, error) {
	if arg != "" {
		return Config{Roots: []string{ExpandTilde(arg, home)}}, nil
	}
	data, err := os.ReadFile(filepath.Join(cwd, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{Roots: []string{cwd}}, nil
		}
		return Config{}, err
	}
	c, err := Parse(data)
	if err != nil {
		return Config{}, err
	}
	for i, r := range c.Roots {
		c.Roots[i] = ExpandTilde(r, home)
	}
	if len(c.Roots) == 0 {
		c.Roots = []string{cwd}
	}
	return c, nil
}
