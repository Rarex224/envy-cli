// Package dotenv parses .env files while preserving line order.
package dotenv

import (
	"os"
	"strings"
)

type Entry struct {
	Key   string
	Value string
	Raw   string
}

type File struct {
	Entries []Entry
}

// Parse understands export, quotes, = in values, inline # comments, and
// multi-line quoted values (e.g. PEM keys).
func Parse(data []byte) *File {
	f := &File{}
	lines := strings.Split(string(data), "\n")
	for i := 0; i < len(lines); i++ {
		raw := strings.TrimRight(lines[i], "\r")
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			f.Entries = append(f.Entries, Entry{Raw: raw})
			continue
		}

		work := strings.TrimPrefix(trimmed, "export ")
		eq := strings.IndexByte(work, '=')
		if eq < 0 {
			f.Entries = append(f.Entries, Entry{Raw: raw})
			continue
		}

		key := strings.TrimSpace(work[:eq])
		rest := strings.TrimLeft(work[eq+1:], " \t")

		if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
			q := rest[0]
			body := rest[1:]
			if idx := strings.IndexByte(body, q); idx >= 0 {
				f.Entries = append(f.Entries, Entry{Key: key, Value: body[:idx], Raw: raw})
				continue
			}
			var sb strings.Builder
			sb.WriteString(body)
			for i+1 < len(lines) {
				i++
				l := strings.TrimRight(lines[i], "\r")
				if idx := strings.IndexByte(l, q); idx >= 0 {
					sb.WriteString("\n" + l[:idx])
					break
				}
				sb.WriteString("\n" + l)
			}
			f.Entries = append(f.Entries, Entry{Key: key, Value: sb.String(), Raw: raw})
			continue
		}

		val := strings.TrimSpace(stripInlineComment(rest))
		f.Entries = append(f.Entries, Entry{Key: key, Value: val, Raw: raw})
	}
	return f
}

// stripInlineComment drops a trailing # comment only when the # follows
// whitespace, so values like http://x#frag stay intact.
func stripInlineComment(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '#' && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t') {
			return s[:i]
		}
	}
	return s
}

func ParseFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data), nil
}

func (f *File) Keys() []string {
	var keys []string
	for _, e := range f.Entries {
		if e.Key != "" {
			keys = append(keys, e.Key)
		}
	}
	return keys
}

func (f *File) Has(key string) bool {
	for _, e := range f.Entries {
		if e.Key == key {
			return true
		}
	}
	return false
}
