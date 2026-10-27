package floydwarshall

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestFloydWarshallAllPairsDistance(t *testing.T) {
	nodes := []string{"A", "B", "C", "D"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 3, Directed: true},
		{From: "A", To: "C", Weight: 8, Directed: true},
		{From: "B", To: "C", Weight: 2, Directed: true},
		{From: "B", To: "D", Weight: 5, Directed: true},
		{From: "C", To: "D", Weight: 1, Directed: true},
		{From: "D", To: "A", Weight: 2, Directed: true},
	}

	dist, hasNegCycle := AllPairsShortestPath(nodes, edges)
	if hasNegCycle {
		t.Fatalf("expected no negative cycle, got hasNegCycle=true")
	}

	expected := [][]int{
		{0, 3, 5, 6},
		{5, 0, 2, 3},
		{3, 6, 0, 1},
		{2, 5, 7, 0},
	}

	if !reflect.DeepEqual(dist, expected) {
		t.Fatalf("dist = %v; want %v", dist, expected)
	}
}

func TestFloydWarshallNegativeCycle(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 1, Directed: true},
		{From: "B", To: "C", Weight: -3, Directed: true},
		{From: "C", To: "A", Weight: -1, Directed: true},
	}

	_, hasNegCycle := AllPairsShortestPath(nodes, edges)
	if !hasNegCycle {
		t.Fatalf("expected negative cycle detection, got false")
	}
}

func TestFloydWarshallUnreachableAndUndirected(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 4, Directed: false},
	}

	dist, hasNegCycle := AllPairsShortestPath(nodes, edges)
	if hasNegCycle {
		t.Fatalf("unexpected negative cycle")
	}

	if dist[0][1] != 4 || dist[1][0] != 4 {
		t.Errorf("expected undirected edge dist[0][1] and dist[1][0] to be 4, got %d and %d", dist[0][1], dist[1][0])
	}
	if dist[0][2] != math.MaxInt {
		t.Errorf("expected unreachable dist[0][2] to be math.MaxInt, got %d", dist[0][2])
	}
}

