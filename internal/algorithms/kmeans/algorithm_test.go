package kmeans

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestClusterConvergence(t *testing.T) {
	// 3 clearly separated clusters of 4 points each
	points := []Point{
		// Cluster 0 near (100, 100)
		{X: 90, Y: 95},
		{X: 105, Y: 100},
		{X: 95, Y: 105},
		{X: 110, Y: 90},
		// Cluster 1 near (500, 100)
		{X: 490, Y: 95},
		{X: 510, Y: 100},
		{X: 500, Y: 105},
		{X: 505, Y: 90},
		// Cluster 2 near (300, 400)
		{X: 290, Y: 395},
		{X: 310, Y: 400},
		{X: 300, Y: 405},
		{X: 305, Y: 390},
	}

	k := 3
	maxIters := 20
	centroids, assignments := Cluster(points, k, maxIters)

	if len(centroids) != k {
		t.Fatalf("expected %d centroids, got %d", k, len(centroids))
	}
	if len(assignments) != len(points) {
		t.Fatalf("expected %d assignments, got %d", len(points), len(assignments))
	}

	// Verify points in same ground-truth cluster are assigned to same cluster
	for i := 0; i < 4; i++ {
		if assignments[i] != assignments[0] {
			t.Errorf("point %d not in cluster 0 group: got %d, want %d", i, assignments[i], assignments[0])
		}
	}
	for i := 4; i < 8; i++ {
		if assignments[i] != assignments[4] {
			t.Errorf("point %d not in cluster 1 group: got %d, want %d", i, assignments[i], assignments[4])
		}
	}
	for i := 8; i < 12; i++ {
		if assignments[i] != assignments[8] {
			t.Errorf("point %d not in cluster 2 group: got %d, want %d", i, assignments[i], assignments[8])
		}
	}

	// Verify all 3 clusters are distinct
	if assignments[0] == assignments[4] || assignments[0] == assignments[8] || assignments[4] == assignments[8] {
		t.Errorf("clusters not separated: assignments were %d, %d, %d", assignments[0], assignments[4], assignments[8])
	}

	// Verify centroids are close to true cluster means
	expectedMeans := []Point{
		{X: 100, Y: 97.5},
		{X: 501.25, Y: 97.5},
		{X: 301.25, Y: 396.25},
	}
	for _, mean := range expectedMeans {
		found := false
		for _, c := range centroids {
			dx := c.X - mean.X
			dy := c.Y - mean.Y
			if math.Sqrt(dx*dx+dy*dy) < 5.0 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected centroid near (%f, %f), but not found in %v", mean.X, mean.Y, centroids)
		}
	}
}

