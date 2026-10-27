package kmp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestComputeLPS(t *testing.T) {
	tests := []struct {
		pattern  string
		expected []int
	}{
		{"ABABCABAB", []int{0, 0, 1, 2, 0, 1, 2, 3, 4}},
		{"AAAA", []int{0, 1, 2, 3}},
		{"ABCD", []int{0, 0, 0, 0}},
		{"AABAACAABAA", []int{0, 1, 0, 1, 2, 0, 1, 2, 3, 4, 5}},
		{"AAABAAAA", []int{0, 1, 2, 0, 1, 2, 3, 3}},
		{"", []int{}},
	}

	for _, tt := range tests {
		got := ComputeLPS(tt.pattern)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("ComputeLPS(%q) = %v; want %v", tt.pattern, got, tt.expected)
		}
	}
}

func TestKMPSearchCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected int
	}{
		{
			name:     "pattern at start",
			text:     "HELLO WORLD",
			pattern:  "HELLO",
			expected: 0,
		},
		{
			name:     "pattern in middle",
			text:     "ABABDABACDABABCABAB",
			pattern:  "ABABCABAB",
			expected: 10,
		},
		{
			name:     "pattern at end",
			text:     "SEARCH TARGET",
			pattern:  "TARGET",
			expected: 7,
		},
		{
			name:     "pattern not found",
			text:     "ABABDABACDABABCABAB",
			pattern:  "XYZ",
			expected: -1,
		},
		{
			name:     "full match",
			text:     "ABC",
			pattern:  "ABC",
			expected: 0,
		},
		{
			name:     "empty pattern",
			text:     "ABC",
			pattern:  "",
			expected: 0,
		},
		{
			name:     "pattern longer than text",
			text:     "A",
			pattern:  "AAA",
			expected: -1,
		},
		{
			name:     "repeating characters with overlap",
			text:     "AAAAAB",
			pattern:  "AAAAB",
			expected: 1,
		},
		{
			name:     "repeated prefix pattern AAAA in AAAAAAAA",
			text:     "AAAAAAAA",
			pattern:  "AAAA",
			expected: 0,
		},
		{
			name:     "repeated prefix pattern AAAA in BAAAA",
			text:     "BAAAA",
			pattern:  "AAAA",
			expected: 1,
		},
		{
			name:     "repeating partial overlap ABABABAC",
			text:     "ABABABAC",
			pattern:  "ABAC",
			expected: 4,
		},
		{
			name:     "single character match",
			text:     "X",
			pattern:  "X",
			expected: 0,
		},
		{
			name:     "single character mismatch",
			text:     "X",
			pattern:  "Y",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Search(tt.text, tt.pattern)
			if got != tt.expected {
				t.Fatalf("Search(%q, %q) = %d; want %d", tt.text, tt.pattern, got, tt.expected)
			}
		})
	}
}

