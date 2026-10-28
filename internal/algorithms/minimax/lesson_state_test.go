package minimax

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestMinimaxLessonReturnAndCutoff(t *testing.T) {
	steps := Trace([]int{3, 5, 2, 9}, 2)
	last := steps[len(steps)-1].State.(renderers.MinimaxLessonState)
	value, child := Solve(BuildTree([]int{3, 5, 2, 9}, 2))
	if last.BestValue != value || last.BestChildID != child.ID || value != 3 {
		t.Fatalf("bad result %+v", last)
	}
	found := false
	for _, step := range steps {
		s := step.State.(renderers.MinimaxLessonState)
		if step.LineID == "L3" && s.ActiveID == "n0-2-2" {
			t.Fatal("visited pruned leaf")
		}
		if step.LineID == "L20" && s.Cutoff {
			found = s.Alpha == 3 && s.Beta == 2 && reflect.DeepEqual(s.PrunedIDs, []string{"n0-2-2"})
		}
	}
	if !found {
		t.Fatal("missing exact cutoff")
	}
}

func TestMinimaxLessonRenderedSnapshots(t *testing.T) {
	for _, scores := range [][]int{nil, {7}, {3, 5, 2}, {3, 5, 2, 9}, make([]int, 64)} {
		steps := Trace(scores, 2)
		if len(scores) == 0 && !strings.Contains(strings.ToLower(steps[len(steps)-1].Explanation), "empty input") {
			t.Fatal("missing empty input label")
		}
		for _, step := range steps {
			svg := renderers.RenderMinimaxLessonSVG(step, 320, 640)
			if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
				t.Fatal("nonfinite SVG")
			}
		}
	}
}

func TestMinimaxLessonDefaultMove(t *testing.T) {
	for _, tc := range []struct {
		s    []int
		v    int
		move string
	}{{[]int{-1, 4, 2, 6, -3, -5, 0, 7}, 4, "n0-1"}, {[]int{1, 1, 1, 1}, 1, "n0-1"}, {[]int{7}, 7, ""}, {[]int{3, 5, 2}, 3, "n0-1"}} {
		steps := Trace(tc.s, 2)
		s := steps[len(steps)-1].State.(renderers.MinimaxLessonState)
		if !s.BestValueValid || s.BestValue != tc.v || s.BestChildID != tc.move {
			t.Fatalf("bad result %+v", s)
		}
	}
}

func TestMinimaxLessonValidation(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `[]`, `[null]`, `{"scores":null}`, `{"scores":[]}`, `{"scores":[1000000]}`, `{"scores":[-1000000]}`, `{"branching":1}`, `{"branching":9}`, `{"branching":null}`, `{"scores":"bad"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"branching":8}`, `[-999999,999999]`, `{"scores":[7],"branching":2}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Errorf("rejected %s: %v", raw, err)
		}
	}
	for _, n := range []int{64, 65} {
		raw, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 64) {
			t.Fatalf("score count %d: %v", n, err)
		}
	}
	for _, scores := range [][]int{{1000000}, make([]int, 65)} {
		steps := Trace(scores, 2)
		if len(steps) != 1 || steps[0].State.(renderers.MinimaxLessonState).Phase != "invalid-input" {
			t.Fatal("invalid direct trace")
		}
	}
}

func TestMinimaxLessonSnapshots(t *testing.T) {
	a, b := Trace([]int{3, 5, 2, 9}, 2), Trace([]int{3, 5, 2, 9}, 2)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("nondeterministic")
	}
	if err := catalog.ValidateSourceLineIDs(Lesson{}, a); err != nil {
		t.Fatal(err)
	}
	early := a[0].State.(renderers.MinimaxLessonState)
	last := a[len(a)-1].State.(renderers.MinimaxLessonState)
	last.Nodes[0].Value = 999
	last.TreeState.Nodes[0].Val = 999
	last.Stack[0] = "changed"
	last.PrunedIDs[0] = "changed"
	if early.Nodes[0].Value == 999 || early.TreeState.Nodes[0].Val == 999 || early.Stack[0] == "changed" || len(early.PrunedIDs) != 0 || early.Nodes[0].Evaluated {
		t.Fatal("snapshot alias/leak")
	}
	pair := false
	for i, s := range a {
		if s.HoldMS != 0 && s.HoldMS != 350 && s.HoldMS != 600 && s.HoldMS != 1200 && s.HoldMS != 1800 {
			t.Fatal("invalid hold")
		}
		if i+1 < len(a) && s.LineID == "L2" && a[i+1].LineID == "L3" {
			pair = s.HoldMS+a[i+1].HoldMS == 1800
		}
	}
	if !pair {
		t.Fatal("no compare/return budget")
	}
}
