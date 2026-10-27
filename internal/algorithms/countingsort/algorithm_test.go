package countingsort

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "default input",
			input:    []int{4, 2, 2, 8, 3, 3, 1},
			expected: []int{1, 2, 2, 3, 3, 4, 8},
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
			name:     "with duplicates",
			input:    []int{3, 1, 2, 3, 1},
			expected: []int{1, 1, 2, 3, 3},
		},
		{
			name:     "with zeros",
			input:    []int{0, 5, 0, 3, 2, 1},
			expected: []int{0, 0, 1, 2, 3, 5},
		},
		{
			name:     "all identical",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
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

func TestSortStability(t *testing.T) {
	// Stable sort invariant: for equal keys, earlier elements in input appear earlier in output.
	arr := []int{4, 2, 2, 8, 3, 3, 1}
	steps := Trace(arr)

	// Collect mapping of input index -> output index for step L21 (output placement)
	targetIndices := make(map[int]int)
	for _, step := range steps {
		if step.LineID == "L21" {
			inIdx := step.Variables["i"].(int)
			outIdx := step.Variables["targetIndex"].(int)
			targetIndices[inIdx] = outIdx
		}
	}

	// Verify for duplicate value 2 at input indices 1 and 2:
	// input index 1 must map to lower output index than input index 2.
	if out1, out2 := targetIndices[1], targetIndices[2]; out1 >= out2 {
		t.Fatalf("stability violated for value 2: inIdx 1 -> %d, inIdx 2 -> %d (expected out1 < out2)", out1, out2)
	}

	// Verify for duplicate value 3 at input indices 4 and 5:
	// input index 4 must map to lower output index than input index 5.
	if out4, out5 := targetIndices[4], targetIndices[5]; out4 >= out5 {
		t.Fatalf("stability violated for value 3: inIdx 4 -> %d, inIdx 5 -> %d (expected out4 < out5)", out4, out5)
	}

	// Verify final step state IDs reflect duplicate stability
	first := steps[0].State.(renderers.ArrayState)
	last := steps[len(steps)-1].State.(renderers.ArrayState)
	for _, val := range []int{2, 3} {
		var before, after []string
		for i, v := range first.Values {
			if v == val {
				before = append(before, first.IDs[i])
			}
		}
		for i, v := range last.Values {
			if v == val {
				after = append(after, last.IDs[i])
			}
		}
		if !slices.Equal(before, after) {
			t.Fatalf("stability violated for val %d: before=%v after=%v", val, before, after)
		}
	}
}

func TestCountingSortLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "countingsort" {
		t.Fatalf("expected ID %q, got %q", "countingsort", meta.ID)
	}
	if meta.Name != "Counting Sort" {
		t.Fatalf("expected Name %q, got %q", "Counting Sort", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Fatalf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Fatalf("expected Renderer %q, got %q", "array", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n + k)" {
		t.Fatalf("expected TimeComplexity %q, got %q", "O(n + k)", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(k)" {
		t.Fatalf("expected SpaceComplexity %q, got %q", "O(k)", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Easy" {
		t.Fatalf("expected Difficulty %q, got %q", "Easy", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Non-Comparison Sort",
		"Frequency Tally",
		"Prefix Sum Accumulation",
		"Stable Output Mapping",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Fatalf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.HowItWorks == "" {
		t.Fatal("expected non-empty HowItWorks")
	}
}

func TestCountingSortCatalogRegistration(t *testing.T) {
	lesson, ok := catalog.Get("countingsort")
	if !ok {
		t.Fatal("countingsort not found in catalog")
	}
	if lesson.Meta().ID != "countingsort" {
		t.Fatalf("expected ID countingsort, got %s", lesson.Meta().ID)
	}
}

func TestCountingSortTraceValidation(t *testing.T) {
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

	// Verify count frequencies in variables
	hasCount := false
	for _, step := range steps {
		if _, ok := step.Variables["count"]; ok {
			hasCount = true
			break
		}
	}
	if !hasCount {
		t.Fatal("expected at least one step with count in variables")
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
	if !reflect.DeepEqual(state.Values, []int{1, 2, 2, 3, 3, 4, 8}) {
		t.Fatalf("expected sorted [1 2 2 3 3 4 8], got %v", state.Values)
	}

	// Test edge cases
	for _, edgeArr := range [][]int{{}, {42}} {
		edgeSteps := Trace(edgeArr)
		if err := catalog.ValidateSourceLineIDs(lesson, edgeSteps); err != nil {
			t.Fatalf("edge case trace validation failed: %v", err)
		}
	}
}

func TestCountingSortTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := json.RawMessage(`[5, 4, 3, 2, 1]`)
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
	if !reflect.DeepEqual(state.Values, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("expected [1 2 3 4 5], got %v", state.Values)
	}
}

func TestCountingSortStrictInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid inputs
	validInputs := []struct {
		name  string
		input json.RawMessage
		want  int
	}{
		{"nil input", nil, 7},
		{"empty raw message", json.RawMessage(""), 7},
		{"whitespace message", json.RawMessage("   "), 7},
		{"empty array", json.RawMessage("[]"), 0},
		{"single element", json.RawMessage("[42]"), 1},
		{"valid numbers", json.RawMessage("[3, 1, 2]"), 3},
	}
	for _, tc := range validInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err != nil {
				t.Fatalf("expected success for %s, got error: %v", tc.name, err)
			}
			last := steps[len(steps)-1].State.(renderers.ArrayState)
			if len(last.Values) != tc.want {
				t.Fatalf("expected %d elements, got %d", tc.want, len(last.Values))
			}
		})
	}

	// Invalid inputs to reject strictly
	invalidInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{"null literal", json.RawMessage("null")},
		{"number scalar", json.RawMessage("123")},
		{"string scalar", json.RawMessage(`"hello"`)},
		{"boolean scalar", json.RawMessage("true")},
		{"object", json.RawMessage(`{"arr": [1, 2]}`)},
		{"malformed json unclosed", json.RawMessage("[1, 2,")},
		{"malformed json trailing comma", json.RawMessage("[1, 2,]")},
		{"non-integer float", json.RawMessage("[1, 2.5, 3]")},
		{"non-number element", json.RawMessage(`[1, "two", 3]`)},
		{"negative number", json.RawMessage("[4, -2, 3]")},
		{"trailing data", json.RawMessage("[1, 2] 3")},
	}
	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected rejection for %s (%s), but got %d steps", tc.name, string(tc.input), len(steps))
			}
		})
	}
}

