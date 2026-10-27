package bfs

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
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

func TestBFSInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid nil input uses defaults
	steps, err := lesson.Build(nil)
	if err != nil || len(steps) == 0 {
		t.Fatalf("Build(nil) expected success, got steps=%d err=%v", len(steps), err)
	}

	// Null input must be rejected
	if _, err := lesson.Build([]byte("null")); err == nil {
		t.Fatalf("Build(null) expected error, got nil")
	}
	if _, err := lesson.Build([]byte("  null  ")); err == nil {
		t.Fatalf("Build('  null  ') expected error, got nil")
	}
	if _, err := lesson.Build([]byte(`{"start": null}`)); err == nil {
		t.Fatalf("Build({'start': null}) expected error, got nil")
	}

	// Malformed JSON must be rejected
	if _, err := lesson.Build([]byte("{bad json")); err == nil {
		t.Fatalf("Build(malformed json) expected error, got nil")
	}

	// Unknown start node must be rejected
	if _, err := lesson.Build([]byte(`{"start": "UNKNOWN"}`)); err == nil {
		t.Fatalf("Build(unknown start) expected error, got nil")
	}
	if _, err := lesson.Build([]byte(`"UNKNOWN"`)); err == nil {
		t.Fatalf("Build(unknown start string) expected error, got nil")
	}

	// Valid custom start
	customSteps, err := lesson.Build([]byte(`{"start": "B"}`))
	if err != nil || len(customSteps) == 0 {
		t.Fatalf("Build({'start': 'B'}) failed: %v", err)
	}
}

func TestBFSTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenStates := make(map[string]bool)
	sawActiveEdge := false
	sawExploreCaption := false

	for _, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step state is %T, want renderers.GraphState", step.State)
		}
		if state.Caption == "" {
			t.Errorf("step %s (%s) has empty Caption", step.LineID, step.Action)
		}
		if strings.Contains(state.Caption, "Dequeue front node") && strings.Contains(state.Caption, "inspect unvisited neighbors") {
			sawExploreCaption = true
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
	if !sawExploreCaption {
		t.Errorf("expected caption explaining unvisited neighbors inspection")
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

func TestBFSWorklistContents(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	for i, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.GraphState", i, step.State)
		}

		if state.WorklistLabel != "Queue (FIFO)" {
			t.Errorf("step %d WorklistLabel = %q; want 'Queue (FIFO)'", i, state.WorklistLabel)
		}

		// When queue variable is recorded in step variables, worklist must match in exact FIFO order
		if qVar, ok := step.Variables["queue"]; ok {
			var qSlice []string
			switch v := qVar.(type) {
			case []string:
				qSlice = v
			case []any:
				for _, item := range v {
					if str, ok := item.(string); ok {
						qSlice = append(qSlice, str)
					}
				}
			}

			if len(state.Worklist) != len(qSlice) {
				t.Fatalf("step %d: worklist length %d != queue length %d", i, len(state.Worklist), len(qSlice))
			}
			for pos, item := range state.Worklist {
				if item.ID != qSlice[pos] || item.Label != qSlice[pos] {
					t.Errorf("step %d worklist[%d] = {%s, %s}; want %s", i, pos, item.ID, item.Label, qSlice[pos])
				}
			}
		}
	}

	// Final step worklist must be empty
	lastState := steps[len(steps)-1].State.(renderers.GraphState)
	if len(lastState.Worklist) != 0 {
		t.Errorf("final step worklist has %d items; want 0", len(lastState.Worklist))
	}
}

func TestBFSSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps1, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	steps2, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Mutate steps1 snapshots
	for _, step := range steps1 {
		state := step.State.(renderers.GraphState)
		if len(state.Nodes) > 0 {
			state.Nodes[0].State = "MUTATED"
			state.Nodes[0].Label = "MUTATED"
		}
		if len(state.Edges) > 0 {
			state.Edges[0].Active = true
		}
		if len(state.Worklist) > 0 {
			state.Worklist[0].ID = "MUTATED"
		}
	}

	// Verify steps2 was completely unaffected
	for i, step := range steps2 {
		state := step.State.(renderers.GraphState)
		if len(state.Nodes) > 0 {
			if state.Nodes[0].State == "MUTATED" || state.Nodes[0].Label == "MUTATED" {
				t.Fatalf("step %d Nodes mutated via shared reference", i)
			}
		}
		if len(state.Worklist) > 0 {
			if state.Worklist[0].ID == "MUTATED" {
				t.Fatalf("step %d Worklist mutated via shared reference", i)
			}
		}
	}
}

func TestBFSVisitOrder(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	var visitedOrder []string
	for _, step := range steps {
		if step.Action == trace.ActionVisit {
			if curr, ok := step.Variables["curr"].(string); ok {
				visitedOrder = append(visitedOrder, curr)
			}
		}
	}

	expected := []string{"A", "B", "C", "D", "E", "F"}
	if !reflect.DeepEqual(visitedOrder, expected) {
		t.Fatalf("visited order got %v, want %v", visitedOrder, expected)
	}
}

func TestBFSDisconnectedGraph(t *testing.T) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A"},
		{ID: "B", Label: "B"},
		{ID: "C", Label: "C"},
		{ID: "D", Label: "D"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "C", To: "D"},
	}

	input, err := json.Marshal(buildInput{
		Nodes: nodes,
		Edges: edges,
		Start: func() *string { s := "A"; return &s }(),
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	lesson := Lesson{}
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("0 steps produced")
	}

	lastState := steps[len(steps)-1].State.(renderers.GraphState)
	for _, n := range lastState.Nodes {
		if n.ID == "A" || n.ID == "B" {
			if n.State != "visited" {
				t.Errorf("reachable node %s state = %q; want 'visited'", n.ID, n.State)
			}
		} else {
			if n.State != "unvisited" {
				t.Errorf("unreachable node %s state = %q; want 'unvisited'", n.ID, n.State)
			}
		}
	}

	var visitOrder []string
	for _, step := range steps {
		if step.Action == trace.ActionVisit {
			if curr, ok := step.Variables["curr"].(string); ok {
				visitOrder = append(visitOrder, curr)
			}
		}
	}
	expectedOrder := []string{"A", "B"}
	if !reflect.DeepEqual(visitOrder, expectedOrder) {
		t.Fatalf("disconnected graph visit order = %v; want %v", visitOrder, expectedOrder)
	}
}

func TestBFSSingleNode(t *testing.T) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A"},
	}
	edges := []renderers.GraphEdge{}

	steps := Trace(nodes, edges, "A")
	if len(steps) == 0 {
		t.Fatal("expected steps for single node BFS")
	}

	var visitOrder []string
	for _, step := range steps {
		if step.Action == trace.ActionVisit {
			if curr, ok := step.Variables["curr"].(string); ok {
				visitOrder = append(visitOrder, curr)
			}
		}
	}
	expectedOrder := []string{"A"}
	if !reflect.DeepEqual(visitOrder, expectedOrder) {
		t.Fatalf("single node visit order = %v; want %v", visitOrder, expectedOrder)
	}

	lastState := steps[len(steps)-1].State.(renderers.GraphState)
	if lastState.Nodes[0].State != "visited" {
		t.Errorf("node A final state = %q; want 'visited'", lastState.Nodes[0].State)
	}
	if len(lastState.Worklist) != 0 {
		t.Errorf("final worklist has %d items; want 0", len(lastState.Worklist))
	}
}
