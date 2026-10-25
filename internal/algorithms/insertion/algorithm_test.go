package insertion

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func TestInsertionSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "standard input",
			input:    []int{12, 11, 13, 5, 6},
			expected: []int{5, 6, 11, 12, 13},
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

func TestInsertionSortTraceValidation(t *testing.T) {
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
		t.Error("expected non-zero comparisons")
	}
	if lastStep.Metrics.Operations == 0 {
		t.Error("expected non-zero operations")
	}
	if lastStep.Metrics.Swaps == 0 {
		t.Error("expected non-zero swaps")
	}

	// Test edge cases
	for _, edgeArr := range [][]int{{}, {42}} {
		edgeSteps := Trace(edgeArr)
		if err := catalog.ValidateSourceLineIDs(lesson, edgeSteps); err != nil {
			t.Fatalf("edge case trace validation failed: %v", err)
		}
	}
}

func TestInsertionSortBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build([]byte(`[5, 4, 3, 2, 1]`))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("custom input line validation failed: %v", err)
	}
}

func TestInsertionSortCatalogRegistration(t *testing.T) {
	l, ok := catalog.Get("insertion")
	if !ok {
		t.Fatal("expected 'insertion' to be registered in catalog")
	}
	if l.Meta().ID != "insertion" {
		t.Fatalf("expected ID 'insertion', got %q", l.Meta().ID)
	}
}

func TestInsertionSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "insertion" {
		t.Errorf("expected ID 'insertion', got %q", meta.ID)
	}
	if meta.Name != "Insertion Sort" {
		t.Errorf("expected Name 'Insertion Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
}
