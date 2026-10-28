package kmeans

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	Points []Point `json:"points"`
	K      int     `json:"k"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "kmeans",
		Name:     "K-Means Clustering",
		Category: catalog.CategoryAI,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "An unsupervised machine learning algorithm that partitions n observations into k clusters by iteratively assigning points to nearest centroids and recalculating centroid positions as cluster means.",
			TimeComplexity:  "O(k · n · I)",
			SpaceComplexity: "O(n + k)",
			Difficulty:      "Medium",
			KeyConcepts:     []string{"Unsupervised Learning", "Centroid Optimization", "Euclidean Distance", "Voronoi Partitioning"},
			HowItWorks:      "Begins by selecting k initial centroids from the dataset. In the assignment step, each 2D point is assigned to its nearest centroid according to Euclidean distance. In the update step, each centroid is recalculated as the arithmetic mean of all points assigned to its cluster. This iterative process repeats until point assignments and centroid coordinates stabilize or maximum iterations are reached.",
		},
		DefaultInput: `{"k":3,"points":[{"x":150,"y":120},{"x":220,"y":100},{"x":180,"y":180},{"x":250,"y":160},{"x":550,"y":110},{"x":620,"y":130},{"x":580,"y":190},{"x":660,"y":170},{"x":340,"y":320},{"x":420,"y":310},{"x":370,"y":380},{"x":450,"y":370}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Cluster(points []Point, k, maxIters int) ([]Point, []int) {"},
		{ID: "L2", Number: 2, Code: "    centroids := initCentroids(points, k)"},
		{ID: "L3", Number: 3, Code: "    assignments := unassigned(len(points)) // all -1"},
		{ID: "L4", Number: 4, Code: "    for iter := 0; iter < maxIters; iter++ {"},
		{ID: "L5", Number: 5, Code: "        changed := false"},
		{ID: "L6", Number: 6, Code: "        for i, p := range points {"},
		{ID: "L7", Number: 7, Code: "            bestCluster := nearestCentroid(p, centroids)"},
		{ID: "L8", Number: 8, Code: "            if assignments[i] != bestCluster {"},
		{ID: "L9", Number: 9, Code: "                assignments[i] = bestCluster"},
		{ID: "L10", Number: 10, Code: "                changed = true"},
		{ID: "L11", Number: 11, Code: "            }"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "        centroids, moved := updateCentroids(points, assignments, centroids)"},
		{ID: "L14", Number: 14, Code: "        if !changed && !moved { break }"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    return centroids, assignments"},
		{ID: "L17", Number: 17, Code: "}"},
	}
}

func defaultPoints() []Point {
	return []Point{
		// Cluster 0
		{X: 150, Y: 120},
		{X: 220, Y: 100},
		{X: 180, Y: 180},
		{X: 250, Y: 160},
		// Cluster 1
		{X: 550, Y: 110},
		{X: 620, Y: 130},
		{X: 580, Y: 190},
		{X: 660, Y: 170},
		// Cluster 2
		{X: 340, Y: 320},
		{X: 420, Y: 310},
		{X: 370, Y: 380},
		{X: 450, Y: 370},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	pts := defaultPoints()
	k := 3

	if len(input) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		if fields == nil {
			return nil, fmt.Errorf("input must be an object")
		}
		if raw, ok := fields["points"]; ok {
			var entries []map[string]json.RawMessage
			if err := json.Unmarshal(raw, &entries); err != nil {
				return nil, err
			}
			for _, entry := range entries {
				if entry == nil {
					return nil, fmt.Errorf("point must be an object")
				}
				for _, key := range []string{"x", "y"} {
					if value, exists := entry[key]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
						return nil, fmt.Errorf("point coordinate must be numeric")
					}
				}
			}
			if err := json.Unmarshal(raw, &pts); err != nil {
				return nil, err
			}
		}
		if raw, ok := fields["k"]; ok {
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("k must be an integer")
			}
			if err := json.Unmarshal(raw, &k); err != nil {
				return nil, err
			}
		}
	}
	if err := validateKMeansInput(pts, k); err != nil {
		return nil, err
	}

	return Trace(pts, k, 20), nil
}

func init() {
	catalog.Register(Lesson{})
}
