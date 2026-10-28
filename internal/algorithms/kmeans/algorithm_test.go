package kmeans

import (
	"encoding/json"
	"math"
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
	state, ok := firstStep.State.(renderers.KMeansLessonState)
	if !ok {
		t.Fatalf("expected state to be renderers.GraphState, got %T", firstStep.State)
	}
	if len(state.Points) == 0 {
		t.Error("expected lesson points to be non-empty")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Errorf("expected last step action ActionDone, got %s", lastStep.Action)
	}
	lastState, ok := lastStep.State.(renderers.KMeansLessonState)
	if !ok {
		t.Fatalf("expected last step state to be renderers.GraphState, got %T", lastStep.State)
	}

	found := make(map[int]int)
	for _, cluster := range lastState.Assignments {
		found[cluster]++
	}
	for cluster := 0; cluster < 3; cluster++ {
		if found[cluster] == 0 {
			t.Errorf("cluster %d has no assigned points", cluster)
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
