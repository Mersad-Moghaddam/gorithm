package bfs

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBFSTraversalCorrectness(t *testing.T) {
	adj := map[string][]string{
		"A": {"B", "C"},
		"B": {"D", "E"},
		"C": {"E"},
		"D": {"F"},
		"E": {"F"},
		"F": {},
	}

	got := Traverse(adj, "A")
	expected := []string{"A", "B", "C", "D", "E", "F"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Traverse got %v, want %v", got, expected)
	}

	// Cycle graph test to ensure termination and correct order
	adjCycle := map[string][]string{
		"A": {"B"},
		"B": {"C", "A"},
		"C": {"A"},
	}
	gotCycle := Traverse(adjCycle, "A")
	expectedCycle := []string{"A", "B", "C"}
	if !reflect.DeepEqual(gotCycle, expectedCycle) {
		t.Fatalf("cycle Traverse got %v, want %v", gotCycle, expectedCycle)
	}

	// Single node
	gotSingle := Traverse(map[string][]string{"A": {}}, "A")
	if !reflect.DeepEqual(gotSingle, []string{"A"}) {
		t.Fatalf("single node got %v, want [A]", gotSingle)
	}

	// Empty start
	if gotEmpty := Traverse(adj, ""); gotEmpty != nil {
		t.Fatalf("empty start got %v, want nil", gotEmpty)
	}
}

func TestBFSLessonValidation(t *testing.T) {
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
	if meta.ID != "bfs" {
		t.Errorf("meta.ID = %q; want 'bfs'", meta.ID)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "graph" {
		t.Errorf("meta.Renderer = %q; want 'graph'", meta.Renderer)
	}
}

func TestBFSTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Check that node states transition through unvisited, queued, visiting, visited
	seenStates := make(map[string]bool)
	sawActiveEdge := false

	for _, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step state is %T, want renderers.GraphState", step.State)
		}
		for _, n := range state.Nodes {
			seenStates[n.State] = true
		}
		for _, e := range state.Edges {
			if e.Active {
				sawActiveEdge = true
			}
		}
	}

	for _, requiredState := range []string{"unvisited", "queued", "visiting", "visited"} {
		if !seenStates[requiredState] {
			t.Errorf("missing expected node state %q across trace steps", requiredState)
		}
	}

	if !sawActiveEdge {
		t.Errorf("expected at least one active edge during traversal")
	}

	// Final step should have all reachable nodes visited and no active edges
	lastState := steps[len(steps)-1].State.(renderers.GraphState)
	for _, n := range lastState.Nodes {
		if n.State != "visited" {
			t.Errorf("node %s final state = %q; want 'visited'", n.ID, n.State)
		}
	}
	for _, e := range lastState.Edges {
		if e.Active {
			t.Errorf("edge %s->%s is active at final step", e.From, e.To)
		}
	}
}
