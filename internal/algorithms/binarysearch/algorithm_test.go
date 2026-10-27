package binarysearch

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
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

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last step action 'done', got %q", lastStep.Action)
	}
	if lastStep.Variables["result"] != 5 {
		t.Errorf("expected target 47 found at index 5, got %v", lastStep.Variables["result"])
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

func TestBinarySearchCustomBuild(t *testing.T) {
	lesson := Lesson{}

	validCases := []struct {
		name  string
		input string
	}{
		{
			name:  "standard with array field",
			input: `{"array": [10, 20, 30, 40, 50], "target": 30}`,
		},
		{
			name:  "standard with arr field",
			input: `{"arr": [10, 20, 30, 40, 50], "target": 10}`,
		},
		{
			name:  "empty array",
			input: `{"array": [], "target": 5}`,
		},
		{
			name:  "single element",
			input: `{"array": [42], "target": 42}`,
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(json.RawMessage(tc.input))
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.input, err)
			}
			if len(steps) == 0 {
				t.Fatal("trace produced 0 steps")
			}
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("line validation failed: %v", err)
			}
		})
	}
}

func TestBinarySearchInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidCases := []struct {
		name      string
		input     string
		errSubstr string
	}{
		{
			name:      "unsorted array",
			input:     `{"array": [30, 10, 20], "target": 20}`,
			errSubstr: "sorted",
		},
		{
			name:      "descending array",
			input:     `{"array": [5, 4, 3, 2, 1], "target": 3}`,
			errSubstr: "sorted",
		},
		{
			name:      "malformed JSON",
			input:     `{"array": [1, 2, 3]`,
			errSubstr: "invalid input format",
		},
		{
			name:      "missing target",
			input:     `{"array": [1, 2, 3]}`,
			errSubstr: "target is required",
		},
		{
			name:      "null target",
			input:     `{"array": [1, 2, 3], "target": null}`,
			errSubstr: "target is required",
		},
		{
			name:      "plain array rejected (missing target)",
			input:     `[1, 2, 3, 4]`,
			errSubstr: "invalid input format",
		},
		{
			name:      "non-integer float target",
			input:     `{"array": [1, 2, 3], "target": 2.5}`,
			errSubstr: "invalid input format",
		},
		{
			name:      "non-integer string target",
			input:     `{"array": [1, 2, 3], "target": "two"}`,
			errSubstr: "invalid input format",
		},
		{
			name:      "non-integer float in array",
			input:     `{"array": [1, 2.5, 3], "target": 2}`,
			errSubstr: "invalid input format",
		},
		{
			name:      "non-integer string in array",
			input:     `{"array": ["a", "b"], "target": 2}`,
			errSubstr: "invalid input format",
		},
		{
			name:      "missing array",
			input:     `{"target": 10}`,
			errSubstr: "array is required",
		},
		{
			name:      "null array",
			input:     `{"array": null, "target": 10}`,
			errSubstr: "array is required",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
			if !strings.Contains(err.Error(), tc.errSubstr) {
				t.Fatalf("expected error to contain %q, got %q", tc.errSubstr, err.Error())
			}
		})
	}
}

func TestBinarySearchUniqueIDs(t *testing.T) {
	cases := [][]int{
		{10, 20, 30, 40, 50},
		{7},
		{},
	}

	for _, arr := range cases {
		steps := Trace(arr, 30)
		for stepIdx, step := range steps {
			state, ok := step.State.(renderers.ArrayState)
			if !ok {
				t.Fatalf("step %d state is not ArrayState: %T", stepIdx, step.State)
			}
			if len(state.IDs) != len(arr) {
				t.Fatalf("step %d: expected %d IDs, got %d", stepIdx, len(arr), len(state.IDs))
			}
			seen := make(map[string]bool)
			for i, id := range state.IDs {
				if id == "" {
					t.Fatalf("step %d: empty ID at index %d", stepIdx, i)
				}
				expectedID := fmt.Sprintf("binarysearch-%d", i)
				if id != expectedID {
					t.Fatalf("step %d: expected ID %q, got %q", stepIdx, expectedID, id)
				}
				if seen[id] {
					t.Fatalf("step %d: duplicate ID %q", stepIdx, id)
				}
				seen[id] = true
			}
		}
	}
}

