package kosaraju

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Kosaraju's Strongly Connected Components algorithm.
// ponytail: two-pass recursive DFS trace; upgrade to iterative trace if recursion exceeds stack limits on large graphs.
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
		clonedEdges[i].Directed = true
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

	worklistLabel := "Finish Stack (LIFO)"
	worklist := make([]renderers.WorklistItem, 0, len(clonedNodes))

	snapshot := func(caption string) renderers.GraphState {
		snNodes := make([]renderers.GraphNode, len(clonedNodes))
		copy(snNodes, clonedNodes)
		snEdges := make([]renderers.GraphEdge, len(clonedEdges))
		copy(snEdges, clonedEdges)
		snWorklist := make([]renderers.WorklistItem, len(worklist))
		copy(snWorklist, worklist)
		return renderers.GraphState{
			Caption:       caption,
			WorklistLabel: worklistLabel,
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
	}

	components := make([][]string, 0)
	sccStates := []string{"visited", "path", "queued", "visiting"}

	b.AddStep("L1", trace.ActionInit, "Initialize Kosaraju's algorithm to compute Strongly Connected Components.", snapshot("Phase 1: Forward DFS on original graph"), map[string]any{
		"phase":        "Phase 1: Forward DFS",
		"pass":         1,
		"components":   cloneComponents(components),
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
	})

	b.AddStep("L2", trace.ActionSet, "Build forward adjacency list and transposed (reversed) adjacency list.", snapshot("Phase 1: Forward DFS on original graph"), map[string]any{
		"phase":        "Phase 1: Forward DFS",
		"pass":         1,
		"components":   cloneComponents(components),
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
	})

	visited := make(map[string]bool, len(clonedNodes))
	b.AddStep("L3", trace.ActionSet, "Initialize visited map and empty finish stack for forward DFS.", snapshot("Phase 1: Forward DFS on original graph"), map[string]any{
		"phase":        "Phase 1: Forward DFS",
		"pass":         1,
		"components":   cloneComponents(components),
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
	})

	var dfs1 func(u string)
	dfs1 = func(u string) {
		visited[u] = true
		clonedNodes[nodeIdx[u]].State = "visiting"
		b.IncOperations()
		b.AddStep("L5", trace.ActionVisit, fmt.Sprintf("Phase 1: Visit vertex %s in forward DFS.", u), snapshot(fmt.Sprintf("Phase 1: Visiting vertex %s", u)), map[string]any{
			"phase":        "Phase 1: Forward DFS",
			"pass":         1,
			"u":            u,
			"finish_stack": worklistIDs(worklist),
			"order":        worklistIDs(worklist),
			"components":   cloneComponents(components),
		})

		for _, v := range adj[u] {
			if _, exists := nodeIdx[v]; !exists {
				continue
			}
			b.IncComparisons()
			setEdgeActive(u, v, true)
			b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Phase 1: Explore outgoing edge (%s -> %s). Target visited: %t.", u, v, visited[v]), snapshot(fmt.Sprintf("Phase 1: Exploring edge %s -> %s", u, v)), map[string]any{
				"phase":        "Phase 1: Forward DFS",
				"pass":         1,
				"u":            u,
				"v":            v,
				"visited":      visited[v],
				"finish_stack": worklistIDs(worklist),
				"order":        worklistIDs(worklist),
				"components":   cloneComponents(components),
			})

			if !visited[v] {
				dfs1(v)
			}
			setEdgeActive(u, v, false)
		}

		clonedNodes[nodeIdx[u]].State = "visited"
		worklist = append(worklist, renderers.WorklistItem{ID: u, Label: u})
		b.IncOperations()
		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Phase 1: Vertex %s finished (all descendants explored). Push onto Finish Stack: %v.", u, worklistIDs(worklist)), snapshot(fmt.Sprintf("Phase 1: Pushed %s to finish stack", u)), map[string]any{
			"phase":        "Phase 1: Forward DFS",
			"pass":         1,
			"u":            u,
			"finish_stack": worklistIDs(worklist),
			"order":        worklistIDs(worklist),
			"components":   cloneComponents(components),
		})
	}

	for _, n := range clonedNodes {
		u := n.ID
		b.IncComparisons()
		b.AddStep("L4", trace.ActionCompare, fmt.Sprintf("Phase 1: Inspect root vertex %s. Visited: %t.", u, visited[u]), snapshot("Phase 1: Checking root vertices"), map[string]any{
			"phase":        "Phase 1: Forward DFS",
			"pass":         1,
			"u":            u,
			"visited":      visited[u],
			"finish_stack": worklistIDs(worklist),
			"order":        worklistIDs(worklist),
			"components":   cloneComponents(components),
		})
		if !visited[u] {
			dfs1(u)
		}
	}

	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Phase 1 complete. Finish stack populated (LIFO): %v.", worklistIDs(worklist)), snapshot("Phase 1 complete: Finish stack populated"), map[string]any{
		"phase":        "Phase 1: Forward DFS",
		"pass":         1,
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
		"components":   cloneComponents(components),
	})

	// Phase 2: Graph transposition — all directed edges are reversed!
	for i := range clonedNodes {
		clonedNodes[i].State = "unvisited"
	}
	for i := range clonedEdges {
		clonedEdges[i].From, clonedEdges[i].To = clonedEdges[i].To, clonedEdges[i].From
		clonedEdges[i].Active = false
		clonedEdges[i].Directed = true
	}

	b.AddStep("L7", trace.ActionSet, "Phase 2: Graph transposition — reversed all directed edges! In the component DAG, reversing edges flips cross-component paths (source SCCs become sinks). Traversal starting from the highest-finish vertex is strictly trapped inside each SCC.", snapshot("Phase 2: Graph transposition (all directed edges reversed)"), map[string]any{
		"phase":        "Phase 2: Graph Transposition",
		"pass":         2,
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
		"components":   cloneComponents(components),
	})

	// Phase 3: Backward DFS in reverse finish order
	visitedRev := make(map[string]bool, len(clonedNodes))
	b.AddStep("L8", trace.ActionSet, "Phase 3 setup: Initialize second visited map and components list for backward DFS decomposition.", snapshot("Phase 3: Backward DFS on transposed graph"), map[string]any{
		"phase":        "Phase 3: Backward DFS",
		"pass":         3,
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
		"components":   cloneComponents(components),
	})

	var dfs2 func(u string, comp *[]string, compState string, compIdx int)
	dfs2 = func(u string, comp *[]string, compState string, compIdx int) {
		visitedRev[u] = true
		clonedNodes[nodeIdx[u]].State = compState
		*comp = append(*comp, u)
		b.IncOperations()
		b.AddStep("L13", trace.ActionVisit, fmt.Sprintf("Phase 3: Reached vertex %s in transposed graph. Added to SCC #%d: %v.", u, compIdx+1, *comp), snapshot(fmt.Sprintf("Phase 3: Expanding SCC #%d with vertex %s", compIdx+1, u)), map[string]any{
			"phase":             "Phase 3: Backward DFS",
			"pass":              3,
			"u":                 u,
			"current_component": cloneSlice(*comp),
			"components":        cloneComponents(components),
			"finish_stack":      worklistIDs(worklist),
			"order":             worklistIDs(worklist),
		})

		for _, v := range revAdj[u] {
			if _, exists := nodeIdx[v]; !exists {
				continue
			}
			b.IncComparisons()
			setEdgeActive(u, v, true)
			b.AddStep("L13", trace.ActionHighlight, fmt.Sprintf("Phase 3: Traverse reversed edge (%s -> %s). Target visited: %t.", u, v, visitedRev[v]), snapshot(fmt.Sprintf("Phase 3: Exploring reversed edge %s -> %s", u, v)), map[string]any{
				"phase":             "Phase 3: Backward DFS",
				"pass":              3,
				"from":              u,
				"to":                v,
				"visited":           visitedRev[v],
				"current_component": cloneSlice(*comp),
				"components":        cloneComponents(components),
				"finish_stack":      worklistIDs(worklist),
				"order":             worklistIDs(worklist),
			})

			if !visitedRev[v] {
				dfs2(v, comp, compState, compIdx)
			}
			setEdgeActive(u, v, false)
		}
	}

	for len(worklist) > 0 {
		b.IncComparisons()
		b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Phase 3: Finish stack has %d item(s) remaining. Continue backward DFS loop.", len(worklist)), snapshot(fmt.Sprintf("Phase 3: %d vertex/vertices on finish stack", len(worklist))), map[string]any{
			"phase":        "Phase 3: Backward DFS",
			"pass":         3,
			"finish_stack": worklistIDs(worklist),
			"order":        worklistIDs(worklist),
			"components":   cloneComponents(components),
		})

		topIdx := len(worklist) - 1
		u := worklist[topIdx].ID
		worklist = worklist[:topIdx]

		b.IncOperations()
		b.AddStep("L10", trace.ActionPop, fmt.Sprintf("Phase 3: Popped vertex %s from finish stack (latest finish time among remaining vertices).", u), snapshot(fmt.Sprintf("Phase 3: Popped %s from finish stack", u)), map[string]any{
			"phase":             "Phase 3: Backward DFS",
			"pass":              3,
			"u":                 u,
			"visited_in_pass_3": visitedRev[u],
			"finish_stack":      worklistIDs(worklist),
			"order":             worklistIDs(worklist),
			"components":        cloneComponents(components),
		})

		b.IncComparisons()
		b.AddStep("L11", trace.ActionCompare, fmt.Sprintf("Phase 3: Check if vertex %s is already assigned to an SCC. Visited: %t.", u, visitedRev[u]), snapshot(fmt.Sprintf("Phase 3: Checking vertex %s", u)), map[string]any{
			"phase":             "Phase 3: Backward DFS",
			"pass":              3,
			"u":                 u,
			"visited_in_pass_3": visitedRev[u],
			"finish_stack":      worklistIDs(worklist),
			"order":             worklistIDs(worklist),
			"components":        cloneComponents(components),
		})

		if !visitedRev[u] {
			compIdx := len(components)
			compState := sccStates[compIdx%len(sccStates)]

			b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Phase 3: Start discovering SCC #%d seeded by root vertex %s (marked %s).", compIdx+1, u, compState), snapshot(fmt.Sprintf("Phase 3: Starting SCC #%d at %s", compIdx+1, u)), map[string]any{
				"phase":        "Phase 3: Backward DFS",
				"pass":         3,
				"u":            u,
				"components":   cloneComponents(components),
				"finish_stack": worklistIDs(worklist),
				"order":        worklistIDs(worklist),
			})

			var comp []string
			dfs2(u, &comp, compState, compIdx)
			components = append(components, cloneSlice(comp))

			b.AddStep("L14", trace.ActionPush, fmt.Sprintf("Phase 3: SCC #%d complete: %v (marked with state %q).", len(components), comp, compState), snapshot(fmt.Sprintf("SCC #%d complete: %v", len(components), comp)), map[string]any{
				"phase":        "Phase 3: Backward DFS",
				"pass":         3,
				"component":    cloneSlice(comp),
				"components":   cloneComponents(components),
				"finish_stack": worklistIDs(worklist),
				"order":        worklistIDs(worklist),
			})
		}
	}

	// Restore original edge orientations for final graph display
	for i := range clonedEdges {
		clonedEdges[i].From, clonedEdges[i].To = clonedEdges[i].To, clonedEdges[i].From
		clonedEdges[i].Active = false
		clonedEdges[i].Directed = true
	}

	b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Kosaraju's algorithm complete. Identified %d strongly connected component(s): %v.", len(components), components), snapshot(fmt.Sprintf("Complete: %d SCC(s) identified", len(components))), map[string]any{
		"phase":        "Complete",
		"components":   cloneComponents(components),
		"finish_stack": worklistIDs(worklist),
		"order":        worklistIDs(worklist),
		"count":        len(components),
		"status":       "complete",
	})

	return b.Steps()
}

func worklistIDs(items []renderers.WorklistItem) []string {
	if items == nil {
		return []string{}
	}
	res := make([]string, len(items))
	for i, item := range items {
		res[i] = item.ID
	}
	return res
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
