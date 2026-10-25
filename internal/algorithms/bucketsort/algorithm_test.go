package bucketsort

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func TestBucketSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "default input",
			input:    []int{78, 17, 39, 26, 72, 94, 21, 12, 23, 68},
			expected: []int{12, 17, 21, 23, 26, 39, 68, 72, 78, 94},
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
			name:     "all identical",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
		},
		{
			name:     "with negatives",
			input:    []int{-5, -1, -10, 0, 3, -2},
			expected: []int{-10, -5, -2, -1, 0, 3},
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

func TestBucketSortLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "bucketsort" {
		t.Errorf("expected ID 'bucketsort', got %q", meta.ID)
	}
	if meta.Name != "Bucket Sort" {
		t.Errorf("expected Name 'Bucket Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n + k) average, O(n²) worst" {
		t.Errorf("expected TimeComplexity 'O(n + k) average, O(n²) worst', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n + k)" {
		t.Errorf("expected SpaceComplexity 'O(n + k)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}
	expectedConcepts := []string{"Scatter-Gather", "Uniform Distribution", "Local Insertion Sorting", "Range Hashing"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}
}

func TestBucketSortCatalogRegistration(t *testing.T) {
	l, ok := catalog.Get("bucketsort")
	if !ok {
		t.Fatal("expected 'bucketsort' to be registered in catalog")
	}
	if l.Meta().ID != "bucketsort" {
		t.Fatalf("expected ID 'bucketsort', got %q", l.Meta().ID)
	}
}

func TestBucketSortTraceValidation(t *testing.T) {
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
	if lastStep.Metrics.Operations == 0 {
		t.Error("expected non-zero operations")
	}

	for _, edgeArr := range [][]int{{}, {42}, {5, 5, 5}} {
		edgeSteps := Trace(edgeArr)
		if err := catalog.ValidateSourceLineIDs(lesson, edgeSteps); err != nil {
			t.Fatalf("edge case trace validation failed: %v", err)
		}
	}
}

func TestBucketSortBuildCustomInput(t *testing.T) {
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