func TestBinarySearchSnapshotImmutability(t *testing.T) {
	steps := Trace([]int{10, 20, 30, 40, 50}, 40)
	if len(steps) < 3 {
		t.Fatalf("expected at least 3 steps, got %d", len(steps))
	}

	first := steps[0].State.(renderers.ArrayState)
	var midStep renderers.ArrayState
	for _, s := range steps {
		if s.LineID == "L4" {
			midStep = s.State.(renderers.ArrayState)
			break
		}
	}
	last := steps[len(steps)-1].State.(renderers.ArrayState)

	origFirstValues := slicesClone(first.Values)
	origFirstIDs := slicesClone(first.IDs)
	origMidValues := slicesClone(midStep.Values)
	origMidIDs := slicesClone(midStep.IDs)

	// Mutate last snapshot
	last.Values[0] = 999
	last.IDs[0] = "corrupted"
	if len(last.Sorted) > 0 {
		last.Sorted[0] = 999
	}
	last.Comparing = append(last.Comparing, 888)

	// Check first snapshot unaffected
	if !reflect.DeepEqual(first.Values, origFirstValues) {
		t.Fatalf("first values mutated: got %v, expected %v", first.Values, origFirstValues)
	}
	if !reflect.DeepEqual(first.IDs, origFirstIDs) {
		t.Fatalf("first IDs mutated: got %v, expected %v", first.IDs, origFirstIDs)
	}

	// Check mid snapshot unaffected
	if !reflect.DeepEqual(midStep.Values, origMidValues) {
		t.Fatalf("mid values mutated: got %v, expected %v", midStep.Values, origMidValues)
	}
	if !reflect.DeepEqual(midStep.IDs, origMidIDs) {
		t.Fatalf("mid IDs mutated: got %v, expected %v", midStep.IDs, origMidIDs)
	}
}

func TestBinarySearchTargetFoundPositions(t *testing.T) {
	arr := []int{10, 20, 30, 40, 50}
	positions := []struct {
		name        string
		target      int
		expectedIdx int
	}{
		{"start", 10, 0},
		{"mid", 30, 2},
		{"end", 50, 4},
	}

	lesson := Lesson{}
	for _, p := range positions {
		t.Run(p.name, func(t *testing.T) {
			steps := Trace(arr, p.target)
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("line validation failed: %v", err)
			}

			lastStep := steps[len(steps)-1]
			if lastStep.Action != trace.ActionDone {
				t.Fatalf("expected last step action 'done', got %q", lastStep.Action)
			}
			if lastStep.LineID != "L6" {
				t.Fatalf("expected last step LineID 'L6', got %q", lastStep.LineID)
			}
			if lastStep.Variables["found"] != true {
				t.Errorf("expected found == true, got %v", lastStep.Variables["found"])
			}
			if lastStep.Variables["result"] != p.expectedIdx {
				t.Errorf("expected result index %d, got %v", p.expectedIdx, lastStep.Variables["result"])
			}

			state := lastStep.State.(renderers.ArrayState)
			if !reflect.DeepEqual(state.Sorted, []int{p.expectedIdx}) {
				t.Errorf("expected Sorted == [%d] for green highlight, got %v", p.expectedIdx, state.Sorted)
			}
			if state.Current != p.expectedIdx {
				t.Errorf("expected Current == %d, got %d", p.expectedIdx, state.Current)
			}
			if !strings.Contains(state.Caption, "found") {
				t.Errorf("expected caption to contain 'found', got %q", state.Caption)
			}
		})
	}
}

func TestBinarySearchTargetNotFound(t *testing.T) {
	arr := []int{10, 20, 30}
	cases := []struct {
		name   string
		target int
	}{
		{"smaller than min", 5},
		{"larger than max", 50},
		{"between elements", 15},
	}

	lesson := Lesson{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := Trace(arr, tc.target)
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("line validation failed: %v", err)
			}

			lastStep := steps[len(steps)-1]
			if lastStep.Action != trace.ActionDone {
				t.Fatalf("expected last step action 'done', got %q", lastStep.Action)
			}
			if lastStep.LineID != "L13" {
				t.Fatalf("expected last step LineID 'L13', got %q", lastStep.LineID)
			}
			if lastStep.Variables["found"] != false {
				t.Errorf("expected found == false, got %v", lastStep.Variables["found"])
			}
			if lastStep.Variables["result"] != -1 {
				t.Errorf("expected result == -1, got %v", lastStep.Variables["result"])
			}

			// Verify elimination explanations occurred
			var hasElimination bool
			for _, s := range steps {
				if (s.LineID == "L7" || s.LineID == "L9") && strings.Contains(s.Explanation, "cannot be in") {
					hasElimination = true
				}
			}
			if !hasElimination {
				t.Error("expected at least one elimination step explaining why search space was halved")
			}

			// Verify not found step explains low > high and empty search space
			if !strings.Contains(lastStep.Explanation, "search space empty") {
				t.Errorf("expected not found description to mention empty search space, got %q", lastStep.Explanation)
			}
		})
	}
}

