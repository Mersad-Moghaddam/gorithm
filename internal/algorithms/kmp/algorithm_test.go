package kmp

import (
	"encoding/json"
	"reflect"
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
			name:     "found in middle",
			text:     "ABABDABACDABABCABAB",
			pattern:  "ABABCABAB",
			expected: 10,
		},
		{
			name:     "not found",
			text:     "ABABDABACDABABCABAB",
			pattern:  "XYZ",
			expected: -1,
		},
		{
			name:     "prefix match",
			text:     "HELLO WORLD",
			pattern:  "HELLO",
			expected: 0,
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

func TestKMPTraceValidation(t *testing.T) {
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

	// Verify last step is ActionDone
	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
}

func TestKMPTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"text":"AABAACAADAABAABA","pattern":"AABA"}`)
	steps, err := lesson.Build(json.RawMessage(input))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
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
	if foundIdx, ok := lastStep.Variables["matchIndex"].(int); !ok || foundIdx != 0 {
		t.Fatalf("expected matchIndex 0, got %v", lastStep.Variables["matchIndex"])
	}
}

func TestKMPTraceNotFound(t *testing.T) {
	steps := Trace("ABCDEF", "XYZ")
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
	if foundIdx, ok := lastStep.Variables["matchIndex"].(int); !ok || foundIdx != -1 {
		t.Fatalf("expected matchIndex -1, got %v", lastStep.Variables["matchIndex"])
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
		svg := renderers.RenderKMPLessonSVG(step, 800, 450)
		if svg == "" {
			t.Errorf("step %d produced empty SVG", i)
		}
	}
}