func TestCountingSortUniqueIDsAcrossSteps(t *testing.T) {
	testCases := [][]int{
		{},
		{42},
		{4, 2, 2, 8, 3, 3, 1},
		{1, 2, 3, 4},
		{5, 4, 3, 2, 1},
		{0, 0, 0},
		{7, 7, 7, 7},
	}

	for _, input := range testCases {
		t.Run(fmt.Sprintf("input_%v", input), func(t *testing.T) {
			steps := Trace(input)
			if len(steps) == 0 {
				t.Fatal("trace produced 0 steps")
			}
			for stepIdx, step := range steps {
				state, ok := step.State.(renderers.ArrayState)
				if !ok {
					t.Fatalf("step %d state is not ArrayState: %T", stepIdx, step.State)
				}
				if len(state.IDs) != len(state.Values) {
					t.Fatalf("step %d (%s): len(IDs)=%d != len(Values)=%d", stepIdx, step.LineID, len(state.IDs), len(state.Values))
				}
				seen := make(map[string]bool)
				for i, id := range state.IDs {
					if id == "" {
						t.Fatalf("step %d (%s) index %d has empty ID", stepIdx, step.LineID, i)
					}
					if seen[id] {
						t.Fatalf("step %d (%s) duplicate ID %q at index %d", stepIdx, step.LineID, id, i)
					}
					seen[id] = true
				}
			}
		})
	}
}

func TestCountingSortDuplicateStability(t *testing.T) {
	testCases := [][]int{
		{4, 2, 2, 8, 3, 3, 1},
		{2, 1, 2, 1},
		{5, 5, 5, 5},
		{0, 3, 0, 2, 3, 0},
	}

	for _, input := range testCases {
		t.Run(fmt.Sprintf("input_%v", input), func(t *testing.T) {
			steps := Trace(input)
			first := steps[0].State.(renderers.ArrayState)
			last := steps[len(steps)-1].State.(renderers.ArrayState)

			expected := Sort(input)
			if !slices.Equal(last.Values, expected) {
				t.Fatalf("final values mismatch: got %v, want %v", last.Values, expected)
			}

			// For each unique value, extract IDs in initial and final state
			valSet := make(map[int]bool)
			for _, v := range input {
				valSet[v] = true
			}

			for val := range valSet {
				var before, after []string
				for i, v := range first.Values {
					if v == val {
						before = append(before, first.IDs[i])
					}
				}
				for i, v := range last.Values {
					if v == val {
						after = append(after, last.IDs[i])
					}
				}
				if !slices.Equal(before, after) {
					t.Fatalf("stability violated for value %d: before=%v after=%v", val, before, after)
				}
			}
		})
	}
}

func TestCountingSortSnapshotImmutability(t *testing.T) {
	steps := Trace([]int{4, 2, 2, 8, 3, 3, 1})
	if len(steps) < 5 {
		t.Fatalf("too few steps: %d", len(steps))
	}

	first := steps[0].State.(renderers.ArrayState)
	firstValuesCopy := slices.Clone(first.Values)
	firstIDsCopy := slices.Clone(first.IDs)

	midIdx := len(steps) / 2
	mid := steps[midIdx].State.(renderers.ArrayState)
	midValuesCopy := slices.Clone(mid.Values)
	midIDsCopy := slices.Clone(mid.IDs)

	last := steps[len(steps)-1].State.(renderers.ArrayState)

	// Mutate last snapshot
	last.Values[0] = 9999
	last.IDs[0] = "corrupted"
	if len(last.Sorted) > 0 {
		last.Sorted[0] = 9999
	}

	// Verify first snapshot was not altered
	if !slices.Equal(first.Values, firstValuesCopy) {
		t.Fatalf("first snapshot Values mutated: got %v, want %v", first.Values, firstValuesCopy)
	}
	if !slices.Equal(first.IDs, firstIDsCopy) {
		t.Fatalf("first snapshot IDs mutated: got %v, want %v", first.IDs, firstIDsCopy)
	}

	// Verify mid snapshot was not altered
	if !slices.Equal(mid.Values, midValuesCopy) {
		t.Fatalf("mid snapshot Values mutated: got %v, want %v", mid.Values, midValuesCopy)
	}
	if !slices.Equal(mid.IDs, midIDsCopy) {
		t.Fatalf("mid snapshot IDs mutated: got %v, want %v", mid.IDs, midIDsCopy)
	}
}