func TestKMPTraceScenarios(t *testing.T) {
	scenarios := []struct {
		name              string
		text              string
		pattern           string
		expectedMatchIdx  int
		expectedFound     bool
		requireBacktrackMsg bool
	}{
		{
			name:              "pattern at start",
			text:              "HELLO WORLD",
			pattern:           "HELLO",
			expectedMatchIdx:  0,
			expectedFound:     true,
			requireBacktrackMsg: false,
		},
		{
			name:              "pattern in middle",
			text:              "ABABDABACDABABCABAB",
			pattern:           "ABABCABAB",
			expectedMatchIdx:  10,
			expectedFound:     true,
			requireBacktrackMsg: true,
		},
		{
			name:              "pattern at end",
			text:              "FIND THE END",
			pattern:           "END",
			expectedMatchIdx:  9,
			expectedFound:     true,
			requireBacktrackMsg: true,
		},
		{
			name:              "pattern not found",
			text:              "ABCDEF",
			pattern:           "XYZ",
			expectedMatchIdx:  -1,
			expectedFound:     false,
			requireBacktrackMsg: false,
		},
		{
			name:              "repeated prefix AAAA at offset",
			text:              "BAAAAA",
			pattern:           "AAAA",
			expectedMatchIdx:  1,
			expectedFound:     true,
			requireBacktrackMsg: false,
		},
		{
			name:              "repeated prefix AAAA in AAAAAB",
			text:              "AAAAAB",
			pattern:           "AAAAB",
			expectedMatchIdx:  1,
			expectedFound:     true,
			requireBacktrackMsg: true,
		},
	}

	lesson := Lesson{}
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			steps := Trace(sc.text, sc.pattern)
			if len(steps) == 0 {
				t.Fatal("trace produced 0 steps")
			}

			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("line validation failed: %v", err)
			}

			lastStep := steps[len(steps)-1]
			if lastStep.Action != trace.ActionDone {
				t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
			}

			matchIdx, ok := lastStep.Variables["matchIndex"].(int)
			if !ok || matchIdx != sc.expectedMatchIdx {
				t.Fatalf("expected matchIndex %d, got %v", sc.expectedMatchIdx, lastStep.Variables["matchIndex"])
			}

			sawBacktrackExplanation := false
			for i, step := range steps {
				ts, ok := step.State.(renderers.TableState)
				if !ok {
					t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
				}
				if ts.Caption == "" {
					t.Errorf("step %d has empty Caption", i)
				}
				if len(ts.RowHeaders) != 3 {
					t.Errorf("step %d expected 3 row headers, got %d", i, len(ts.RowHeaders))
				}
				if len(ts.ColHeaders) < len(sc.text) {
					t.Errorf("step %d expected at least %d col headers, got %d", i, len(sc.text), len(ts.ColHeaders))
				}
				for _, cell := range ts.Cells {
					switch cell.State {
					case "normal", "active", "computed", "source":
					default:
						t.Errorf("step %d invalid cell state %q", i, cell.State)
					}
				}
				if strings.Contains(strings.ToLower(step.Explanation), "backtrack") {
					sawBacktrackExplanation = true
				}
			}

			if sc.requireBacktrackMsg && !sawBacktrackExplanation {
				t.Errorf("scenario %s expected explanation explaining why text pointer does not backtrack", sc.name)
			}
		})
	}
}

func TestSnapshotImmutability(t *testing.T) {
	steps := Trace("ABABDABACDABABCABAB", "ABABCABAB")
	if len(steps) < 5 {
		t.Fatalf("trace produced too few steps: %d", len(steps))
	}

	firstState := steps[0].State.(renderers.TableState)
	midIdx := len(steps) / 2
	midState := steps[midIdx].State.(renderers.TableState)
	lastIdx := len(steps) - 1
	lastState := steps[lastIdx].State.(renderers.TableState)

	// Save copies of initial state contents
	firstRow0 := firstState.RowHeaders[0]
	firstCol0 := firstState.ColHeaders[0]
	firstCell0Val := firstState.Cells[0].Value
	firstCell0State := firstState.Cells[0].State

	midRow0 := midState.RowHeaders[0]
	midCol0 := midState.ColHeaders[0]
	midCell0Val := midState.Cells[0].Value
	midCell0State := midState.Cells[0].State

	// Mutate final step's slices
	lastState.RowHeaders[0] = "CORRUPTED"
	lastState.ColHeaders[0] = "9999"
	lastState.Cells[0].Value = "MUTATED"
	lastState.Cells[0].State = "corrupted"

	// Verify earlier steps were not affected
	if firstState.RowHeaders[0] != firstRow0 || firstState.ColHeaders[0] != firstCol0 ||
		firstState.Cells[0].Value != firstCell0Val || firstState.Cells[0].State != firstCell0State {
		t.Fatalf("mutating final step corrupted step 0: %+v", firstState)
	}

	if midState.RowHeaders[0] != midRow0 || midState.ColHeaders[0] != midCol0 ||
		midState.Cells[0].Value != midCell0Val || midState.Cells[0].State != midCell0State {
		t.Fatalf("mutating final step corrupted intermediate step %d: %+v", midIdx, midState)
	}

	// Mutate step 0 and verify intermediate step is not affected
	firstState.RowHeaders[1] = "CHANGED"
	firstState.ColHeaders[1] = "8888"
	firstState.Cells[1].Value = "CHANGED"
	if midState.RowHeaders[1] == "CHANGED" || midState.ColHeaders[1] == "8888" || midState.Cells[1].Value == "CHANGED" {
		t.Fatalf("mutating step 0 corrupted intermediate step %d", midIdx)
	}
}

func TestKMPInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid inputs: nil or empty defaults to standard sample
	for _, validEmpty := range []json.RawMessage{nil, []byte(""), []byte("   ")} {
		steps, err := lesson.Build(validEmpty)
		if err != nil {
			t.Fatalf("expected nil error for empty input %q, got %v", string(validEmpty), err)
		}
		if len(steps) == 0 {
			t.Fatal("expected steps for default input")
		}
	}

	// Valid custom object input
	objInput := []byte(`{"text":"ABCDE","pattern":"CDE"}`)
	steps, err := lesson.Build(objInput)
	if err != nil {
		t.Fatalf("expected nil error for valid object input, got %v", err)
	}
	if last := steps[len(steps)-1]; last.Variables["matchIndex"] != 2 {
		t.Fatalf("expected matchIndex 2, got %v", last.Variables["matchIndex"])
	}

	// Valid custom array input
	arrInput := []byte(`["ABCDE","CDE"]`)
	steps, err = lesson.Build(arrInput)
	if err != nil {
		t.Fatalf("expected nil error for valid array input, got %v", err)
	}
	if last := steps[len(steps)-1]; last.Variables["matchIndex"] != 2 {
		t.Fatalf("expected matchIndex 2, got %v", last.Variables["matchIndex"])
	}

	// Invalid inputs that must be rejected
	invalidCases := []struct {
		name  string
		input []byte
	}{
		{"null literal", []byte("null")},
		{"whitespace null", []byte("   null   ")},
		{"malformed JSON syntax", []byte("{ invalid json")},
		{"malformed JSON array syntax", []byte("[1, 2,")},
		{"numeric literal", []byte("12345")},
		{"string literal", []byte(`"just a string"`)},
		{"empty text in object", []byte(`{"text":"","pattern":"abc"}`)},
		{"empty pattern in object", []byte(`{"text":"abc","pattern":""}`)},
		{"empty text and pattern in object", []byte(`{"text":"","pattern":""}`)},
		{"empty JSON object", []byte(`{}`)},
		{"array with fewer than 2 elements", []byte(`["only_text"]`)},
		{"empty text in array", []byte(`["","pattern"]`)},
		{"empty pattern in array", []byte(`["text",""]`)},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected error for input %s (%s), got nil", tc.name, string(tc.input))
			}
		})
	}
}

func TestKMPMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "kmp" {
		t.Errorf("expected ID 'kmp', got %q", meta.ID)
	}
	if meta.Name != "KMP String Matching" {
		t.Errorf("expected Name 'KMP String Matching', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySearching {
		t.Errorf("expected Category %q, got %q", catalog.CategorySearching, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n + m)" {
		t.Errorf("expected TimeComplexity 'O(n + m)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(m)" {
		t.Errorf("expected SpaceComplexity 'O(m)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Longest Proper Prefix/Suffix", "No Backtracking", "Preprocessing", "Linear Search in Strings"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("kmp")
	if !ok {
		t.Fatal("lesson 'kmp' not registered in catalog")
	}
	if l.Meta().ID != "kmp" {
		t.Errorf("expected registered ID 'kmp', got %q", l.Meta().ID)
	}
}

func TestKMPRenderTableSVG(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	for i, step := range steps {
		svg := renderers.RenderTableSVG(step, 800, 450)
		if svg == "" {
			t.Errorf("step %d produced empty SVG", i)
		}
	}
}
