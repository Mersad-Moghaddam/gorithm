package insertion

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
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
		{
			name:     "negative values",
			input:    []int{0, -5, 3, -2, 4},
			expected: []int{-5, -2, 0, 3, 4},
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

func TestTeachingTrace(t *testing.T) {
	cases := [][]int{
		nil,
		{},
		{7},
		{2, 1, 2, 1},
		{1, 2, 3},
		{3, 2, 1},
		{0, -1, -1},
		{5, 4, 3, 2, 1},
	}

	for _, input := range cases {
		steps := Trace(input)
		expected := Sort(input)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatalf("line validation failed for input %v: %v", input, err)
		}

		first := steps[0].State.(renderers.ArrayState)
		if !slices.Equal(first.Values, input) && !(len(input) == 0 && len(first.Values) == 0) {
			t.Fatalf("initial values mismatch: got %v, want %v", first.Values, input)
		}
		if len(first.IDs) != len(input) {
			t.Fatalf("initial IDs count mismatch: got %d, want %d", len(first.IDs), len(input))
		}

		objects := make(map[string]int, len(first.IDs))
		for i, id := range first.IDs {
			if id == "" {
				t.Fatalf("empty ID at index %d", i)
			}
			if _, exists := objects[id]; exists {
				t.Fatalf("duplicate initial ID: %s", id)
			}
			objects[id] = input[i]
		}

		for stepIdx, step := range steps {
			s := step.State.(renderers.ArrayState)
			if s.Caption == "" {
				t.Fatalf("step %d missing caption", stepIdx)
			}
			if len(s.IDs) != len(input) {
				t.Fatalf("step %d lost IDs: got %d, want %d", stepIdx, len(s.IDs), len(input))
			}

			seenIDs := make(map[string]bool, len(s.IDs))
			for i, id := range s.IDs {
				if id == "" || seenIDs[id] {
					t.Fatalf("step %d has duplicate or empty ID %q", stepIdx, id)
				}
				seenIDs[id] = true
				if objects[id] != s.Values[i] {
					t.Fatalf("object %s changed value from %d to %d", id, objects[id], s.Values[i])
				}
			}

			for _, sortedIdx := range s.Sorted {
				if sortedIdx < 0 || sortedIdx >= len(input) {
					t.Fatalf("step %d has out-of-bounds sorted index %d", stepIdx, sortedIdx)
				}
			}

			if len(input) > 0 {
				svg := renderers.RenderArraySVG(step, 800, 400)
				for _, id := range s.IDs {
					if strings.Count(svg, `data-key="array-`+id+`"`) != 1 {
						t.Fatalf("step %d SVG missing or repeated element ID %q", stepIdx, id)
					}
				}
			}
		}

		last := steps[len(steps)-1].State.(renderers.ArrayState)
		if !slices.Equal(last.Values, expected) && !(len(expected) == 0 && len(last.Values) == 0) {
			t.Fatalf("final state not sorted: got %v, want %v", last.Values, expected)
		}
		if len(last.Sorted) != len(input) {
			t.Fatalf("final state sorted count mismatch: got %d, want %d", len(last.Sorted), len(input))
		}

		// Stability check: duplicate values retain initial relative ID order
		for _, v := range input {
			var beforeIDs, afterIDs []string
			for i, val := range first.Values {
				if val == v {
					beforeIDs = append(beforeIDs, first.IDs[i])
				}
			}
			for i, val := range last.Values {
				if val == v {
					afterIDs = append(afterIDs, last.IDs[i])
				}
			}
			if !slices.Equal(beforeIDs, afterIDs) {
				t.Fatalf("stability broken for value %d: before=%v after=%v", v, beforeIDs, afterIDs)
			}
		}

		// Immutability check
		if len(input) > 0 {
			last.Values[0] = 999
			last.IDs[0] = "corrupted"
			if len(last.Sorted) > 0 {
				last.Sorted[0] = 999
			}
			if slices.Equal(first.Values, last.Values) || first.IDs[0] == "corrupted" {
				t.Fatalf("snapshot history aliased final state for input %v", input)
			}
		}
	}
}