func TestCountingSortEdgeCases(t *testing.T) {
	lesson := Lesson{}

	// Edge case 1: Empty slice
	t.Run("empty", func(t *testing.T) {
		steps := Trace([]int{})
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		if len(steps) != 3 {
			t.Fatalf("expected 3 steps for empty array, got %d", len(steps))
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if len(last.Values) != 0 || len(last.IDs) != 0 {
			t.Fatalf("expected empty values and ids, got %+v", last)
		}
	})

	// Edge case 2: Single element
	t.Run("single_element", func(t *testing.T) {
		steps := Trace([]int{42})
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		if len(steps) != 4 {
			t.Fatalf("expected 4 steps for single element, got %d", len(steps))
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !reflect.DeepEqual(last.Values, []int{42}) {
			t.Fatalf("expected [42], got %v", last.Values)
		}
		if !reflect.DeepEqual(last.IDs, []string{"countingsort-0"}) {
			t.Fatalf("expected ['countingsort-0'], got %v", last.IDs)
		}
		if !reflect.DeepEqual(last.Sorted, []int{0}) {
			t.Fatalf("expected Sorted [0], got %v", last.Sorted)
		}
	})

	// Edge case 3: Already sorted
	t.Run("already_sorted", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		steps := Trace(input)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !reflect.DeepEqual(last.Values, input) {
			t.Fatalf("expected %v, got %v", input, last.Values)
		}
		expectedIDs := []string{
			"countingsort-0", "countingsort-1", "countingsort-2", "countingsort-3", "countingsort-4",
		}
		if !reflect.DeepEqual(last.IDs, expectedIDs) {
			t.Fatalf("expected IDs %v, got %v", expectedIDs, last.IDs)
		}
	})

	// Edge case 4: Duplicates
	t.Run("duplicates", func(t *testing.T) {
		input := []int{2, 2, 2}
		steps := Trace(input)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !reflect.DeepEqual(last.Values, input) {
			t.Fatalf("expected %v, got %v", input, last.Values)
		}
		expectedIDs := []string{"countingsort-0", "countingsort-1", "countingsort-2"}
		if !reflect.DeepEqual(last.IDs, expectedIDs) {
			t.Fatalf("expected IDs %v, got %v", expectedIDs, last.IDs)
		}
	})
}

func TestCountingSortTeachingPhasesAndCaptions(t *testing.T) {
	steps := Trace([]int{4, 2, 2, 8, 3, 3, 1})

	var foundP1, foundP2, foundP3, foundP4 bool
	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		if state.Caption == "" {
			t.Fatalf("step %s has empty Caption", step.LineID)
		}
		if step.Explanation == "" {
			t.Fatalf("step %s has empty Explanation", step.LineID)
		}

		if strings.Contains(state.Caption, "Phase 1") {
			foundP1 = true
		}
		if strings.Contains(state.Caption, "Phase 2") {
			foundP2 = true
		}
		if strings.Contains(state.Caption, "Phase 3") {
			foundP3 = true
		}
		if strings.Contains(state.Caption, "Phase 4") {
			foundP4 = true
		}
	}

	if !foundP1 || !foundP2 || !foundP3 || !foundP4 {
		t.Fatalf("missing phases in captions: P1=%v P2=%v P3=%v P4=%v", foundP1, foundP2, foundP3, foundP4)
	}

	// Verify descriptive content requirements
	hasCountStoreExplanation := false
	hasPrefixSumPositionExplanation := false
	hasOutputStabilityExplanation := false

	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		if strings.Contains(state.Caption, "tally") || strings.Contains(state.Caption, "frequency") {
			hasCountStoreExplanation = true
		}
		if strings.Contains(state.Caption, "1-based") || strings.Contains(state.Caption, "prefix sum") {
			hasPrefixSumPositionExplanation = true
		}
		if strings.Contains(state.Caption, "stability") || strings.Contains(state.Caption, "stable") {
			hasOutputStabilityExplanation = true
		}
	}

	if !hasCountStoreExplanation {
		t.Fatal("expected caption explaining what count array stores")
	}
	if !hasPrefixSumPositionExplanation {
		t.Fatal("expected caption explaining prefix sums and 1-based positions")
	}
	if !hasOutputStabilityExplanation {
		t.Fatal("expected caption explaining output placement and stability")
	}
}
