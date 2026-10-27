package countingsort

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "default input",
			input:    []int{4, 2, 2, 8, 3, 3, 1},
			expected: []int{1, 2, 2, 3, 3, 4, 8},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "reverse sorted",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "with duplicates",
			input:    []int{3, 1, 2, 3, 1},
			expected: []int{1, 1, 2, 3, 3},
		},
		{
			name:     "with zeros",
			input:    []int{0, 5, 0, 3, 2, 1},
			expected: []int{0, 0, 1, 2, 3, 5},
		},
		{
			name:     "all identical",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Sort(tc.input)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("Sort(%v) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestSortStability(t *testing.T) {
	// Stable sort invariant: for equal keys, earlier elements in input appear earlier in output.
	arr := []int{4, 2, 2, 8, 3, 3, 1}
	steps := Trace(arr)

	// Collect mapping of input index -> output index for step L21 (output placement)
	targetIndices := make(map[int]int)
	for _, step := range steps {
		if step.LineID == "L21" {
			inIdx := step.Variables["i"].(int)
			outIdx := step.Variables["targetIndex"].(int)
			targetIndices[inIdx] = outIdx
		}
	}

	// Verify for duplicate value 2 at input indices 1 and 2:
	// input index 1 must map to lower output index than input index 2.
	if out1, out2 := targetIndices[1], targetIndices[2]; out1 >= out2 {
		t.Fatalf("stability violated for value 2: inIdx 1 -> %d, inIdx 2 -> %d (expected out1 < out2)", out1, out2)
	}

	// Verify for duplicate value 3 at input indices 4 and 5:
	// input index 4 must map to lower output index than input index 5.
	if out4, out5 := targetIndices[4], targetIndices[5]; out4 >= out5 {
		t.Fatalf("stability violated for value 3: inIdx 4 -> %d, inIdx 5 -> %d (expected out4 < out5)", out4, out5)
	}
}

func TestCountingSortLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "countingsort" {
		t.Fatalf("expected ID %q, got %q", "countingsort", meta.ID)
	}
	if meta.Name != "Counting Sort" {
		t.Fatalf("expected Name %q, got %q", "Counting Sort", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Fatalf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Fatalf("expected Renderer %q, got %q", "array", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n + k)" {
		t.Fatalf("expected TimeComplexity %q, got %q", "O(n + k)", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(k)" {
		t.Fatalf("expected SpaceComplexity %q, got %q", "O(k)", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Easy" {
		t.Fatalf("expected Difficulty %q, got %q", "Easy", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Non-Comparison Sort",
		"Frequency Tally",
		"Prefix Sum Accumulation",
		"Stable Output Mapping",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Fatalf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.HowItWorks == "" {
		t.Fatal("expected non-empty HowItWorks")
	}
}

func TestCountingSortCatalogRegistration(t *testing.T) {
	lesson, ok := catalog.Get("countingsort")
	if !ok {
		t.Fatal("countingsort not found in catalog")
	}
	if lesson.Meta().ID != "countingsort" {
		t.Fatalf("expected ID countingsort, got %s", lesson.Meta().ID)
	}
}

func TestCountingSortTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}

	// Verify count frequencies in variables
	hasCount := false
	for _, step := range steps {
		if _, ok := step.Variables["count"]; ok {
			hasCount = true
			break
		}
	}
	if !hasCount {
		t.Fatal("expected at least one step with count in variables")
	}

	// Verify final step state values are sorted
	finalStep := steps[len(steps)-1]
	state, ok := finalStep.State.(renderers.ArrayState)
	if !ok {
		t.Fatalf("expected final state to be ArrayState, got %T", finalStep.State)
	}
	if !sort.IntsAreSorted(state.Values) {
		t.Fatalf("final state values are not sorted: %v", state.Values)
	}
	if !reflect.DeepEqual(state.Values, []int{1, 2, 2, 3, 3, 4, 8}) {
		t.Fatalf("expected sorted [1 2 2 3 3 4 8], got %v", state.Values)
	}

	// Test edge cases
	for _, edgeArr := range [][]int{{}, {42}} {
		edgeSteps := Trace(edgeArr)
		if err := catalog.ValidateSourceLineIDs(lesson, edgeSteps); err != nil {
			t.Fatalf("edge case trace validation failed: %v", err)
		}
	}
}

func TestCountingSortTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := json.RawMessage(`[5, 4, 3, 2, 1]`)
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed on custom input: %v", err)
	}

	finalStep := steps[len(steps)-1]
	state, ok := finalStep.State.(renderers.ArrayState)
	if !ok {
		t.Fatalf("expected final state to be ArrayState, got %T", finalStep.State)
	}
	if !reflect.DeepEqual(state.Values, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("expected [1 2 3 4 5], got %v", state.Values)
	}
}
