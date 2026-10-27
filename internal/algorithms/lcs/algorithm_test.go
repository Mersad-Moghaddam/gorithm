package lcs

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

func TestLCSInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input string
	}{
		{"null input", `null`},
		{"malformed json object", `{"s1": "A"`},
		{"malformed json array", `["A"`},
		{"numeric json", `123`},
		{"boolean json", `true`},
		{"string json", `"hello"`},
		{"empty object", `{}`},
		{"empty s1 in object", `{"s1": "", "s2": "ABC"}`},
		{"empty s2 in object", `{"s1": "ABC", "s2": ""}`},
		{"both empty in object", `{"s1": "", "s2": ""}`},
		{"empty text in object", `{"text": "", "pattern": "ABC"}`},
		{"empty pattern in object", `{"text": "ABC", "pattern": ""}`},
		{"empty array", `[]`},
		{"single element array", `["ABC"]`},
		{"empty s1 in array", `["", "ABC"]`},
		{"empty s2 in array", `["ABC", ""]`},
		{"both empty in array", `["", ""]`},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Errorf("Build(%s) expected error, got nil", tc.input)
			}
		})
	}

	validInputs := []struct {
		name    string
		input   string
		wantLCS string
	}{
		{"valid s1/s2 object", `{"s1": "ABC", "s2": "AC"}`, "AC"},
		{"valid text/pattern object", `{"text": "ABC", "pattern": "BC"}`, "BC"},
		{"valid array", `["ABC", "BC"]`, "BC"},
		{"nil input uses default", ``, "GTAB"},
	}

	for _, tc := range validInputs {
		t.Run(tc.name, func(t *testing.T) {
			var raw json.RawMessage
			if tc.input != "" {
				raw = json.RawMessage(tc.input)
			}
			steps, err := lesson.Build(raw)
			if err != nil {
				t.Fatalf("Build(%s) unexpected error: %v", tc.input, err)
			}
			lastStep := steps[len(steps)-1]
			if lcsVal, ok := lastStep.Variables["lcs"].(string); !ok || lcsVal != tc.wantLCS {
				t.Errorf("Build(%s) got LCS %v, want %q", tc.input, lastStep.Variables["lcs"], tc.wantLCS)
			}
		})
	}
}

