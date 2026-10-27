package dfs

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestTraverseCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		adj      map[string][]string
		start    string
		expected []string
	}{
		{
			name: "canonical tree-like graph",
			adj: map[string][]string{
				"A": {"B", "C"},
				"B": {"D", "E"},
				"C": {"E", "F"},
				"D": {},
				"E": {},
				"F": {},
			},
			start:    "A",
			expected: []string{"A", "B", "D", "E", "C", "F"},
		},
		{
			name: "graph with cycle",
			adj: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			start:    "A",
			expected: []string{"A", "B", "C"},
		},
		{
			name: "disconnected graph",
			adj: map[string][]string{
				"A": {"B"},
				"B": {},
				"C": {"D"},
				"D": {},
			},
			start:    "A",
			expected: []string{"A", "B"},
		},
		{
			name: "single node",
			adj: map[string][]string{
				"A": {},
			},
			start:    "A",
			expected: []string{"A"},
		},
		{
			name:     "empty start",
			adj:      map[string][]string{"A": {"B"}},
			start:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Traverse(tt.adj, tt.start)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestDFSTraceValidation(t *testing.T) {
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
	if meta.ID != "dfs" {
		t.Errorf("meta.ID = %q; want 'dfs'", meta.ID)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "graph" {
		t.Errorf("meta.Renderer = %q; want 'graph'", meta.Renderer)
	}
}

func TestDFSWorklistAndStates(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	sawDiveCaption := false
	sawBacktrackCaption := false
	sawActiveEdgeDuringDescent := false
	maxStackDepth := 0

	for i, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.GraphState", i, step.State)
		}

		if state.WorklistLabel != "Call Stack (LIFO)" {
			t.Errorf("step %d WorklistLabel = %q; want 'Call Stack (LIFO)'", i, state.WorklistLabel)
		}

		if len(state.Worklist) > maxStackDepth {
			maxStackDepth = len(state.Worklist)
		}

		// When stack is non-empty, verify top element matches active visiting node
		if len(state.Worklist) > 0 {
			top := state.Worklist[len(state.Worklist)-1]
			if !strings.HasPrefix(top.Label, "dfs(") || !strings.HasSuffix(top.Label, ")") {
				t.Errorf("step %d invalid stack frame label %q", i, top.Label)
			}

			// Top frame node should be in "visiting" or "visited" state
			found := false
			for _, n := range state.Nodes {
				if n.ID == top.ID {
					found = true
					if n.State != "visiting" && n.State != "visited" {
						t.Errorf("step %d: top of stack node %s has unexpected state %q", i, top.ID, n.State)
					}
					break
				}
			}
			if !found {
				t.Errorf("step %d: top of stack node %s not found in graph nodes", i, top.ID)
			}
		}

		if state.Caption == "Dive deeper: call DFS(neighbor); push to stack" {
			sawDiveCaption = true
			// During descent step, exactly one edge must be active
			activeCount := 0
			for _, e := range state.Edges {
				if e.Active {
					activeCount++
				}
			}
			if activeCount == 1 {
				sawActiveEdgeDuringDescent = true
			} else {
				t.Errorf("step %d (descent) expected exactly 1 active edge, got %d", i, activeCount)
			}
		}

		if state.Caption == "No unvisited neighbors: backtrack; pop from stack" {
			sawBacktrackCaption = true
		}
	}

	if maxStackDepth < 2 {
		t.Errorf("max stack depth was %d; expected at least 2 for recursive DFS", maxStackDepth)
	}
	if !sawDiveCaption {
		t.Errorf("missing expected caption 'Dive deeper: call DFS(neighbor); push to stack'")
	}
	if !sawBacktrackCaption {
		t.Errorf("missing expected caption 'No unvisited neighbors: backtrack; pop from stack'")
	}
	if !sawActiveEdgeDuringDescent {
		t.Errorf("expected active edge highlighted during descent")
	}

	// Initial step should have empty worklist and no active edges
	firstState := steps[0].State.(renderers.GraphState)
	if len(firstState.Worklist) != 0 {
		t.Errorf("initial step worklist length = %d; want 0", len(firstState.Worklist))
	}
	for _, e := range firstState.Edges {
		if e.Active {
			t.Errorf("initial step edge %s->%s is active", e.From, e.To)
		}
	}

	// Final step should have all reachable nodes visited, empty stack, no active edges
	finalState := steps[len(steps)-1].State.(renderers.GraphState)
	if len(finalState.Worklist) != 0 {
		t.Errorf("final step worklist length = %d; want 0", len(finalState.Worklist))
	}
	for _, node := range finalState.Nodes {
		if node.State != "visited" {
			t.Errorf("expected node %s to be visited, got %q", node.ID, node.State)
		}
	}
	for _, edge := range finalState.Edges {
		if edge.Active {
			t.Errorf("expected edge (%s -> %s) to be inactive at end", edge.From, edge.To)
		}
	}
}

func TestDFSSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	// Capture step 5 (inside traversal with active nodes and worklist)
	targetStepIdx := 5
	if len(steps) <= targetStepIdx {
		t.Fatalf("not enough steps: %d", len(steps))
	}

	state5 := steps[targetStepIdx].State.(renderers.GraphState)
	origNode0State := state5.Nodes[0].State
	origEdge0Active := state5.Edges[0].Active

	// Mutate step 5
	state5.Nodes[0].State = "corrupted_state"
	state5.Edges[0].Active = !origEdge0Active
	if len(state5.Worklist) > 0 {
		state5.Worklist[0].Label = "corrupted_label"
	}

	// Rebuild and verify step 5 in a fresh build matches original values
	freshSteps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build fresh trace: %v", err)
	}
	freshState5 := freshSteps[targetStepIdx].State.(renderers.GraphState)
	if freshState5.Nodes[0].State != origNode0State {
		t.Errorf("snapshot mutation leaked: got %q, want %q", freshState5.Nodes[0].State, origNode0State)
	}
	if freshState5.Edges[0].Active != origEdge0Active {
		t.Errorf("snapshot edge mutation leaked")
	}

	// Cross-step immutability check: mutating step i must not alter step j
	for i := 0; i < len(steps)-1; i++ {
		stI := steps[i].State.(renderers.GraphState)
		stNext := steps[i+1].State.(renderers.GraphState)

		// Record stNext properties
		nextEdge0Active := stNext.Edges[0].Active

		// Mutate stI
		stI.Nodes[0].State = "mutated_in_test"
		stI.Edges[0].Active = !stI.Edges[0].Active
		if len(stI.Worklist) > 0 {
			stI.Worklist[0].ID = "mutated"
		}

		// Verify stNext is unaffected
		if stNext.Edges[0].Active != nextEdge0Active {
			t.Fatalf("step %d mutation modified step %d edge state", i, i+1)
		}
	}
}

