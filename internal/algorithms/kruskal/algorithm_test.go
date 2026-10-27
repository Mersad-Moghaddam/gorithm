package kruskal

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
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

	lesson := Lesson{}
	if _, err := lesson.Build(json.RawMessage(`{"nodes":[{"id":"A"}],"edges":[]}`)); err == nil {
		t.Fatal("expected error for empty edges in custom input")
	}

	singleSteps := Trace([]renderers.GraphNode{{ID: "A", Label: "A"}}, nil)
	if len(singleSteps) != 1 {
		t.Fatalf("expected 1 step for single node, got %d", len(singleSteps))
	}
	if singleSteps[0].Action != trace.ActionDone {
		t.Fatalf("expected ActionDone, got %v", singleSteps[0].Action)
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

func TestKruskalWorklist(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	hasWorklistLabel := false
	hasAcceptedItem := false
	hasRejectedItem := false
	acceptedCount := 0
	rejectedCount := 0

	for _, s := range steps {
		gs, ok := s.State.(renderers.GraphState)
		if !ok {
			continue
		}
		if gs.WorklistLabel == "Sorted Edges by Weight" {
			hasWorklistLabel = true
		}
		for _, item := range gs.Worklist {
			if strings.HasPrefix(item.Label, "✓") {
				hasAcceptedItem = true
			}
			if strings.HasPrefix(item.Label, "✗") {
				hasRejectedItem = true
			}
		}
	}

	if !hasWorklistLabel {
		t.Fatal("expected WorklistLabel 'Sorted Edges by Weight'")
	}
	if !hasAcceptedItem {
		t.Fatal("expected at least one accepted item (✓) in worklist")
	}
	if !hasRejectedItem {
		t.Fatal("expected at least one rejected item (✗) in worklist")
	}

	// Verify final step worklist has 5 accepted edges (for 6 vertices)
	finalState := steps[len(steps)-1].State.(renderers.GraphState)
	for _, item := range finalState.Worklist {
		if strings.HasPrefix(item.Label, "✓") {
			acceptedCount++
		}
		if strings.HasPrefix(item.Label, "✗") {
			rejectedCount++
		}
	}
	if acceptedCount != 5 {
		t.Fatalf("expected 5 accepted edges in final worklist, got %d", acceptedCount)
	}
	if rejectedCount < 1 {
		t.Fatalf("expected at least 1 rejected edge in final worklist, got %d", rejectedCount)
	}
}

func TestKruskalSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	// Verify each step has distinct slice backing arrays for Nodes, Edges, Worklist
	for i := 0; i < len(steps)-1; i++ {
		s1, ok1 := steps[i].State.(renderers.GraphState)
		s2, ok2 := steps[i+1].State.(renderers.GraphState)
		if !ok1 || !ok2 {
			continue
		}
		if len(s1.Nodes) > 0 && len(s2.Nodes) > 0 {
			if &s1.Nodes[0] == &s2.Nodes[0] {
				t.Fatalf("steps %d and %d share Nodes backing array", i, i+1)
			}
		}
		if len(s1.Edges) > 0 && len(s2.Edges) > 0 {
			if &s1.Edges[0] == &s2.Edges[0] {
				t.Fatalf("steps %d and %d share Edges backing array", i, i+1)
			}
		}
		if len(s1.Worklist) > 0 && len(s2.Worklist) > 0 {
			if &s1.Worklist[0] == &s2.Worklist[0] {
				t.Fatalf("steps %d and %d share Worklist backing array", i, i+1)
			}
		}
	}

	// Verify mutating one snapshot does not affect earlier snapshot
	s0 := steps[0].State.(renderers.GraphState)
	origNodeLabel := s0.Nodes[0].Label
	origEdgeActive := s0.Edges[0].Active

	lastStep := steps[len(steps)-1].State.(renderers.GraphState)
	lastStep.Nodes[0].Label = "MUTATED"
	lastStep.Edges[0].Active = !lastStep.Edges[0].Active
	if len(lastStep.Worklist) > 0 {
		lastStep.Worklist[0].Label = "MUTATED"
	}

	s0Check := steps[0].State.(renderers.GraphState)
	if s0Check.Nodes[0].Label != origNodeLabel {
		t.Fatalf("snapshot mutation leaked to step 0: got %s, want %s", s0Check.Nodes[0].Label, origNodeLabel)
	}
	if s0Check.Edges[0].Active != origEdgeActive {
		t.Fatalf("snapshot edge mutation leaked to step 0")
	}
}

