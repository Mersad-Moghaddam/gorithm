package heap

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestHeapSortCorrectness(t *testing.T) {
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
		{name: "default input", input: []int{12, 11, 13, 5, 6, 7}},
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
			if !reflect.DeepEqual(tc.input, orig) {
				t.Fatalf("input was mutated: expected %v, got %v", orig, tc.input)
			}
		})
	}
}

func TestHeapSortTraceValidation(t *testing.T) {
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
	if lastStep.Metrics.Swaps == 0 {
		t.Fatalf("expected swaps > 0, got %d", lastStep.Metrics.Swaps)
	}
}

func TestHeapSortCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput, _ := json.Marshal([]int{9, 4, 7, 1})
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
}

func TestHeapSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "heap" {
		t.Errorf("expected ID 'heap', got %q", meta.ID)
	}
	if meta.Name != "Heap Sort" {
		t.Errorf("expected Name 'Heap Sort', got %q", meta.Name)
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
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}
	expectedConcepts := []string{"Binary Heap", "Max-Heapify", "In-Place", "Priority Queue Duality"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}

func TestHeapSortRegistration(t *testing.T) {
	l, ok := catalog.Get("heap")
	if !ok {
		t.Fatal("lesson 'heap' not registered in catalog")
	}
	if l.Meta().ID != "heap" {
		t.Errorf("expected ID 'heap', got %q", l.Meta().ID)
	}
}

func TestHeapSortBuildValidation(t *testing.T) {
	lesson := Lesson{}
	invalidInputs := []json.RawMessage{
		[]byte(`not json`),
		[]byte(`{"bad":1}`),
		[]byte(`[1,"two"]`),
		[]byte(`null`),
		[]byte(`[1.5, 2]`),
		[]byte(`[1.0]`),
		[]byte(`[1, 2] trailing`),
		[]byte(`   `),
	}
	for _, input := range invalidInputs {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("expected error for invalid input %q, got nil", string(input))
		}
	}

	validInputs := []json.RawMessage{
		nil,
		[]byte(``),
		[]byte(`[]`),
		[]byte(`[5, 1, 4]`),
	}
	for _, input := range validInputs {
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("unexpected error for valid input %q: %v", string(input), err)
		}
		if len(steps) == 0 {
			t.Fatalf("expected steps for input %q", string(input))
		}
	}

	emptySteps, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	lastState := emptySteps[len(emptySteps)-1].State.(renderers.ArrayState)
	if len(lastState.Values) != 0 {
		t.Fatalf("empty input should produce 0 values, got %d", len(lastState.Values))
	}
}