func TestLCSTableStateAndDependencies(t *testing.T) {
	s1 := "AGGTAB"
	s2 := "GXTXAYB"
	steps := Trace(s1, s2)

	m := len(s1)
	n := len(s2)

	foundMatch := false
	foundMismatch := false
	foundBacktrack := false

	for stepIdx, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", stepIdx, step.State)
		}

		// Verify headers
		if len(tableState.RowHeaders) != m+1 {
			t.Fatalf("step %d: RowHeaders len = %d, want %d", stepIdx, len(tableState.RowHeaders), m+1)
		}
		if tableState.RowHeaders[0] != "-" {
			t.Fatalf("step %d: RowHeaders[0] = %q, want '-'", stepIdx, tableState.RowHeaders[0])
		}
		for i := 0; i < m; i++ {
			if tableState.RowHeaders[i+1] != string(s1[i]) {
				t.Fatalf("step %d: RowHeaders[%d] = %q, want %q", stepIdx, i+1, tableState.RowHeaders[i+1], string(s1[i]))
			}
		}

		if len(tableState.ColHeaders) != n+1 {
			t.Fatalf("step %d: ColHeaders len = %d, want %d", stepIdx, len(tableState.ColHeaders), n+1)
		}
		if tableState.ColHeaders[0] != "-" {
			t.Fatalf("step %d: ColHeaders[0] = %q, want '-'", stepIdx, tableState.ColHeaders[0])
		}
		for j := 0; j < n; j++ {
			if tableState.ColHeaders[j+1] != string(s2[j]) {
				t.Fatalf("step %d: ColHeaders[%d] = %q, want %q", stepIdx, j+1, tableState.ColHeaders[j+1], string(s2[j]))
			}
		}

		// Verify cells count
		expectedCellCount := (m + 1) * (n + 1)
		if len(tableState.Cells) != expectedCellCount {
			t.Fatalf("step %d: Cells len = %d, want %d", stepIdx, len(tableState.Cells), expectedCellCount)
		}

		// Verify Match cell visualization (L7 ActionCompare when s1[i-1] == s2[j-1])
		if step.LineID == "L7" && strings.HasPrefix(tableState.Caption, "Match") {
			foundMatch = true
			if len(tableState.Dependencies) != 1 {
				t.Fatalf("step %d: match step dependencies len = %d, want 1", stepIdx, len(tableState.Dependencies))
			}
			dep := tableState.Dependencies[0]
			i := step.Variables["i"].(int)
			j := step.Variables["j"].(int)
			if dep.FromRow != i-1 || dep.FromCol != j-1 || dep.ToRow != i || dep.ToCol != j {
				t.Errorf("step %d: match arrow mismatch: got (%d,%d)->(%d,%d), want (%d,%d)->(%d,%d)",
					stepIdx, dep.FromRow, dep.FromCol, dep.ToRow, dep.ToCol, i-1, j-1, i, j)
			}
			expectedCaption := fmt.Sprintf("Match '%s' == '%s': extend diagonal LCS length by 1",
				step.Variables["char1"], step.Variables["char2"])
			if tableState.Caption != expectedCaption {
				t.Errorf("step %d: match caption = %q, want %q", stepIdx, tableState.Caption, expectedCaption)
			}
		}

		// Verify Mismatch cell visualization (L7 ActionCompare when s1[i-1] != s2[j-1])
		if step.LineID == "L7" && strings.HasPrefix(tableState.Caption, "Mismatch") {
			foundMismatch = true
			if len(tableState.Dependencies) != 2 {
				t.Fatalf("step %d: mismatch step dependencies len = %d, want 2", stepIdx, len(tableState.Dependencies))
			}
			i := step.Variables["i"].(int)
			j := step.Variables["j"].(int)

			topDep := tableState.Dependencies[0]
			leftDep := tableState.Dependencies[1]

			if topDep.FromRow != i-1 || topDep.FromCol != j || topDep.ToRow != i || topDep.ToCol != j {
				t.Errorf("step %d: top arrow mismatch: got (%d,%d)->(%d,%d), want (%d,%d)->(%d,%d)",
					stepIdx, topDep.FromRow, topDep.FromCol, topDep.ToRow, topDep.ToCol, i-1, j, i, j)
			}
			if leftDep.FromRow != i || leftDep.FromCol != j-1 || leftDep.ToRow != i || leftDep.ToCol != j {
				t.Errorf("step %d: left arrow mismatch: got (%d,%d)->(%d,%d), want (%d,%d)->(%d,%d)",
					stepIdx, leftDep.FromRow, leftDep.FromCol, leftDep.ToRow, leftDep.ToCol, i, j-1, i, j)
			}
		}

		// Verify Backtrack step
		if step.LineID == "L17" {
			foundBacktrack = true
			if len(tableState.Dependencies) != 1 {
				t.Fatalf("step %d: backtrack match dependencies len = %d, want 1", stepIdx, len(tableState.Dependencies))
			}
			i := step.Variables["i"].(int)
			j := step.Variables["j"].(int)
			dep := tableState.Dependencies[0]
			if dep.FromRow != i || dep.FromCol != j || dep.ToRow != i-1 || dep.ToCol != j-1 {
				t.Errorf("step %d: backtrack diagonal arrow got (%d,%d)->(%d,%d), want (%d,%d)->(%d,%d)",
					stepIdx, dep.FromRow, dep.FromCol, dep.ToRow, dep.ToCol, i, j, i-1, j-1)
			}
		}
	}

	if !foundMatch {
		t.Error("no match step evaluated in trace")
	}
	if !foundMismatch {
		t.Error("no mismatch step evaluated in trace")
	}
	if !foundBacktrack {
		t.Error("no backtrack match step evaluated in trace")
	}

	// Verify final step state
	lastStep := steps[len(steps)-1]
	lastState := lastStep.State.(renderers.TableState)
	computedCount := 0
	for _, cell := range lastState.Cells {
		if cell.State == "computed" {
			computedCount++
		}
		if cell.State == "active" {
			t.Errorf("final step should have no active cell, found at (%d,%d)", cell.Row, cell.Col)
		}
	}
	if computedCount != 4 { // "GTAB" length is 4
		t.Errorf("final step computed cells = %d, want 4", computedCount)
	}
}

