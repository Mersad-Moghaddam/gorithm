package quick

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
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
			name:     "all equal",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
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
		{
			name:     "with negatives",
			input:    []int{-5, 10, -2, 0, 7, -9},
			expected: []int{-9, -5, -2, 0, 7, 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCopy := slices.Clone(tt.input)
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
	if lastStep.Metrics.Swaps == 0 {
		t.Errorf("expected swaps > 0, got %d", lastStep.Metrics.Swaps)
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

func TestQuickSortInputValidation(t *testing.T) {
	lesson := Lesson{}
	invalidInputs := []json.RawMessage{
		[]byte(`not json`),
		[]byte(`{"bad":1}`),
		[]byte(`123`),
		[]byte(`"string"`),
		[]byte(`true`),
		[]byte(`null`),
		[]byte(`[null]`),
		[]byte(`[1, null, 2]`),
		[]byte(`[1, 2.5]`),
		[]byte(`[3.14]`),
		[]byte(`[1, "two"]`),
	}

	for _, input := range invalidInputs {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("expected error for invalid input %s, got nil", input)
		}
	}

	validEmpty, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatalf("unexpected error for empty array: %v", err)
	}
	emptyState := validEmpty[len(validEmpty)-1].State.(renderers.ArrayState)
	if len(emptyState.Values) != 0 {
		t.Fatalf("expected empty values, got %v", emptyState.Values)
	}

	validCustom, err := lesson.Build(json.RawMessage(`[5, 2, 8]`))
	if err != nil {
		t.Fatalf("unexpected error for custom input: %v", err)
	}
	customLast := validCustom[len(validCustom)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(customLast.Values, []int{2, 5, 8}) {
		t.Fatalf("expected sorted [2, 5, 8], got %v", customLast.Values)
	}

	validDefault, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("unexpected error for nil input: %v", err)
	}
	defaultLast := validDefault[len(validDefault)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(defaultLast.Values, []int{10, 30, 40, 50, 70, 80, 90}) {
		t.Fatalf("expected default sorted, got %v", defaultLast.Values)
	}
}

func TestQuickSortTeachingTraceAndIdentity(t *testing.T) {
	cases := [][]int{
		nil,
		{},
		{42},
		{7, 7, 7, 7},
		{5, 4, 3, 2, 1},
		{1, 2, 3, 4, 5},
		{3, 1, 2, 3, 1},
		{10, 80, 30, 90, 40, 50, 70},
	}

	for _, input := range cases {
		steps := Trace(input)
		if len(steps) == 0 {
			t.Fatalf("trace produced 0 steps for input %v", input)
		}

		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatalf("source line validation failed for input %v: %v", input, err)
		}

		first := steps[0].State.(renderers.ArrayState)
		n := len(input)

		if len(first.IDs) != n || len(first.Values) != n {
			t.Fatalf("initial state mismatch for input %v: %+v", input, first)
		}

		origValues := make(map[string]int)
		for i, id := range first.IDs {
			expectedID := fmt.Sprintf("quick-%d", i)
			if id != expectedID {
				t.Fatalf("expected ID %q, got %q", expectedID, id)
			}
			origValues[id] = first.Values[i]
		}

		for stepIdx, step := range steps {
			state := step.State.(renderers.ArrayState)
			if state.Caption == "" {
				t.Fatalf("step %d missing caption", stepIdx)
			}
			if len(state.IDs) != n || len(state.Values) != n {
				t.Fatalf("step %d lost array size: ids=%d values=%d expected=%d", stepIdx, len(state.IDs), len(state.Values), n)
			}

			seen := make(map[string]bool)
			for i, id := range state.IDs {
				if id == "" || seen[id] {
					t.Fatalf("step %d has duplicate or empty ID %q: %+v", stepIdx, id, state.IDs)
				}
				seen[id] = true
				if origValues[id] != state.Values[i] {
					t.Fatalf("step %d object %q changed value from %d to %d", stepIdx, id, origValues[id], state.Values[i])
				}
			}

			for _, sIdx := range state.Sorted {
				if sIdx < 0 || sIdx >= n {
					t.Fatalf("step %d has out-of-range sorted index %d", stepIdx, sIdx)
				}
			}
		}

		last := steps[len(steps)-1].State.(renderers.ArrayState)
		expectedSorted := Sort(input)
		if !slices.Equal(last.Values, expectedSorted) {
			t.Fatalf("final values not sorted: got %v, expected %v", last.Values, expectedSorted)
		}
		if len(last.Sorted) != n {
			t.Fatalf("final state sorted count = %d, expected %d", len(last.Sorted), n)
		}

		if n > 0 {
			origVal0 := first.Values[0]
			origID0 := first.IDs[0]
			last.Values[0] = 9999
			last.IDs[0] = "mutated-id"
			last.Sorted[0] = 9999

			if first.Values[0] != origVal0 || first.IDs[0] != origID0 {
				t.Fatalf("snapshot mutation leaked into earlier step: first=%+v", first)
			}
		}
	}
}

func TestQuickSortTeachingDecisions(t *testing.T) {
	input := []int{10, 80, 30, 90, 40, 50, 70}
	steps := Trace(input)

	var hasCompareLess, hasCompareGreater bool
	var hasPivotSwap, hasPartitionComplete bool

	for _, step := range steps {
		state := step.State.(renderers.ArrayState)

		// 1. Pivot visibility during partition
		if step.LineID == "L9" || step.LineID == "L10" || step.LineID == "L11" || step.LineID == "L12" || step.LineID == "L13" || step.LineID == "L14" {
			pivotIdx, ok := step.Variables["pivotIndex"].(int)
			if !ok {
				t.Fatalf("step %s missing pivotIndex in variables", step.LineID)
			}
			if state.Current != pivotIdx {
				t.Fatalf("step %s pivot element must be highlighted as Current: expected %d, got %d", step.LineID, pivotIdx, state.Current)
			}
		}

		// 2. Active partition range and scanning pointer j
		if step.LineID == "L11" || step.LineID == "L12" {
			j, ok := step.Variables["j"].(int)
			if !ok {
				t.Fatalf("step %s missing j in variables", step.LineID)
			}
			if !reflect.DeepEqual(state.Comparing, []int{j}) {
				t.Fatalf("step %s scanning pointer j must be highlighted in Comparing: expected [%d], got %v", step.LineID, j, state.Comparing)
			}
			if _, ok := step.Variables["low"]; !ok {
				t.Fatalf("step %s missing low in variables", step.LineID)
			}
			if _, ok := step.Variables["high"]; !ok {
				t.Fatalf("step %s missing high in variables", step.LineID)
			}
			if _, ok := step.Variables["i"]; !ok {
				t.Fatalf("step %s missing i in variables", step.LineID)
			}
		}

		// 3. Swap when arr[j] < pivot
		if step.LineID == "L14" && step.Action == trace.ActionSwap {
			hasCompareLess = true
			if len(state.Swapped) == 0 {
				t.Fatalf("step L14 swap must have Swapped indices set")
			}
			if !strings.Contains(state.Caption, "smaller") && !strings.Contains(step.Explanation, "smaller") {
				t.Fatalf("step L14 caption must explain smaller element movement: %q", state.Caption)
			}
		}

		if step.LineID == "L12" && step.Action == trace.ActionCompare && strings.Contains(state.Caption, "stays") {
			hasCompareGreater = true
		}

		// 4. Final swap of pivot and settled status
		if step.LineID == "L17" && step.Action == trace.ActionSwap {
			hasPivotSwap = true
			p, ok := step.Variables["p"].(int)
			if !ok {
				t.Fatalf("step L17 missing p in variables")
			}
			if !slices.Contains(state.Sorted, p) {
				t.Fatalf("step L17 pivot at index %d must be marked settled in Sorted: %v", p, state.Sorted)
			}
			if !slices.Contains(state.Swapped, p) {
				t.Fatalf("step L17 Swapped must contain pivot index %d: %v", p, state.Swapped)
			}
		}

		if step.LineID == "L18" && step.Action == trace.ActionSet {
			hasPartitionComplete = true
			p := step.Variables["p"].(int)
			if state.Current != p {
				t.Fatalf("step L18 Current must be pivot %d, got %d", p, state.Current)
			}
		}
	}

	if !hasCompareLess {
		t.Fatal("expected at least one swap step when arr[j] < pivot (L14)")
	}
	if !hasCompareGreater {
		t.Fatal("expected at least one branch where arr[j] >= pivot")
	}
	if !hasPivotSwap {
		t.Fatal("expected final pivot placement swap step (L17)")
	}
	if !hasPartitionComplete {
		t.Fatal("expected partition complete step (L18)")
	}
}

func TestQuickSortVisualRoles(t *testing.T) {
	steps := Trace([]int{10, 80, 30, 90, 40, 50, 70})

	var verifiedL12, verifiedL14, verifiedL17 bool

	for _, step := range steps {
		state := step.State.(renderers.ArrayState)

		if step.LineID == "L12" && len(state.Comparing) == 1 && state.Current >= 0 {
			svg := renderers.RenderArraySVG(step, 800, 400)
			pivotKey := fmt.Sprintf(`data-key="array-%s"`, state.IDs[state.Current])
			scanKey := fmt.Sprintf(`data-key="array-%s"`, state.IDs[state.Comparing[0]])

			if !strings.Contains(svg, pivotKey) || !strings.Contains(svg, scanKey) {
				t.Fatalf("missing data-key in SVG for step %d", step.Index)
			}

			pivotGroup := svg[strings.Index(svg, pivotKey):]
			pivotGroup = strings.SplitN(pivotGroup, "</g>", 2)[0]
			if !strings.Contains(pivotGroup, `fill="url(#bar-current)"`) {
				t.Fatalf("pivot element must have url(#bar-current) fill: %s", pivotGroup)
			}

			scanGroup := svg[strings.Index(svg, scanKey):]
			scanGroup = strings.SplitN(scanGroup, "</g>", 2)[0]
			if !strings.Contains(scanGroup, `fill="url(#bar-comp)"`) {
				t.Fatalf("scanning pointer element must have url(#bar-comp) fill: %s", scanGroup)
			}
			verifiedL12 = true
		}

		if step.LineID == "L14" && len(state.Swapped) == 2 && state.Swapped[0] != state.Swapped[1] {
			svg := renderers.RenderArraySVG(step, 800, 400)
			for _, swapIdx := range state.Swapped {
				swapKey := fmt.Sprintf(`data-key="array-%s"`, state.IDs[swapIdx])
				swapGroup := svg[strings.Index(svg, swapKey):]
				swapGroup = strings.SplitN(swapGroup, "</g>", 2)[0]
				if !strings.Contains(swapGroup, `fill="url(#bar-swap)"`) {
					t.Fatalf("swapped element %d must have url(#bar-swap) fill: %s", swapIdx, swapGroup)
				}
			}
			verifiedL14 = true
		}

		if step.LineID == "L17" {
			p := step.Variables["p"].(int)
			svg := renderers.RenderArraySVG(step, 800, 400)
			pivotKey := fmt.Sprintf(`data-key="array-%s"`, state.IDs[p])
			pivotGroup := svg[strings.Index(svg, pivotKey):]
			pivotGroup = strings.SplitN(pivotGroup, "</g>", 2)[0]
			if !strings.Contains(pivotGroup, `fill="url(#bar-sorted)"`) {
				t.Fatalf("settled pivot element must have url(#bar-sorted) fill: %s", pivotGroup)
			}
			verifiedL17 = true
		}
	}

	if !verifiedL12 {
		t.Fatal("failed to verify L12 visual roles in SVG")
	}
	if !verifiedL14 {
		t.Fatal("failed to verify L14 visual roles in SVG")
	}
	if !verifiedL17 {
		t.Fatal("failed to verify L17 visual roles in SVG")
	}
}