func TestHeapSortTeachingDecisionsAndIdentity(t *testing.T) {
	lesson := Lesson{}
	input := []int{1, 3, 2}
	steps := Trace(input)

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatal(err)
	}

	var (
		foundParentChildrenComp bool
		foundSiftDownSwap       bool
		foundExtractionSwap     bool
		foundSettledStep        bool
	)

	for _, step := range steps {
		state := step.State.(renderers.ArrayState)

		// Verify unique non-empty IDs across all steps
		seenIDs := make(map[string]bool)
		for _, id := range state.IDs {
			if id == "" || seenIDs[id] {
				t.Fatalf("non-unique or empty ID %q in step %+v", id, state.IDs)
			}
			seenIDs[id] = true
		}
		if len(seenIDs) != len(input) {
			t.Fatalf("expected %d IDs, got %d in state %+v", len(input), len(seenIDs), state.IDs)
		}

		// 1. Max-heapify phase: parent and children in Comparing at L12
		if step.LineID == "L12" && len(state.Comparing) == 3 && reflect.DeepEqual(state.Comparing, []int{0, 1, 2}) {
			foundParentChildrenComp = true
		}

		// 1. Sift-down swap at L16
		if step.LineID == "L16" && step.Action == trace.ActionSwap && reflect.DeepEqual(state.Swapped, []int{0, 1}) {
			foundSiftDownSwap = true
		}

		// 2. Extraction phase swap: root 0 with last element 2 at L7 before settling
		if step.LineID == "L7" && step.Action == trace.ActionSwap && reflect.DeepEqual(state.Swapped, []int{0, 2}) {
			for _, s := range state.Sorted {
				if s == 2 {
					t.Fatalf("index 2 should not be in sorted slice during extraction swap step")
				}
			}
			foundExtractionSwap = true
		}

		// 2. Settled element joining sorted suffix at L7 (in green)
		if step.LineID == "L7" && step.Action == trace.ActionSet && len(state.Sorted) > 0 && state.Sorted[0] == 2 {
			foundSettledStep = true
		}
	}

	if !foundParentChildrenComp {
		t.Error("missing parent-children Comparing step at L12")
	}
	if !foundSiftDownSwap {
		t.Error("missing sift-down swap step at L16")
	}
	if !foundExtractionSwap {
		t.Error("missing extraction swap step at L7")
	}
	if !foundSettledStep {
		t.Error("missing settled element joining sorted suffix step at L7")
	}

	first := steps[0].State.(renderers.ArrayState)
	last := steps[len(steps)-1].State.(renderers.ArrayState)

	if !reflect.DeepEqual(first.Values, []int{1, 3, 2}) || !reflect.DeepEqual(first.IDs, []string{"heap-0", "heap-1", "heap-2"}) {
		t.Fatalf("unexpected first step state: %+v", first)
	}
	if !reflect.DeepEqual(last.Values, []int{1, 2, 3}) || !reflect.DeepEqual(last.Sorted, []int{0, 1, 2}) {
		t.Fatalf("unexpected final step state: %+v", last)
	}

	// Snapshot immutability check
	var intermediate renderers.ArrayState
	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		if step.LineID == "L7" && step.Action == trace.ActionSet && reflect.DeepEqual(state.Sorted, []int{2}) {
			intermediate = state
			break
		}
	}
	if len(intermediate.Values) == 0 {
		t.Fatal("missing intermediate snapshot")
	}

	last.Values[0] = 999
	last.IDs[0] = "corrupted"
	last.Sorted[0] = 999

	if first.Values[0] == 999 || first.IDs[0] == "corrupted" {
		t.Fatal("mutating last snapshot mutated first snapshot")
	}
	if intermediate.Values[0] == 999 || intermediate.IDs[0] == "corrupted" || intermediate.Sorted[0] == 999 {
		t.Fatal("mutating last snapshot mutated intermediate snapshot")
	}
}

func TestHeapSortDuplicateStability(t *testing.T) {
	input := []int{5, 2, 5, 2}
	steps := Trace(input)

	initialIDs := []string{"heap-0", "heap-1", "heap-2", "heap-3"}
	first := steps[0].State.(renderers.ArrayState)
	if !reflect.DeepEqual(first.IDs, initialIDs) {
		t.Fatalf("expected initial IDs %v, got %v", initialIDs, first.IDs)
	}

	for i, step := range steps {
		state := step.State.(renderers.ArrayState)
		seen := make(map[string]bool)
		for _, id := range state.IDs {
			if seen[id] {
				t.Fatalf("step %d has duplicate ID %q", i, id)
			}
			seen[id] = true
		}
		if len(seen) != len(input) {
			t.Fatalf("step %d has missing IDs: %v", i, state.IDs)
		}
	}

	last := steps[len(steps)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(last.Values, []int{2, 2, 5, 5}) {
		t.Fatalf("expected sorted values [2, 2, 5, 5], got %v", last.Values)
	}
}

func TestHeapSortEdgeCases(t *testing.T) {
	lesson := Lesson{}
	testCases := []struct {
		name     string
		input    []int
		expected []int
	}{
		{name: "empty", input: []int{}, expected: []int{}},
		{name: "single", input: []int{42}, expected: []int{42}},
		{name: "all equal", input: []int{7, 7, 7, 7}, expected: []int{7, 7, 7, 7}},
		{name: "reverse", input: []int{5, 4, 3, 2, 1}, expected: []int{1, 2, 3, 4, 5}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			steps := Trace(tc.input)
			if len(steps) == 0 {
				t.Fatalf("expected steps for %s", tc.name)
			}
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("validation failed for %s: %v", tc.name, err)
			}
			last := steps[len(steps)-1].State.(renderers.ArrayState)
			if !reflect.DeepEqual(last.Values, tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, last.Values)
			}
			if len(last.IDs) != len(tc.input) {
				t.Fatalf("expected %d IDs, got %d", len(tc.input), len(last.IDs))
			}
			if len(last.Sorted) != len(tc.input) {
				t.Fatalf("expected %d sorted indices, got %d", len(tc.input), len(last.Sorted))
			}
		})
	}
}
