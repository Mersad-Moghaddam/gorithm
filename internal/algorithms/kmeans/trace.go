package kmeans

import (
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func clusterColor(c int) string {
	if c < 0 {
		return "unvisited"
	}
	switch c % 4 {
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
	if pts == nil {
		return nil
	}
	c := make([]Point, len(pts))
	copy(c, pts)
	return c
}

func cloneSlice(s []int) []int {
	if s == nil {
		return nil
	}
	c := make([]int, len(s))
	copy(c, s)
	return c
}

func cloneGraphState(s renderers.GraphState) renderers.GraphState {
	var nodes []renderers.GraphNode
	if s.Nodes != nil {
		nodes = make([]renderers.GraphNode, len(s.Nodes))
		copy(nodes, s.Nodes)
	}
	var edges []renderers.GraphEdge
	if s.Edges != nil {
		edges = make([]renderers.GraphEdge, len(s.Edges))
		copy(edges, s.Edges)
	}
	var worklist []renderers.WorklistItem
	if s.Worklist != nil {
		worklist = make([]renderers.WorklistItem, len(s.Worklist))
		copy(worklist, s.Worklist)
	}
	return renderers.GraphState{
		Caption:       s.Caption,
		WorklistLabel: s.WorklistLabel,
		Worklist:      worklist,
		Nodes:         nodes,
		Edges:         edges,
	}
}

func makeGraphState(points []Point, centroids []Point, assignments []int, activePoint int, activeCentroid int, activeDist float64, caption string) renderers.GraphState {
	nodes := make([]renderers.GraphNode, 0, len(points)+len(centroids))
	edges := make([]renderers.GraphEdge, 0, len(points)+1)

	// Add point nodes
	for i, p := range points {
		st := "unvisited"
		if i < len(assignments) && assignments[i] >= 0 {
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
		if i < len(assignments) && assignments[i] >= 0 && assignments[i] < len(centroids) {
			edges = append(edges, renderers.GraphEdge{
				From:   fmt.Sprintf("P%d", i),
				To:     fmt.Sprintf("C%d", assignments[i]),
				Active: false,
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

	// Add active comparison edge if evaluating distance
	if activePoint >= 0 && activeCentroid >= 0 && activeCentroid < len(centroids) {
		found := false
		for eIdx := range edges {
			if edges[eIdx].From == fmt.Sprintf("P%d", activePoint) && edges[eIdx].To == fmt.Sprintf("C%d", activeCentroid) {
				edges[eIdx].Active = true
				edges[eIdx].Weight = int(math.Round(activeDist))
				found = true
				break
			}
		}
		if !found {
			edges = append(edges, renderers.GraphEdge{
				From:   fmt.Sprintf("P%d", activePoint),
				To:     fmt.Sprintf("C%d", activeCentroid),
				Weight: int(math.Round(activeDist)),
				Active: true,
			})
		}
	}

	return cloneGraphState(renderers.GraphState{
		Caption: caption,
		Nodes:   nodes,
		Edges:   edges,
	})
}

// Trace executes K-Means clustering and records every step for visualization.
func Trace(points []Point, k, maxIters int) []trace.Step {
	b := trace.NewBuilder()

	if len(points) == 0 || k <= 0 {
		b.AddStep("L1", trace.ActionInit, "Empty dataset or invalid k; clustering cannot proceed.", cloneGraphState(renderers.GraphState{Caption: "Empty dataset or invalid k; clustering cannot proceed."}), nil)
		b.AddStep("L16", trace.ActionDone, "Terminated with empty result.", cloneGraphState(renderers.GraphState{Caption: "Terminated with empty result."}), nil)
		return b.Steps()
	}

	pts := clonePoints(points)
	if k > len(pts) {
		k = len(pts)
	}
	if maxIters <= 0 {
		maxIters = 100
	}

	assignments := make([]int, len(pts))
	for i := range assignments {
		assignments[i] = -1
	}

	caption := fmt.Sprintf("Initialize K-Means clustering with %d points and k=%d.", len(pts), k)
	b.AddStep("L1", trace.ActionInit, caption,
		makeGraphState(pts, nil, assignments, -1, -1, 0, caption),
		map[string]any{"k": k, "numPoints": len(pts)})

	centroids := make([]Point, k)
	step := len(pts) / k
	for i := 0; i < k; i++ {
		centroids[i] = pts[i*step]
	}
	b.IncOperations()

	caption = fmt.Sprintf("Select %d initial centroids from the dataset.", k)
	b.AddStep("L2", trace.ActionSet, caption,
		makeGraphState(pts, centroids, assignments, -1, -1, 0, caption),
		map[string]any{"centroids": clonePoints(centroids)})

	caption = "Initialize cluster assignment array with unassigned points."
	b.AddStep("L3", trace.ActionInit, caption,
		makeGraphState(pts, centroids, assignments, -1, -1, 0, caption),
		map[string]any{"assignments": cloneSlice(assignments)})

	threshold := 0.001
	var finalIter int
	for iter := 0; iter < maxIters; iter++ {
		finalIter = iter + 1
		caption = fmt.Sprintf("Iteration %d: begin assignment phase.", iter+1)
		b.AddStep("L4", trace.ActionHighlight, caption,
			makeGraphState(pts, centroids, assignments, -1, -1, 0, caption),
			map[string]any{"iteration": iter + 1})

		changed := false
		caption = fmt.Sprintf("Iteration %d: reset changed flag to false.", iter+1)
		b.AddStep("L5", trace.ActionSet, caption,
			makeGraphState(pts, centroids, assignments, -1, -1, 0, caption),
			map[string]any{"iteration": iter + 1, "changed": false})

		for i, p := range pts {
			caption = fmt.Sprintf("Evaluating nearest centroid for Point P%d at (%v, %v).", i, p.X, p.Y)
			b.AddStep("L6", trace.ActionVisit, caption,
				makeGraphState(pts, centroids, assignments, i, -1, 0, caption),
				map[string]any{"point": i, "x": p.X, "y": p.Y})

			bestCluster := 0
			minDist := math.Inf(1)
			for cIdx, c := range centroids {
				d := math.Sqrt(distSq(p, c))
				b.IncComparisons()
				if d < minDist {
					minDist = d
					bestCluster = cIdx
				}
				distCaption := fmt.Sprintf("Point P%d (%v, %v) distance to Centroid C%d is %.2f", i, p.X, p.Y, cIdx, d)
				b.AddStep("L7", trace.ActionCompare, distCaption,
					makeGraphState(pts, centroids, assignments, i, cIdx, d, distCaption),
					map[string]any{"point": i, "centroid": cIdx, "dist": d})
			}

			closestCaption := fmt.Sprintf("Point (%v, %v) closest to Centroid %d (dist=%.2f); assign to Cluster %d", p.X, p.Y, bestCluster, minDist, bestCluster)
			b.AddStep("L7", trace.ActionCompare, closestCaption,
				makeGraphState(pts, centroids, assignments, i, bestCluster, minDist, closestCaption),
				map[string]any{"point": i, "closestCentroid": bestCluster, "dist": minDist})

			checkCaption := fmt.Sprintf("Check if cluster assignment changed for Point P%d: current=%d, new=%d", i, assignments[i], bestCluster)
			b.AddStep("L8", trace.ActionCompare, checkCaption,
				makeGraphState(pts, centroids, assignments, i, -1, 0, checkCaption),
				map[string]any{"point": i, "current": assignments[i], "new": bestCluster})

			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
				b.IncSwaps()

				assignCaption := fmt.Sprintf("Point (%v, %v) closest to Centroid %d (dist=%.2f); assign to Cluster %d", p.X, p.Y, bestCluster, minDist, bestCluster)
				b.AddStep("L9", trace.ActionSet, assignCaption,
					makeGraphState(pts, centroids, assignments, -1, -1, 0, assignCaption),
					map[string]any{"point": i, "cluster": bestCluster, "dist": minDist})

				b.AddStep("L10", trace.ActionSet, "Mark changed flag as true.",
					makeGraphState(pts, centroids, assignments, -1, -1, 0, "Mark changed flag as true."),
					map[string]any{"changed": true})
			}
		}

		counts := make([]int, k)
		sums := make([]Point, k)
		for idx, p := range pts {
			c := assignments[idx]
			if c >= 0 && c < k {
				counts[c]++
				sums[c].X += p.X
				sums[c].Y += p.Y
			}
		}

		maxMovement := 0.0
		for j := 0; j < k; j++ {
			oldCentroid := centroids[j]
			if counts[j] > 0 {
				newCentroid := Point{
					X: sums[j].X / float64(counts[j]),
					Y: sums[j].Y / float64(counts[j]),
				}
				movement := math.Sqrt(distSq(oldCentroid, newCentroid))
				if movement > maxMovement {
					maxMovement = movement
				}
				centroids[j] = newCentroid
				b.IncOperations()

				var updateCaption string
				if movement < threshold {
					updateCaption = fmt.Sprintf("Centroid C%d converged: movement %.2f < threshold %.3f at (%.2f, %.2f)", j, movement, threshold, newCentroid.X, newCentroid.Y)
				} else {
					updateCaption = fmt.Sprintf("Centroid C%d moved from (%.2f, %.2f) to mean (%.2f, %.2f) of %d points (movement=%.2f)", j, oldCentroid.X, oldCentroid.Y, newCentroid.X, newCentroid.Y, counts[j], movement)
				}

				b.AddStep("L13", trace.ActionSet, updateCaption,
					makeGraphState(pts, centroids, assignments, -1, -1, 0, updateCaption),
					map[string]any{
						"centroid":    j,
						"oldPos":      oldCentroid,
						"newPos":      newCentroid,
						"movement":    movement,
						"pointsCount": counts[j],
					})
			}
		}

		converged := !changed && maxMovement < threshold
		if converged || !changed {
			convCaption := fmt.Sprintf("Convergence check at iteration %d: centroids stabilized (max movement %.2f < threshold %.3f).", iter+1, maxMovement, threshold)
			b.AddStep("L14", trace.ActionCompare, convCaption,
				makeGraphState(pts, centroids, assignments, -1, -1, 0, convCaption),
				map[string]any{"converged": true, "iteration": iter + 1, "changed": changed, "maxMovement": maxMovement})
			break
		} else {
			contCaption := fmt.Sprintf("Iteration %d complete: changed=%t, max movement=%.2f; continuing to iteration %d.", iter+1, changed, maxMovement, iter+2)
			b.AddStep("L14", trace.ActionCompare, contCaption,
				makeGraphState(pts, centroids, assignments, -1, -1, 0, contCaption),
				map[string]any{"converged": false, "iteration": iter + 1, "changed": changed, "maxMovement": maxMovement})
		}
	}

	doneCaption := fmt.Sprintf("K-Means clustering complete after %d iteration(s): centroids stabilized, final clusters clearly distinguished.", finalIter)
	b.AddStep("L16", trace.ActionDone, doneCaption,
		makeGraphState(pts, centroids, assignments, -1, -1, 0, doneCaption),
		map[string]any{"converged": true, "iterations": finalIter, "centroids": clonePoints(centroids), "assignments": cloneSlice(assignments)})

	return b.Steps()
}
