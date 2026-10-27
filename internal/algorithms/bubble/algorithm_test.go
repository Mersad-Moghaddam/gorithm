package bubble

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBubbleSortCorrectness(t *testing.T) {
	input := []int{64, 34, 25, 12, 22, 11, 90}
	expected := []int{11, 12, 22, 25, 34, 64, 90}
	got := Sort(input)

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestTeachingTrace(t *testing.T) {
	for _, input := range [][]int{nil, {7}, {2, 1, 2, 1}, {1, 2, 3}, {3, 2, 1}, {0, -1, -1}} {
		steps := Trace(input)
		expected := Sort(input)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		first := steps[0].State.(renderers.ArrayState)
		if !slices.Equal(first.Values, input) || len(first.IDs) != len(input) {
			t.Fatalf("initial objects lost: %+v", first)
		}
		objects := map[string]int{}
		for i, id := range first.IDs {
			if _, exists := objects[id]; exists || id == "" {
				t.Fatalf("nonunique identity %q", id)
			}
			objects[id] = input[i]
		}
		for k, step := range steps {
			s := step.State.(renderers.ArrayState)
			if s.Caption == "" || s.Current != -1 {
				t.Fatalf("missing caption or unrelated current highlight: %+v", s)
			}
			for i, id := range s.IDs {
				if objects[id] != s.Values[i] {
					t.Fatalf("object changed value: %+v", s)
				}
			}
			for i, index := range s.Sorted {
				if index != len(input)-len(s.Sorted)+i || index < 0 || index >= len(input) {
					t.Fatalf("invalid settled suffix: %v", s.Sorted)
				}
				if s.Values[index] != expected[index] {
					t.Fatal("unsettled value marked final")
				}
			}
			if len(s.IDs) != len(input) {
				t.Fatal("snapshot lost identities")
			}
			svg := renderers.RenderArraySVG(step, 800, 400)
			for _, id := range s.IDs {
				if strings.Count(svg, `data-key="array-`+id+`"`) != 1 {
					t.Fatal("rendered identity missing or repeated")
				}
			}
			if step.Action == trace.ActionCompare {
				if len(s.Comparing) != 2 || s.Comparing[1] != s.Comparing[0]+1 {
					t.Fatalf("comparison not adjacent: %+v", s)
				}
				j := s.Comparing[0]
				next := steps[k+1].State.(renderers.ArrayState)
				wantIDs := slices.Clone(s.IDs)
				if s.Values[j] > s.Values[j+1] {
					wantIDs[j], wantIDs[j+1] = wantIDs[j+1], wantIDs[j]
					if steps[k+1].Action != trace.ActionSwap {
						t.Fatal("missing swap decision")
					}
				} else if !strings.Contains(next.Caption, "Keep") {
					t.Fatal("missing keep-order decision")
				}
				if !slices.Equal(next.IDs, wantIDs) {
					t.Fatalf("wrong object movement: %v -> %v", s.IDs, next.IDs)
				}
			}
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !slices.Equal(last.Values, Sort(input)) || len(last.Sorted) != len(input) {
			t.Fatalf("untruthful completion: %+v", last)
		}
		for _, value := range input {
			var before, after []string
			for i, v := range first.Values {
				if v == value {
					before = append(before, first.IDs[i])
				}
			}
			for i, v := range last.Values {
				if v == value {
					after = append(after, last.IDs[i])
				}
			}
			if !slices.Equal(before, after) {
				t.Fatal("equal objects reordered")
			}
		}
		if len(input) > 0 {
			last.Values[0], last.IDs[0], last.Sorted[0] = 999, "changed", 999
			if !slices.Equal(first.Values, input) || first.IDs[0] == "changed" {
				t.Fatal("history aliases final snapshot")
			}
		}
	}
}

func TestBubbleSortTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}
