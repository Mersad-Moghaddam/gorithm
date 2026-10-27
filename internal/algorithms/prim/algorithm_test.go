package prim

import (
	"fmt"
	"reflect"
	"strings"
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

	state, ok := lastStep.State.(renderers.GraphState)
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

func TestPrimWorklist(t *testing.T) {
	nodes, edges, start := DefaultGraph()
	steps := Trace(nodes, edges, start)

	foundWorklist := false
	for i, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is %T, want GraphState", i, step.State)
		}

		if state.WorklistLabel != "Priority Queue (Crossing Edges: edge, weight)" {
			t.Errorf("step %d: expected worklist label %q, got %q",
				i, "Priority Queue (Crossing Edges: edge, weight)", state.WorklistLabel)
		}

		if len(state.Worklist) > 0 {
			foundWorklist = true
			// Verify items are ordered by weight
			for j := 0; j < len(state.Worklist)-1; j++ {
				var w1, w2 int
				_, err1 := fmt.Sscanf(state.Worklist[j].Label[strings.LastIndex(state.Worklist[j].Label, ",")+1:], "%d", &w1)
				_, err2 := fmt.Sscanf(state.Worklist[j+1].Label[strings.LastIndex(state.Worklist[j+1].Label, ",")+1:], "%d", &w2)
				if err1 == nil && err2 == nil && w1 > w2 {
					t.Errorf("step %d: worklist not ordered by weight: %s (%d) > %s (%d)",
						i, state.Worklist[j].Label, w1, state.Worklist[j+1].Label, w2)
				}
			}
			for _, item := range state.Worklist {
				if item.ID == "" || item.Label == "" {
					t.Errorf("step %d: worklist item has empty field: %+v", i, item)
				}
			}
		}
	}

	if !foundWorklist {
		t.Error("expected to find at least one step with non-empty worklist")
	}
}

func TestPrimSnapshotImmutability(t *testing.T) {
	nodes, edges, start := DefaultGraph()
	steps := Trace(nodes, edges, start)

	if len(steps) < 2 {
		t.Fatalf("expected multiple steps, got %d", len(steps))
	}

	// Verify deep copy: mutating state in step i must not alter step j
	for i := 0; i < len(steps)-1; i++ {
		s1 := steps[i].State.(renderers.GraphState)
		s2 := steps[i+1].State.(renderers.GraphState)

		if len(s1.Nodes) > 0 && len(s2.Nodes) > 0 {
			origState := s2.Nodes[0].State
			s1.Nodes[0].State = "MUTATED_TEST_VAL"
			if s2.Nodes[0].State == "MUTATED_TEST_VAL" {
				t.Fatalf("steps %d and %d share nodes slice backing array", i, i+1)
			}
			s2.Nodes[0].State = origState
		}

		if len(s1.Edges) > 0 && len(s2.Edges) > 0 {
			origActive := s2.Edges[0].Active
			s1.Edges[0].Active = !origActive
			if s2.Edges[0].Active == s1.Edges[0].Active {
				t.Fatalf("steps %d and %d share edges slice backing array", i, i+1)
			}
			s2.Edges[0].Active = origActive
		}

		if len(s1.Worklist) > 0 && len(s2.Worklist) > 0 {
			origID := s2.Worklist[0].ID
			s1.Worklist[0].ID = "MUTATED_WORKLIST_ID"
			if s2.Worklist[0].ID == "MUTATED_WORKLIST_ID" {
				t.Fatalf("steps %d and %d share worklist slice backing array", i, i+1)
			}
			s2.Worklist[0].ID = origID
		}
	}
}

func TestPrimMSTEdgeCount(t *testing.T) {
	nodes, edges, start := DefaultGraph()
	steps := Trace(nodes, edges, start)
	lastState := steps[len(steps)-1].State.(renderers.GraphState)

	activeCount := 0
	for _, e := range lastState.Edges {
		if e.Active {
			activeCount++
		}
	}
	expectedEdges := len(nodes) - 1
	if activeCount != expectedEdges {
		t.Fatalf("expected exactly %d (V-1) active edges in final MST, got %d", expectedEdges, activeCount)
	}

	if !strings.Contains(lastState.Caption, "MST complete") || !strings.Contains(lastState.Caption, "13") {
		t.Errorf("expected final MST caption to include total weight 13, got %q", lastState.Caption)
	}
}

