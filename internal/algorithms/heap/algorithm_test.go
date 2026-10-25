package heap

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
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