func TestDFSDisconnectedGraph(t *testing.T) {
	customNodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 100, Y: 100},
		{ID: "B", Label: "B", X: 200, Y: 100},
		{ID: "C", Label: "C", X: 300, Y: 100},
		{ID: "D", Label: "D", X: 400, Y: 100},
	}
	customEdges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "C", To: "D"},
	}

	steps := Trace(customNodes, customEdges, "A")
	if len(steps) == 0 {
		t.Fatal("expected steps for disconnected graph")
	}

	finalStep := steps[len(steps)-1]
	finalState := finalStep.State.(renderers.GraphState)

	visitedCount := 0
	unvisitedCount := 0
	for _, n := range finalState.Nodes {
		switch n.State {
		case "visited":
			visitedCount++
			if n.ID != "A" && n.ID != "B" {
				t.Errorf("unexpected visited node: %s", n.ID)
			}
		case "unvisited":
			unvisitedCount++
			if n.ID != "C" && n.ID != "D" {
				t.Errorf("unexpected unvisited node: %s", n.ID)
			}
		default:
			t.Errorf("node %s has unexpected final state %q", n.ID, n.State)
		}
	}

	if visitedCount != 2 || unvisitedCount != 2 {
		t.Fatalf("expected 2 visited and 2 unvisited nodes, got %d visited, %d unvisited", visitedCount, unvisitedCount)
	}

	// Final order should be ["A", "B"]
	finalOrder, ok := finalStep.Variables["order"].([]string)
	if !ok || !reflect.DeepEqual(finalOrder, []string{"A", "B"}) {
		t.Errorf("expected final order [A, B], got %v", finalStep.Variables["order"])
	}
}

func TestDFSSingleNode(t *testing.T) {
	singleNode := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 200, Y: 200},
	}
	singleEdge := []renderers.GraphEdge{}

	steps := Trace(singleNode, singleEdge, "A")
	if len(steps) == 0 {
		t.Fatal("expected steps for single node")
	}

	finalStep := steps[len(steps)-1]
	finalState := finalStep.State.(renderers.GraphState)
	if len(finalState.Nodes) != 1 || finalState.Nodes[0].State != "visited" {
		t.Fatalf("expected single node visited, got %v", finalState.Nodes)
	}
	if len(finalState.Worklist) != 0 {
		t.Errorf("expected empty worklist at end of single node trace")
	}
	finalOrder := finalStep.Variables["order"]
	if !reflect.DeepEqual(finalOrder, []string{"A"}) {
		t.Errorf("expected final order [A], got %v", finalOrder)
	}
}

func TestDFSInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid nil input (defaults to A)
	if steps, err := lesson.Build(nil); err != nil || len(steps) == 0 {
		t.Fatalf("build with nil failed: %v", err)
	}

	// Valid string input
	if steps, err := lesson.Build(json.RawMessage(`"B"`)); err != nil || len(steps) == 0 {
		t.Fatalf("build with 'B' failed: %v", err)
	}

	// Valid object input
	if steps, err := lesson.Build(json.RawMessage(`{"start": "B"}`)); err != nil || len(steps) == 0 {
		t.Fatalf("build with {'start': 'B'} failed: %v", err)
	}

	// Rejections
	invalidCases := []struct {
		name  string
		input string
	}{
		{"null input", "null"},
		{"null start", `{"start": null}`},
		{"empty start", `{"start": ""}`},
		{"unknown start string", `"UNKNOWN"`},
		{"unknown start object", `{"start": "UNKNOWN"}`},
		{"malformed JSON syntax", `{"start": `},
		{"numeric input", `123`},
		{"array input", `[1, 2, 3]`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Fatalf("expected error for %s (%s), but got nil", tc.name, tc.input)
			}
		})
	}
}
