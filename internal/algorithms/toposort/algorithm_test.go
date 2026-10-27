package toposort

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSortValidDAG(t *testing.T) {
	nodes := []string{"code", "test", "lint", "build", "deploy"}
	edges := [][2]string{
		{"code", "lint"},
		{"code", "test"},
		{"lint", "build"},
		{"test", "build"},
		{"build", "deploy"},
	}

	order, ok := Sort(nodes, edges)
	if !ok {
		t.Fatalf("expected valid DAG, got ok = false")
	}
	if len(order) != len(nodes) {
		t.Fatalf("expected %d nodes in order, got %d (%v)", len(nodes), len(order), order)
	}

	// Verify topological order: for every edge (u, v), pos(u) < pos(v)
	pos := make(map[string]int)
	for i, n := range order {
		pos[n] = i
	}
	for _, e := range edges {
		u, v := e[0], e[1]
		if pos[u] >= pos[v] {
			t.Errorf("topological order violated: %s (pos %d) came after or equal to %s (pos %d)", u, pos[u], v, pos[v])
		}
	}
}

func TestSortCycleDetection(t *testing.T) {
	nodes := []string{"A", "B", "C"}
	edges := [][2]string{
		{"A", "B"},
		{"B", "C"},
		{"C", "A"},
	}

	order, ok := Sort(nodes, edges)
	if ok {
		t.Fatalf("expected cycle to be detected (ok = false), got ok = true with order %v", order)
	}
	if len(order) == len(nodes) {
		t.Fatalf("expected order length < node count on cycle, got %d", len(order))
	}
}

