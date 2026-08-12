package drift

import (
	"testing"

	"github.com/Rarex224/envy-cli/internal/dotenv"
)

func TestDiff(t *testing.T) {
	schema := dotenv.Parse([]byte("A=\nB=\nC=\n"))
	actual := dotenv.Parse([]byte("A=1\nC=3\nZ=9\n"))
	got := Diff(schema, actual)
	want := []Change{
		{"A", Ok},
		{"B", Missing},
		{"C", Ok},
		{"Z", Extra},
	}
	if len(got) != len(want) {
		t.Fatalf("Diff len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Diff[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestSummarize(t *testing.T) {
	changes := []Change{{"A", Ok}, {"B", Missing}, {"C", Missing}, {"Z", Extra}}
	missing, extra := Summarize(changes)
	if missing != 2 || extra != 1 {
		t.Fatalf("Summarize = (%d,%d), want (2,1)", missing, extra)
	}
}

func TestKindString(t *testing.T) {
	if Missing.String() != "missing" || Extra.String() != "extra" || Ok.String() != "ok" {
		t.Fatalf("Kind.String mismatch: %s %s %s", Ok, Missing, Extra)
	}
}
