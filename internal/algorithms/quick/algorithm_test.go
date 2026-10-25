package quick

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func TestQuickSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "standard input",
			input:    []int{10, 80, 30, 90, 40, 50, 70},
			expected: []int{10, 30, 40, 50, 70, 80, 90},
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
			inputCopy := make([]int, len(tt.input))
			copy(inputCopy, tt.input)
			got := Sort(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
			if !reflect.DeepEqual(tt.input, inputCopy) {
				t.Fatalf("Sort mutated input array")
			}
		})
	}
}

func TestQuickSortTraceValidation(t *testing.T) {
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
		t.Errorf("expected comparisons > 0, got %d", lastStep.Metrics.Comparisons)
	}

	customSteps, err := lesson.Build([]byte("[5, 2, 8]"))
	if err != nil {
		t.Fatalf("failed to build custom trace: %v", err)
	}
	if len(customSteps) == 0 {
		t.Fatal("custom trace produced 0 steps")
	}

	regLesson, ok := catalog.Get("quick")
	if !ok {
		t.Fatal("quick lesson not found in catalog")
	}
	if regLesson.Meta().ID != "quick" {
		t.Errorf("expected registered ID 'quick', got %q", regLesson.Meta().ID)
	}
}

func TestQuickSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "quick" {
		t.Errorf("expected ID 'quick', got %q", meta.ID)
	}
	if meta.Name != "Quick Sort" {
		t.Errorf("expected Name 'Quick Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n log n) avg, O(n²) worst" {
		t.Errorf("expected TimeComplexity 'O(n log n) avg, O(n²) worst', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(log n)" {
		t.Errorf("expected SpaceComplexity 'O(log n)', got %q", meta.Documentation.SpaceComplexity)
	}
	expectedConcepts := []string{"Divide and Conquer", "Partitioning", "Pivot Selection", "In-Place"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}
