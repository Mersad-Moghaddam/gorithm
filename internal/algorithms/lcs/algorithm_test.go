package lcs

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSolveCorrectness(t *testing.T) {
	tests := []struct {
		name       string
		s1         string
		s2         string
		wantLCS    string
		wantLength int
	}{
		{
			name:       "classic AGGTAB and GXTXAYB",
			s1:         "AGGTAB",
			s2:         "GXTXAYB",
			wantLCS:    "GTAB",
			wantLength: 4,
		},
		{
			name:       "identical strings",
			s1:         "DYNAMIC",
			s2:         "DYNAMIC",
			wantLCS:    "DYNAMIC",
			wantLength: 7,
		},
		{
			name:       "completely disjoint",
			s1:         "ABC",
			s2:         "DEF",
			wantLCS:    "",
			wantLength: 0,
		},
		{
			name:       "first empty",
			s1:         "",
			s2:         "ABC",
			wantLCS:    "",
			wantLength: 0,
		},
		{
			name:       "second empty",
			s1:         "XYZ",
			s2:         "",
			wantLCS:    "",
			wantLength: 0,
		},
		{
			name:       "both empty",
			s1:         "",
			s2:         "",
			wantLCS:    "",
			wantLength: 0,
		},
		{
			name:       "multiple valid LCS length check",
			s1:         "ABCBDAB",
			s2:         "BDCABA",
			wantLength: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLCS, gotDP := Solve(tt.s1, tt.s2)

			if len(gotDP) != len(tt.s1)+1 {
				t.Fatalf("Solve(%q, %q) DP rows = %d, want %d", tt.s1, tt.s2, len(gotDP), len(tt.s1)+1)
			}
			for r, row := range gotDP {
				if len(row) != len(tt.s2)+1 {
					t.Fatalf("Solve(%q, %q) DP row %d len = %d, want %d", tt.s1, tt.s2, r, len(row), len(tt.s2)+1)
				}
			}

			endVal := gotDP[len(tt.s1)][len(tt.s2)]
			if endVal != tt.wantLength {
				t.Errorf("Solve(%q, %q) DP[%d][%d] = %d, want %d", tt.s1, tt.s2, len(tt.s1), len(tt.s2), endVal, tt.wantLength)
			}
			if len(gotLCS) != tt.wantLength {
				t.Errorf("Solve(%q, %q) got LCS length = %d (%q), want %d", tt.s1, tt.s2, len(gotLCS), gotLCS, tt.wantLength)
			}
			if tt.wantLCS != "" && gotLCS != tt.wantLCS {
				t.Errorf("Solve(%q, %q) got LCS = %q, want %q", tt.s1, tt.s2, gotLCS, tt.wantLCS)
			}
		})
	}
}

func TestLCSLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "lcs" {
		t.Errorf("expected ID 'lcs', got %q", meta.ID)
	}
	if meta.Name != "Longest Common Subsequence" {
		t.Errorf("expected Name 'Longest Common Subsequence', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(m · n)" {
		t.Errorf("expected TimeComplexity 'O(m · n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(m · n)" {
		t.Errorf("expected SpaceComplexity 'O(m · n)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{
		"2D Table",
		"Optimal Substructure",
		"Character Match/Mismatch Transition",
		"Backtracking",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("lcs")
	if !ok {
		t.Fatal("lesson 'lcs' not registered in catalog")
	}
	if l.Meta().ID != "lcs" {
		t.Errorf("expected registered ID 'lcs', got %q", l.Meta().ID)
	}
}

func TestLCSTraceValidation(t *testing.T) {
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
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
	if lcsVal, ok := lastStep.Variables["lcs"].(string); !ok || lcsVal != "GTAB" {
		t.Fatalf("expected lcs 'GTAB', got %v", lastStep.Variables["lcs"])
	}
}

func TestLCSTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"s1": "ABCDGH", "s2": "AEDFHR"}`)
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
	if lcsVal, ok := lastStep.Variables["lcs"].(string); !ok || lcsVal != "ADH" {
		t.Fatalf("expected lcs 'ADH', got %v", lastStep.Variables["lcs"])
	}
}

func TestLCSRenderTableSVG(t *testing.T) {
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
