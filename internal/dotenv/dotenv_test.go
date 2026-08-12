package dotenv

import (
	"reflect"
	"testing"
)

func TestParseKeys(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"simple", "A=1\nB=2\n", []string{"A", "B"}},
		{"export prefix", "export A=1\n", []string{"A"}},
		{"comment and blank", "# note\n\nA=1\n", []string{"A"}},
		{"quoted value with equals", `A="postgres://u:p@h/db?x=1"` + "\n", []string{"A"}},
		{"spaces around key", "  B = 2 \n", []string{"B"}},
		{"empty", "", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Parse([]byte(c.in)).Keys()
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("Keys() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestParseValueUnquoted(t *testing.T) {
	f := Parse([]byte(`A="postgres://u:p@h/db?x=1"` + "\nB='raw'\nC=plain\n"))
	want := map[string]string{"A": "postgres://u:p@h/db?x=1", "B": "raw", "C": "plain"}
	for k, v := range want {
		var got string
		for _, e := range f.Entries {
			if e.Key == k {
				got = e.Value
			}
		}
		if got != v {
			t.Fatalf("%s value = %q, want %q", k, got, v)
		}
	}
}

func TestInlineComments(t *testing.T) {
	f := Parse([]byte("PORT=8080 # the http port\nHASH=ab#cd\nURL=http://x#frag\nQ=\"a # b\"\n"))
	want := map[string]string{
		"PORT": "8080",          // inline comment stripped
		"HASH": "ab#cd",         // no space before # -> kept
		"URL":  "http://x#frag", // fragment kept
		"Q":    "a # b",         // # inside quotes kept
	}
	for k, v := range want {
		got := valueOf(f, k)
		if got != v {
			t.Fatalf("%s = %q, want %q", k, got, v)
		}
	}
}

func TestMultilineQuotedValue(t *testing.T) {
	in := "PRIVATE_KEY=\"line1\nline2=stillvalue\nline3\"\nNEXT=ok\n"
	f := Parse([]byte(in))
	if got := f.Keys(); len(got) != 2 || got[0] != "PRIVATE_KEY" || got[1] != "NEXT" {
		t.Fatalf("Keys() = %v, want [PRIVATE_KEY NEXT] (no spurious keys from inner =)", got)
	}
	if got := valueOf(f, "PRIVATE_KEY"); got != "line1\nline2=stillvalue\nline3" {
		t.Fatalf("PRIVATE_KEY = %q", got)
	}
}

func valueOf(f *File, key string) string {
	for _, e := range f.Entries {
		if e.Key == key {
			return e.Value
		}
	}
	return ""
}

func TestHas(t *testing.T) {
	f := Parse([]byte("A=1\n"))
	if !f.Has("A") {
		t.Fatal("Has(A) = false, want true")
	}
	if f.Has("B") {
		t.Fatal("Has(B) = true, want false")
	}
}
