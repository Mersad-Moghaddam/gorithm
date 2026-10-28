package astar

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestAStarPathCorrectness(t *testing.T) {
	// Simple grid-like graph
	// A(0,0) -> B(1,0) [wt 1] -> C(2,0) [wt 1] -> D(3,0) [wt 1]
	// A(0,0) -> E(0,1) [wt 3] -> D(3,0) [wt 10]
	nodes := []AStarNode{
		{ID: "A", Label: "A", X: 0, Y: 0},
		{ID: "B", Label: "B", X: 1, Y: 0},
		{ID: "C", Label: "C", X: 2, Y: 0},
		{ID: "D", Label: "D", X: 3, Y: 0},
		{ID: "E", Label: "E", X: 0, Y: 1},
	}
	edges := []AStarEdge{
		{From: "A", To: "B", Weight: 1, Undirected: true},
		{From: "B", To: "C", Weight: 1, Undirected: true},
		{From: "C", To: "D", Weight: 1, Undirected: true},
		{From: "A", To: "E", Weight: 3, Undirected: true},
		{From: "E", To: "D", Weight: 10, Undirected: true},
	}

	cost, path := FindPath(nodes, edges, "A", "D")
	if cost != 3 {
		t.Fatalf("expected cost 3, got %d", cost)
	}
	expectedPath := []string{"A", "B", "C", "D"}
	if !reflect.DeepEqual(path, expectedPath) {
		t.Fatalf("expected path %v, got %v", expectedPath, path)
	}
}

func TestAStarStartEqualsGoal(t *testing.T) {
	nodes := []AStarNode{
		{ID: "A", Label: "A", X: 0, Y: 0},
		{ID: "B", Label: "B", X: 1, Y: 1},
	}
	edges := []AStarEdge{
		{From: "A", To: "B", Weight: 5, Undirected: true},
	}

	cost, path := FindPath(nodes, edges, "A", "A")
	if cost != 0 {
		t.Fatalf("expected cost 0, got %d", cost)
	}
	expectedPath := []string{"A"}
	if !reflect.DeepEqual(path, expectedPath) {
		t.Fatalf("expected path %v, got %v", expectedPath, path)
	}
}

func TestAStarUnreachable(t *testing.T) {
	nodes := []AStarNode{
		{ID: "A", Label: "A", X: 0, Y: 0},
		{ID: "B", Label: "B", X: 5, Y: 5},
	}
	edges := []AStarEdge{}

	cost, path := FindPath(nodes, edges, "A", "B")
	if cost != -1 {
		t.Fatalf("expected cost -1, got %d", cost)
	}
	if path != nil {
		t.Fatalf("expected nil path, got %v", path)
	}
}

func TestAStarHeuristics(t *testing.T) {
	n1 := AStarNode{ID: "1", X: 0, Y: 0}
	n2 := AStarNode{ID: "2", X: 3, Y: 4}

	eDist := EuclideanDistance(n1, n2)
	if eDist != 5 {
		t.Fatalf("expected Euclidean distance 5, got %d", eDist)
	}

	mDist := ManhattanDistance(n1, n2)
	if mDist != 7 {
		t.Fatalf("expected Manhattan distance 7, got %d", mDist)
	}
}

func TestAStarLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "astar" {
		t.Fatalf("expected ID 'astar', got %q", meta.ID)
	}
	if meta.Name != "A* Pathfinding" {
		t.Fatalf("expected Name 'A* Pathfinding', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Fatalf("expected Category %q, got %q", catalog.CategoryGraphs, meta.Category)
	}
	if meta.Renderer != "graph" {
		t.Fatalf("expected Renderer 'graph', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(E)" {
		t.Fatalf("expected TimeComplexity 'O(E)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V)" {
		t.Fatalf("expected SpaceComplexity 'O(V)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Heuristic Search", "f(n)=g(n)+h(n)", "Admissible Heuristic", "Optimal Path"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Fatalf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Fatal("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Fatal("expected non-empty HowItWorks")
	}
}

func TestAStarDefaultBuildValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build default trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for default build: %v", err)
	}

	// Verify last step is ActionDone
	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last step ActionDone, got %q", lastStep.Action)
	}

	// Verify GraphState is emitted
	if _, ok := lastStep.State.(renderers.AStarLessonState); !ok {
		t.Fatalf("expected state type renderers.AStarLessonState, got %T", lastStep.State)
	}

	// Verify f/g/h values present in step variables
	hasFGH := false
	for _, s := range steps {
		if s.Variables != nil {
			_, hasG := s.Variables["g"]
			_, hasH := s.Variables["h"]
			_, hasF := s.Variables["f"]
			if hasG && hasH && hasF {
				hasFGH = true
				break
			}
		}
	}
	if !hasFGH {
		t.Fatal("expected trace steps to contain f/g/h variables")
	}
}

func TestAStarBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"nodes": [
			{"id": "S", "label": "Start", "x": 100, "y": 200},
			{"id": "M", "label": "Mid", "x": 300, "y": 200},
			{"id": "G", "label": "Goal", "x": 500, "y": 200}
		],
		"edges": [
			{"from": "S", "to": "M", "weight": 200, "undirected": true},
			{"from": "M", "to": "G", "weight": 200, "undirected": true}
		],
		"start": "S",
		"goal": "G"
	}`)

	steps, err := lesson.Build(json.RawMessage(customInput))
	if err != nil {
		t.Fatalf("failed to build custom trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps for custom input")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for custom input: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last step ActionDone, got %q", lastStep.Action)
	}
}

func TestAStarRegistration(t *testing.T) {
	l, ok := catalog.Get("astar")
	if !ok {
		t.Fatal("lesson 'astar' not registered in catalog")
	}
	if l.Meta().ID != "astar" {
		t.Fatalf("expected registered ID 'astar', got %q", l.Meta().ID)
	}
}

func TestAStarManhattanHeuristicPath(t *testing.T) {
	nodes := []AStarNode{
		{ID: "A", X: 0, Y: 0},
		{ID: "B", X: 1, Y: 0},
		{ID: "C", X: 1, Y: 1},
	}
	edges := []AStarEdge{
		{From: "A", To: "B", Weight: 10, Undirected: true},
		{From: "B", To: "C", Weight: 10, Undirected: true},
		{From: "A", To: "C", Weight: 25, Undirected: true},
	}

	cost, path := FindPathWithHeuristic(nodes, edges, "A", "C", ManhattanDistance)
	if cost != 20 {
		t.Fatalf("expected cost 20, got %d", cost)
	}
	expected := []string{"A", "B", "C"}
	if !reflect.DeepEqual(path, expected) {
		t.Fatalf("expected %v, got %v", expected, path)
	}
}

func TestAStarTraceStartEqualsGoal(t *testing.T) {
	nodes := []AStarNode{{ID: "A", X: 100, Y: 100}}
	edges := []AStarEdge{}
	steps := Trace(nodes, edges, "A", "A")
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace")
	}
	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected ActionDone, got %q", lastStep.Action)
	}
	if lastStep.Variables["cost"] != 0 {
		t.Fatalf("expected cost 0, got %v", lastStep.Variables["cost"])
	}
}

func TestAStarTraceUnreachable(t *testing.T) {
	nodes := []AStarNode{
		{ID: "A", X: 100, Y: 100},
		{ID: "B", X: 400, Y: 400},
	}
	edges := []AStarEdge{}
	steps := Trace(nodes, edges, "A", "B")
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace")
	}
	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
	lastStep := steps[len(steps)-1]
	if lastStep.Variables["found"] != false {
		t.Fatalf("expected found=false, got %v", lastStep.Variables["found"])
	}
}

func TestAStarTraceInvalidParams(t *testing.T) {
	nodes := []AStarNode{{ID: "A", X: 100, Y: 100}}
	steps := Trace(nodes, nil, "X", "Y")
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace for invalid parameters")
	}
	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	emptySteps := Trace(nil, nil, "A", "B")
	if len(emptySteps) != 0 {
		t.Fatalf("expected empty trace for empty nodes, got %d steps", len(emptySteps))
	}
}
