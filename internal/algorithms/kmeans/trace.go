package kmeans

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func clusterColor(c int) string {
	switch c {
	case 0:
		return "visiting" // cyan/blue
	case 1:
		return "visited"  // green
	case 2:
		return "path"     // purple
	default:
		return "queued"   // yellow
	}
}

func clonePoints(pts []Point) []Point {
	c := make([]Point, len(pts))
	copy(c, pts)
	return c
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}

func makeGraphState(points []Point, centroids []Point, assignments []int, activePoint int) renderers.GraphState {
	nodes := make([]renderers.GraphNode, 0, len(points)+len(centroids))
	edges := make([]renderers.GraphEdge, 0, len(points))

	// Add point nodes
	for i, p := range points {
		st := "unvisited"
		if assignments[i] >= 0 {
			st = clusterColor(assignments[i])
		}
		if i == activePoint {
			st = "queued"
		}

		nodes = append(nodes, renderers.GraphNode{
			ID:    fmt.Sprintf("P%d", i),
			Label: fmt.Sprintf("P%d", i),
			X:     p.X,
			Y:     p.Y,
			State: st,
		})

		// Add edge to assigned centroid if centroid exists
		if assignments[i] >= 0 && assignments[i] < len(centroids) {
			edges = append(edges, renderers.GraphEdge{
				From:   fmt.Sprintf("P%d", i),
				To:     fmt.Sprintf("C%d", assignments[i]),
				Active: true,
			})
		}
	}

	// Add centroid nodes
	for j, c := range centroids {
		nodes = append(nodes, renderers.GraphNode{
			ID:    fmt.Sprintf("C%d", j),
			Label: fmt.Sprintf("C%d", j),
			X:     c.X,
			Y:     c.Y,
			State: clusterColor(j),
		})
	}

	return renderers.GraphState{
		Nodes: nodes,
		Edges: edges,
	}
}

// Trace executes K-Means clustering and records every step for visualization.
func Trace(points []Point, k, maxIters int) []trace.Step {
	b := trace.NewBuilder()

	if len(points) == 0 || k <= 0 {
		b.AddStep("L1", trace.ActionInit, "Empty dataset or invalid k; clustering cannot proceed.", renderers.GraphState{}, nil)
		b.AddStep("L16", trace.ActionDone, "Terminated with empty result.", renderers.GraphState{}, nil)
		return b.Steps()
	}

	if k > len(points) {
		k = len(points)
	}
	if maxIters <= 0 {
		maxIters = 100
	}

	assignments := make([]int, len(points))
	for i := range assignments {
		assignments[i] = -1
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize K-Means clustering with %d points and k=%d.", len(points), k),
		makeGraphState(points, nil, assignments, -1),
		map[string]any{"k": k, "numPoints": len(points)})

	centroids := make([]Point, k)
	step := len(points) / k
	for i := 0; i < k; i++ {
		centroids[i] = points[i*step]
	}
	b.IncOperations()

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Select %d initial centroids from the dataset.", k),
		makeGraphState(points, centroids, assignments, -1),
		map[string]any{"centroids": clonePoints(centroids)})

	b.AddStep("L3", trace.ActionInit, "Initialize cluster assignment array.",
		makeGraphState(points, centroids, assignments, -1),
		map[string]any{"assignments": cloneSlice(assignments)})

	var finalIter int
	for iter := 0; iter < maxIters; iter++ {
		finalIter = iter + 1
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Iteration %d: begin assignment phase.", iter+1),
			makeGraphState(points, centroids, assignments, -1),
			map[string]any{"iteration": iter + 1})

		changed := false
		b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Reset changed flag to false for iteration %d.", iter+1),
			makeGraphState(points, centroids, assignments, -1),
			map[string]any{"iteration": iter + 1, "changed": false})

		for i, p := range points {
			b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Evaluating nearest centroid for point P%d at (%.1f, %.1f).", i, p.X, p.Y),
				makeGraphState(points, centroids, assignments, i),
				map[string]any{"point": i})

			for range centroids {
				b.IncComparisons()
			}
			bestCluster := nearestCentroid(p, centroids)

			b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("Point P%d is closest to centroid C%d.", i, bestCluster),
				makeGraphState(points, centroids, assignments, i),
				map[string]any{"point": i, "closestCentroid": bestCluster})

			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Check if cluster assignment changed for point P%d.", i),
				makeGraphState(points, centroids, assignments, i),
				map[string]any{"current": assignments[i], "new": bestCluster})

			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
				b.IncSwaps()

				b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Assign point P%d to cluster %d (Centroid C%d).", i, bestCluster, bestCluster),
					makeGraphState(points, centroids, assignments, -1),
					map[string]any{"point": i, "cluster": bestCluster})

				b.AddStep("L10", trace.ActionSet, "Mark changed flag as true.",
					makeGraphState(points, centroids, assignments, -1),
					map[string]any{"changed": true})
			}
		}

		counts := make([]int, k)
		sums := make([]Point, k)
		for i, p := range points {
			c := assignments[i]
			counts[c]++
			sums[c].X += p.X
			sums[c].Y += p.Y
		}

		moved := false
		for j := 0; j < k; j++ {
			if counts[j] > 0 {
				newCentroid := Point{
					X: sums[j].X / float64(counts[j]),
					Y: sums[j].Y / float64(counts[j]),
				}
				if distSq(centroids[j], newCentroid) > 1e-9 {
					moved = true
				}
				centroids[j] = newCentroid
			}
		}
		b.IncOperations()

		b.AddStep("L13", trace.ActionSet, fmt.Sprintf("Iteration %d: recalculate centroid positions to cluster means.", iter+1),
			makeGraphState(points, centroids, assignments, -1),
			map[string]any{"centroids": clonePoints(centroids)})

		b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Convergence check at iteration %d: changed = %t.", iter+1, changed),
			makeGraphState(points, centroids, assignments, -1),
			map[string]any{"changed": changed, "iteration": iter + 1})

		if !changed && !moved {
			break
		}
	}

	b.AddStep("L16", trace.ActionDone, fmt.Sprintf("K-Means clustering complete after %d iteration(s).", finalIter),
		makeGraphState(points, centroids, assignments, -1),
		map[string]any{"converged": true, "iterations": finalIter})

	return b.Steps()
}
