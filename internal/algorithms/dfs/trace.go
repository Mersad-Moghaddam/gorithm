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

	var worklist []renderers.WorklistItem

	snapshot := func(caption string) renderers.GraphState {
		snNodes := make([]renderers.GraphNode, len(clonedNodes))
		copy(snNodes, clonedNodes)
		snEdges := make([]renderers.GraphEdge, len(clonedEdges))
		copy(snEdges, clonedEdges)
		snWorklist := make([]renderers.WorklistItem, len(worklist))
		copy(snWorklist, worklist)
		return renderers.GraphState{
			Caption:       caption,
			WorklistLabel: "Call Stack (LIFO)",
			Worklist:      snWorklist,
			Nodes:         snNodes,
			Edges:         snEdges,
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

	clearActiveEdges := func() {
		for i := range clonedEdges {
			clonedEdges[i].Active = false
		}
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Start DFS traversal from start node %s.", start), snapshot(fmt.Sprintf("Start DFS traversal from node %s", start)), map[string]any{
		"start": start,
	})

	visited := make(map[string]bool, len(clonedNodes))
	b.AddStep("L2", trace.ActionSet, "Initialize visited map to track visited vertices.", snapshot("Initialize visited set"), map[string]any{
		"visited": []string{},
	})

	order := make([]string, 0, len(clonedNodes))
	b.AddStep("L3", trace.ActionSet, "Initialize order slice to record traversal sequence.", snapshot("Initialize traversal order"), map[string]any{
		"order": []string{},
	})

	worklist = append(worklist, renderers.WorklistItem{ID: start, Label: fmt.Sprintf("dfs(%s)", start)})
	clonedNodes[nodeIdx[start]].State = "visiting"
	b.AddStep("L14", trace.ActionVisit, fmt.Sprintf("Call dfs(%s) to begin exploration.", start), snapshot(fmt.Sprintf("Call DFS(%s) to begin exploration", start)), map[string]any{
		"target": start,
	})

	var dfs func(u string)
	dfs = func(u string) {
		clearActiveEdges()
		visited[u] = true
		b.IncOperations()
		b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Mark node %s as visiting.", u), snapshot(fmt.Sprintf("Visiting %s (active on top of stack)", u)), map[string]any{
			"u":     u,
			"state": "visiting",
		})

		order = append(order, u)
		b.AddStep("L7", trace.ActionPush, fmt.Sprintf("Append node %s to traversal order: %v.", u, order), snapshot(fmt.Sprintf("Append %s to traversal order", u)), map[string]any{
			"u":     u,
			"order": cloneSlice(order),
		})

		neighbors := adj[u]
		b.AddStep("L8", trace.ActionHighlight, fmt.Sprintf("Explore outgoing edges from node %s: %v.", u, neighbors), snapshot(fmt.Sprintf("Explore outgoing edges from %s", u)), map[string]any{
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
				var caption string
				if vState == "visiting" {
					reason = fmt.Sprintf("Neighbor %s is currently visiting (cycle detected). Skip edge (%s -> %s).", v, u, v)
					caption = fmt.Sprintf("Cycle detected: neighbor %s is currently on stack", v)
				} else {
					reason = fmt.Sprintf("Neighbor %s is already visited. Skip edge (%s -> %s).", v, u, v)
					caption = fmt.Sprintf("Neighbor %s is already visited; skip edge", v)
				}
				b.AddStep("L9", trace.ActionCompare, reason, snapshot(caption), map[string]any{
					"u":       u,
					"v":       v,
					"visited": true,
				})
				setEdgeActive(u, v, false)
			} else {
				b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Neighbor %s is unvisited. Traverse edge (%s -> %s).", v, u, v), snapshot(fmt.Sprintf("Neighbor %s is unvisited: descend", v)), map[string]any{
					"u":       u,
					"v":       v,
					"visited": false,
				})

				worklist = append(worklist, renderers.WorklistItem{ID: v, Label: fmt.Sprintf("dfs(%s)", v)})
				clonedNodes[vIdx].State = "visiting"
				b.AddStep("L10", trace.ActionPush, fmt.Sprintf("Dive deeper: call dfs(%s); push to stack.", v), snapshot("Dive deeper: call DFS(neighbor); push to stack"), map[string]any{
					"from": u,
					"to":   v,
				})

				dfs(v)

				clearActiveEdges()
				b.AddStep("L11", trace.ActionPop, fmt.Sprintf("Backtrack to node %s from node %s.", u, v), snapshot(fmt.Sprintf("Backtrack to %s from %s", u, v)), map[string]any{
					"u":        u,
					"backFrom": v,
				})
			}
		}

		clonedNodes[nodeIdx[u]].State = "visited"
		if len(worklist) > 0 {
			worklist = worklist[:len(worklist)-1]
		}
		b.IncOperations()
		b.AddStep("L12", trace.ActionPop, fmt.Sprintf("All neighbors of %s explored. Mark %s as visited and pop from stack.", u, u), snapshot("No unvisited neighbors: backtrack; pop from stack"), map[string]any{
			"u":     u,
			"state": "visited",
		})
	}

	dfs(start)

	b.AddStep("L15", trace.ActionDone, fmt.Sprintf("DFS traversal complete. Visited order: %v.", order), snapshot("DFS traversal complete"), map[string]any{
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