func TestClusterEdgeCases(t *testing.T) {
	// Empty points
	c, a := Cluster(nil, 3, 10)
	if c != nil || a != nil {
		t.Errorf("expected nil for empty points, got %v, %v", c, a)
	}

	// k <= 0
	c, a = Cluster([]Point{{X: 1, Y: 1}}, 0, 10)
	if c != nil || a != nil {
		t.Errorf("expected nil for k <= 0, got %v, %v", c, a)
	}

	// k > len(points)
	pts := []Point{{X: 1, Y: 1}, {X: 2, Y: 2}}
	c, a = Cluster(pts, 5, 10)
	if len(c) != 2 || len(a) != 2 {
		t.Fatalf("expected k clamped to %d, got %d centroids", len(pts), len(c))
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "kmeans" {
		t.Errorf("expected ID 'kmeans', got %q", meta.ID)
	}
	if meta.Name != "K-Means Clustering" {
		t.Errorf("expected Name 'K-Means Clustering', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("expected Category %q, got %q", catalog.CategoryAI, meta.Category)
	}
	if meta.Renderer != "graph" {
		t.Errorf("expected Renderer 'graph', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(k · n · I)" {
		t.Errorf("expected TimeComplexity 'O(k · n · I)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n + k)" {
		t.Errorf("expected SpaceComplexity 'O(n + k)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{"Unsupervised Learning", "Centroid Optimization", "Euclidean Distance", "Voronoi Partitioning"}
	if len(meta.Documentation.KeyConcepts) != len(expectedConcepts) {
		t.Errorf("expected %d key concepts, got %d", len(expectedConcepts), len(meta.Documentation.KeyConcepts))
	}

	regLesson, ok := catalog.Get("kmeans")
	if !ok {
		t.Fatal("expected kmeans registered in catalog")
	}
	if regLesson.Meta().ID != "kmeans" {
		t.Errorf("catalog registered wrong ID: %q", regLesson.Meta().ID)
	}
}

func TestLessonValidateSourceLineIDs(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps to be non-empty")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}
}

func TestTraceGraphState(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	firstStep := steps[0]
	state, ok := firstStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected state to be renderers.GraphState, got %T", firstStep.State)
	}
	if len(state.Nodes) == 0 {
		t.Error("expected graph state nodes to be non-empty")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last step action ActionDone, got %s", lastStep.Action)
	}
	lastState, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	// In final step, each of the 12 points should be assigned to one of the 3 cluster colors
	foundColors := make(map[string]int)
	for _, n := range lastState.Nodes {
		foundColors[n.State]++
	}
	// At least visiting, visited, path should be present for the 3 clusters
	for _, requiredState := range []string{"visiting", "visited", "path"} {
		if foundColors[requiredState] == 0 {
			t.Errorf("expected at least one node with state %q, found counts: %v", requiredState, foundColors)
		}
	}
}

func TestLessonBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := json.RawMessage(`{"k":2,"points":[{"x":10,"y":10},{"x":15,"y":12},{"x":80,"y":85},{"x":85,"y":90}]}`)
	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps for custom input")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed on custom input: %v", err)
	}
}

func TestLessonBuildValidation(t *testing.T) {
	lesson := Lesson{}

	// Nil and empty input should succeed with defaults
	if steps, err := lesson.Build(nil); err != nil || len(steps) == 0 {
		t.Fatalf("expected nil input to succeed with defaults, got err: %v", err)
	}
	if steps, err := lesson.Build(json.RawMessage("")); err != nil || len(steps) == 0 {
		t.Fatalf("expected empty input to succeed with defaults, got err: %v", err)
	}
	if steps, err := lesson.Build(json.RawMessage("   ")); err != nil || len(steps) == 0 {
		t.Fatalf("expected whitespace input to succeed with defaults, got err: %v", err)
	}

	// Null input
	if _, err := lesson.Build(json.RawMessage("null")); err == nil {
		t.Fatal("expected error for null input, got nil")
	}

	// Malformed JSON
	if _, err := lesson.Build(json.RawMessage("{not-json")); err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// Empty points
	if _, err := lesson.Build(json.RawMessage(`{"k": 2, "points": []}`)); err == nil {
		t.Fatal("expected error for empty points, got nil")
	}
	if _, err := lesson.Build(json.RawMessage(`{"k": 2}`)); err == nil {
		t.Fatal("expected error for missing points, got nil")
	}
	if _, err := lesson.Build(json.RawMessage(`{"k": 2, "points": null}`)); err == nil {
		t.Fatal("expected error for null points, got nil")
	}

	// k <= 0
	if _, err := lesson.Build(json.RawMessage(`{"k": 0, "points": [{"x": 10, "y": 10}]}`)); err == nil {
		t.Fatal("expected error for k <= 0, got nil")
	}
	if _, err := lesson.Build(json.RawMessage(`{"k": -1, "points": [{"x": 10, "y": 10}]}`)); err == nil {
		t.Fatal("expected error for negative k, got nil")
	}

	// k > len(points)
	if _, err := lesson.Build(json.RawMessage(`{"k": 3, "points": [{"x": 10, "y": 10}]}`)); err == nil {
		t.Fatal("expected error for k > len(points), got nil")
	}

	// Valid custom input
	steps, err := lesson.Build(json.RawMessage(`{"k": 1, "points": [{"x": 10, "y": 10}]}`))
	if err != nil {
		t.Fatalf("expected valid single-point input to succeed, got err: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}
}

func TestAssignmentVisualizationAndCaptions(t *testing.T) {
	pts := []Point{
		{X: 10, Y: 10},
		{X: 12, Y: 14},
		{X: 200, Y: 200},
		{X: 205, Y: 210},
	}
	k := 2
	steps := Trace(pts, k, 10)
	if len(steps) == 0 {
		t.Fatal("expected steps to be generated")
	}

	// Verify initial step L1 has all points unassigned ("unvisited") and no centroids
	firstState := steps[0].State.(renderers.GraphState)
	if len(firstState.Nodes) != len(pts) {
		t.Fatalf("expected initial step to have %d nodes, got %d", len(pts), len(firstState.Nodes))
	}
	for _, n := range firstState.Nodes {
		if n.State != "unvisited" {
			t.Errorf("expected initial point node to be unvisited, got %s", n.State)
		}
	}

	// Verify L2 step introduces initial k centroids with IDs C0, C1
	var l2State renderers.GraphState
	var foundL2 bool
	for _, s := range steps {
		if s.LineID == "L2" {
			l2State = s.State.(renderers.GraphState)
			foundL2 = true
			break
		}
	}
	if !foundL2 {
		t.Fatal("missing L2 step for initial centroid selection")
	}
	centroidCount := 0
	for _, n := range l2State.Nodes {
		if n.ID == "C0" || n.ID == "C1" {
			centroidCount++
		}
	}
	if centroidCount != k {
		t.Fatalf("expected %d centroid nodes in L2 state, found %d", k, centroidCount)
	}

	// Verify distance comparisons and assignment captions
	var foundDistCompare bool
	var foundAssignmentCaption bool
	expectedCaptionPattern := "closest to Centroid"

	for _, s := range steps {
		gs, ok := s.State.(renderers.GraphState)
		if !ok {
			continue
		}

		if s.LineID == "L7" && s.Action == trace.ActionCompare {
			if strings.Contains(s.Explanation, "distance to Centroid") {
				foundDistCompare = true
				// Check active comparison edge exists
				var hasActiveEdge bool
				for _, e := range gs.Edges {
					if e.Active && e.Weight >= 0 {
						hasActiveEdge = true
						break
					}
				}
				if !hasActiveEdge {
					t.Errorf("expected active comparison edge in step: %s", s.Explanation)
				}
			}
			if strings.Contains(s.Explanation, expectedCaptionPattern) && strings.Contains(s.Explanation, "assign to Cluster") {
				foundAssignmentCaption = true
				if gs.Caption != s.Explanation {
					t.Errorf("expected GraphState caption to match explanation, got %q vs %q", gs.Caption, s.Explanation)
				}
			}
		}

		if s.LineID == "L9" && s.Action == trace.ActionSet {
			// Assigned point node should match cluster color
			vars := s.Variables
			ptIdx, okPt := vars["point"].(int)
			clusterIdx, okCl := vars["cluster"].(int)
			if okPt && okCl {
				for _, n := range gs.Nodes {
					if n.ID == fmt.Sprintf("P%d", ptIdx) {
						expectedState := clusterColor(clusterIdx)
						if n.State != expectedState {
							t.Errorf("expected point P%d to have state %s after L9, got %s", ptIdx, expectedState, n.State)
						}
					}
				}
			}
		}
	}

	if !foundDistCompare {
		t.Error("expected to find centroid distance comparison steps")
	}
	if !foundAssignmentCaption {
		t.Error("expected to find assignment caption matching required pattern")
	}
}

func TestCentroidUpdatesAndConvergence(t *testing.T) {
	pts := defaultPoints()
	steps := Trace(pts, 3, 20)

	var foundCentroidUpdate bool
	var foundConvergenceExplanation bool
	var foundFinalStep bool

	for _, s := range steps {
		gs := s.State.(renderers.GraphState)
		if s.LineID == "L13" && s.Action == trace.ActionSet {
			foundCentroidUpdate = true
			if !strings.Contains(s.Explanation, "Centroid C") {
				t.Errorf("expected L13 explanation to mention Centroid, got %q", s.Explanation)
			}
			if strings.Contains(s.Explanation, "converged") && strings.Contains(s.Explanation, "threshold") {
				foundConvergenceExplanation = true
			}
		}
		if s.LineID == "L16" && s.Action == trace.ActionDone {
			foundFinalStep = true
			if !strings.Contains(s.Explanation, "centroids stabilized") && !strings.Contains(gs.Caption, "centroids stabilized") {
				t.Errorf("expected L16 to explain stabilized centroids, got explanation %q, caption %q", s.Explanation, gs.Caption)
			}
		}
	}

	if !foundCentroidUpdate {
		t.Error("expected L13 centroid update step")
	}
	if !foundConvergenceExplanation {
		t.Error("expected L13 to explain convergence when movement < threshold")
	}
	if !foundFinalStep {
		t.Error("expected L16 final step")
	}
}

func TestSnapshotImmutability(t *testing.T) {
	pts := []Point{
		{X: 10, Y: 10},
		{X: 15, Y: 15},
		{X: 80, Y: 80},
		{X: 85, Y: 85},
	}
	steps := Trace(pts, 2, 10)
	if len(steps) < 5 {
		t.Fatalf("expected at least 5 steps, got %d", len(steps))
	}

	firstStepState := steps[0].State.(renderers.GraphState)
	origFirstNode0 := firstStepState.Nodes[0]

	midIdx := len(steps) / 2
	midStepState := steps[midIdx].State.(renderers.GraphState)
	origMidNode0 := midStepState.Nodes[0]

	lastIdx := len(steps) - 1
	lastStepState := steps[lastIdx].State.(renderers.GraphState)

	// Mutate lastStepState's nodes and edges
	lastStepState.Nodes[0].X = 999999.0
	lastStepState.Nodes[0].State = "corrupted_state"
	if len(lastStepState.Edges) > 0 {
		lastStepState.Edges[0].Weight = 888888
		lastStepState.Edges[0].Active = !lastStepState.Edges[0].Active
	}

	// Verify firstStepState and midStepState were not affected
	if firstStepState.Nodes[0].X == 999999.0 || firstStepState.Nodes[0].State == "corrupted_state" {
		t.Fatal("mutating last step mutated first step nodes")
	}
	if firstStepState.Nodes[0] != origFirstNode0 {
		t.Fatalf("first step node altered: got %+v, want %+v", firstStepState.Nodes[0], origFirstNode0)
	}
	if midStepState.Nodes[0].X == 999999.0 || midStepState.Nodes[0].State == "corrupted_state" {
		t.Fatal("mutating last step mutated mid step nodes")
	}
	if midStepState.Nodes[0] != origMidNode0 {
		t.Fatalf("mid step node altered: got %+v, want %+v", midStepState.Nodes[0], origMidNode0)
	}

	// Verify modifying input points array does not affect generated steps
	pts[0].X = -12345.0
	if steps[0].State.(renderers.GraphState).Nodes[0].X == -12345.0 {
		t.Fatal("mutating input points array after Trace mutated step snapshot")
	}
}