func TestSortEmpty(t *testing.T) {
	order, ok := Sort(nil, nil)
	if !ok || len(order) != 0 {
		t.Fatalf("expected empty valid sort, got %v, %v", order, ok)
	}

	order2, ok2 := Sort([]string{}, [][2]string{})
	if !ok2 || len(order2) != 0 {
		t.Fatalf("expected empty valid sort from empty slice, got %v, %v", order2, ok2)
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "toposort" {
		t.Errorf("meta.ID = %q, want 'toposort'", meta.ID)
	}
	if meta.Name != "Topological Sort (DAG)" {
		t.Errorf("meta.Name = %q, want 'Topological Sort (DAG)'", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q, want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "graph" {
		t.Errorf("meta.Renderer = %q, want 'graph'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(V + E)" {
		t.Errorf("timeComplexity = %q, want 'O(V + E)'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V)" {
		t.Errorf("spaceComplexity = %q, want 'O(V)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("difficulty = %q, want 'Medium'", meta.Documentation.Difficulty)
	}

	requiredConcepts := []string{
		"Directed Acyclic Graph",
		"In-Degree Zero Queue",
		"Dependency Resolution",
		"Cycle Detection",
	}
	conceptMap := make(map[string]bool)
	for _, c := range meta.Documentation.KeyConcepts {
		conceptMap[c] = true
	}
	for _, req := range requiredConcepts {
		if !conceptMap[req] {
			t.Errorf("missing key concept %q in %v", req, meta.Documentation.KeyConcepts)
		}
	}
}

func TestLessonRegistered(t *testing.T) {
	registered, ok := catalog.Get("toposort")
	if !ok || registered == nil {
		t.Fatal("toposort lesson not registered in catalog")
	}
}

func TestTeachingTraceDecisions(t *testing.T) {
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

	var sawInitialInDegrees bool
	var sawVisiting bool
	var sawVisited bool
	var sawActiveEdge bool
	var sawInDegreeZeroAdded bool

	for _, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.GraphState", step.Index, step.State)
		}

		if state.WorklistLabel != "Ready Queue (In-Degree 0)" {
			t.Fatalf("step %d WorklistLabel = %q, want 'Ready Queue (In-Degree 0)'", step.Index, state.WorklistLabel)
		}

		// Verify worklist items match queue variable
		if qVar, ok := step.Variables["queue"]; ok {
			qSlice := qVar.([]string)
			if len(state.Worklist) != len(qSlice) {
				t.Fatalf("step %d worklist len %d != queue var len %d", step.Index, len(state.Worklist), len(qSlice))
			}
			for i, item := range state.Worklist {
				if item.ID != qSlice[i] || item.Label != qSlice[i] {
					t.Fatalf("step %d worklist item %d mismatch: got %+v, want %q", step.Index, i, item, qSlice[i])
				}
			}
		}

		// Check L2 initial in-degrees caption
		if step.LineID == "L2" {
			if strings.Contains(state.Caption, "clean: 0") && strings.Contains(state.Caption, "package: 3") {
				sawInitialInDegrees = true
			}
		}

		// Check L7 dequeuing marks visiting
		if step.LineID == "L7" && step.Action == trace.ActionPop {
			curr := step.Variables["curr"].(string)
			for _, n := range state.Nodes {
				if n.ID == curr && n.State == "visiting" {
					sawVisiting = true
				}
			}
		}

		// Check L8 order append marks visited
		if step.LineID == "L8" && step.Action == trace.ActionVisit {
			curr := step.Variables["curr"].(string)
			for _, n := range state.Nodes {
				if n.ID == curr && n.State == "visited" {
					sawVisited = true
				}
			}
		}

		// Check L9 / L10 active outgoing edge
		if step.LineID == "L9" || step.LineID == "L10" {
			curr := step.Variables["curr"].(string)
			neighbor := step.Variables["neighbor"].(string)
			for _, e := range state.Edges {
				if e.From == curr && e.To == neighbor && e.Active {
					sawActiveEdge = true
				}
			}
		}

		// Check L12 in-degree reached 0 caption
		if step.LineID == "L12" && step.Action == trace.ActionPush {
			if state.Caption == "In-degree reached 0; added to ready queue" {
				sawInDegreeZeroAdded = true
			}
		}
	}

	if !sawInitialInDegrees {
		t.Error("expected step L2 to show initial in-degrees in caption")
	}
	if !sawVisiting {
		t.Error("expected step L7 to mark current node 'visiting'")
	}
	if !sawVisited {
		t.Error("expected step L8 to mark current node 'visited'")
	}
	if !sawActiveEdge {
		t.Error("expected step L9/L10 to show active outgoing edge")
	}
	if !sawInDegreeZeroAdded {
		t.Error("expected step L12 to have caption 'In-degree reached 0; added to ready queue'")
	}

	// Verify final step caption shows complete topological order
	last := steps[len(steps)-1]
	lastState := last.State.(renderers.GraphState)
	if !strings.HasPrefix(lastState.Caption, "Topological order:") {
		t.Fatalf("final step caption = %q, want prefix 'Topological order:'", lastState.Caption)
	}
	if last.Variables["status"] != "complete" || last.Variables["valid"] != true {
		t.Fatalf("final step status/valid wrong: %v, %v", last.Variables["status"], last.Variables["valid"])
	}
}

func TestSnapshotImmutability(t *testing.T) {
	steps := Trace([]renderers.GraphNode{
		{ID: "A", Label: "A", State: "unvisited"},
		{ID: "B", Label: "B", State: "unvisited"},
	}, []renderers.GraphEdge{
		{From: "A", To: "B", Directed: true},
	})

	if len(steps) < 2 {
		t.Fatalf("expected multiple steps, got %d", len(steps))
	}

	firstState := steps[0].State.(renderers.GraphState)
	lastState := steps[len(steps)-1].State.(renderers.GraphState)

	// Mutate last state's slices
	lastState.Nodes[0].State = "corrupted"
	lastState.Nodes[0].Label = "corrupted"
	lastState.Edges[0].Active = true
	if len(lastState.Worklist) > 0 {
		lastState.Worklist[0].Label = "corrupted"
	}

	// Verify first state was NOT affected
	if firstState.Nodes[0].State == "corrupted" || firstState.Nodes[0].Label == "corrupted" {
		t.Fatal("mutating last step node corrupted first step node")
	}
	if firstState.Edges[0].Active == true {
		t.Fatal("mutating last step edge corrupted first step edge")
	}

	// Test intermediate step worklist immutability
	for _, step := range steps {
		s := step.State.(renderers.GraphState)
		if len(s.Worklist) > 0 {
			origID := s.Worklist[0].ID
			s.Worklist[0].ID = "mutated"
			s.Worklist[0].Label = "mutated"
			freshSteps := Trace([]renderers.GraphNode{
				{ID: "A", Label: "A", State: "unvisited"},
				{ID: "B", Label: "B", State: "unvisited"},
			}, []renderers.GraphEdge{
				{From: "A", To: "B", Directed: true},
			})
			freshState := freshSteps[step.Index].State.(renderers.GraphState)
			if freshState.Worklist[0].ID != origID {
				t.Fatalf("worklist aliased across runs: got %q, want %q", freshState.Worklist[0].ID, origID)
			}
			break
		}
	}
}

func TestTraceCycleDetection(t *testing.T) {
	// Full cycle
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", State: "unvisited"},
		{ID: "B", Label: "B", State: "unvisited"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B", Directed: true},
		{From: "B", To: "A", Directed: true},
	}

	steps := Trace(nodes, edges)
	if len(steps) == 0 {
		t.Fatal("expected trace steps for cyclic graph, got 0")
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("cyclic trace line validation failed: %v", err)
	}

	last := steps[len(steps)-1]
	lastState := last.State.(renderers.GraphState)

	if lastState.Caption != "Cycle detected! Remaining vertices cannot be resolved" {
		t.Errorf("last caption = %q, want 'Cycle detected! Remaining vertices cannot be resolved'", lastState.Caption)
	}
	if last.Variables["status"] != "cycle_detected" {
		t.Errorf("last step status = %v, want 'cycle_detected'", last.Variables["status"])
	}
	if last.Variables["valid"] != false {
		t.Errorf("last step valid = %v, want false", last.Variables["valid"])
	}

	// Partial cycle: C -> A -> B -> A
	nodesPartial := []renderers.GraphNode{
		{ID: "C", Label: "C", State: "unvisited"},
		{ID: "A", Label: "A", State: "unvisited"},
		{ID: "B", Label: "B", State: "unvisited"},
	}
	edgesPartial := []renderers.GraphEdge{
		{From: "C", To: "A", Directed: true},
		{From: "A", To: "B", Directed: true},
		{From: "B", To: "A", Directed: true},
	}
	stepsPartial := Trace(nodesPartial, edgesPartial)
	if err := catalog.ValidateSourceLineIDs(lesson, stepsPartial); err != nil {
		t.Fatalf("partial cyclic trace line validation failed: %v", err)
	}
	lastPartial := stepsPartial[len(stepsPartial)-1]
	lastPartialState := lastPartial.State.(renderers.GraphState)
	if lastPartialState.Caption != "Cycle detected! Remaining vertices cannot be resolved" {
		t.Errorf("partial cycle caption = %q, want 'Cycle detected! Remaining vertices cannot be resolved'", lastPartialState.Caption)
	}
	if lastPartial.Variables["status"] != "cycle_detected" || lastPartial.Variables["valid"] != false {
		t.Errorf("partial cycle status/valid wrong: %v, %v", lastPartial.Variables["status"], lastPartial.Variables["valid"])
	}
}

func TestTraceEmptyGraph(t *testing.T) {
	for _, steps := range [][]trace.Step{
		Trace(nil, nil),
		Trace([]renderers.GraphNode{}, []renderers.GraphEdge{}),
	} {
		if len(steps) != 1 {
			t.Fatalf("expected 1 step for empty graph, got %d", len(steps))
		}
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatalf("empty graph line validation failed: %v", err)
		}
		step := steps[0]
		state := step.State.(renderers.GraphState)
		if state.WorklistLabel != "Ready Queue (In-Degree 0)" {
			t.Errorf("worklist label = %q, want 'Ready Queue (In-Degree 0)'", state.WorklistLabel)
		}
		if len(state.Worklist) != 0 {
			t.Errorf("worklist len = %d, want 0", len(state.Worklist))
		}
		if state.Caption != "Topological order: []" {
			t.Errorf("caption = %q, want 'Topological order: []'", state.Caption)
		}
		if step.Variables["status"] != "complete" || step.Variables["valid"] != true {
			t.Errorf("empty graph status/valid wrong: %v, %v", step.Variables["status"], step.Variables["valid"])
		}
	}

	// Test Build with empty nodes
	lesson := Lesson{}
	steps, err := lesson.Build(json.RawMessage(`{"nodes": []}`))
	if err != nil {
		t.Fatalf("failed to build empty graph: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step for empty graph build, got %d", len(steps))
	}
}

func TestLessonBuildValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid default inputs
	for _, input := range []json.RawMessage{nil, []byte(""), []byte("   "), []byte("{}")} {
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("expected valid build for default input %q, got err: %v", string(input), err)
		}
		if len(steps) == 0 {
			t.Fatalf("expected steps for default input %q, got 0", string(input))
		}
	}

	// Strict invalid inputs: null, malformed JSON, wrong types
	invalidInputs := []json.RawMessage{
		[]byte("null"),
		[]byte("   null   "),
		[]byte("not json"),
		[]byte("[1, 2, 3]"),
		[]byte(`"string"`),
		[]byte(`{"nodes": "not an array"}`),
		[]byte(`{"nodes": [{"id": ""}]}`),
		[]byte(`{"nodes": [{"id": "A"}], "edges": [{"from": "", "to": "A"}]}`),
		[]byte(`{"nodes": [{"id": "A"}], "edges": [{"from": "A", "to": "B"}]}`),
	}

	for _, invalid := range invalidInputs {
		_, err := lesson.Build(invalid)
		if err == nil {
			t.Fatalf("expected build error for invalid input: %s", string(invalid))
		}
	}

	// Valid custom input
	customJSON := []byte(`{
		"nodes": [{"id": "CS101", "label": "CS101", "x": 50, "y": 50}, {"id": "CS102", "label": "CS102", "x": 150, "y": 50}],
		"edges": [{"from": "CS101", "to": "CS102"}]
	}`)
	steps, err := lesson.Build(customJSON)
	if err != nil {
		t.Fatalf("custom Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected trace steps from custom input, got 0")
	}
	last := steps[len(steps)-1]
	if last.Variables["valid"] != true {
		t.Errorf("expected valid DAG for custom schedule, got valid = false")
	}
}
