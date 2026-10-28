package knapsack

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestKnapsackBuildRejectsInvalidInput(t *testing.T) {
	for _, raw := range []string{`{`, `null`, `[]`, `{"weights":[-1],"values":[1],"capacity":1}`, `{"weights":[1],"values":[],"capacity":1}`, `{"weights":[1],"values":[1],"capacity":-1}`, `{"weights":[201],"values":[1],"capacity":1}`, `{"weights":[1],"values":[1000001],"capacity":1}`, `{"weights":[1],"values":[-1],"capacity":1}`, `{"weights":[1],"values":[1],"capacity":201}`, `{"weights":null}`, `{"capacity":null}`, `{"weights":"bad"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("Build(%s) accepted invalid input", raw)
		}
	}
}

func TestKnapsackAnimationSnapshots(t *testing.T) {
	for _, tc := range []struct {
		w, v []int
		c    int
	}{
		{[]int{2, 3, 4, 5}, []int{3, 4, 5, 6}, 5}, {[]int{2}, []int{3}, 4},
		{[]int{0}, []int{3}, 0}, {[]int{1}, []int{-2}, -1},
		{[]int{1, 2}, []int{4}, 2}, {nil, nil, 0},
	} {
		steps := Trace(tc.w, tc.v, tc.c)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		want, table := Solve(tc.w, tc.v, tc.c)
		last, ok := steps[len(steps)-1].State.(renderers.KnapsackLessonState)
		if !ok || !last.HasResult || last.Result != want || !reflect.DeepEqual(last.DP, table) {
			t.Fatalf("wrong final snapshot: %#v", last)
		}
		for i, step := range steps {
			s := step.State.(renderers.KnapsackLessonState)
			if step.LineID == "L8" {
				if step.HoldMS != 600 || steps[i+1].HoldMS != 1200 {
					t.Fatal("compare/write budget")
				}
				if s.Sources[0] != [2]int{s.Item - 1, s.Column} {
					t.Fatal("exclude must use previous row")
				}
				if s.Fits && s.Sources[1] != [2]int{s.Item - 1, s.Column - s.Weights[s.Item-1]} {
					t.Fatal("include must use previous row")
				}
				if s.Known[s.Item][s.Column] {
					t.Fatal("candidate marked written")
				}
			}
			if step.LineID == "L9" || step.LineID == "L11" {
				if !s.Known[s.Item][s.Column] || s.Chosen != s.DP[s.Item][s.Column] || s.Chosen != max(s.Exclude, s.Include) {
					t.Fatal("incorrect commit")
				}
			}
			if step.LineID == "L5" && s.Item > 1 && !s.Known[s.Item-1][s.Capacity] {
				t.Fatal("previous row hidden")
			}
		}
		first := steps[0].State.(renderers.KnapsackLessonState)
		last.DP[0][0] = 99
		last.Known[0][0] = false
		last.RowHeaders[0] = "changed"
		last.Cells[0].Value = "changed"
		if first.DP[0][0] != 0 || !first.Known[0][0] || first.RowHeaders[0] == "changed" || first.Cells[0].Value == "changed" {
			t.Fatal("aliased snapshot")
		}
	}
}

func TestKnapsackBuildBoundaries(t *testing.T) {
	for _, raw := range []string{``, `{}`, `{"weights":[0],"values":[0],"capacity":0}`, `{"weights":[200],"values":[1000000],"capacity":200}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
	}
	for _, n := range []int{0, 20, 21} {
		w, v := make([]int, n), make([]int, n)
		raw, _ := json.Marshal(map[string]any{"weights": w, "values": v, "capacity": 0})
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 20) {
			t.Fatalf("item count %d: %v", n, err)
		}
	}
}
