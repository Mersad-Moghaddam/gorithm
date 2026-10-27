package selection

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func TestSelectionSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "standard input",
			input:    []int{64, 25, 12, 22, 11},
			expected: []int{11, 12, 22, 25, 64},
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
			name:     "single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sort(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestSelectionSortTraceValidation(t *testing.T) {
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

func TestSelectionSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "selection" {
		t.Errorf("expected ID 'selection', got %q", meta.ID)
	}
	if meta.Name != "Selection Sort" {
		t.Errorf("expected Name 'Selection Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n²)" {
		t.Errorf("expected TimeComplexity 'O(n²)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Min Element Scanning", "In-Place", "Unstable Swapping"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}

func TestSelectionSortCustomBuild(t *testing.T) {
	lesson := Lesson{}
	customInput := json.RawMessage(`[5, 1, 4]`)
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

func TestSelectionSortTraceMetrics(t *testing.T) {
	// For [3, 2, 1]:
	// i=0: compares (0,1)->2, (1,2)->1; swap (0,2) -> [1, 2, 3] (2 comp, 1 swap)
	// i=1: compares (1,2)->3 > 2; no swap (1 comp, 0 swap)
	// Total comparisons: 3, total swaps: 1
	steps := Trace([]int{3, 2, 1})
	lastStep := steps[len(steps)-1]

	if lastStep.Metrics.Comparisons != 3 {
		t.Errorf("expected 3 comparisons, got %d", lastStep.Metrics.Comparisons)
	}
	if lastStep.Metrics.Swaps != 1 {
		t.Errorf("expected 1 swap, got %d", lastStep.Metrics.Swaps)
	}
}

func TestSelectionSortEdgeCaseTraces(t *testing.T) {
	lesson := Lesson{}
	for _, input := range [][]int{nil, {}, {42}} {
		steps := Trace(input)
		if len(steps) == 0 {
			t.Fatalf("expected steps for input %v", input)
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed for input %v: %v", input, err)
		}
	}
}

func TestSelectionSortRegistration(t *testing.T) {
	l, ok := catalog.Get("selection")
	if !ok {
		t.Fatal("lesson 'selection' not registered in catalog")
	}
	if l.Meta().ID != "selection" {
		t.Errorf("expected ID 'selection', got %q", l.Meta().ID)
	}
}