func TestBinarySearchSingleElement(t *testing.T) {
	lesson := Lesson{}

	// Single element found
	stepsFound := Trace([]int{42}, 42)
	if err := catalog.ValidateSourceLineIDs(lesson, stepsFound); err != nil {
		t.Fatal(err)
	}
	lastFound := stepsFound[len(stepsFound)-1]
	if lastFound.LineID != "L6" || lastFound.Variables["result"] != 0 {
		t.Errorf("expected single element found at index 0, got %v", lastFound.Variables["result"])
	}
	stateFound := lastFound.State.(renderers.ArrayState)
	if !reflect.DeepEqual(stateFound.Sorted, []int{0}) {
		t.Errorf("expected Sorted == [0], got %v", stateFound.Sorted)
	}

	// Single element not found
	stepsNotFound := Trace([]int{42}, 7)
	if err := catalog.ValidateSourceLineIDs(lesson, stepsNotFound); err != nil {
		t.Fatal(err)
	}
	lastNotFound := stepsNotFound[len(stepsNotFound)-1]
	if lastNotFound.LineID != "L13" || lastNotFound.Variables["result"] != -1 {
		t.Errorf("expected single element not found result -1, got %v", lastNotFound.Variables["result"])
	}
}

func TestBinarySearchEvenAndOddLengths(t *testing.T) {
	lesson := Lesson{}

	cases := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"odd length 5", []int{1, 3, 5, 7, 9}, 7, 3},
		{"odd length 7", []int{1, 3, 5, 7, 9, 11, 13}, 3, 1},
		{"even length 2", []int{10, 20}, 20, 1},
		{"even length 4", []int{2, 4, 6, 8}, 6, 2},
		{"even length 6", []int{2, 4, 6, 8, 10, 12}, 10, 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := Trace(tc.arr, tc.target)
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("line validation failed: %v", err)
			}

			// Verify mid calculation and comparison steps
			for _, s := range steps {
				if s.LineID == "L4" {
					low := s.Variables["low"].(int)
					high := s.Variables["high"].(int)
					mid := s.Variables["mid"].(int)
					expectedMid := low + (high-low)/2
					if mid != expectedMid {
						t.Errorf("mid mismatch: got %d, expected %d", mid, expectedMid)
					}
					state := s.State.(renderers.ArrayState)
					if state.Current != mid {
						t.Errorf("expected Current == mid (%d), got %d", mid, state.Current)
					}
				}
				if s.LineID == "L5" {
					mid := s.Variables["mid"].(int)
					state := s.State.(renderers.ArrayState)
					if !reflect.DeepEqual(state.Comparing, []int{mid}) {
						t.Errorf("expected Comparing == [%d], got %v", mid, state.Comparing)
					}
				}
				// Every step must have target variable
				if _, ok := s.Variables["target"]; !ok {
					t.Errorf("step %s missing target variable", s.LineID)
				}
			}

			last := steps[len(steps)-1]
			if last.Variables["result"] != tc.expected {
				t.Fatalf("expected result %d, got %v", tc.expected, last.Variables["result"])
			}
		})
	}
}

func TestBinarySearchEmptyArray(t *testing.T) {
	lesson := Lesson{}
	steps := Trace([]int{}, 10)
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on empty array: %v", err)
	}
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps for empty array, got %d", len(steps))
	}
	last := steps[len(steps)-1]
	if last.LineID != "L13" || last.Variables["result"] != -1 {
		t.Errorf("expected result -1 for empty array, got %v", last.Variables["result"])
	}
}

func slicesClone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	c := make([]T, len(s))
	copy(c, s)
	return c
}
