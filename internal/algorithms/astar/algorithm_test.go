package astar

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
	if _, ok := lastStep.State.(renderers.GraphState); !ok {
		t.Fatalf("expected state type renderers.GraphState, got %T", lastStep.State)
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

func TestAStarWorklistAndCaptions(t *testing.T) {
	nodes, edges, start, goal := defaultGraph()
	steps := Trace(nodes, edges, start, goal)
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace")
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	var hasWorklistItems bool
	var hasLowestFCaption bool
	var hasTentativeGCaption bool

	for _, step := range steps {
		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("expected step state of type GraphState, got %T", step.State)
		}

		if state.WorklistLabel != "Open Set (Min-Heap: node, f = g + h)" {
			t.Fatalf("unexpected WorklistLabel %q", state.WorklistLabel)
		}
		if state.Caption != step.Explanation {
			t.Fatalf("state.Caption %q != step.Explanation %q", state.Caption, step.Explanation)
		}

		if strings.Contains(state.Caption, "Pop node") && strings.Contains(state.Caption, "lowest total estimated cost f(n)=g+h") {
			hasLowestFCaption = true
		}
		if strings.Contains(state.Caption, "Update neighbor") && strings.Contains(state.Caption, "tentative g is lower") {
			hasTentativeGCaption = true
		}

		if len(state.Worklist) > 0 {
			hasWorklistItems = true
			var prevF int
			for i, item := range state.Worklist {
				if item.ID == "" {
					t.Fatalf("worklist item has empty ID at step %s", step.LineID)
				}
				// Format: "NODE: f=NUM (g=NUM, h=NUM)"
				var id string
				var f, g, h int
				n, err := fmt.Sscanf(item.Label, "%s f=%d (g=%d, h=%d)", &id, &f, &g, &h)
				if err != nil || n != 4 {
					// Also handle without trailing colon in node token
					id = strings.TrimSuffix(id, ":")
				}
				if f != g+h {
					t.Fatalf("worklist item fScore mismatch: f=%d != g=%d + h=%d in %q", f, g, h, item.Label)
				}
				if i > 0 && f < prevF {
					t.Fatalf("worklist items not sorted by fScore: %d < %d", f, prevF)
				}
				prevF = f
			}
		}
	}

	if !hasWorklistItems {
		t.Fatal("expected trace steps to contain worklist items")
	}
	if !hasLowestFCaption {
		t.Fatal("expected caption explaining Pop node with lowest f(n)=g+h")
	}
	if !hasTentativeGCaption {
		t.Fatal("expected caption explaining Update neighbor when tentative g is lower")
	}
}

func TestAStarSnapshotImmutability(t *testing.T) {
	nodes, edges, start, goal := defaultGraph()
	steps := Trace(nodes, edges, start, goal)
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	// Verify step snapshots have independent slice allocations
	for i := 0; i < len(steps)-1; i++ {
		s1 := steps[i].State.(renderers.GraphState)
		s2 := steps[i+1].State.(renderers.GraphState)
		if len(s1.Nodes) > 0 && len(s2.Nodes) > 0 && &s1.Nodes[0] == &s2.Nodes[0] {
			t.Fatalf("step %d and %d share Nodes backing array", i, i+1)
		}
		if len(s1.Edges) > 0 && len(s2.Edges) > 0 && &s1.Edges[0] == &s2.Edges[0] {
			t.Fatalf("step %d and %d share Edges backing array", i, i+1)
		}
	}

	// Check start node state transitions across distinct steps
	step0State := steps[0].State.(renderers.GraphState)
	if step0State.Nodes[0].State != "unvisited" {
		t.Fatalf("expected start node to be unvisited at step 0, got %q", step0State.Nodes[0].State)
	}
	var foundQueued, foundVisiting, foundPath bool
	for _, s := range steps {
		g := s.State.(renderers.GraphState)
		for _, n := range g.Nodes {
			if n.ID == start {
				if n.State == "queued" {
					foundQueued = true
				}
				if n.State == "visiting" {
					foundVisiting = true
				}
				if n.State == "path" {
					foundPath = true
				}
			}
		}
	}
	if !foundQueued {
		t.Fatal("expected start node to transition to 'queued'")
	}
	if !foundVisiting {
		t.Fatal("expected start node to transition to 'visiting'")
	}
	if !foundPath {
		t.Fatal("expected start node to transition to 'path'")
	}

	// Verify mutating step 0 does not alter subsequent steps
	s0 := steps[0].State.(renderers.GraphState)
	s1Before := steps[1].State.(renderers.GraphState).Nodes[0].State

	s0.Nodes[0].State = "MUTATED"
	s0.Edges[0].Active = true
	s0.Worklist = append(s0.Worklist, renderers.WorklistItem{ID: "HACK", Label: "HACK"})

	s1After := steps[1].State.(renderers.GraphState).Nodes[0].State
	if s1After != s1Before {
		t.Fatalf("step 1 state mutated by modifying step 0: before=%q, after=%q", s1Before, s1After)
	}
}

