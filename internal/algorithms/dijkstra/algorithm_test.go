package dijkstra

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestDijkstraShortestPathCorrectness(t *testing.T) {
	nodes := []string{"A", "B", "C", "D", "E"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "C", Weight: 2},
		{From: "B", To: "C", Weight: 1},
		{From: "B", To: "D", Weight: 5},
		{From: "C", To: "D", Weight: 8},
		{From: "C", To: "E", Weight: 10},
		{From: "D", To: "E", Weight: 2},
	}

	dist, path := ShortestPath(nodes, edges, "A", "E")

	expectedDist := map[string]int{
		"A": 0,
		"B": 3,
		"C": 2,
		"D": 8,
		"E": 10,
	}

	for node, want := range expectedDist {
		if got := dist[node]; got != want {
			t.Errorf("dist[%s] = %d; want %d", node, got, want)
		}
	}

	expectedPath := []string{"A", "C", "B", "D", "E"}
	if !reflect.DeepEqual(path, expectedPath) {
		t.Fatalf("path = %v; want %v", path, expectedPath)
	}
}

func TestDijkstraDirectedGraph(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 1, Directed: true},
		{From: "B", To: "C", Weight: 2, Directed: true},
		{From: "C", To: "A", Weight: 5, Directed: true},
	}

	dist, path := ShortestPath(nodes, edges, "A", "C")
	if dist["C"] != 3 {
		t.Errorf("dist[C] = %d; want 3", dist["C"])
	}
	expectedPath := []string{"A", "B", "C"}
	if !reflect.DeepEqual(path, expectedPath) {
		t.Fatalf("path = %v; want %v", path, expectedPath)
	}

	// Reverse direction should not use backwards edges
	distRev, pathRev := ShortestPath(nodes, edges, "C", "B")
	if distRev["B"] != 6 {
		t.Errorf("distRev[B] = %d; want 6", distRev["B"])
	}
	expectedPathRev := []string{"C", "A", "B"}
	if !reflect.DeepEqual(pathRev, expectedPathRev) {
		t.Fatalf("pathRev = %v; want %v", pathRev, expectedPathRev)
	}
}

func TestDijkstraUnreachableAndSameNode(t *testing.T) {
	nodes := []string{"A", "B", "Z"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 3, Directed: true},
	}

	// Same node
	distSame, pathSame := ShortestPath(nodes, edges, "A", "A")
	if distSame["A"] != 0 {
		t.Errorf("distSame[A] = %d; want 0", distSame["A"])
	}
	if !reflect.DeepEqual(pathSame, []string{"A"}) {
		t.Errorf("pathSame = %v; want [A]", pathSame)
	}

	// Unreachable
	_, pathUnreachable := ShortestPath(nodes, edges, "A", "Z")
	if len(pathUnreachable) != 0 {
		t.Errorf("expected empty path for unreachable node, got %v", pathUnreachable)
	}
}

func TestDijkstraLessonValidation(t *testing.T) {
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
}

func TestDijkstraTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	pathNodes := 0
	for _, n := range state.Nodes {
		if n.State == "path" {
			pathNodes++
		}
	}
	if pathNodes != 5 {
		t.Errorf("expected 5 path nodes in final state, got %d", pathNodes)
	}

	// Check distance table in variables
	if _, ok := lastStep.Variables["distances"]; !ok {
		t.Errorf("expected 'distances' table in step variables, got %v", lastStep.Variables)
	}
}
