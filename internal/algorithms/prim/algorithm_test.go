package prim

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestPrimMSTCorrectness(t *testing.T) {
	t.Run("default 6-node graph", func(t *testing.T) {
		nodes := []string{"A", "B", "C", "D", "E", "F"}
		edges := []WeightedEdge{
			{From: "A", To: "B", Weight: 4},
			{From: "A", To: "C", Weight: 2},
			{From: "B", To: "C", Weight: 1},
			{From: "B", To: "D", Weight: 5},
			{From: "C", To: "D", Weight: 8},
			{From: "C", To: "E", Weight: 10},
			{From: "D", To: "E", Weight: 2},
			{From: "D", To: "F", Weight: 6},
			{From: "E", To: "F", Weight: 3},
		}

		mstEdges, totalWeight := MST(nodes, edges, "A")
		if len(mstEdges) != len(nodes)-1 {
			t.Fatalf("expected %d edges in MST, got %d", len(nodes)-1, len(mstEdges))
		}
		if totalWeight != 13 {
			t.Fatalf("expected total weight 13, got %d", totalWeight)
		}

		// Verify connected & acyclic with different start
		altEdges, altWeight := MST(nodes, edges, "F")
		if len(altEdges) != len(nodes)-1 {
			t.Fatalf("expected %d edges with start F, got %d", len(nodes)-1, len(altEdges))
		}
		if altWeight != 13 {
			t.Fatalf("expected weight 13 with start F, got %d", altWeight)
		}
	})

	t.Run("4-node graph", func(t *testing.T) {
		nodes := []string{"A", "B", "C", "D"}
		edges := []WeightedEdge{
			{From: "A", To: "B", Weight: 1},
			{From: "B", To: "C", Weight: 2},
			{From: "A", To: "C", Weight: 4},
			{From: "C", To: "D", Weight: 3},
			{From: "B", To: "D", Weight: 5},
		}

		mstEdges, totalWeight := MST(nodes, edges, "A")
		if len(mstEdges) != len(nodes)-1 {
			t.Fatalf("expected %d edges in MST, got %d", len(nodes)-1, len(mstEdges))
		}
		if totalWeight != 6 {
			t.Fatalf("expected total weight 6, got %d", totalWeight)
		}
	})

	t.Run("single node graph", func(t *testing.T) {
		mstEdges, totalWeight := MST([]string{"A"}, nil, "A")
		if len(mstEdges) != 0 {
			t.Fatalf("expected 0 edges, got %d", len(mstEdges))
		}
		if totalWeight != 0 {
			t.Fatalf("expected 0 weight, got %d", totalWeight)
		}
	})

	t.Run("empty graph", func(t *testing.T) {
		mstEdges, totalWeight := MST(nil, nil, "")
		if len(mstEdges) != 0 {
			t.Fatalf("expected 0 edges, got %d", len(mstEdges))
		}
		if totalWeight != 0 {
			t.Fatalf("expected 0 weight, got %d", totalWeight)
		}
	})
}

func TestPrimLessonValidation(t *testing.T) {
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

	firstStep := steps[0]
	if firstStep.Action != trace.ActionInit {
		t.Errorf("expected first action to be init, got %s", firstStep.Action)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last action to be done, got %s", lastStep.Action)
	}

	state, ok := lastStep.State.(renderers.PrimLessonState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	activeCount := 0
	for _, e := range state.Edges {
		if e.Active {
			activeCount++
		}
	}
	if activeCount != 5 {
		t.Errorf("expected 5 active MST edges, got %d", activeCount)
	}

	for _, n := range state.Nodes {
		if n.State != "path" {
			t.Errorf("expected node %s to be 'path', got %q", n.ID, n.State)
		}
	}

	// Verify catalog registration
	regLesson, ok := catalog.Get("prim")
	if !ok {
		t.Fatal("prim lesson not registered in catalog")
	}
	if regLesson.Meta().ID != "prim" {
		t.Errorf("expected registered ID 'prim', got %q", regLesson.Meta().ID)
	}
}

func TestPrimCustomInput(t *testing.T) {
	lesson := Lesson{}
	customJSON := []byte(`{
		"start": "1",
		"edges": [
			{"from": "1", "to": "2", "weight": 10},
			{"from": "2", "to": "3", "weight": 15},
			{"from": "1", "to": "3", "weight": 5}
		]
	}`)

	steps, err := lesson.Build(customJSON)
	if err != nil {
		t.Fatalf("failed to build custom trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("custom trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for custom input: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last step done, got %s", lastStep.Action)
	}
}

func TestPrimMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "prim" {
		t.Errorf("expected ID 'prim', got %q", meta.ID)
	}
	if meta.Name != "Prim's Minimum Spanning Tree" {
		t.Errorf("expected Name 'Prim's Minimum Spanning Tree', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("expected Category %q, got %q", catalog.CategoryGraphs, meta.Category)
	}
	if meta.Renderer != "graph" {
		t.Errorf("expected Renderer 'graph', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(E log V)" {
		t.Errorf("expected TimeComplexity 'O(E log V)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V)" {
		t.Errorf("expected SpaceComplexity 'O(V)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Greedy", "Cut Property", "Growing Tree", "Priority Queue"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}
