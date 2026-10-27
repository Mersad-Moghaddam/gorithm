package radixsort

import (
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"strings"
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
		{
			name:     "varying digit lengths",
			input:    []int{1000, 5, 23, 456, 12, 7890},
			expected: []int{5, 12, 23, 456, 1000, 7890},
		},
		{
			name:     "all equal",
			input:    []int{7, 7, 7, 7, 7},
			expected: []int{7, 7, 7, 7, 7},
		},
		{
			name:     "all zeroes",
			input:    []int{0, 0, 0},
			expected: []int{0, 0, 0},
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
	state, ok := finalStep.State.(renderers.ArrayState)
	if !ok {
		t.Fatalf("expected final state to be ArrayState, got %T", finalStep.State)
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
	state, ok := finalStep.State.(renderers.ArrayState)
	if !ok {
		t.Fatalf("expected final state to be ArrayState, got %T", finalStep.State)
	}
	if !reflect.DeepEqual(state.Values, []int{10, 20, 30, 40, 50}) {
		t.Fatalf("expected [10 20 30 40 50], got %v", state.Values)
	}
}

func TestRadixSortIdentityAndStability(t *testing.T) {
	input := []int{170, 45, 75, 90, 802, 24, 2, 66, 45, 24, 170}
	steps := Trace(input)
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}

	first := steps[0].State.(renderers.ArrayState)
	if len(first.IDs) != len(input) {
		t.Fatalf("expected %d IDs, got %d", len(input), len(first.IDs))
	}

	// Map each value to its sequence of assigned IDs in original order
	expectedOrder := make(map[int][]string)
	for i, v := range input {
		expectedOrder[v] = append(expectedOrder[v], first.IDs[i])
	}

	// Check each step maintains unique IDs and valid captions
	for k, step := range steps {
		s := step.State.(renderers.ArrayState)
		if s.Caption == "" {
			t.Fatalf("step %d missing caption", k)
		}
		if len(s.IDs) != len(input) {
			t.Fatalf("step %d lost IDs: want %d, got %d", k, len(input), len(s.IDs))
		}
		unique := make(map[string]bool)
		for _, id := range s.IDs {
			if id == "" || unique[id] {
				t.Fatalf("step %d non-unique or empty ID: %q", k, id)
			}
			unique[id] = true
		}

		// Ensure SVG renders each unique data-key exactly once
		svg := renderers.RenderArraySVG(step, 800, 400)
		for _, id := range s.IDs {
			if strings.Count(svg, `data-key="array-`+id+`"`) != 1 {
				t.Fatalf("step %d: SVG missing or duplicate key for ID %s", k, id)
			}
		}

		// At pass-completion steps (L5), relative order of duplicate elements must be preserved
		if step.LineID == "L5" {
			currentOrder := make(map[int][]string)
			for i, v := range s.Values {
				currentOrder[v] = append(currentOrder[v], s.IDs[i])
			}
			for v, initialIDs := range expectedOrder {
				if !slices.Equal(currentOrder[v], initialIDs) {
					t.Fatalf("step %d (L5): stability violated for value %d: want %v, got %v", k, v, initialIDs, currentOrder[v])
				}
			}
		}
	}

	// Verify final step stability and correctness
	last := steps[len(steps)-1].State.(renderers.ArrayState)
	expectedSorted := Sort(input)
	if !slices.Equal(last.Values, expectedSorted) {
		t.Fatalf("final values = %v, want %v", last.Values, expectedSorted)
	}
	finalOrder := make(map[int][]string)
	for i, v := range last.Values {
		finalOrder[v] = append(finalOrder[v], last.IDs[i])
	}
	for v, initialIDs := range expectedOrder {
		if !slices.Equal(finalOrder[v], initialIDs) {
			t.Fatalf("final stability violated for value %d: want %v, got %v", v, initialIDs, finalOrder[v])
		}
	}
}

func TestRadixSortSnapshotImmutability(t *testing.T) {
	input := []int{170, 45, 75, 90, 802, 24, 2, 66}
	steps := Trace(input)
	if len(steps) < 5 {
		t.Fatalf("expected at least 5 steps, got %d", len(steps))
	}

	first := steps[0].State.(renderers.ArrayState)
	firstValuesOrig := slices.Clone(first.Values)
	firstIDsOrig := slices.Clone(first.IDs)

	midIdx := len(steps) / 2
	mid := steps[midIdx].State.(renderers.ArrayState)
	midValuesOrig := slices.Clone(mid.Values)
	midIDsOrig := slices.Clone(mid.IDs)

	last := steps[len(steps)-1].State.(renderers.ArrayState)

	// Mutate last snapshot
	last.Values[0] = 999999
	last.IDs[0] = "radixsort-corrupted"
	if len(last.Sorted) > 0 {
		last.Sorted[0] = 999
	}

	// Verify first and intermediate snapshots were not mutated by changing last
	if !slices.Equal(first.Values, firstValuesOrig) || !slices.Equal(first.IDs, firstIDsOrig) {
		t.Fatalf("first snapshot mutated: want %v, got %v", firstValuesOrig, first.Values)
	}
	if !slices.Equal(mid.Values, midValuesOrig) || !slices.Equal(mid.IDs, midIDsOrig) {
		t.Fatalf("intermediate snapshot mutated by modifying last: want %v, got %v", midValuesOrig, mid.Values)
	}

	// Mutate first snapshot and ensure mid is not affected
	first.Values[0] = 777777
	first.IDs[0] = "radixsort-first-corrupted"
	if !slices.Equal(mid.Values, midValuesOrig) || !slices.Equal(mid.IDs, midIDsOrig) {
		t.Fatalf("intermediate snapshot mutated by modifying first: want %v, got %v", midValuesOrig, mid.Values)
	}
}

func TestRadixSortEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		input []int
	}{
		{name: "nil", input: nil},
		{name: "empty", input: []int{}},
		{name: "single item", input: []int{42}},
		{name: "varying digit lengths", input: []int{10000, 1, 100, 10}},
		{name: "all equal", input: []int{9, 9, 9, 9}},
		{name: "all zeroes", input: []int{0, 0, 0}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := Trace(tc.input)
			if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
				t.Fatalf("ValidateSourceLineIDs failed: %v", err)
			}
			if len(steps) == 0 {
				t.Fatal("expected at least 1 step")
			}

			last := steps[len(steps)-1].State.(renderers.ArrayState)
			expected := Sort(tc.input)
			if !slices.Equal(last.Values, expected) && !(len(tc.input) == 0 && len(last.Values) == 0) {
				t.Fatalf("expected sorted %v, got %v", expected, last.Values)
			}
			if len(last.IDs) != len(tc.input) {
				t.Fatalf("expected %d IDs, got %d", len(tc.input), len(last.IDs))
			}

			for _, step := range steps {
				s := step.State.(renderers.ArrayState)
				if s.Caption == "" {
					t.Fatalf("step %s missing caption", step.LineID)
				}
				unique := make(map[string]bool)
				for _, id := range s.IDs {
					if id == "" || unique[id] {
						t.Fatalf("non-unique or empty id %q", id)
					}
					unique[id] = true
				}
			}
		})
	}
}

func TestRadixSortInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid inputs
	for _, tc := range []struct {
		name     string
		input    json.RawMessage
		wantSize int
	}{
		{name: "nil input", input: nil, wantSize: 8},
		{name: "empty slice input", input: []byte(""), wantSize: 8},
		{name: "empty array json", input: []byte("[]"), wantSize: 0},
		{name: "custom array", input: []byte("[10, 5, 20]"), wantSize: 3},
	} {
		steps, err := lesson.Build(tc.input)
		if err != nil {
			t.Fatalf("expected %s to succeed, got error: %v", tc.name, err)
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if len(last.Values) != tc.wantSize {
			t.Fatalf("%s: expected size %d, got %d", tc.name, tc.wantSize, len(last.Values))
		}
	}

	// Invalid inputs to reject
	invalidInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{name: "whitespace only", input: []byte("   ")},
		{name: "null literal", input: []byte("null")},
		{name: "malformed json", input: []byte("[1, 2,")},
		{name: "object instead of array", input: []byte(`{"a": 1}`)},
		{name: "number instead of array", input: []byte("42")},
		{name: "boolean instead of array", input: []byte("true")},
		{name: "string instead of array", input: []byte(`"hello"`)},
		{name: "float in array", input: []byte("[1, 2.5, 3]")},
		{name: "string in array", input: []byte(`[1, "2", 3]`)},
		{name: "null in array", input: []byte("[1, null, 3]")},
		{name: "negative integer", input: []byte("[-5, 10]")},
		{name: "negative integer inside array", input: []byte("[1, 2, -3]")},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected %s (%s) to be rejected, but it succeeded", tc.name, tc.input)
			}
		})
	}
}

func TestRadixSortEducationalCaptions(t *testing.T) {
	steps := Trace([]int{170, 45, 75, 90, 802, 24, 2, 66})

	var foundPass1, foundPass10, foundGroupCaption bool
	for _, step := range steps {
		s := step.State.(renderers.ArrayState)
		if strings.Contains(s.Caption, "Pass exp=1 (ones digit)") {
			foundPass1 = true
		}
		if strings.Contains(s.Caption, "Pass exp=10 (tens digit)") {
			foundPass10 = true
		}
		if strings.Contains(s.Caption, "group by ones digit:") && strings.Contains(s.Caption, "170(0)") {
			foundGroupCaption = true
		}
	}

	if !foundPass1 {
		t.Error("expected caption with 'Pass exp=1 (ones digit)'")
	}
	if !foundPass10 {
		t.Error("expected caption with 'Pass exp=10 (tens digit)'")
	}
	if !foundGroupCaption {
		t.Error("expected educational caption grouping by digit: 'group by ones digit: [170(0)...]'")
	}
}