func TestAStarPathReconstructionInTrace(t *testing.T) {
	nodes, edges, start, goal := defaultGraph()
	steps := Trace(nodes, edges, start, goal)
	lastStep := steps[len(steps)-1]

	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last step ActionDone, got %q", lastStep.Action)
	}

	rawPath, ok := lastStep.Variables["path"].([]string)
	if !ok {
		t.Fatalf("expected []string path in variables, got %T", lastStep.Variables["path"])
	}
	expectedPath := []string{"A", "B", "D", "F"}
	if !reflect.DeepEqual(rawPath, expectedPath) {
		t.Fatalf("expected path %v, got %v", expectedPath, rawPath)
	}

	state := lastStep.State.(renderers.GraphState)
	pathNodeMap := make(map[string]bool)
	for _, p := range rawPath {
		pathNodeMap[p] = true
	}

	// Verify all nodes on path have state "path"
	for _, n := range state.Nodes {
		if pathNodeMap[n.ID] {
			if n.State != "path" {
				t.Fatalf("node %s on path has state %q, expected 'path'", n.ID, n.State)
			}
		}
	}

	// Verify path edges are active and non-path edges are inactive
	pathEdgeMap := make(map[string]bool)
	for i := 0; i < len(rawPath)-1; i++ {
		u, v := rawPath[i], rawPath[i+1]
		pathEdgeMap[u+"->"+v] = true
		pathEdgeMap[v+"->"+u] = true
	}

	for _, e := range state.Edges {
		isPathEdge := pathEdgeMap[e.From+"->"+e.To]
		if isPathEdge && !e.Active {
			t.Fatalf("path edge (%s -> %s) is not active", e.From, e.To)
		}
		if !isPathEdge && e.Active {
			t.Fatalf("non-path edge (%s -> %s) is active", e.From, e.To)
		}
	}
}

func TestAStarStartEqualsGoalComplete(t *testing.T) {
	nodes := []AStarNode{{ID: "A", Label: "A", X: 100, Y: 100}}
	steps := Trace(nodes, nil, "A", "A")

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
	path := lastStep.Variables["path"].([]string)
	if !reflect.DeepEqual(path, []string{"A"}) {
		t.Fatalf("expected path [A], got %v", path)
	}

	state := lastStep.State.(renderers.GraphState)
	if state.Nodes[0].State != "path" {
		t.Fatalf("expected start node to be 'path', got %q", state.Nodes[0].State)
	}
}

func TestAStarUnreachableGoalDirected(t *testing.T) {
	nodes := []AStarNode{
		{ID: "A", Label: "A", X: 100, Y: 100},
		{ID: "B", Label: "B", X: 300, Y: 100},
	}
	edges := []AStarEdge{
		{From: "A", To: "B", Weight: 10, Undirected: false},
	}

	// Search from B to A (impossible since edge is directed A -> B)
	steps := Trace(nodes, edges, "B", "A")
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
	if lastStep.LineID != "L19" {
		t.Fatalf("expected LineID L19, got %q", lastStep.LineID)
	}
	if lastStep.Variables["found"] != false {
		t.Fatalf("expected found=false, got %v", lastStep.Variables["found"])
	}
	if lastStep.Variables["cost"] != -1 {
		t.Fatalf("expected cost -1, got %v", lastStep.Variables["cost"])
	}
}

func TestAStarBuildStrictValidation(t *testing.T) {
	lesson := Lesson{}

	// Nil and empty input should succeed with default graph
	if _, err := lesson.Build(nil); err != nil {
		t.Fatalf("Build(nil) failed: %v", err)
	}
	if _, err := lesson.Build([]byte("")); err != nil {
		t.Fatalf("Build('') failed: %v", err)
	}
	if _, err := lesson.Build([]byte("   \n\t")); err != nil {
		t.Fatalf("Build(whitespace) failed: %v", err)
	}

	// Rejections
	cases := []struct {
		name  string
		input string
	}{
		{"null input", `null`},
		{"malformed JSON", `{not valid json`},
		{"empty JSON object", `{}`},
		{"unknown start node", `{"start": "UNKNOWN", "goal": "F"}`},
		{"unknown goal node", `{"start": "A", "goal": "UNKNOWN"}`},
		{"disconnected node", `{"nodes": [{"id":"A","x":0,"y":0},{"id":"B","x":1,"y":1},{"id":"C","x":2,"y":2}],"edges":[{"from":"A","to":"B","weight":10,"undirected":true}],"start":"A","goal":"B"}`},
		{"isolated nodes with empty edges", `{"nodes": [{"id":"A","x":0,"y":0},{"id":"B","x":1,"y":1}],"edges":[],"start":"A","goal":"B"}`},
		{"unknown node in edge from", `{"nodes": [{"id":"A","x":0,"y":0},{"id":"B","x":1,"y":1}],"edges":[{"from":"Z","to":"B","weight":10}],"start":"A","goal":"B"}`},
		{"unknown node in edge to", `{"nodes": [{"id":"A","x":0,"y":0},{"id":"B","x":1,"y":1}],"edges":[{"from":"A","to":"Z","weight":10}],"start":"A","goal":"B"}`},
		{"negative edge weight", `{"nodes": [{"id":"A","x":0,"y":0},{"id":"B","x":1,"y":1}],"edges":[{"from":"A","to":"B","weight":-5}],"start":"A","goal":"B"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build([]byte(tc.input))
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}
