package bellmanford

import (
	"math"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBellmanFordShortestPathCorrectness(t *testing.T) {
	nodes := []string{"A", "B", "C", "D", "E"}
	edges := []DirectedEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "C", Weight: 2},
		{From: "B", To: "D", Weight: 2},
		{From: "C", To: "D", Weight: 5},
		{From: "C", To: "E", Weight: 4},
		{From: "D", To: "E", Weight: -3},
	}

	dist, hasNegCycle := ShortestPath(nodes, edges, "A")
	if hasNegCycle {
		t.Fatalf("expected no negative cycle, got hasNegCycle=true")
	}

	expected := map[string]int{
		"A": 0,
		"B": 4,
		"C": 2,
		"D": 6,
		"E": 3,
	}

	for node, want := range expected {
		if got := dist[node]; got != want {
			t.Errorf("dist[%s] = %d; want %d", node, got, want)
		}
	}
}

func TestBellmanFordNegativeCycle(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := []DirectedEdge{
		{From: "A", To: "B", Weight: 1},
		{From: "B", To: "C", Weight: -3},
		{From: "C", To: "A", Weight: -1},
	}

	_, hasNegCycle := ShortestPath(nodes, edges, "A")
	if !hasNegCycle {
		t.Fatalf("expected negative cycle, got hasNegCycle=false")
	}
}

func TestBellmanFordUnreachable(t *testing.T) {
	nodes := []string{"A", "B", "Z"}
	edges := []DirectedEdge{
		{From: "A", To: "B", Weight: 5},
	}

	dist, hasNegCycle := ShortestPath(nodes, edges, "A")
	if hasNegCycle {
		t.Fatalf("unexpected negative cycle")
	}
	if dist["A"] != 0 {
		t.Errorf("dist[A] = %d; want 0", dist["A"])
	}
	if dist["B"] != 5 {
		t.Errorf("dist[B] = %d; want 5", dist["B"])
	}
	if dist["Z"] != math.MaxInt {
		t.Errorf("dist[Z] = %d; want math.MaxInt", dist["Z"])
	}
}

func TestBellmanFordLessonValidation(t *testing.T) {
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
	if meta.ID != "bellmanford" {
		t.Errorf("meta.ID = %q; want 'bellmanford'", meta.ID)
	}
	if meta.Name != "Bellman-Ford Shortest Path" {
		t.Errorf("meta.Name = %q; want 'Bellman-Ford Shortest Path'", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "graph" {
		t.Errorf("meta.Renderer = %q; want 'graph'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(V · E)" {
		t.Errorf("meta.Documentation.TimeComplexity = %q; want 'O(V · E)'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V)" {
		t.Errorf("meta.Documentation.SpaceComplexity = %q; want 'O(V)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("meta.Documentation.Difficulty = %q; want 'Medium'", meta.Documentation.Difficulty)
	}
	expectedConcepts := []string{
		"Dynamic Programming on Graphs",
		"Edge Relaxation Passes",
		"Negative Weight Handling",
		"Negative Cycle Detection",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("meta.Documentation.KeyConcepts = %v; want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}
}

func TestBellmanFordTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	lastStep := steps[len(steps)-1]
	_, ok := lastStep.State.(renderers.BellmanFordLessonState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	if _, ok := lastStep.Variables["distances"]; !ok {
		t.Errorf("expected 'distances' in step variables, got %v", lastStep.Variables)
	}
}

func TestBellmanFordTraceNegativeCycle(t *testing.T) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A"},
		{ID: "B", Label: "B"},
		{ID: "C", Label: "C"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B", Weight: 1},
		{From: "B", To: "C", Weight: -3},
		{From: "C", To: "A", Weight: -1},
	}

	steps := Trace(nodes, edges, "A")
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.LineID != "L15" {
		t.Fatalf("expected last step LineID 'L15', got %q", lastStep.LineID)
	}
	if lastStep.Variables["negativeCycle"] != true {
		t.Fatalf("expected negativeCycle=true in last step variables")
	}
}

func TestBellmanFordTraceInvalidStart(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A", Label: "A"}}
	edges := []renderers.GraphEdge{}

	steps := Trace(nodes, edges, "Z")
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps for invalid start, got %d", len(steps))
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}

func TestBellmanFordBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build([]byte(`{"start": "B"}`))
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}