func TestLCSSnapshotImmutability(t *testing.T) {
	steps := Trace("AGGTAB", "GXTXAYB")
	if len(steps) < 5 {
		t.Fatalf("too few steps: %d", len(steps))
	}

	firstStep := steps[0]
	firstState := firstStep.State.(renderers.TableState)

	var matchStep trace.Step
	var mismatchStep trace.Step
	matchIdx := -1
	mismatchIdx := -1
	for idx, s := range steps {
		st := s.State.(renderers.TableState)
		if len(st.Dependencies) == 1 && s.LineID == "L7" && matchIdx == -1 {
			matchStep = s
			matchIdx = idx
		}
		if len(st.Dependencies) == 2 && s.LineID == "L7" && mismatchIdx == -1 {
			mismatchStep = s
			mismatchIdx = idx
		}
	}
	if matchIdx == -1 || mismatchIdx == -1 {
		t.Fatal("missing match or mismatch step for immutability test")
	}

	matchState := matchStep.State.(renderers.TableState)
	mismatchState := mismatchStep.State.(renderers.TableState)
	lastState := steps[len(steps)-1].State.(renderers.TableState)

	origFirstCellVal := firstState.Cells[0].Value
	origFirstCellState := firstState.Cells[0].State
	origMatchDepFromRow := matchState.Dependencies[0].FromRow
	origMismatchDepFromCol := mismatchState.Dependencies[0].FromCol

	// Mutate lastState Cells
	lastState.Cells[0].Value = "MUTATED"
	lastState.Cells[0].State = "corrupted"

	if firstState.Cells[0].Value != origFirstCellVal || firstState.Cells[0].State != origFirstCellState {
		t.Fatalf("mutating lastState Cells modified firstState Cells: got val=%q, state=%q",
			firstState.Cells[0].Value, firstState.Cells[0].State)
	}
	if matchState.Cells[0].Value == "MUTATED" || matchState.Cells[0].State == "corrupted" {
		t.Fatal("mutating lastState Cells modified matchState Cells")
	}

	// Mutate Dependencies
	matchState.Dependencies[0].FromRow = 999
	mismatchState.Dependencies[0].FromCol = 888

	if origMatchDepFromRow == 999 {
		t.Fatal("unexpected initial match dep fromRow")
	}
	for i, s := range steps {
		st := s.State.(renderers.TableState)
		// Check that other steps were not mutated
		if i != matchIdx && len(st.Dependencies) > 0 {
			if st.Dependencies[0].FromRow == 999 {
				t.Fatalf("step %d dependencies mutated by matchStep mutation", i)
			}
		}
		if i != mismatchIdx && len(st.Dependencies) > 0 {
			if st.Dependencies[0].FromCol == 888 && origMismatchDepFromCol != 888 {
				t.Fatalf("step %d dependencies mutated by mismatchStep mutation", i)
			}
		}
	}
}

func TestLCSStringReconstruction(t *testing.T) {
	testCases := []struct {
		s1      string
		s2      string
		wantLCS string
	}{
		{"AGGTAB", "GXTXAYB", "GTAB"},
		{"ABCDGH", "AEDFHR", "ADH"},
		{"DYNAMIC", "DYNAMIC", "DYNAMIC"},
		{"ABC", "DEF", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.s1+"_"+tc.s2, func(t *testing.T) {
			steps := Trace(tc.s1, tc.s2)
			lastStep := steps[len(steps)-1]

			if lastStep.Action != trace.ActionDone {
				t.Fatalf("expected last action ActionDone, got %s", lastStep.Action)
			}

			lcsVal, ok := lastStep.Variables["lcs"].(string)
			if !ok {
				t.Fatalf("last step missing lcs variable: %v", lastStep.Variables)
			}
			if lcsVal != tc.wantLCS {
				t.Errorf("Trace(%q, %q) reconstructed LCS = %q, want %q", tc.s1, tc.s2, lcsVal, tc.wantLCS)
			}

			lcsLen, ok := lastStep.Variables["lcsLength"].(int)
			if !ok || lcsLen != len(tc.wantLCS) {
				t.Errorf("Trace(%q, %q) lcsLength = %v, want %d", tc.s1, tc.s2, lcsLen, len(tc.wantLCS))
			}

			// Verify progressive reconstruction during backtrack match steps
			reconstructedSoFar := ""
			for _, step := range steps {
				if step.LineID == "L17" {
					ch := step.Variables["char"].(string)
					reconstructedSoFar = ch + reconstructedSoFar
					stepLCS := step.Variables["lcs"].(string)
					if stepLCS != reconstructedSoFar {
						t.Errorf("step L17 lcs variable = %q, want %q", stepLCS, reconstructedSoFar)
					}
				}
			}
			if reconstructedSoFar != tc.wantLCS {
				t.Errorf("total reconstructed from L17 steps = %q, want %q", reconstructedSoFar, tc.wantLCS)
			}
		})
	}
}