func TestKruskalCycleRejectionAndComponents(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	hasCycleCaption := false
	hasComponentMergeExplanation := false
	hasNodeComponentLabel := false

	for _, s := range steps {
		gs, ok := s.State.(renderers.GraphState)
		if !ok {
			continue
		}
		if gs.Caption == "Cycle detected: both endpoints in same component; reject edge" {
			hasCycleCaption = true
			if s.Variables == nil || s.Variables["rejected"] != true {
				t.Fatalf("expected rejected variable on cycle step")
			}
		}
		for _, n := range gs.Nodes {
			if strings.Contains(n.Label, "[") && strings.Contains(n.Label, "]") {
				hasNodeComponentLabel = true
			}
		}
		if s.Variables != nil && s.Variables["componentA"] != nil && s.Variables["componentB"] != nil {
			hasComponentMergeExplanation = true
		}
	}

	if !hasCycleCaption {
		t.Fatal("expected step with caption 'Cycle detected: both endpoints in same component; reject edge'")
	}
	if !hasComponentMergeExplanation {
		t.Fatal("expected DSU union step explaining component merge")
	}
	if !hasNodeComponentLabel {
		t.Fatal("expected node labels showing component information (e.g. A[A])")
	}
}

func TestKruskalFinalCaption(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	finalStep := steps[len(steps)-1]
	gs, ok := finalStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected GraphState in final step")
	}

	if !strings.Contains(gs.Caption, "12") || !strings.Contains(gs.Caption, "weight") {
		t.Fatalf("expected total weight in final caption, got %q", gs.Caption)
	}
	if !strings.Contains(gs.Caption, "edges") {
		t.Fatalf("expected edges in final caption, got %q", gs.Caption)
	}
}

func TestKruskalInputValidation(t *testing.T) {
	lesson := Lesson{}

	// null input
	if _, err := lesson.Build(json.RawMessage("null")); err == nil {
		t.Fatal("expected error for null input")
	}

	// malformed JSON
	if _, err := lesson.Build(json.RawMessage("{bad-json}")); err == nil {
		t.Fatal("expected error for malformed JSON")
	}

	// empty graph (no nodes)
	if _, err := lesson.Build(json.RawMessage(`{"nodes": []}`)); err == nil {
		t.Fatal("expected error for empty nodes")
	}

	// empty edges
	if _, err := lesson.Build(json.RawMessage(`{"nodes": [{"id":"A"}], "edges": []}`)); err == nil {
		t.Fatal("expected error for empty edges")
	}

	// invalid edge reference
	if _, err := lesson.Build(json.RawMessage(`{"nodes": [{"id":"A"}], "edges": [{"from":"A","to":"B","weight":1}]}`)); err == nil {
		t.Fatal("expected error for non-existent node reference in edge")
	}

	// duplicate node ID
	if _, err := lesson.Build(json.RawMessage(`{"nodes": [{"id":"A"},{"id":"A"}], "edges": [{"from":"A","to":"A","weight":1}]}`)); err == nil {
		t.Fatal("expected error for duplicate node ID")
	}

	// nil and empty input should succeed with default graph
	if steps, err := lesson.Build(nil); err != nil || len(steps) == 0 {
		t.Fatalf("expected success for nil input, got err: %v, steps: %d", err, len(steps))
	}
	if steps, err := lesson.Build(json.RawMessage("")); err != nil || len(steps) == 0 {
		t.Fatalf("expected success for empty input, got err: %v, steps: %d", err, len(steps))
	}
}

func TestKruskalBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"nodes": [{"id":"1"},{"id":"2"},{"id":"3"},{"id":"4"}],
		"edges": [
			{"from":"1","to":"2","weight":10},
			{"from":"2","to":"3","weight":20},
			{"from":"1","to":"3","weight":25},
			{"from":"3","to":"4","weight":30}
		]
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

	// Verify cycle rejection occurs on edge 1-3
	hasCycle := false
	for _, s := range steps {
		if s.Variables != nil && s.Variables["rejected"] == true && s.Variables["from"] == "1" && s.Variables["to"] == "3" {
			hasCycle = true
		}
	}
	if !hasCycle {
		t.Fatal("expected cycle rejection for edge 1-3")
	}
}
