package linearsearch

import (
	"encoding/json"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSearchCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{
			name:     "found in middle",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   33,
			expected: 4,
		},
		{
			name:     "found at start",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   15,
			expected: 0,
		},
		{
			name:     "found at end",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   55,
			expected: 6,
		},
		{
			name:     "not found",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   999,
			expected: -1,
		},
		{
			name:     "empty array",
			arr:      []int{},
			target:   10,
			expected: -1,
		},
		{
			name:     "single element found",
			arr:      []int{42},
			target:   42,
			expected: 0,
		},
		{
			name:     "single element not found",
			arr:      []int{42},
			target:   7,
			expected: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Search(tc.arr, tc.target)
			if got != tc.expected {
				t.Fatalf("Search(%v, %d) = %d; want %d", tc.arr, tc.target, got, tc.expected)
			}
		})
	}
}

func TestLinearSearchTraceValidation(t *testing.T) {
	lesson := Lesson{}

	// Test default build
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on default build: %v", err)
	}

	// Verify found element is marked in Sorted/Green
	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.ArrayState)
	if !ok {
		t.Fatalf("expected state to be ArrayState, got %T", lastStep.State)
	}
	if len(state.Sorted) == 0 || state.Sorted[0] != 4 {
		t.Fatalf("expected target index 4 in Sorted, got %v", state.Sorted)
	}

	// Test not found case
	customInput, _ := json.Marshal(map[string]any{
		"arr":    []int{1, 2, 3},
		"target": 99,
	})
	notFoundSteps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("failed to build custom trace: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, notFoundSteps); err != nil {
		t.Fatalf("line validation failed on not found trace: %v", err)
	}
	lastNotFound := notFoundSteps[len(notFoundSteps)-1]
	notFoundState := lastNotFound.State.(renderers.ArrayState)
	if len(notFoundState.Sorted) != 0 {
		t.Fatalf("expected empty Sorted on not found, got %v", notFoundState.Sorted)
	}

	// Test empty array trace
	emptySteps := Trace([]int{}, 10)
	if err := catalog.ValidateSourceLineIDs(lesson, emptySteps); err != nil {
		t.Fatalf("line validation failed on empty array: %v", err)
	}

	// Test single element found trace
	singleSteps := Trace([]int{10}, 10)
	if err := catalog.ValidateSourceLineIDs(lesson, singleSteps); err != nil {
		t.Fatalf("line validation failed on single element found: %v", err)
	}

	// Test single element not found trace
	singleNotFoundSteps := Trace([]int{10}, 20)
	if err := catalog.ValidateSourceLineIDs(lesson, singleNotFoundSteps); err != nil {
		t.Fatalf("line validation failed on single element not found: %v", err)
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "linearsearch" {
		t.Errorf("expected ID 'linearsearch', got %q", meta.ID)
	}
	if meta.Name != "Linear Search" {
		t.Errorf("expected Name 'Linear Search', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySearching {
		t.Errorf("expected Category %q, got %q", catalog.CategorySearching, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n)" {
		t.Errorf("expected TimeComplexity 'O(n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	// Verify catalog registration
	registered, ok := catalog.Get("linearsearch")
	if !ok || registered == nil {
		t.Fatal("linearsearch not found in catalog registry")
	}
}
