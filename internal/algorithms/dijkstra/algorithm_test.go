package dijkstra

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
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

func TestDijkstraBuildStrictValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input []byte
	}{
		{"null literal", []byte("null")},
		{"malformed JSON syntax", []byte("{start: A")},
		{"JSON array instead of object", []byte(`["A", "B"]`)},
		{"JSON number", []byte("42")},
		{"non-existent start node", []byte(`{"start": "Z"}`)},
		{"non-existent target node", []byte(`{"target": "Z"}`)},
		{"negative edge weight", []byte(`{"edges": [{"from": "A", "to": "B", "weight": -5}]}`)},
		{"unknown edge from node", []byte(`{"edges": [{"from": "UNKNOWN", "to": "B", "weight": 2}]}`)},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}

	validInputs := []struct {
		name  string
		input []byte
	}{
		{"empty object uses defaults", []byte(`{}`)},
		{"explicit valid start and target", []byte(`{"start": "A", "target": "C"}`)},
		{"same start and target", []byte(`{"start": "B", "target": "B"}`)},
	}

	for _, tc := range validInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
			if len(steps) == 0 {
				t.Fatalf("expected steps for %s, got 0", tc.name)
			}
		})
	}
}

func TestDijkstraTracePropertiesAndFinalStep(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last step action %v, got %v", trace.ActionDone, lastStep.Action)
	}
	if lastStep.LineID != "L17" {
		t.Fatalf("expected last step LineID L17, got %s", lastStep.LineID)
	}

	state, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	// Final cost in caption
	if !strings.Contains(state.Caption, "10") {
		t.Errorf("expected final cost 10 in caption, got: %q", state.Caption)
	}

	// Final path nodes in "path" state
	pathNodes := 0
	for _, n := range state.Nodes {
		if n.State == "path" {
			pathNodes++
		}
	}
	if pathNodes != 5 {
		t.Errorf("expected 5 path nodes in final state, got %d", pathNodes)
	}

	// Final path edges must be active
	activeEdges := 0
	for _, e := range state.Edges {
		if e.Active {
			activeEdges++
		}
	}
	if activeEdges != 4 {
		t.Errorf("expected 4 active path edges in final state, got %d", activeEdges)
	}

	// Distances table in variables
	if _, ok := lastStep.Variables["distances"]; !ok {
		t.Errorf("expected 'distances' table in step variables, got %v", lastStep.Variables)
	}
}

