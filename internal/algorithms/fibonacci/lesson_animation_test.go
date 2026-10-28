package fibonacci

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"testing"
)

func TestFibonacciInputBounds(t *testing.T) {
	for _, input := range []string{`-1`, `47`, `null`, `{"n":null}`, `{"n":"5"}`, `[]`, `{`, `1.5`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	for _, input := range []string{`0`, `46`, `{"n":0}`, `{"n":46}`, `{}`, ``} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err != nil {
			t.Errorf("rejected %s: %v", input, err)
		}
	}
}

func TestFibonacciAnimationSnapshots(t *testing.T) {
	for _, n := range []int{-1, 0, 1, 2, 5, 8, 46} {
		steps := Trace(n)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		want, table := Compute(n)
		final, ok := steps[len(steps)-1].State.(renderers.FibonacciLessonState)
		if !ok {
			t.Fatalf("missing typed snapshot: %T", steps[len(steps)-1].State)
		}
		if !final.HasResult || final.Result != want || !reflect.DeepEqual(final.Values, table) {
			t.Fatalf("n=%d final=%+v want=%d %v", n, final, want, table)
		}
		for i, step := range steps {
			s := step.State.(renderers.FibonacciLessonState)
			original := renderers.RenderFibonacciLessonSVG(step, 320, 450)
			for j := len(steps) - 1; j >= 0; j-- {
				renderers.RenderFibonacciLessonSVG(steps[j], 320, 450)
			}
			if original != renderers.RenderFibonacciLessonSVG(step, 320, 450) {
				t.Fatal("render changes after reverse seek")
			}
			if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
				t.Fatal("invalid hold")
			}
			if step.LineID == "L8" {
				if n == 1 || s.Known[s.Index] || s.Sources != [2]int{s.Index - 2, s.Index - 1} {
					t.Fatalf("future value exposed: %+v", s)
				}
				for _, source := range s.Sources {
					if !s.Known[source] {
						t.Fatal("unknown source")
					}
				}
				next := steps[i+1].State.(renderers.FibonacciLessonState)
				if steps[i+1].LineID != "L9" || step.HoldMS+steps[i+1].HoldMS != 1800 || next.Values[s.Index] != s.Values[s.Index-2]+s.Values[s.Index-1] {
					t.Fatal("invalid commit")
				}
				for j := range s.Values {
					if j != s.Index && (s.Values[j] != next.Values[j] || s.Known[j] != next.Known[j]) {
						t.Fatal("unrelated cell changed")
					}
				}
			}
		}
		if n > 0 && (len(steps) != 4+2*(n-1) || steps[len(steps)-1].Metrics.Operations != 2*n) {
			t.Fatal("event order/count or metrics changed")
		}
		if len(steps) > 2 {
			final.Values[0] = 999
			final.Known[0] = false
			final.Cells[0].Value = "mutated"
			final.ColHeaders[0] = "mutated"
			final.RowHeaders[0] = "mutated"
			after := steps[1].State.(renderers.FibonacciLessonState)
			if after.Values[0] != 0 || !after.Known[0] || after.Cells[0].Value != "0" || after.ColHeaders[0] != "0" || after.RowHeaders[0] != "dp" {
				t.Fatal("snapshots alias")
			}
		}
	}
}
