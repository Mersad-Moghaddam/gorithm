package floydwarshall

import (
	"encoding/json"
	"math"
	"reflect"
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

func TestFloydWarshallSaturatesFiniteOverflow(t *testing.T) {
	nodes := []string{"A", "B", "C", "D"}
	for _, tc := range []struct{ weight, want int }{{math.MaxInt - 10, math.MaxInt - 1}, {math.MinInt + 10, math.MinInt}} {
		dist, _ := AllPairsShortestPath(nodes, []WeightedEdge{{"A", "B", tc.weight, true}, {"B", "C", tc.weight, true}})
		if dist[0][2] != tc.want || dist[0][3] != math.MaxInt {
			t.Fatalf("weight %d: got %d and unreachable %d", tc.weight, dist[0][2], dist[0][3])
		}
	}
}

func TestFloydWarshallExactIntermediateDistances(t *testing.T) {
	nodes := []string{"A", "B", "C", "D"}
	edges := []WeightedEdge{{"A", "B", math.MaxInt - 1, true}, {"B", "C", 10, true}, {"C", "A", -math.MaxInt, true}}
	dist, negative := AllPairsShortestPath(nodes, edges)
	if negative || dist[0][0] != 0 || dist[1][0] != 10-math.MaxInt || dist[0][2] != math.MaxInt-1 || dist[0][3] != math.MaxInt {
		t.Fatalf("positive cycle: negative=%v distances=%v", negative, dist)
	}
	// A direct MaxInt edge is finite, not the unreachable sentinel.
	dist, negative = AllPairsShortestPath(nodes, []WeightedEdge{{"A", "B", math.MaxInt, true}, {"B", "C", -math.MaxInt, true}})
	if negative || dist[0][1] != math.MaxInt-1 || dist[0][2] != 0 || dist[0][3] != math.MaxInt {
		t.Fatalf("MaxInt edge: negative=%v distances=%v", negative, dist)
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

	for i, step := range steps {
		tableState, ok := step.State.(renderers.FloydWarshallLessonState)
		if !ok {
			t.Fatalf("step %d: expected TableState, got %T", i, step.State)
		}
		if len(tableState.RowHeaders) != 4 || len(tableState.ColHeaders) != 4 {
			t.Errorf("step %d: headers length mismatch: rows=%d, cols=%d", i, len(tableState.RowHeaders), len(tableState.ColHeaders))
		}
		if len(tableState.Cells) != 16 {
			t.Errorf("step %d: expected 16 cells, got %d", i, len(tableState.Cells))
		}

		if step.Action == "init" {
			hasInit = true
		}
		if step.Action == "compare" {
			hasCompare = true
		}
		if step.Action == "set" {
			hasSet = true
		}
		if step.Action == "done" {
			hasDone = true
		}
	}

	if !hasInit || !hasCompare || !hasSet || !hasDone {
		t.Errorf("missing expected actions: init=%t, compare=%t, set=%t, done=%t", hasInit, hasCompare, hasSet, hasDone)
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
	tableState, ok := lastStep.State.(renderers.FloydWarshallLessonState)
	if !ok {
		t.Fatalf("expected TableState in last step, got %T", lastStep.State)
	}
	if len(tableState.RowHeaders) != 3 {
		t.Fatalf("expected 3 row headers for custom input, got %d", len(tableState.RowHeaders))
	}
}