func TestDijkstraTraceStartEqualsTarget(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"start": "A", "target": "A"}`)
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected renderers.GraphState, got %T", lastStep.State)
	}

	if !strings.Contains(state.Caption, "final cost: 0") {
		t.Errorf("expected final cost 0 in caption, got %q", state.Caption)
	}

	pathNodes := 0
	for _, n := range state.Nodes {
		if n.State == "path" {
			pathNodes++
			if n.ID != "A" {
				t.Errorf("expected path node to be A, got %s", n.ID)
			}
		}
	}
	if pathNodes != 1 {
		t.Errorf("expected exactly 1 path node, got %d", pathNodes)
	}
}

func TestDijkstraTraceUnreachableTarget(t *testing.T) {
	lesson := Lesson{}
	customInput := LessonInput{
		Start:  "A",
		Target: "Z",
		Nodes: []renderers.GraphNode{
			{ID: "A", Label: "A", X: 100, Y: 100, State: "unvisited"},
			{ID: "B", Label: "B", X: 200, Y: 100, State: "unvisited"},
			{ID: "Z", Label: "Z", X: 300, Y: 100, State: "unvisited"},
		},
		Edges: []renderers.GraphEdge{
			{From: "A", To: "B", Weight: 2, Directed: true},
		},
	}
	raw, err := json.Marshal(customInput)
	if err != nil {
		t.Fatalf("failed to marshal custom input: %v", err)
	}

	steps, err := lesson.Build(raw)
	if err != nil {
		t.Fatalf("unexpected error for unreachable target: %v", err)
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected renderers.GraphState, got %T", lastStep.State)
	}

	if !strings.Contains(strings.ToLower(state.Caption), "unreachable") {
		t.Errorf("expected caption to indicate unreachable, got %q", state.Caption)
	}

	for _, n := range state.Nodes {
		if n.State == "path" {
			t.Errorf("node %s unexpectedly marked as path in unreachable search", n.ID)
		}
	}
}

func TestDijkstraWorklistContents(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	expectedWorklistLabel := "Priority Queue (Min-Heap: node, dist)"
	itemRegex := regexp.MustCompile(`^[A-Z] \(d=(\d+)\)$`)

	sawNonEmptyWorklist := false

	for i, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is not GraphState", i)
		}

		if state.WorklistLabel != expectedWorklistLabel {
			t.Errorf("step %d: worklistLabel = %q, want %q", i, state.WorklistLabel, expectedWorklistLabel)
		}

		if len(state.Worklist) > 0 {
			sawNonEmptyWorklist = true
			var prevDist int
			for j, item := range state.Worklist {
				matches := itemRegex.FindStringSubmatch(item.Label)
				if len(matches) != 2 {
					t.Errorf("step %d item %d: label %q does not match expected pattern 'Node (d=X)'", i, j, item.Label)
					continue
				}
				d, _ := strconv.Atoi(matches[1])
				if j > 0 && d < prevDist {
					t.Errorf("step %d: worklist is not sorted by distance: item %d has d=%d, previous had d=%d", i, j, d, prevDist)
				}
				prevDist = d
			}
		}
	}

	if !sawNonEmptyWorklist {
		t.Fatal("expected at least one step with a non-empty worklist")
	}

	// Verify L4 step (pushing start node) has start node with d=0
	foundStartPush := false
	for _, step := range steps {
		if step.LineID == "L4" {
			state := step.State.(renderers.GraphState)
			if len(state.Worklist) == 1 && state.Worklist[0].Label == "A (d=0)" {
				foundStartPush = true
				break
			}
		}
	}
	if !foundStartPush {
		t.Error("expected step L4 to have Worklist with 'A (d=0)'")
	}
}

func TestDijkstraSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	// Snapshot all step states
	type savedState struct {
		nodes    []renderers.GraphNode
		edges    []renderers.GraphEdge
		worklist []renderers.WorklistItem
	}

	saved := make([]savedState, len(steps))
	for i, step := range steps {
		gs := step.State.(renderers.GraphState)
		nodesCopy := make([]renderers.GraphNode, len(gs.Nodes))
		copy(nodesCopy, gs.Nodes)
		edgesCopy := make([]renderers.GraphEdge, len(gs.Edges))
		copy(edgesCopy, gs.Edges)
		wlCopy := make([]renderers.WorklistItem, len(gs.Worklist))
		copy(wlCopy, gs.Worklist)
		saved[i] = savedState{nodes: nodesCopy, edges: edgesCopy, worklist: wlCopy}
	}

	// Mutate every step's slices in-place
	for _, step := range steps {
		gs := step.State.(renderers.GraphState)
		if len(gs.Nodes) > 0 {
			gs.Nodes[0].State = "CORRUPTED_NODE"
		}
		if len(gs.Edges) > 0 {
			gs.Edges[0].Active = !gs.Edges[0].Active
		}
		if len(gs.Worklist) > 0 {
			gs.Worklist[0].Label = "CORRUPTED_WORKLIST"
		}
	}

	// Check that a fresh build matches the original saved states
	freshSteps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build fresh steps: %v", err)
	}

	for i := range freshSteps {
		gs := freshSteps[i].State.(renderers.GraphState)
		if !reflect.DeepEqual(gs.Nodes, saved[i].nodes) {
			t.Fatalf("step %d nodes mutated across builds or shared references", i)
		}
		if !reflect.DeepEqual(gs.Edges, saved[i].edges) {
			t.Fatalf("step %d edges mutated across builds or shared references", i)
		}
		if !reflect.DeepEqual(gs.Worklist, saved[i].worklist) {
			t.Fatalf("step %d worklist mutated across builds or shared references", i)
		}
	}

	// Also verify that adjacent steps do not share underlying slice memory
	for i := 0; i < len(freshSteps)-1; i++ {
		gs1 := freshSteps[i].State.(renderers.GraphState)
		gs2 := freshSteps[i+1].State.(renderers.GraphState)

		if len(gs1.Nodes) > 0 && len(gs2.Nodes) > 0 && &gs1.Nodes[0] == &gs2.Nodes[0] {
			t.Errorf("steps %d and %d share Nodes slice memory", i, i+1)
		}
		if len(gs1.Edges) > 0 && len(gs2.Edges) > 0 && &gs1.Edges[0] == &gs2.Edges[0] {
			t.Errorf("steps %d and %d share Edges slice memory", i, i+1)
		}
		if len(gs1.Worklist) > 0 && len(gs2.Worklist) > 0 && &gs1.Worklist[0] == &gs2.Worklist[0] {
			t.Errorf("steps %d and %d share Worklist slice memory", i, i+1)
		}
	}
}

func TestDijkstraEdgeRelaxationVisualization(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	sawRelaxationStep := false
	for _, step := range steps {
		if step.LineID == "L10" {
			state := step.State.(renderers.GraphState)
			if relaxed, ok := step.Variables["relaxed"].(bool); ok && relaxed {
				sawRelaxationStep = true
				// In relaxation step, an active edge should be shown
				hasActiveEdge := false
				for _, e := range state.Edges {
					if e.Active {
						hasActiveEdge = true
						break
					}
				}
				if !hasActiveEdge {
					t.Errorf("step %s relaxation has no active edge: %q", step.LineID, state.Caption)
				}
				if !strings.Contains(state.Caption, "shorter path found") {
					t.Errorf("step %s relaxation caption missing explanation: %q", step.LineID, state.Caption)
				}
			}
		}
	}

	if !sawRelaxationStep {
		t.Fatal("expected at least one edge relaxation step in trace")
	}
}