func TestInsertionVisualRoles(t *testing.T) {
	steps := Trace([]int{3, 1, 2})
	var verifiedCompare, verifiedSwap, verifiedSortedPrefix bool

	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		svg := renderers.RenderArraySVG(step, 800, 400)

		if step.LineID == "L6" && step.Action == trace.ActionCompare && len(state.Comparing) == 1 && state.Current >= 0 {
			heldKeyG := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Current]+`"`):]
			compG := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Comparing[0]]+`"`):]

			if strings.Contains(strings.SplitN(heldKeyG, "</g>", 2)[0], `fill="url(#bar-current)"`) &&
				strings.Contains(strings.SplitN(compG, "</g>", 2)[0], `fill="url(#bar-comp)"`) {
				verifiedCompare = true
			}
		}

		if step.LineID == "L7" && step.Action == trace.ActionSwap && len(state.Swapped) == 2 {
			g1 := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Swapped[0]]+`"`):]
			g2 := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Swapped[1]]+`"`):]

			if strings.Contains(strings.SplitN(g1, "</g>", 2)[0], `fill="url(#bar-swap)"`) &&
				strings.Contains(strings.SplitN(g2, "</g>", 2)[0], `fill="url(#bar-swap)"`) {
				verifiedSwap = true
			}
		}

		if step.LineID == "L11" && len(state.Sorted) > 0 {
			allGreen := true
			for _, sortedIdx := range state.Sorted {
				sortedG := svg[strings.Index(svg, `data-key="array-`+state.IDs[sortedIdx]+`"`):]
				if !strings.Contains(strings.SplitN(sortedG, "</g>", 2)[0], `fill="url(#bar-sorted)"`) {
					allGreen = false
					break
				}
			}
			if allGreen {
				verifiedSortedPrefix = true
			}
		}
	}

	if !verifiedCompare {
		t.Error("held key (purple) and compared element (yellow) visual roles not verified")
	}
	if !verifiedSwap {
		t.Error("shifted elements (red) visual roles not verified")
	}
	if !verifiedSortedPrefix {
		t.Error("growing sorted prefix (green) visual roles not verified")
	}
}

func TestInsertionSortTraceMetrics(t *testing.T) {
	steps := Trace([]int{3, 2, 1})
	lastStep := steps[len(steps)-1]

	if lastStep.Metrics.Comparisons != 3 {
		t.Errorf("expected 3 comparisons for [3, 2, 1], got %d", lastStep.Metrics.Comparisons)
	}
	if lastStep.Metrics.Swaps != 3 {
		t.Errorf("expected 3 swaps/shifts for [3, 2, 1], got %d", lastStep.Metrics.Swaps)
	}
	if lastStep.Metrics.Operations != 8 {
		t.Errorf("expected 8 total operations for [3, 2, 1], got %d", lastStep.Metrics.Operations)
	}
}

func TestInsertionBuildInput(t *testing.T) {
	lesson := Lesson{}
	invalidInputs := []json.RawMessage{
		[]byte(`not json`),
		[]byte(`{"bad":1}`),
		[]byte(`[1,"two"]`),
		[]byte(`null`),
		[]byte(`[1, 2.5]`),
		[]byte(`[null]`),
		[]byte(`[1, null]`),
	}

	for _, input := range invalidInputs {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("expected error for invalid input %s, got nil", input)
		}
	}

	// Valid empty array
	emptySteps, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatalf("failed to build trace with empty array: %v", err)
	}
	if len(emptySteps) == 0 {
		t.Fatal("empty input produced 0 steps")
	}
	if got := emptySteps[len(emptySteps)-1].State.(renderers.ArrayState).Values; len(got) != 0 {
		t.Fatalf("empty input produced non-empty state: %v", got)
	}

	// Default input
	defaultSteps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace with default input: %v", err)
	}
	if len(defaultSteps) == 0 {
		t.Fatal("default input produced 0 steps")
	}

	// Custom valid input
	customSteps, err := lesson.Build(json.RawMessage(`[5, 4, 3, 2, 1]`))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(customSteps) == 0 {
		t.Fatal("custom input produced 0 steps")
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
	if meta.Documentation.TimeComplexity != "O(n²)" {
		t.Errorf("expected TimeComplexity 'O(n²)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}
	expectedConcepts := []string{"Incremental Build", "Adaptive", "In-Place", "Stable"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}
