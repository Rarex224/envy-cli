// Package drift compares a schema against an actual .env by key.
package drift

import "github.com/Rarex224/envy-cli/internal/dotenv"

type Kind int

const (
	Ok Kind = iota
	Missing
	Extra
)

func (k Kind) String() string {
	switch k {
	case Missing:
		return "missing"
	case Extra:
		return "extra"
	default:
		return "ok"
	}
}

type Change struct {
	Key  string
	Kind Kind
}

// Diff lists schema keys first (Ok/Missing), then keys only in actual (Extra).
func Diff(schema, actual *dotenv.File) []Change {
	var out []Change
	for _, k := range schema.Keys() {
		if actual.Has(k) {
			out = append(out, Change{k, Ok})
		} else {
			out = append(out, Change{k, Missing})
		}
	}
	for _, k := range actual.Keys() {
		if !schema.Has(k) {
			out = append(out, Change{k, Extra})
		}
	}
	return out
}

func Summarize(changes []Change) (missing, extra int) {
	for _, c := range changes {
		switch c.Kind {
		case Missing:
			missing++
		case Extra:
			extra++
		}
	}
	return missing, extra
}
