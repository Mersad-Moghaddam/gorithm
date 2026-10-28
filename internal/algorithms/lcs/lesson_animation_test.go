package lcs

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestLCSAnimation(t *testing.T) {
	for _, pair := range [][2]string{{"ABC", "AC"}, {"AGGTAB", "GXTXAYB"}, {"AB", "BA"}, {"", ""}, {"", "A"}, {"A", ""}, {"ABC", "DEF"}, {"AAA", "AA"}, {"ABC", "ABC"}, {"é", "ê"}, {"<>&\"", "<&\""}} {
		steps := Trace(pair[0], pair[1])
		want, dp := Solve(pair[0], pair[1])
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		last, ok := steps[len(steps)-1].State.(renderers.LCSLessonState)
		if !ok {
			t.Fatalf("typed LCS snapshot missing: %T", steps[0].State)
		}
		if last.Sequence != want || !reflect.DeepEqual(last.DP, dp) {
			t.Fatalf("oracle mismatch %q: %q", pair, last.Sequence)
		}
		prev := ""
		for i, step := range steps {
			s := step.State.(renderers.LCSLessonState)
			out := renderers.RenderLCSLessonSVG(step, 320, 450)
			if !strings.Contains(out, `data-phase="`+s.Phase+`"`) {
				t.Fatal("real trace render phase")
			}
			if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
				t.Fatal("hold", step.HoldMS)
			}
			if step.LineID == "L7" || step.LineID == "L16" {
				if step.HoldMS+steps[i+1].HoldMS != 1800 {
					t.Fatal("pair budget")
				}
			}
			if step.LineID == "L17" {
				if s.To != [2]int{s.From[0] - 1, s.From[1] - 1} || s.Sequence != pair[0][s.From[0]-1:s.From[0]]+prev {
					t.Fatal("prepend/diagonal", s)
				}
				prev = s.Sequence
			}
			if step.LineID == "L19" && s.To != [2]int{s.From[0] - 1, s.From[1]} {
				t.Fatal("up")
			}
			if step.LineID == "L19" && s.DP[s.From[0]-1][s.From[1]] < s.DP[s.From[0]][s.From[1]-1] {
				t.Fatal("up tie rule")
			}
			if step.LineID == "L21" && s.To != [2]int{s.From[0], s.From[1] - 1} {
				t.Fatal("left")
			}
		}
		for i, p := range last.Matches {
			if pair[0][p[0]] != pair[1][p[1]] || (i > 0 && (last.Matches[i-1][0] >= p[0] || last.Matches[i-1][1] >= p[1])) {
				t.Fatal("match identity")
			}
		}
		first := steps[0].State.(renderers.LCSLessonState)
		before := renderers.RenderLCSLessonSVG(steps[0], 320, 450)
		last.DP[0][0] = 999
		last.Filled[0][0] = false
		last.Cells[0].Value = "changed"
		if first.DP[0][0] != 0 || !first.Filled[0][0] || first.Cells[0].Value != "0" {
			t.Fatal("aliased snapshot")
		}
		if before != renderers.RenderLCSLessonSVG(steps[0], 320, 450) {
			t.Fatal("reverse render mutated")
		}
		if len(pair[0]) > 0 && len(pair[1]) > 0 && first.Filled[1][1] {
			t.Fatal("uncomputed cell known")
		}
	}
}

func TestLCSBuildBounds(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `[]`, `["a"]`, `{"s1":3}`, `{"s1":null}`, `{"pattern":false}`, `42`, `"ABC"`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, n := range []int{0, 32, 33} {
		for side := 0; side < 2; side++ {
			pair := []string{"", ""}
			pair[side] = strings.Repeat("a", n)
			raw, _ := json.Marshal(pair)
			_, err := (Lesson{}).Build(raw)
			if (err != nil) != (n > 32) {
				t.Fatalf("length %d: %v", n, err)
			}
		}
	}
	for _, raw := range []string{`{}`, `{"s1":"","s2":""}`, `{"text":"ABC","pattern":"AC"}`, `["",""]`, `{"s1":"é","s2":"ê"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Fatal(raw, err)
		}
	}
}

func TestLCSBuildExplicitEmptyStrings(t *testing.T) {
	for _, raw := range []string{`{"s1":"","s2":""}`, `{"text":"","pattern":""}`} {
		steps, err := (Lesson{}).Build(json.RawMessage(raw))
		if err != nil {
			t.Fatal(err)
		}
		state := steps[0].State.(renderers.LCSLessonState)
		if state.S1 != "" || state.S2 != "" {
			t.Fatalf("%s: got %q, %q", raw, state.S1, state.S2)
		}
	}
}
