package kruskal

import (
	"encoding/json"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestKruskalMSTCorrectness(t *testing.T) {
	nodes := []string{"A", "B", "C", "D"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 1},
		{From: "B", To: "C", Weight: 2},
		{From: "A", To: "C", Weight: 4},
		{From: "C", To: "D", Weight: 3},
		{From: "B", To: "D", Weight: 5},
	}

	mstEdges, totalWeight := MST(nodes, edges)

	if len(mstEdges) != len(nodes)-1 {
		t.Fatalf("expected %d edges in MST, got %d", len(nodes)-1, len(mstEdges))
	}
	if totalWeight != 6 {
		t.Fatalf("expected total weight 6, got %d", totalWeight)
	}

	dsu := NewDSU(nodes)
	for _, e := range mstEdges {
		if !dsu.Union(e.From, e.To) {
			t.Fatalf("cycle detected with edge %+v", e)
		}
	}
}

func TestKruskalDefaultGraph(t *testing.T) {
	nodes, edges := defaultGraph()
	nodeIDs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}

	mstEdges, totalWeight := MST(nodeIDs, edges)
	if len(mstEdges) != len(nodeIDs)-1 {
		t.Fatalf("expected %d edges in MST, got %d", len(nodeIDs)-1, len(mstEdges))
	}
	if totalWeight != 12 {
		t.Fatalf("expected total weight 12, got %d", totalWeight)
	}
}

func TestKruskalSingleNode(t *testing.T) {
	mst, weight := MST([]string{"A"}, nil)
	if len(mst) != 0 || weight != 0 {
		t.Fatalf("expected 0 edges and 0 weight, got %d edges and %d weight", len(mst), weight)
	}
}

func TestKruskalLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "kruskal" {
		t.Fatalf("expected ID kruskal, got %s", meta.ID)
	}
	if meta.Name != "Kruskal's Minimum Spanning Tree" {
		t.Fatalf("unexpected name: %s", meta.Name)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Fatalf("unexpected category: %s", meta.Category)
	}
	if meta.Renderer != "graph" {
		t.Fatalf("unexpected renderer: %s", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(E log E)" {
		t.Fatalf("unexpected time complexity: %s", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(V)" {
		t.Fatalf("unexpected space complexity: %s", meta.Documentation.SpaceComplexity)
	}
	expectedConcepts := []string{"Greedy", "Edge Sorting", "Disjoint Set Union", "Cycle Detection"}
	if len(meta.Documentation.KeyConcepts) != len(expectedConcepts) {
		t.Fatalf("expected %d key concepts, got %d", len(expectedConcepts), len(meta.Documentation.KeyConcepts))
	}
}

func TestKruskalLessonValidation(t *testing.T) {
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

	// Verify cycle rejection step exists
	hasCycleStep := false
	hasIncludeStep := false
	for _, s := range steps {
		if s.Variables != nil && s.Variables["rejected"] == true {
			hasCycleStep = true
		}
		if s.Action == trace.ActionPush {
			hasIncludeStep = true
		}
	}
	if !hasCycleStep {
		t.Fatal("expected at least one cycle rejection step in trace")
	}
	if !hasIncludeStep {
		t.Fatal("expected at least one tree inclusion step in trace")
	}
}

func TestKruskalBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"nodes": [{"id":"1"},{"id":"2"},{"id":"3"}],
		"edges": [{"from":"1","to":"2","weight":10},{"from":"2","to":"3","weight":20},{"from":"1","to":"3","weight":30}]
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
}
