package bellmanford

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func cloneNodes(nodes []renderers.GraphNode) []renderers.GraphNode {
	c := make([]renderers.GraphNode, len(nodes))
	copy(c, nodes)
	return c
}

func cloneEdges(edges []renderers.GraphEdge) []renderers.GraphEdge {
	c := make([]renderers.GraphEdge, len(edges))
	copy(c, edges)
	return c
}

func cloneDist(dist map[string]int) map[string]int {
	c := make(map[string]int, len(dist))
	for k, v := range dist {
		c[k] = v
	}
	return c
}

func findNodeIndex(nodes []renderers.GraphNode, id string) int {
	for i, n := range nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func formatDist(d int) string {
	if d == math.MaxInt {
		return "∞"
	}
	return strconv.Itoa(d)
}

// Trace executes Bellman-Ford and records each step for visualization.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	currNodes := cloneNodes(nodes)
	currEdges := cloneEdges(edges)

	dist := make(map[string]int, len(nodes))
	for _, n := range nodes {
		dist[n.ID] = math.MaxInt
	}

	startIdx := findNodeIndex(currNodes, start)
	if startIdx < 0 {
		b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Start node %q not found in graph.", start), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})
		b.AddStep("L18", trace.ActionDone, "Search terminated: invalid start node.", renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Bellman-Ford algorithm from source node %s.", start), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"start": start, "distances": cloneDist(dist)})

	dist[start] = 0
	currNodes[startIdx].State = "visiting"
	b.IncOperations()
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Initialize distances: dist[%s] = 0, all other nodes = ∞.", start), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"current": start, "distances": cloneDist(dist)})

	numPasses := len(nodes) - 1
	for pass := 1; pass <= numPasses; pass++ {
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting relaxation pass %d of %d.", pass, numPasses), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"pass": pass, "distances": cloneDist(dist)})

		relaxed := false
		b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Pass %d: reset relaxed flag to false.", pass), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"pass": pass, "relaxed": false, "distances": cloneDist(dist)})

		for eIdx := range currEdges {
			edge := currEdges[eIdx]
			u := edge.From
			v := edge.To
			w := edge.Weight

			currEdges[eIdx].Active = true
			b.AddStep("L5", trace.ActionVisit, fmt.Sprintf("Pass %d: inspecting directed edge (%s -> %s, weight %d).", pass, u, v, w), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"pass": pass, "from": u, "to": v, "weight": w, "distances": cloneDist(dist)})

			b.IncComparisons()
			if dist[u] != math.MaxInt && dist[u]+w < dist[v] {
				newDist := dist[u] + w
				oldDistStr := formatDist(dist[v])
				b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Relaxation check passed: dist[%s] (%d) + weight (%d) = %d < dist[%s] (%s).", u, dist[u], w, newDist, v, oldDistStr), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"pass": pass, "from": u, "to": v, "newDist": newDist, "relaxed": true, "distances": cloneDist(dist)})

				dist[v] = newDist
				b.IncOperations()

				vIdx := findNodeIndex(currNodes, v)
				if vIdx >= 0 {
					currNodes[vIdx].State = "visited"
				}

				b.AddStep("L7", trace.ActionSet, fmt.Sprintf("Updated shortest distance to %s: dist[%s] = %d.", v, v, newDist), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"pass": pass, "updatedNode": v, "distances": cloneDist(dist)})

				relaxed = true
				b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Edge relaxed in pass %d. Set relaxed = true.", pass), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"pass": pass, "relaxed": true, "distances": cloneDist(dist)})
			} else {
				reason := "predecessor distance is ∞"
				if dist[u] != math.MaxInt {
					reason = fmt.Sprintf("candidate %d >= dist[%s] (%d)", dist[u]+w, v, dist[v])
				}
				b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Edge (%s -> %s) cannot be relaxed: %s.", u, v, reason), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"pass": pass, "from": u, "to": v, "distances": cloneDist(dist)})
			}

			currEdges[eIdx].Active = false
		}

		b.AddStep("L11", trace.ActionCompare, fmt.Sprintf("Pass %d complete. Any edges relaxed: %t.", pass, relaxed), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"pass": pass, "relaxed": relaxed, "distances": cloneDist(dist)})

		if !relaxed {
			b.AddStep("L11", trace.ActionHighlight, "No edge was relaxed in this pass. Distances have converged early.", renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"pass": pass, "distances": cloneDist(dist)})
			break
		}
	}

	// Negative cycle detection pass
	b.AddStep("L13", trace.ActionHighlight, "Checking for reachable negative-weight cycles with a final edge scan.", renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"distances": cloneDist(dist)})

	for eIdx := range currEdges {
		edge := currEdges[eIdx]
		u := edge.From
		v := edge.To
		w := edge.Weight

		currEdges[eIdx].Active = true
		b.IncComparisons()

		if dist[u] != math.MaxInt && dist[u]+w < dist[v] {
			b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Negative cycle check: edge (%s -> %s) can still be relaxed (%d + %d = %d < %s).", u, v, dist[u], w, dist[u]+w, formatDist(dist[v])), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"from": u, "to": v, "distances": cloneDist(dist), "negativeCycle": true})

			b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Negative cycle detected reachable via edge (%s -> %s)!", u, v), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"from": u, "to": v, "distances": cloneDist(dist), "negativeCycle": true})

			return b.Steps()
		}

		b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Negative cycle check: edge (%s -> %s) is stable.", u, v), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"from": u, "to": v, "distances": cloneDist(dist)})

		currEdges[eIdx].Active = false
	}

	b.AddStep("L18", trace.ActionDone, fmt.Sprintf("Bellman-Ford completed! Optimal shortest paths from %s found with no negative cycles.", start), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"start": start, "distances": cloneDist(dist), "negativeCycle": false})

	return b.Steps()
}
