package binarysearch

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func TestSearchCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{
			name:     "found in odd length middle",
			arr:      []int{1, 3, 5, 7, 9},
			target:   5,
			expected: 2,
		},
		{
			name:     "found in odd length left",
			arr:      []int{1, 3, 5, 7, 9},
			target:   1,
			expected: 0,
		},
		{
			name:     "found in odd length right",
			arr:      []int{1, 3, 5, 7, 9},
			target:   9,
			expected: 4,
		},
		{
			name:     "found in even length left half",
			arr:      []int{2, 4, 6, 8, 10, 12},
			target:   4,
			expected: 1,
		},
		{
			name:     "found in even length right half",
			arr:      []int{2, 4, 6, 8, 10, 12},
			target:   10,
			expected: 4,
		},
		{
			name:     "not found smaller than min",
			arr:      []int{2, 4, 6, 8, 10},
			target:   1,
			expected: -1,
		},
		{
			name:     "not found larger than max",
			arr:      []int{2, 4, 6, 8, 10},
			target:   15,
			expected: -1,
		},
		{
			name:     "not found between elements",
			arr:      []int{2, 4, 6, 8, 10},
			target:   5,
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
		{
			name:     "empty slice",
			arr:      []int{},
			target:   10,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Search(tt.arr, tt.target)
			if got != tt.expected {
				t.Fatalf("Search(%v, %d) = %d, expected %d", tt.arr, tt.target, got, tt.expected)
			}
		})
	}
}

func TestBinarySearchMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "binarysearch" {
		t.Errorf("expected ID 'binarysearch', got %q", meta.ID)
	}
	if meta.Name != "Binary Search" {
		t.Errorf("expected Name 'Binary Search', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySearching {
		t.Errorf("expected Category %q, got %q", catalog.CategorySearching, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n)" {
		t.Errorf("expected TimeComplexity 'O(log n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Divide and Conquer", "Sorted Precondition", "Range Bisection", "Logarithmic Search"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}
}

func TestBinarySearchTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build default trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for default build: %v", err)
	}
}

func TestBinarySearchCustomBuild(t *testing.T) {
	lesson := Lesson{}

	// Test with JSON object containing array and target
	customInput := json.RawMessage(`{"array": [10, 20, 30, 40, 50], "target": 30}`)
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

	// Test with plain array JSON
	arrayInput := json.RawMessage(`[1, 2, 3, 4]`)
	steps2, err := lesson.Build(arrayInput)
	if err != nil {
		t.Fatalf("failed to build trace with array input: %v", err)
	}
	if len(steps2) == 0 {
		t.Fatal("trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps2); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}

func TestBinarySearchEdgeCaseTraces(t *testing.T) {
	lesson := Lesson{}
	cases := []struct {
		arr    []int
		target int
	}{
		{nil, 5},
		{[]int{}, 5},
		{[]int{10}, 10},
		{[]int{10}, 5},
		{[]int{1, 2, 3, 4, 5, 6}, 1},
		{[]int{1, 2, 3, 4, 5, 6}, 6},
		{[]int{1, 2, 3, 4, 5, 6}, 7},
	}

	for _, c := range cases {
		steps := Trace(c.arr, c.target)
		if len(steps) == 0 {
			t.Fatalf("expected steps for input %v, target %d", c.arr, c.target)
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed for input %v, target %d: %v", c.arr, c.target, err)
		}
	}
}

func TestBinarySearchRegistration(t *testing.T) {
	l, ok := catalog.Get("binarysearch")
	if !ok {
		t.Fatal("lesson 'binarysearch' not registered in catalog")
	}
	if l.Meta().ID != "binarysearch" {
		t.Errorf("expected ID 'binarysearch', got %q", l.Meta().ID)
	}
}

func TestBinarySearchDefaultBuildValues(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) error: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != "done" {
		t.Fatalf("expected last step action 'done', got %q", lastStep.Action)
	}
	if lastStep.Variables["result"] != 5 {
		t.Errorf("expected target 47 found at index 5, got %v", lastStep.Variables["result"])
	}
}

func TestBinarySearchTracePointersAndState(t *testing.T) {
	steps := Trace([]int{10, 20, 30, 40, 50}, 40)
	var foundMid, foundPointers bool
	for _, s := range steps {
		if s.LineID == "L4" && !foundMid {
			foundMid = true
			if mid, ok := s.Variables["mid"].(int); !ok || mid != 2 {
				t.Errorf("expected initial mid = 2, got %v", s.Variables["mid"])
			}
		}
		if s.LineID == "L5" {
			if _, ok := s.Variables["low"]; ok {
				if _, ok := s.Variables["high"]; ok {
					foundPointers = true
				}
			}
		}
	}
	if !foundMid {
		t.Error("expected mid calculation step L4 in trace")
	}
	if !foundPointers {
		t.Error("expected low and high pointers in trace variables")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Variables["found"] != true {
		t.Errorf("expected found == true, got %v", lastStep.Variables["found"])
	}
	if lastStep.Variables["result"] != 3 {
		t.Errorf("expected result index 3, got %v", lastStep.Variables["result"])
	}
}

