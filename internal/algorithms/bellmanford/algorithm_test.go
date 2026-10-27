package bellmanford

import (
	"math"
	"reflect"
	"strings"
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
	gs, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	if gs.WorklistLabel != "Pass Edges to Relax" {
		t.Errorf("expected WorklistLabel 'Pass Edges to Relax', got %q", gs.WorklistLabel)
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

func TestBellmanFordInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Null input
	if _, err := lesson.Build([]byte("null")); err == nil {
		t.Error("expected error for null input, got nil")
	}

	// Malformed JSON
	if _, err := lesson.Build([]byte("{not-json")); err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}

	// Unknown start node on default graph
	if _, err := lesson.Build([]byte(`{"start": "UNKNOWN"}`)); err == nil {
		t.Error("expected error for unknown start node, got nil")
	}

	// Unknown start node on custom graph
	if _, err := lesson.Build([]byte(`{"nodes": [{"id": "X"}], "start": "Z"}`)); err == nil {
		t.Error("expected error for start node not in custom nodes, got nil")
	}

	// Valid inputs should pass
	if _, err := lesson.Build(nil); err != nil {
		t.Errorf("expected success for nil input, got: %v", err)
	}
	if _, err := lesson.Build([]byte(`{"start": "B"}`)); err != nil {
		t.Errorf("expected success for valid start node 'B', got: %v", err)
	}
}

func TestBellmanFordWorklist(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundWorklist := false
	for i, step := range steps {
		gs, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is not GraphState", i)
		}
		if gs.WorklistLabel != "Pass Edges to Relax" {
			t.Errorf("step %d: WorklistLabel = %q; want %q", i, gs.WorklistLabel, "Pass Edges to Relax")
		}
		if len(gs.Worklist) > 0 {
			foundWorklist = true
			for _, item := range gs.Worklist {
				if item.ID == "" || item.Label == "" {
					t.Errorf("step %d: worklist item has empty field: %+v", i, item)
				}
			}
		}
	}
	if !foundWorklist {
		t.Fatal("expected at least one step with non-empty worklist")
	}
}

func TestBellmanFordSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) < 3 {
		t.Fatalf("need at least 3 steps, got %d", len(steps))
	}

	// Capture step 0 values
	s0 := steps[0].State.(renderers.GraphState)
	origNode0State := s0.Nodes[0].State

	// Mutate step 1
	s1 := steps[1].State.(renderers.GraphState)
	s1.Nodes[0].State = "MUTATED_STATE"
	if len(s1.Edges) > 0 {
		s1.Edges[0].Active = !s1.Edges[0].Active
	}
	if len(s1.Worklist) > 0 {
		s1.Worklist[0].Label = "MUTATED_LABEL"
	}

	// Verify step 0 is unchanged
	s0After := steps[0].State.(renderers.GraphState)
	if s0After.Nodes[0].State != origNode0State {
		t.Errorf("step 0 was mutated when step 1 was modified: got %s, want %s", s0After.Nodes[0].State, origNode0State)
	}

	// Verify slice backing arrays are distinct across adjacent steps
	for i := 0; i < len(steps)-1; i++ {
		si := steps[i].State.(renderers.GraphState)
		sj := steps[i+1].State.(renderers.GraphState)

		if len(si.Nodes) > 0 && len(sj.Nodes) > 0 && &si.Nodes[0] == &sj.Nodes[0] {
			t.Errorf("steps %d and %d share Nodes backing array", i, i+1)
		}
		if len(si.Edges) > 0 && len(sj.Edges) > 0 && &si.Edges[0] == &sj.Edges[0] {
			t.Errorf("steps %d and %d share Edges backing array", i, i+1)
		}
		if len(si.Worklist) > 0 && len(sj.Worklist) > 0 && &si.Worklist[0] == &sj.Worklist[0] {
			t.Errorf("steps %d and %d share Worklist backing array", i, i+1)
		}
	}
}

func TestBellmanFordNegativeCycleHighlighting(t *testing.T) {
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

	lastStep := steps[len(steps)-1]
	if lastStep.LineID != "L15" {
		t.Fatalf("expected last step LineID 'L15', got %q", lastStep.LineID)
	}

	gs := lastStep.State.(renderers.GraphState)
	if !strings.Contains(gs.Caption, "Negative cycle detected") {
		t.Errorf("expected caption to announce negative cycle, got %q", gs.Caption)
	}

	highlightedNodes := 0
	for _, n := range gs.Nodes {
		if n.State == "path" {
			highlightedNodes++
		}
	}
	if highlightedNodes == 0 {
		t.Error("expected negative cycle nodes to be highlighted with State='path'")
	}

	activeEdges := 0
	for _, e := range gs.Edges {
		if e.Active {
			activeEdges++
		}
	}
	if activeEdges == 0 {
		t.Error("expected cycle edges to be highlighted with Active=true")
	}
}

func TestBellmanFordDisconnectedGraph(t *testing.T) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A"},
		{ID: "B", Label: "B"},
		{ID: "C", Label: "C"},
		{ID: "D", Label: "D"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B", Weight: 3},
		{From: "C", To: "D", Weight: 2},
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
	if lastStep.LineID != "L18" {
		t.Fatalf("expected last step L18, got %q", lastStep.LineID)
	}

	dist := lastStep.Variables["distances"].(map[string]int)
	if dist["A"] != 0 {
		t.Errorf("dist[A] = %d; want 0", dist["A"])
	}
	if dist["B"] != 3 {
		t.Errorf("dist[B] = %d; want 3", dist["B"])
	}
	if dist["C"] != math.MaxInt {
		t.Errorf("dist[C] = %d; want math.MaxInt", dist["C"])
	}
	if dist["D"] != math.MaxInt {
		t.Errorf("dist[D] = %d; want math.MaxInt", dist["D"])
	}

	foundUnreachableExplanation := false
	for _, s := range steps {
		if s.LineID == "L6" {
			if reason, ok := s.Variables["reason"].(string); ok && strings.Contains(reason, "∞") {
				foundUnreachableExplanation = true
				break
			}
		}
	}
	if !foundUnreachableExplanation {
		t.Error("expected non-relaxation step explaining unreachable node with ∞")
	}
}

func TestBellmanFordPassVisualization(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenPasses := make(map[int]bool)
	for _, s := range steps {
		if p, ok := s.Variables["pass"].(int); ok && p > 0 {
			seenPasses[p] = true
		}
	}

	if !seenPasses[1] {
		t.Error("pass 1 not found in step variables")
	}
	if !seenPasses[2] {
		t.Error("pass 2 not found in step variables")
	}
}
