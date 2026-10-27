package dfs

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Depth-First Search on a graph.
// ponytail: DFS trace with adjacency map built from edges; upgrade to handle multigraphs or parallel edges if needed.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}

	clonedNodes := make([]renderers.GraphNode, len(nodes))
	copy(clonedNodes, nodes)
	clonedEdges := make([]renderers.GraphEdge, len(edges))
	copy(clonedEdges, edges)

	for i := range clonedNodes {
		clonedNodes[i].State = "unvisited"
	}
	for i := range clonedEdges {
		clonedEdges[i].Active = false
	}

	nodeIdx := make(map[string]int, len(clonedNodes))
	for i, n := range clonedNodes {
		nodeIdx[n.ID] = i
	}

	if start == "" {
		start = clonedNodes[0].ID
	}
	if _, exists := nodeIdx[start]; !exists {
		return b.Steps()
	}

	adj := make(map[string][]string, len(clonedNodes))
	for _, e := range clonedEdges {
		adj[e.From] = append(adj[e.From], e.To)
	}

	snapshot := func() renderers.GraphState {
		snNodes := make([]renderers.GraphNode, len(clonedNodes))
		copy(snNodes, clonedNodes)
		snEdges := make([]renderers.GraphEdge, len(clonedEdges))
		copy(snEdges, clonedEdges)
		return renderers.GraphState{
			Nodes: snNodes,
			Edges: snEdges,
		}
	}

	setEdgeActive := func(from, to string, active bool) {
		for i := range clonedEdges {
			if clonedEdges[i].From == from && clonedEdges[i].To == to {
				clonedEdges[i].Active = active
				return
			}
		}
		for i := range clonedEdges {
			if clonedEdges[i].From == to && clonedEdges[i].To == from {
				clonedEdges[i].Active = active
				return
			}
		}
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Start DFS traversal from start node %s.", start), snapshot(), map[string]any{
		"start": start,
	})

	visited := make(map[string]bool, len(clonedNodes))
	b.AddStep("L2", trace.ActionSet, "Initialize visited map to track visited vertices.", snapshot(), map[string]any{
		"visited": []string{},
	})

	order := make([]string, 0, len(clonedNodes))
	b.AddStep("L3", trace.ActionSet, "Initialize order slice to record traversal sequence.", snapshot(), map[string]any{
		"order": []string{},
	})

	b.AddStep("L14", trace.ActionVisit, fmt.Sprintf("Call dfs(%s) to begin exploration.", start), snapshot(), map[string]any{
		"target": start,
	})

	var dfs func(u string)
	dfs = func(u string) {
		visited[u] = true
		clonedNodes[nodeIdx[u]].State = "visiting"
		b.IncOperations()
		b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Mark node %s as visiting.", u), snapshot(), map[string]any{
			"u":     u,
			"state": "visiting",
		})

		order = append(order, u)
		b.AddStep("L7", trace.ActionPush, fmt.Sprintf("Append node %s to traversal order: %v.", u, order), snapshot(), map[string]any{
			"u":     u,
			"order": cloneSlice(order),
		})

		neighbors := adj[u]
		b.AddStep("L8", trace.ActionHighlight, fmt.Sprintf("Explore outgoing edges from node %s: %v.", u, neighbors), snapshot(), map[string]any{
			"u":         u,
			"neighbors": cloneSlice(neighbors),
		})

		for _, v := range neighbors {
			vIdx, vExists := nodeIdx[v]
			if !vExists {
				continue
			}

			b.IncComparisons()
			setEdgeActive(u, v, true)

			if visited[v] {
				vState := clonedNodes[vIdx].State
				var reason string
				if vState == "visiting" {
					reason = fmt.Sprintf("Neighbor %s is currently visiting (cycle detected). Skip edge (%s -> %s).", v, u, v)
				} else {
					reason = fmt.Sprintf("Neighbor %s is already visited. Skip edge (%s -> %s).", v, u, v)
				}
				b.AddStep("L9", trace.ActionCompare, reason, snapshot(), map[string]any{
					"u":       u,
					"v":       v,
					"visited": true,
				})
				setEdgeActive(u, v, false)
			} else {
				b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Neighbor %s is unvisited. Traverse edge (%s -> %s).", v, u, v), snapshot(), map[string]any{
					"u":       u,
					"v":       v,
					"visited": false,
				})

				b.AddStep("L10", trace.ActionPush, fmt.Sprintf("Recursively call dfs(%s).", v), snapshot(), map[string]any{
					"from": u,
					"to":   v,
				})

				dfs(v)

				setEdgeActive(u, v, false)
				b.AddStep("L11", trace.ActionPop, fmt.Sprintf("Backtrack to node %s from node %s.", u, v), snapshot(), map[string]any{
					"u":        u,
					"backFrom": v,
				})
			}
		}

		clonedNodes[nodeIdx[u]].State = "visited"
		b.IncOperations()
		b.AddStep("L12", trace.ActionSet, fmt.Sprintf("All neighbors of %s explored. Mark %s as visited.", u, u), snapshot(), map[string]any{
			"u":     u,
			"state": "visited",
		})
	}

	dfs(start)

	b.AddStep("L15", trace.ActionDone, fmt.Sprintf("DFS traversal complete. Visited order: %v.", order), snapshot(), map[string]any{
		"order":  cloneSlice(order),
		"status": "complete",
	})

	return b.Steps()
}

func cloneSlice(s []string) []string {
	if s == nil {
		return nil
	}
	c := make([]string, len(s))
	copy(c, s)
	return c
}
