package kosaraju

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Kosaraju's Strongly Connected Components algorithm.
// ponytail: two-pass DFS trace with graph reversal; upgrade to iterative trace if recursion exceeds stack limits.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step {
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

	adj := make(map[string][]string, len(clonedNodes))
	revAdj := make(map[string][]string, len(clonedNodes))
	for _, n := range clonedNodes {
		adj[n.ID] = nil
		revAdj[n.ID] = nil
	}
	for _, e := range clonedEdges {
		adj[e.From] = append(adj[e.From], e.To)
		revAdj[e.To] = append(revAdj[e.To], e.From)
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
	}

	components := make([][]string, 0)
	order := make([]string, 0, len(clonedNodes))

	b.AddStep("L1", trace.ActionInit, "Initialize Kosaraju's algorithm to compute Strongly Connected Components.", snapshot(), map[string]any{
		"components": cloneComponents(components),
		"order":      cloneSlice(order),
		"pass":       1,
	})

	b.AddStep("L2", trace.ActionSet, "Build forward adjacency list and transposed (reversed) adjacency list.", snapshot(), map[string]any{
		"components": cloneComponents(components),
		"order":      cloneSlice(order),
		"pass":       1,
	})

	visited := make(map[string]bool, len(clonedNodes))
	b.AddStep("L3", trace.ActionSet, "Initialize visited map and finish order stack for Pass 1.", snapshot(), map[string]any{
		"components": cloneComponents(components),
		"order":      cloneSlice(order),
		"pass":       1,
	})

	var dfs1 func(u string)
	dfs1 = func(u string) {
		visited[u] = true
		clonedNodes[nodeIdx[u]].State = "visiting"
		b.IncOperations()
		b.AddStep("L5", trace.ActionVisit, fmt.Sprintf("Pass 1: Visit vertex %s and explore unvisited outgoing edges.", u), snapshot(), map[string]any{
			"u":          u,
			"order":      cloneSlice(order),
			"components": cloneComponents(components),
			"pass":       1,
		})

		for _, v := range adj[u] {
			if _, exists := nodeIdx[v]; !exists {
				continue
			}
			b.IncComparisons()
			setEdgeActive(u, v, true)
			b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Pass 1: Explore outgoing edge (%s -> %s). Target visited: %t.", u, v, visited[v]), snapshot(), map[string]any{
				"u":          u,
				"v":          v,
				"visited":    visited[v],
				"order":      cloneSlice(order),
				"components": cloneComponents(components),
				"pass":       1,
			})

			if !visited[v] {
				dfs1(v)
			}
			setEdgeActive(u, v, false)
		}

		order = append(order, u)
		clonedNodes[nodeIdx[u]].State = "visited"
		b.IncOperations()
		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Pass 1: Vertex %s finished. Push to finish order: %v.", u, order), snapshot(), map[string]any{
			"u":          u,
			"order":      cloneSlice(order),
			"components": cloneComponents(components),
			"pass":       1,
		})
	}

	for _, n := range clonedNodes {
		u := n.ID
		b.IncComparisons()
		b.AddStep("L4", trace.ActionCompare, fmt.Sprintf("Pass 1: Check root vertex %s. Visited: %t.", u, visited[u]), snapshot(), map[string]any{
			"u":          u,
			"visited":    visited[u],
			"order":      cloneSlice(order),
			"components": cloneComponents(components),
			"pass":       1,
		})
		if !visited[u] {
			dfs1(u)
		}
	}

	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Pass 1 complete. Finish order computed: %v.", order), snapshot(), map[string]any{
		"order":      cloneSlice(order),
		"components": cloneComponents(components),
		"pass":       1,
	})

	// Transpose graph: reverse edges in state and reset node states
	for i := range clonedNodes {
		clonedNodes[i].State = "unvisited"
	}
	for i := range clonedEdges {
		clonedEdges[i].From, clonedEdges[i].To = clonedEdges[i].To, clonedEdges[i].From
		clonedEdges[i].Active = false
	}

	b.AddStep("L7", trace.ActionSet, "Pass 2 setup: Transpose graph (reverse all directed edges) and reset node states for component decomposition.", snapshot(), map[string]any{
		"order":      cloneSlice(order),
		"components": cloneComponents(components),
		"pass":       2,
	})

	visitedRev := make(map[string]bool, len(clonedNodes))
	var dfs2 func(u string, comp *[]string)
	dfs2 = func(u string, comp *[]string) {
		visitedRev[u] = true
		clonedNodes[nodeIdx[u]].State = "path"
		*comp = append(*comp, u)
		b.IncOperations()
		b.AddStep("L12", trace.ActionVisit, fmt.Sprintf("Pass 2: Reached vertex %s in transposed graph. Add to current SCC: %v.", u, *comp), snapshot(), map[string]any{
			"u":                 u,
			"current_component": cloneSlice(*comp),
			"components":        cloneComponents(components),
			"order":             cloneSlice(order),
			"pass":              2,
		})

		for _, v := range revAdj[u] {
			if _, exists := nodeIdx[v]; !exists {
				continue
			}
			b.IncComparisons()
			setEdgeActive(u, v, true)
			b.AddStep("L12", trace.ActionHighlight, fmt.Sprintf("Pass 2: Traverse transposed edge (%s -> %s). Target visited: %t.", u, v, visitedRev[v]), snapshot(), map[string]any{
				"from":              u,
				"to":                v,
				"visited":           visitedRev[v],
				"current_component": cloneSlice(*comp),
				"components":        cloneComponents(components),
				"order":             cloneSlice(order),
				"pass":              2,
			})

			if !visitedRev[v] {
				dfs2(v, comp)
			}
			setEdgeActive(u, v, false)
		}
	}

	for i := len(order) - 1; i >= 0; i-- {
		u := order[i]
		b.IncComparisons()
		b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Pass 2: Inspect vertex %s with latest finish time. Already assigned: %t.", u, visitedRev[u]), snapshot(), map[string]any{
			"u":                 u,
			"order_index":       i,
			"visited_in_pass_2": visitedRev[u],
			"components":        cloneComponents(components),
			"order":             cloneSlice(order),
			"pass":              2,
		})

		if !visitedRev[u] {
			b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Pass 2: Start new Strongly Connected Component seeded by root vertex %s.", u), snapshot(), map[string]any{
				"u":          u,
				"components": cloneComponents(components),
				"order":      cloneSlice(order),
				"pass":       2,
			})

			var comp []string
			dfs2(u, &comp)
			components = append(components, cloneSlice(comp))

			b.AddStep("L13", trace.ActionSet, fmt.Sprintf("SCC #%d identified: %v.", len(components), comp), snapshot(), map[string]any{
				"component":  cloneSlice(comp),
				"components": cloneComponents(components),
				"order":      cloneSlice(order),
				"pass":       2,
			})
		}
	}

	// Restore original edge orientations for final graph display
	for i := range clonedEdges {
		clonedEdges[i].From, clonedEdges[i].To = clonedEdges[i].To, clonedEdges[i].From
		clonedEdges[i].Active = false
	}

	b.AddStep("L16", trace.ActionDone, fmt.Sprintf("Kosaraju's algorithm complete. Identified %d strongly connected component(s): %v.", len(components), components), snapshot(), map[string]any{
		"components": cloneComponents(components),
		"order":      cloneSlice(order),
		"count":      len(components),
		"status":     "complete",
	})

	return b.Steps()
}

func cloneComponents(c [][]string) [][]string {
	if c == nil {
		return [][]string{}
	}
	out := make([][]string, len(c))
	for i, comp := range c {
		out[i] = make([]string, len(comp))
		copy(out[i], comp)
	}
	return out
}

func cloneSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}