func TestPrimDisconnectedGraph(t *testing.T) {
	discNodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 100, Y: 100},
		{ID: "B", Label: "B", X: 200, Y: 100},
		{ID: "C", Label: "C", X: 300, Y: 300},
		{ID: "D", Label: "D", X: 400, Y: 300},
	}
	discEdges := []WeightedEdge{
		{From: "A", To: "B", Weight: 3},
		{From: "C", To: "D", Weight: 5},
	}

	mstEdges, totalWeight := MST([]string{"A", "B", "C", "D"}, discEdges, "A")
	if len(mstEdges) != 1 {
		t.Fatalf("expected 1 edge in reachable component MST, got %d", len(mstEdges))
	}
	if totalWeight != 3 {
		t.Fatalf("expected total weight 3, got %d", totalWeight)
	}

	steps := Trace(discNodes, discEdges, "A")
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps for disconnected graph")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last action Done, got %s", lastStep.Action)
	}
	lastState := lastStep.State.(renderers.GraphState)
	if !strings.Contains(lastState.Caption, "Disconnected graph") {
		t.Errorf("expected caption to indicate disconnected graph, got %q", lastState.Caption)
	}

	activeCount := 0
	for _, e := range lastState.Edges {
		if e.Active {
			activeCount++
		}
	}
	if activeCount != 1 {
		t.Errorf("expected 1 active edge in disconnected graph, got %d", activeCount)
	}
}

func TestPrimSingleNodeGraph(t *testing.T) {
	singleNode := []renderers.GraphNode{{ID: "A", Label: "A", X: 100, Y: 100}}
	steps := Trace(singleNode, nil, "A")
	if len(steps) == 0 {
		t.Fatal("expected trace for single node")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last step action Done, got %s", lastStep.Action)
	}
	lastState := lastStep.State.(renderers.GraphState)
	if !strings.Contains(lastState.Caption, "MST complete: 1 vertex connected with total weight 0") {
		t.Errorf("unexpected caption for single node: %q", lastState.Caption)
	}
}

func TestPrimCycleRejection(t *testing.T) {
	nodes, edges, start := DefaultGraph()
	steps := Trace(nodes, edges, start)

	foundCycleStep := false
	for _, step := range steps {
		if step.LineID == "L9" {
			foundCycleStep = true
			if !strings.Contains(step.Explanation, "Skip cycle: both endpoints already in tree") {
				t.Errorf("expected cycle explanation 'Skip cycle: both endpoints already in tree', got %q", step.Explanation)
			}
			state, ok := step.State.(renderers.GraphState)
			if !ok {
				t.Fatalf("step state is %T, want GraphState", step.State)
			}
			// Crossing candidate edge should be active during cycle inspection
			activeCount := 0
			for _, e := range state.Edges {
				if e.Active {
					activeCount++
				}
			}
			if activeCount == 0 {
				t.Error("expected candidate edge to be active during cycle rejection step")
			}
		}
	}

	if !foundCycleStep {
		t.Error("expected at least one L9 cycle rejection step in default graph")
	}
}

func TestPrimStrictInputValidation(t *testing.T) {
	lesson := Lesson{}

	// nil should succeed with default graph
	steps, err := lesson.Build(nil)
	if err != nil || len(steps) == 0 {
		t.Fatalf("expected Build(nil) to succeed, got err=%v", err)
	}

	// null should be rejected
	if _, err := lesson.Build([]byte("null")); err == nil {
		t.Error("expected error for 'null' input, got nil")
	}

	// start null should be rejected
	if _, err := lesson.Build([]byte(`{"start": null}`)); err == nil {
		t.Error("expected error for '{\"start\": null}', got nil")
	}

	// malformed JSON should be rejected
	if _, err := lesson.Build([]byte(`{not valid json`)); err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}

	// unknown start node should be rejected
	if _, err := lesson.Build([]byte(`{"start": "Z"}`)); err == nil {
		t.Error("expected error for unknown start node 'Z', got nil")
	}

	// unknown start node with custom edges should be rejected
	customInvalid := []byte(`{
		"edges": [{"from": "X", "to": "Y", "weight": 1}],
		"start": "Z"
	}`)
	if _, err := lesson.Build(customInvalid); err == nil {
		t.Error("expected error for unknown start node in custom graph, got nil")
	}
}