func TestFloydWarshallLessonValidation(t *testing.T) {
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

	meta := lesson.Meta()
	if meta.ID != "floydwarshall" {
		t.Errorf("meta.ID = %q; want 'floydwarshall'", meta.ID)
	}
	if meta.Name != "Floyd-Warshall All-Pairs Path" {
		t.Errorf("meta.Name = %q; want 'Floyd-Warshall All-Pairs Path'", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "table" {
		t.Errorf("meta.Renderer = %q; want 'table'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(V³)" {
		t.Errorf("meta.Documentation.TimeComplexity = %q; want 'O(V³)'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V²)" {
		t.Errorf("meta.Documentation.SpaceComplexity = %q; want 'O(V²)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("meta.Documentation.Difficulty = %q; want 'Hard'", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"All-Pairs Distance",
		"Intermediate Vertex Optimization",
		"Adjacency Matrix DP",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("meta.Documentation.KeyConcepts = %v; want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}

	registered, ok := catalog.Get("floydwarshall")
	if !ok {
		t.Fatal("expected floydwarshall to be registered in catalog")
	}
	if registered.Meta().ID != "floydwarshall" {
		t.Errorf("registered ID = %q; want 'floydwarshall'", registered.Meta().ID)
	}
}

func TestFloydWarshallTraceTableState(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	hasInit := false
	hasCompare := false
	hasSet := false
	hasDone := false
	var relaxationChecked bool

	for i, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d: expected TableState, got %T", i, step.State)
		}
		if len(tableState.RowHeaders) != 4 || len(tableState.ColHeaders) != 4 {
			t.Errorf("step %d: headers length mismatch: rows=%d, cols=%d", i, len(tableState.RowHeaders), len(tableState.ColHeaders))
		}
		if len(tableState.Cells) != 16 {
			t.Errorf("step %d: expected 16 cells, got %d", i, len(tableState.Cells))
		}

		cellMap := make(map[[2]int]renderers.TableCell)
		for _, cell := range tableState.Cells {
			cellMap[[2]int{cell.Row, cell.Col}] = cell
		}

		switch step.Action {
		case "init":
			hasInit = true
		case "compare":
			hasCompare = true
			k := step.Variables["k"].(int)
			rowI := step.Variables["i"].(int)
			colJ := step.Variables["j"].(int)

			if len(tableState.Dependencies) != 2 {
				t.Fatalf("step %d: expected 2 dependencies, got %d", i, len(tableState.Dependencies))
			}
			dep1 := tableState.Dependencies[0]
			dep2 := tableState.Dependencies[1]
			if dep1.FromRow != rowI || dep1.FromCol != k || dep1.ToRow != rowI || dep1.ToCol != colJ {
				t.Errorf("step %d: unexpected dep1: %+v, want (%d,%d)->(%d,%d)", i, dep1, rowI, k, rowI, colJ)
			}
			if dep2.FromRow != k || dep2.FromCol != colJ || dep2.ToRow != rowI || dep2.ToCol != colJ {
				t.Errorf("step %d: unexpected dep2: %+v, want (%d,%d)->(%d,%d)", i, dep2, k, colJ, rowI, colJ)
			}

			// Cell (i, j) active
			if cellMap[[2]int{rowI, colJ}].State != "active" {
				t.Errorf("step %d: cell (%d,%d) state = %q; want 'active'", i, rowI, colJ, cellMap[[2]int{rowI, colJ}].State)
			}
			// Cell (i, k) source (unless i==k and j==k)
			if !(rowI == k && colJ == k) {
				if cellMap[[2]int{rowI, k}].State != "source" && !(rowI == rowI && k == colJ) {
					t.Errorf("step %d: cell (%d,%d) state = %q; want 'source'", i, rowI, k, cellMap[[2]int{rowI, k}].State)
				}
				if cellMap[[2]int{k, colJ}].State != "source" && !(k == rowI && colJ == colJ) {
					t.Errorf("step %d: cell (%d,%d) state = %q; want 'source'", i, k, colJ, cellMap[[2]int{k, colJ}].State)
				}
			}

			// Intermediate vertex k in caption and vars
			if !strings.Contains(tableState.Caption, fmt.Sprintf("k=%s", step.Variables["intermediate"])) {
				t.Errorf("step %d: caption %q missing intermediate vertex", i, tableState.Caption)
			}
		case "set":
			hasSet = true
			relaxationChecked = true
			k := step.Variables["k"].(int)
			rowI := step.Variables["i"].(int)
			colJ := step.Variables["j"].(int)
			inter := step.Variables["intermediate"].(string)

			// Relaxed cell (i, j) is computed
			if cellMap[[2]int{rowI, colJ}].State != "computed" {
				t.Errorf("step %d: cell (%d,%d) state = %q; want 'computed'", i, rowI, colJ, cellMap[[2]int{rowI, colJ}].State)
			}
			// Sources are marked source
			if cellMap[[2]int{rowI, k}].State != "source" {
				t.Errorf("step %d: cell (%d,%d) state = %q; want 'source'", i, rowI, k, cellMap[[2]int{rowI, k}].State)
			}
			if cellMap[[2]int{k, colJ}].State != "source" {
				t.Errorf("step %d: cell (%d,%d) state = %q; want 'source'", i, k, colJ, cellMap[[2]int{k, colJ}].State)
			}

			// Dependencies
			if len(tableState.Dependencies) != 2 {
				t.Fatalf("step %d: expected 2 dependencies, got %d", i, len(tableState.Dependencies))
			}

			// Educational caption: "Path through intermediate k=%s is shorter: %d + %d = %d < %d"
			expectedPrefix := fmt.Sprintf("Path through intermediate k=%s is shorter:", inter)
			if !strings.HasPrefix(tableState.Caption, expectedPrefix) {
				t.Errorf("step %d: caption = %q; want prefix %q", i, tableState.Caption, expectedPrefix)
			}
			if !strings.Contains(tableState.Caption, "<") || !strings.Contains(tableState.Caption, "+") || !strings.Contains(tableState.Caption, "=") {
				t.Errorf("step %d: caption %q does not match arithmetic format", i, tableState.Caption)
			}
		case "done":
			hasDone = true
		}
	}

	if !hasInit || !hasCompare || !hasSet || !hasDone || !relaxationChecked {
		t.Errorf("missing expected actions: init=%t, compare=%t, set=%t, done=%t, relaxationChecked=%t",
			hasInit, hasCompare, hasSet, hasDone, relaxationChecked)
	}
}

func TestFloydWarshallSnapshotImmutability(t *testing.T) {
	steps := Trace([]string{"A", "B", "C"}, []WeightedEdge{
		{From: "A", To: "B", Weight: 3, Directed: true},
		{From: "B", To: "C", Weight: 2, Directed: true},
		{From: "A", To: "C", Weight: 8, Directed: true},
	})
	if len(steps) < 2 {
		t.Fatalf("expected multiple steps, got %d", len(steps))
	}

	for i := 0; i < len(steps)-1; i++ {
		s1 := steps[i].State.(renderers.TableState)
		s2 := steps[i+1].State.(renderers.TableState)

		if len(s1.RowHeaders) > 0 && len(s2.RowHeaders) > 0 && &s1.RowHeaders[0] == &s2.RowHeaders[0] {
			t.Errorf("steps %d and %d share RowHeaders backing array", i, i+1)
		}
		if len(s1.ColHeaders) > 0 && len(s2.ColHeaders) > 0 && &s1.ColHeaders[0] == &s2.ColHeaders[0] {
			t.Errorf("steps %d and %d share ColHeaders backing array", i, i+1)
		}
		if len(s1.Cells) > 0 && len(s2.Cells) > 0 && &s1.Cells[0] == &s2.Cells[0] {
			t.Errorf("steps %d and %d share Cells backing array", i, i+1)
		}
		if len(s1.Dependencies) > 0 && len(s2.Dependencies) > 0 && &s1.Dependencies[0] == &s2.Dependencies[0] {
			t.Errorf("steps %d and %d share Dependencies backing array", i, i+1)
		}
	}

	// Verify mutating step 0 does not mutate step 1
	s0 := steps[0].State.(renderers.TableState)
	s1 := steps[1].State.(renderers.TableState)
	origS1Row := s1.RowHeaders[0]
	origS1CellVal := s1.Cells[0].Value

	s0.RowHeaders[0] = "MODIFIED"
	s0.Cells[0].Value = "MODIFIED"

	if s1.RowHeaders[0] != origS1Row {
		t.Errorf("mutating step 0 RowHeaders mutated step 1 RowHeaders")
	}
	if s1.Cells[0].Value != origS1CellVal {
		t.Errorf("mutating step 0 Cells mutated step 1 Cells")
	}
}

func TestFloydWarshallNegativeCycleTrace(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 1, Directed: true},
		{From: "B", To: "C", Weight: -3, Directed: true},
		{From: "C", To: "A", Weight: -1, Directed: true},
	}

	steps := Trace(nodes, edges)
	var foundNegCycleStep bool
	for _, step := range steps {
		if step.LineID == "L16" {
			foundNegCycleStep = true
			if step.Action != "highlight" {
				t.Errorf("expected action 'highlight' for negative cycle, got %s", step.Action)
			}
			if neg, ok := step.Variables["negativeCycle"].(bool); !ok || !neg {
				t.Errorf("expected negativeCycle=true in vars")
			}
			ts, ok := step.State.(renderers.TableState)
			if !ok {
				t.Fatalf("expected TableState, got %T", step.State)
			}
			rowI := step.Variables["i"].(int)
			colJ := step.Variables["j"].(int)
			for _, cell := range ts.Cells {
				if cell.Row == rowI && cell.Col == colJ {
					if cell.State != "active" {
						t.Errorf("expected diagonal negative cycle cell (%d,%d) to be active, got %s", rowI, colJ, cell.State)
					}
				}
			}
		}
	}

	if !foundNegCycleStep {
		t.Fatal("expected negative cycle detection step L16 in trace")
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("negative cycle trace line validation failed: %v", err)
	}
}

func TestFloydWarshallInputValidationStrict(t *testing.T) {
	lesson := Lesson{}

	// Valid empty/nil inputs default to valid steps
	for _, input := range []json.RawMessage{nil, []byte(""), []byte("   ")} {
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("expected nil/empty input to succeed, got %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected non-empty steps for default input")
		}
	}

	// Invalid inputs that must be rejected
	invalidCases := []struct {
		name  string
		input []byte
	}{
		{"null", []byte("null")},
		{"malformed json", []byte("{not json")},
		{"empty object", []byte("{}")},
		{"empty nodes slice", []byte(`{"nodes": []}`)},
		{"empty graphNodes slice", []byte(`{"graphNodes": []}`)},
		{"node with empty string", []byte(`{"nodes": ["A", " "]}`)},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected error for %s input %s, got nil", tc.name, string(tc.input))
			}
		})
	}
}

func TestFloydWarshallBuildCustomInput(t *testing.T) {
	customJSON := []byte(`{
		"nodes": ["X", "Y", "Z"],
		"edges": [
			{"from": "X", "to": "Y", "weight": 4, "directed": true},
			{"from": "Y", "to": "Z", "weight": -1, "directed": true}
		]
	}`)

	lesson := Lesson{}
	steps, err := lesson.Build(json.RawMessage(customJSON))
	if err != nil {
		t.Fatalf("custom build failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	tableState, ok := lastStep.State.(renderers.TableState)
	if !ok {
		t.Fatalf("expected TableState in last step, got %T", lastStep.State)
	}
	if len(tableState.RowHeaders) != 3 {
		t.Fatalf("expected 3 row headers for custom input, got %d", len(tableState.RowHeaders))
	}
}
