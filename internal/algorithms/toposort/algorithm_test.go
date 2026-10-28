package toposort

import (
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
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

func TestLessonBuildAndValidateSourceLineIDs(t *testing.T) {
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

	// Verify steps contain GraphState and expected variables
	sawInDegrees := false
	sawProcessed := false
	for _, s := range steps {
		if _, ok := s.State.(renderers.ToposortLessonState); !ok {
			t.Fatalf("step %d state is %T, want renderers.ToposortLessonState", s.Index, s.State)
		}
		if _, ok := s.Variables["inDegrees"]; ok {
			sawInDegrees = true
		}
		if _, ok := s.Variables["order"]; ok {
			sawProcessed = true
		} else if _, ok := s.Variables["processed"]; ok {
			sawProcessed = true
		}
	}

	if !sawInDegrees {
		t.Error("expected steps to report inDegrees in variables")
	}
	if !sawProcessed {
		t.Error("expected steps to report processed/order in variables")
	}
}

func TestLessonRegistered(t *testing.T) {
	registered, ok := catalog.Get("toposort")
	if !ok || registered == nil {
		t.Fatal("toposort lesson not registered in catalog")
	}
}

func TestTraceCycleDetection(t *testing.T) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 100, Y: 100, State: "unvisited"},
		{ID: "B", Label: "B", X: 200, Y: 100, State: "unvisited"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "B", To: "A"},
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
	if last.Variables["status"] != "cycle_detected" {
		t.Errorf("last step status = %v, want 'cycle_detected'", last.Variables["status"])
	}
	if last.Variables["valid"] != false {
		t.Errorf("last step valid = %v, want false", last.Variables["valid"])
	}
}

func TestLessonCustomBuild(t *testing.T) {
	lesson := Lesson{}
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
		t.Errorf("expected valid DAG for custom course schedule, got valid = false")
	}
}

func TestTraceEmpty(t *testing.T) {
	steps := Trace(nil, nil)
	if len(steps) != 0 {
		t.Fatalf("expected 0 steps for empty input, got %d", len(steps))
	}
}
