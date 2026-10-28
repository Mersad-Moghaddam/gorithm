package radixsort

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestRadixSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "default input",
			input:    []int{170, 45, 75, 90, 802, 24, 2, 66},
			expected: []int{2, 24, 45, 66, 75, 90, 170, 802},
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
			name:     "duplicates",
			input:    []int{5, 2, 8, 2, 5, 1, 8},
			expected: []int{1, 2, 2, 5, 5, 8, 8},
		},
		{
			name:     "large numbers with zeroes",
			input:    []int{10000, 9999, 10, 0, 500},
			expected: []int{0, 10, 500, 9999, 10000},
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

func TestRadixSortLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "radixsort" {
		t.Fatalf("expected ID %q, got %q", "radixsort", meta.ID)
	}
	if meta.Name != "Radix Sort (LSD)" {
		t.Fatalf("expected Name %q, got %q", "Radix Sort (LSD)", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Fatalf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Fatalf("expected Renderer %q, got %q", "array", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(d · (n + b))" {
		t.Fatalf("expected TimeComplexity %q, got %q", "O(d · (n + b))", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n + b)" {
		t.Fatalf("expected SpaceComplexity %q, got %q", "O(n + b)", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Fatalf("expected Difficulty %q, got %q", "Medium", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{"Digit-by-Digit Processing", "Stable Sub-sorts", "Non-Comparison Sorting", "Fixed Radix Base 10"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Fatalf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}

func TestRadixSortCatalogRegistration(t *testing.T) {
	lesson, ok := catalog.Get("radixsort")
	if !ok {
		t.Fatal("radixsort not found in catalog")
	}
	if lesson.Meta().ID != "radixsort" {
		t.Fatalf("expected ID radixsort, got %s", lesson.Meta().ID)
	}
}

func TestRadixSortTraceValidation(t *testing.T) {
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

	// Verify exponent in variables
	hasExp := false
	for _, step := range steps {
		if _, ok := step.Variables["exp"]; ok {
			hasExp = true
			break
		}
	}
	if !hasExp {
		t.Fatal("expected at least one step with exp in variables")
	}

	// Verify final step state values are sorted
	finalStep := steps[len(steps)-1]
	state, ok := finalStep.State.(renderers.RadixsortLessonState)
	if !ok {
		t.Fatalf("expected final state to be RadixsortLessonState, got %T", finalStep.State)
	}
	if !sort.IntsAreSorted(state.Values) {
		t.Fatalf("final state values are not sorted: %v", state.Values)
	}
}

func TestRadixSortTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := json.RawMessage(`[50, 40, 30, 20, 10]`)
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
	state, ok := finalStep.State.(renderers.RadixsortLessonState)
	if !ok {
		t.Fatalf("expected final state to be RadixsortLessonState, got %T", finalStep.State)
	}
	if !reflect.DeepEqual(state.Values, []int{10, 20, 30, 40, 50}) {
		t.Fatalf("expected [10 20 30 40 50], got %v", state.Values)
	}
}
