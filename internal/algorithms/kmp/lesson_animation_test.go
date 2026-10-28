package kmp

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestKMPLessonLPSConstruction(t *testing.T) {
	for _, p := range []string{"ABABCABAB", "AAAA", "ABABAC"} {
		steps := Trace(p+p, p)
		count := 0
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		for _, step := range steps {
			s, ok := step.State.(renderers.KMPLessonState)
			if !ok {
				t.Fatalf("untyped %T", step.State)
			}
			if s.Phase == "lps" {
				count++
				if step.LineID != "L3" {
					t.Fatal("helper source")
				}
			}
			for j, v := range s.LPS {
				if j > s.LPSKnownThrough {
					if v != -1 {
						t.Fatal("unknown prefilled")
					}
					continue
				}
				want := 0
				for k := 1; k <= j; k++ {
					if p[:k] == p[j-k+1:j+1] {
						want = k
					}
				}
				if v != want {
					t.Fatalf("LPS[%d]=%d want %d", j, v, want)
				}
			}
		}
		last := steps[len(steps)-1].State.(renderers.KMPLessonState)
		if count < len(p) || !reflect.DeepEqual(last.LPS, ComputeLPS(p)) {
			t.Fatal("missing construction", last)
		}
		first := steps[0].State.(renderers.KMPLessonState)
		saved := append([]int(nil), first.LPS...)
		last.LPS[0] = 99
		if !reflect.DeepEqual(first.LPS, saved) {
			t.Fatal("snapshot alias")
		}
	}
}

func TestKMPLessonFallback(t *testing.T) {
	steps := Trace("ABABDABACDABABCABAB", "ABABCABAB")
	count := 0
	paired := false
	for k, step := range steps {
		if step.LineID == "L13" {
			s := step.State.(renderers.KMPLessonState)
			p := steps[k-1].State.(renderers.KMPLessonState)
			if s.I != p.I || s.J != p.LPS[p.J-1] || s.Shift != s.I-s.J || s.FallbackFrom != p.J || s.FallbackTo != s.J {
				t.Fatal(s)
			}
			count++
		}
		switch step.HoldMS {
		case 0, 350, 600, 1200, 1800:
		default:
			t.Fatal("invalid hold")
		}
		if k > 0 && step.HoldMS == 1200 && steps[k-1].HoldMS == 600 {
			paired = true
		}
	}
	last := steps[len(steps)-1].State.(renderers.KMPLessonState)
	if count < 2 || !paired || !last.ResultKnown || last.MatchIndex != 10 {
		t.Fatal(last)
	}
}

func TestKMPLessonEdgeBytes(t *testing.T) {
	for _, v := range [][2]string{{"AAAAA", "AAA"}, {"abc", ""}, {"", "a"}, {"a", "abc"}, {"abc", "z"}, {"é猫é", "猫"}, {"x<&\"y", "<&\""}} {
		steps := Trace(v[0], v[1])
		last := steps[len(steps)-1].State.(renderers.KMPLessonState)
		if !last.ResultKnown || last.MatchIndex != Search(v[0], v[1]) {
			t.Fatal(v, last)
		}
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
	}
}

func TestKMPLessonValidation(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `{"text":2}`, `{"pattern":null}`, `[]`, `["a",2]`, `{"text":"` + strings.Repeat("a", 257) + `"}`, `{"pattern":"` + strings.Repeat("b", 65) + `"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"text":""}`, `["abc",""]`, `{"text":"` + strings.Repeat("a", 256) + `","pattern":"` + strings.Repeat("a", 64) + `"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Fatal(err)
		}
	}
}
