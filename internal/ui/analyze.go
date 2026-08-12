package ui

import (
	"github.com/Rarex224/envy-cli/internal/dotenv"
	"github.com/Rarex224/envy-cli/internal/drift"
	"github.com/Rarex224/envy-cli/internal/scanner"
)

type Summary struct {
	Project  scanner.Project
	Changes  []drift.Change
	Values   map[string]string
	Missing  int
	Extra    int
	NoSchema bool
	Err      error
}

// Analyze diffs a project's env files against its schema; a key present in any
// env file (.env, .env.local, ...) counts as defined.
func Analyze(p scanner.Project) Summary {
	s := Summary{Project: p, Values: map[string]string{}}
	if p.Err != nil {
		s.Err = p.Err
		return s
	}

	merged := &dotenv.File{}
	for _, path := range p.EnvFiles {
		f, err := dotenv.ParseFile(path)
		if err != nil {
			s.Err = err
			return s
		}
		for _, e := range f.Entries {
			if e.Key == "" {
				continue
			}
			if _, ok := s.Values[e.Key]; !ok {
				merged.Entries = append(merged.Entries, e)
			}
			s.Values[e.Key] = e.Value
		}
	}

	if p.SchemaPath == "" {
		s.NoSchema = true
		return s
	}
	schema, err := dotenv.ParseFile(p.SchemaPath)
	if err != nil {
		s.Err = err
		return s
	}
	s.Changes = drift.Diff(schema, merged)
	s.Missing, s.Extra = drift.Summarize(s.Changes)
	return s
}

func MissingKeys(s Summary) []string {
	var out []string
	for _, c := range s.Changes {
		if c.Kind == drift.Missing {
			out = append(out, c.Key)
		}
	}
	return out
}
