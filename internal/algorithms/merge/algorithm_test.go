package merge

import (
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestMergeSortCorrectness(t *testing.T) {
	tests := []struct {
		name  string
		input []int
	}{
		{name: "empty", input: []int{}},
		{name: "single", input: []int{42}},
		{name: "already sorted", input: []int{1, 2, 3, 4, 5}},
		{name: "reverse sorted", input: []int{5, 4, 3, 2, 1}},
		{name: "duplicates", input: []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}},
		{name: "with negatives", input: []int{-5, 10, -2, 0, 7, -9}},
		{name: "default input", input: []int{38, 27, 43, 3, 9, 82, 10}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expected := make([]int, len(tc.input))
			copy(expected, tc.input)
			sort.Ints(expected)

			orig := make([]int, len(tc.input))
			copy(orig, tc.input)

			got := Sort(tc.input)

			if !reflect.DeepEqual(got, expected) {
				t.Fatalf("expected %v, got %v", expected, got)
			}
			// Verify input was not mutated
			if !reflect.DeepEqual(tc.input, orig) {
				t.Fatalf("input was mutated: expected %v, got %v", orig, tc.input)
			}
		})
	}
}

func TestMergeSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "merge" {
		t.Errorf("expected ID 'merge', got %q", meta.ID)
	}
	if meta.Name != "Merge Sort" {
		t.Errorf("expected Name 'Merge Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n log n)" {
		t.Errorf("expected TimeComplexity 'O(n log n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("expected SpaceComplexity 'O(n)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Divide and Conquer", "Stable", "Predictable Performance", "Auxiliary Array"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}

func TestMergeSortRegistration(t *testing.T) {
	l, ok := catalog.Get("merge")
	if !ok {
		t.Fatal("lesson 'merge' not registered in catalog")
	}
	if l.Meta().ID != "merge" {
		t.Errorf("expected ID 'merge', got %q", l.Meta().ID)
	}
}

func TestMergeSortTraceValidation(t *testing.T) {
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

	lastStep := steps[len(steps)-1]
	if lastStep.Metrics.Comparisons == 0 {
		t.Fatalf("expected comparisons > 0, got %d", lastStep.Metrics.Comparisons)
	}
	if lastStep.Metrics.Operations == 0 {
		t.Fatalf("expected operations > 0, got %d", lastStep.Metrics.Operations)
	}
}

func TestMergeSortCustomBuild(t *testing.T) {
	lesson := Lesson{}
	customInput, _ := json.Marshal([]int{5, 2, 8, 1})
	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	last := steps[len(steps)-1].State.(renderers.ArrayState)
	if !slices.Equal(last.Values, []int{1, 2, 5, 8}) {
		t.Fatalf("expected sorted [1, 2, 5, 8], got %v", last.Values)
	}
}

func TestMergeSortInputValidation(t *testing.T) {
	lesson := Lesson{}
	invalidInputs := []json.RawMessage{
		[]byte(`not json`),
		[]byte(`{"bad":1}`),
		[]byte(`[1,"two"]`),
		[]byte(`null`),
		[]byte(`[1.5, 2]`),
		[]byte(`123`),
		[]byte(`[1, null]`),
		[]byte(`[1, 2] junk`),
		[]byte(`   `),
	}

	for _, input := range invalidInputs {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("invalid input %s accepted without error", input)
		}
	}

	// Valid empty array
	steps, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatalf("empty JSON array rejected: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps for empty array")
	}
	finalState := steps[len(steps)-1].State.(renderers.ArrayState)
	if len(finalState.Values) != 0 {
		t.Fatalf("expected empty values, got %v", finalState.Values)
	}
}

func TestMergeTeachingTraceAndStability(t *testing.T) {
	testCases := [][]int{
		nil,
		{},
		{7},
		{2, 1, 2, 1},
		{1, 2, 3},
		{3, 2, 1},
		{0, -1, -1},
		{38, 27, 43, 3, 9, 82, 10},
	}

	for _, input := range testCases {
		steps := Trace(input)
		expected := Sort(input)
		n := len(input)

		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatalf("source line validation failed for input %v: %v", input, err)
		}

		first := steps[0].State.(renderers.ArrayState)
		if !slices.Equal(first.Values, input) || len(first.IDs) != n {
			t.Fatalf("initial snapshot lost input objects: %+v", first)
		}

		// Verify initial IDs are distinct and nonempty
		initialObjects := make(map[string]int)
		for i, id := range first.IDs {
			if id == "" {
				t.Fatalf("empty ID at index %d", i)
			}
			if _, exists := initialObjects[id]; exists {
				t.Fatalf("duplicate initial ID %q", id)
			}
			initialObjects[id] = input[i]
		}

		var sawCompare, sawPick, sawPlace bool

		for k, step := range steps {
			s := step.State.(renderers.ArrayState)

			// Every step must have a descriptive caption
			if s.Caption == "" {
				t.Fatalf("step %d missing caption", k)
			}

			// Objects must retain their values across all steps
			if len(s.IDs) != n {
				t.Fatalf("step %d lost IDs: got %d, expected %d", k, len(s.IDs), n)
			}
			for i, id := range s.IDs {
				if initialObjects[id] != s.Values[i] {
					t.Fatalf("step %d: object %q changed value from %d to %d", k, id, initialObjects[id], s.Values[i])
				}
			}

			// Check SVG rendering
			svg := renderers.RenderArraySVG(step, 800, 400)
			for _, id := range s.IDs {
				if strings.Count(svg, `data-key="array-`+id+`"`) != 1 {
					t.Fatalf("step %d: rendered identity %q missing or repeated in SVG", k, id)
				}
			}

			// Teaching phase checks
			if step.LineID == "L14" && step.Action == trace.ActionCompare {
				if len(s.Comparing) != 2 {
					t.Fatalf("step %d: compare step on L14 must have 2 comparing indices: %+v", k, s.Comparing)
				}
				sawCompare = true
			}
			if (step.LineID == "L15" || step.LineID == "L17") && step.Action == trace.ActionSet {
				if s.Current < 0 || !slices.Contains(s.Swapped, s.Current) {
					t.Fatalf("step %d: pick step must highlight chosen element in Current/Swapped: %+v", k, s)
				}
				sawPick = true
			}
			if step.LineID == "L23" && step.Action == trace.ActionSet {
				if s.Current < 0 || !slices.Contains(s.Swapped, s.Current) {
					t.Fatalf("step %d: placement step must highlight placed element in Current/Swapped: %+v", k, s)
				}
				sawPlace = true
			}
		}

		if n >= 2 {
			if !sawCompare || !sawPick || !sawPlace {
				t.Fatalf("input %v missing teaching states: compare=%v pick=%v place=%v", input, sawCompare, sawPick, sawPlace)
			}
		}

		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !slices.Equal(last.Values, expected) {
			t.Fatalf("input %v: final values %v != expected %v", input, last.Values, expected)
		}
		if n > 0 && len(last.Sorted) != n {
			t.Fatalf("input %v: final step must have all %d items marked sorted, got %v", input, n, last.Sorted)
		}

		// Stability verification: for duplicate values, relative ID order must be strictly preserved
		for _, val := range input {
			var beforeIDs, afterIDs []string
			for i, v := range first.Values {
				if v == val {
					beforeIDs = append(beforeIDs, first.IDs[i])
				}
			}
			for i, v := range last.Values {
				if v == val {
					afterIDs = append(afterIDs, last.IDs[i])
				}
			}
			if !slices.Equal(beforeIDs, afterIDs) {
				t.Fatalf("input %v: stability violated for value %d: initial %v -> final %v", input, val, beforeIDs, afterIDs)
			}
		}

		// Immutability check: mutating final snapshot must not alter earlier snapshots
		if n > 0 {
			last.Values[0] = 999
			last.IDs[0] = "corrupted"
			last.Sorted[0] = 999
			if !slices.Equal(first.Values, input) || first.IDs[0] == "corrupted" {
				t.Fatalf("history aliased final snapshot for input %v", input)
			}
		}
	}
}
