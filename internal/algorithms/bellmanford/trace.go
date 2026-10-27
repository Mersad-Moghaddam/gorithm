package bellmanford

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func cloneNodes(nodes []renderers.GraphNode) []renderers.GraphNode {
	if nodes == nil {
		return nil
	}
	c := make([]renderers.GraphNode, len(nodes))
	copy(c, nodes)
	return c
}

func cloneEdges(edges []renderers.GraphEdge) []renderers.GraphEdge {
	if edges == nil {
		return nil
	}
	c := make([]renderers.GraphEdge, len(edges))
	copy(c, edges)
	return c
}

func cloneWorklist(items []renderers.WorklistItem) []renderers.WorklistItem {
	if items == nil {
		return []renderers.WorklistItem{}
	}
	c := make([]renderers.WorklistItem, len(items))
	copy(c, items)
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

func makeWorklist(edges []renderers.GraphEdge) []renderers.WorklistItem {
	items := make([]renderers.WorklistItem, len(edges))
	for i, e := range edges {
		items[i] = renderers.WorklistItem{
			ID:    fmt.Sprintf("%s->%s", e.From, e.To),
			Label: fmt.Sprintf("%s→%s", e.From, e.To),
		}
	}
	return items
}

func makeGraphState(caption string, worklist []renderers.WorklistItem, nodes []renderers.GraphNode, edges []renderers.GraphEdge) renderers.GraphState {
	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: "Pass Edges to Relax",
		Worklist:      cloneWorklist(worklist),
		Nodes:         cloneNodes(nodes),
		Edges:         cloneEdges(edges),
	}
}

func highlightNegativeCycle(nodes []renderers.GraphNode, edges []renderers.GraphEdge, parent map[string]string, u, v string) {
	for i := range edges {
		edges[i].Active = false
	}
	parent[v] = u

	// Step back |V| times to guarantee entering the cycle
	curr := u
	for i := 0; i < len(nodes); i++ {
		if p, ok := parent[curr]; ok && p != "" {
			curr = p
		}
	}

	cycleNodes := make(map[string]bool)
	cycleEdges := make(map[string]bool)

	startNode := curr
	p := curr
	for {
		cycleNodes[p] = true
		next := parent[p]
		if next == "" {
			break
		}
		cycleEdges[fmt.Sprintf("%s->%s", next, p)] = true
		p = next
		if p == startNode || cycleNodes[p] {
			cycleNodes[p] = true
			break
		}
	}

	// Always ensure triggering edge and endpoints are highlighted
	cycleNodes[u] = true
	cycleNodes[v] = true
	cycleEdges[fmt.Sprintf("%s->%s", u, v)] = true

	for i := range edges {
		key := fmt.Sprintf("%s->%s", edges[i].From, edges[i].To)
		if cycleEdges[key] {
			edges[i].Active = true
		}
	}

	for i := range nodes {
		if cycleNodes[nodes[i].ID] {
			nodes[i].State = "path"
		}
	}
}

// Trace executes Bellman-Ford and records each step for visualization.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	currNodes := cloneNodes(nodes)
	currEdges := cloneEdges(edges)

	for i := range currEdges {
		currEdges[i].Directed = true
	}

	dist := make(map[string]int, len(nodes))
	parent := make(map[string]string, len(nodes))
	for _, n := range nodes {
		dist[n.ID] = math.MaxInt
	}

	numPasses := len(nodes) - 1

	startIdx := findNodeIndex(currNodes, start)
	if startIdx < 0 {
		caption := fmt.Sprintf("Start node %q not found in graph.", start)
		b.AddStep("L1", trace.ActionInit, caption, makeGraphState(caption, nil, currNodes, currEdges), map[string]any{
			"pass":        0,
			"totalPasses": numPasses,
			"distances":   cloneDist(dist),
		})
		termCaption := "Search terminated: invalid start node."
		b.AddStep("L18", trace.ActionDone, termCaption, makeGraphState(termCaption, nil, currNodes, currEdges), map[string]any{
			"pass":        0,
			"totalPasses": numPasses,
			"distances":   cloneDist(dist),
		})
		return b.Steps()
	}

	initCaption := fmt.Sprintf("Pass 0: Initialize Bellman-Ford from source %s. |V|=%d vertices, max relaxation passes |V|-1=%d.", start, len(nodes), numPasses)
	b.AddStep("L1", trace.ActionInit, initCaption, makeGraphState(initCaption, nil, currNodes, currEdges), map[string]any{
		"start":       start,
		"pass":        0,
		"totalPasses": numPasses,
		"distances":   cloneDist(dist),
	})

	dist[start] = 0
	currNodes[startIdx].State = "visiting"
	b.IncOperations()

	setCaption := fmt.Sprintf("Pass 0: Set dist[%s] = 0, all other nodes = ∞. Ready to relax edges.", start)
	b.AddStep("L2", trace.ActionSet, setCaption, makeGraphState(setCaption, nil, currNodes, currEdges), map[string]any{
		"current":     start,
		"pass":        0,
		"totalPasses": numPasses,
		"distances":   cloneDist(dist),
	})

	for pass := 1; pass <= numPasses; pass++ {
		allPassWorklist := makeWorklist(currEdges)

		passStartCaption := fmt.Sprintf("Pass %d of %d: Beginning edge relaxation pass.", pass, numPasses)
		b.AddStep("L3", trace.ActionHighlight, passStartCaption, makeGraphState(passStartCaption, allPassWorklist, currNodes, currEdges), map[string]any{
			"pass":        pass,
			"totalPasses": numPasses,
			"distances":   cloneDist(dist),
		})

		relaxed := false
		resetCaption := fmt.Sprintf("Pass %d: Reset relaxed = false. Checking %d edges for distance improvements.", pass, len(currEdges))
		b.AddStep("L4", trace.ActionSet, resetCaption, makeGraphState(resetCaption, allPassWorklist, currNodes, currEdges), map[string]any{
			"pass":        pass,
			"totalPasses": numPasses,
			"relaxed":     false,
			"distances":   cloneDist(dist),
		})

		for eIdx := range currEdges {
			edge := currEdges[eIdx]
			u := edge.From
			v := edge.To
			w := edge.Weight
			remainingWorklist := makeWorklist(currEdges[eIdx:])

			currEdges[eIdx].Active = true
			visitCaption := fmt.Sprintf("Pass %d: Inspecting edge %s→%s (weight %d).", pass, u, v, w)
			b.AddStep("L5", trace.ActionVisit, visitCaption, makeGraphState(visitCaption, remainingWorklist, currNodes, currEdges), map[string]any{
				"pass":        pass,
				"totalPasses": numPasses,
				"from":        u,
				"to":          v,
				"weight":      w,
				"relaxed":     relaxed,
				"distances":   cloneDist(dist),
			})

			b.IncComparisons()
			if dist[u] != math.MaxInt && dist[u]+w < dist[v] {
				newDist := dist[u] + w
				oldDistStr := formatDist(dist[v])

				relaxCheckCaption := fmt.Sprintf("Pass %d: Relaxation condition met! dist[%s] (%d) + weight (%d) = %d < dist[%s] (%s).", pass, u, dist[u], w, newDist, v, oldDistStr)
				b.AddStep("L6", trace.ActionCompare, relaxCheckCaption, makeGraphState(relaxCheckCaption, remainingWorklist, currNodes, currEdges), map[string]any{
					"pass":        pass,
					"totalPasses": numPasses,
					"from":        u,
					"to":          v,
					"weight":      w,
					"newDist":     newDist,
					"oldDist":     oldDistStr,
					"relaxed":     true,
					"distances":   cloneDist(dist),
				})

				dist[v] = newDist
				parent[v] = u
				b.IncOperations()

				vIdx := findNodeIndex(currNodes, v)
				if vIdx >= 0 {
					currNodes[vIdx].State = "visited"
				}

				updateCaption := fmt.Sprintf("Pass %d: Updated tentative distance dist[%s] = %d (via %s).", pass, v, newDist, u)
				b.AddStep("L7", trace.ActionSet, updateCaption, makeGraphState(updateCaption, remainingWorklist, currNodes, currEdges), map[string]any{
					"pass":              pass,
					"totalPasses":       numPasses,
					"updatedNode":       v,
					"newDist":           newDist,
					"tentativeDistance": newDist,
					"relaxed":           true,
					"distances":         cloneDist(dist),
				})

				relaxed = true
				setRelaxedCaption := fmt.Sprintf("Pass %d: Distance relaxed! Set relaxed = true.", pass)
				b.AddStep("L8", trace.ActionSet, setRelaxedCaption, makeGraphState(setRelaxedCaption, remainingWorklist, currNodes, currEdges), map[string]any{
					"pass":        pass,
					"totalPasses": numPasses,
					"relaxed":     true,
					"distances":   cloneDist(dist),
				})
			} else {
				var nonRelaxCaption string
				var reason string
				if dist[u] == math.MaxInt {
					reason = fmt.Sprintf("dist[%s] is ∞", u)
					nonRelaxCaption = fmt.Sprintf("Pass %d: Edge %s→%s not relaxed: dist[%s] is ∞ (node %s unreachable). dist[%s] stays %s.", pass, u, v, u, u, v, formatDist(dist[v]))
				} else {
					reason = fmt.Sprintf("%d + %d = %d >= %d", dist[u], w, dist[u]+w, dist[v])
					nonRelaxCaption = fmt.Sprintf("Pass %d: Edge %s→%s not relaxed: dist[%s] (%d) + weight (%d) = %d ≥ dist[%s] (%d). Distance stays unchanged.", pass, u, v, u, dist[u], w, dist[u]+w, v, dist[v])
				}
				b.AddStep("L6", trace.ActionCompare, nonRelaxCaption, makeGraphState(nonRelaxCaption, remainingWorklist, currNodes, currEdges), map[string]any{
					"pass":        pass,
					"totalPasses": numPasses,
					"from":        u,
					"to":          v,
					"reason":      reason,
					"relaxed":     relaxed,
					"distances":   cloneDist(dist),
				})
			}

			currEdges[eIdx].Active = false
		}

		passEndCaption := fmt.Sprintf("Pass %d complete. Any edges relaxed: %t.", pass, relaxed)
		b.AddStep("L11", trace.ActionCompare, passEndCaption, makeGraphState(passEndCaption, []renderers.WorklistItem{}, currNodes, currEdges), map[string]any{
			"pass":        pass,
			"totalPasses": numPasses,
			"relaxed":     relaxed,
			"distances":   cloneDist(dist),
		})

		if !relaxed {
			convergedCaption := fmt.Sprintf("Pass %d: No edges relaxed. Distances have converged early! Stopping passes.", pass)
			b.AddStep("L11", trace.ActionHighlight, convergedCaption, makeGraphState(convergedCaption, []renderers.WorklistItem{}, currNodes, currEdges), map[string]any{
				"pass":           pass,
				"totalPasses":    numPasses,
				"convergedEarly": true,
				"distances":      cloneDist(dist),
			})
			break
		}
	}

	// Negative cycle check pass: pass |V|
	cyclePass := len(nodes)
	allCycleWorklist := makeWorklist(currEdges)

	cycleStartCaption := fmt.Sprintf("Pass %d (|V| Cycle Check): Verifying if any distance decreases further to detect negative cycles.", cyclePass)
	b.AddStep("L13", trace.ActionHighlight, cycleStartCaption, makeGraphState(cycleStartCaption, allCycleWorklist, currNodes, currEdges), map[string]any{
		"pass":                  cyclePass,
		"totalPasses":           cyclePass,
		"checkingNegativeCycle": true,
		"distances":             cloneDist(dist),
	})

	for eIdx := range currEdges {
		edge := currEdges[eIdx]
		u := edge.From
		v := edge.To
		w := edge.Weight
		remainingWorklist := makeWorklist(currEdges[eIdx:])

		currEdges[eIdx].Active = true
		b.IncComparisons()

		if dist[u] != math.MaxInt && dist[u]+w < dist[v] {
			cycleCheckCaption := fmt.Sprintf("Pass %d (Cycle Check): Edge %s→%s can still be relaxed (%d + %d = %d < %s)! Negative cycle detected!", cyclePass, u, v, dist[u], w, dist[u]+w, formatDist(dist[v]))
			b.AddStep("L14", trace.ActionCompare, cycleCheckCaption, makeGraphState(cycleCheckCaption, remainingWorklist, currNodes, currEdges), map[string]any{
				"pass":          cyclePass,
				"from":          u,
				"to":            v,
				"negativeCycle": true,
				"distances":     cloneDist(dist),
			})

			highlightNegativeCycle(currNodes, currEdges, parent, u, v)

			cycleDetectedCaption := "Negative cycle detected! Graph contains a reachable negative cycle; shortest paths are undefined."
			b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Negative cycle detected! Reachable cycle involves edge %s→%s.", u, v), makeGraphState(cycleDetectedCaption, remainingWorklist, currNodes, currEdges), map[string]any{
				"pass":          cyclePass,
				"from":          u,
				"to":            v,
				"negativeCycle": true,
				"distances":     cloneDist(dist),
			})

			return b.Steps()
		}

		stableCaption := fmt.Sprintf("Pass %d (Cycle Check): Edge %s→%s is stable. No negative cycle detected here.", cyclePass, u, v)
		b.AddStep("L14", trace.ActionCompare, stableCaption, makeGraphState(stableCaption, remainingWorklist, currNodes, currEdges), map[string]any{
			"pass":          cyclePass,
			"from":          u,
			"to":            v,
			"negativeCycle": false,
			"distances":     cloneDist(dist),
		})

		currEdges[eIdx].Active = false
	}

	doneCaption := fmt.Sprintf("Bellman-Ford complete! Optimal shortest paths from %s found with no negative cycles.", start)
	b.AddStep("L18", trace.ActionDone, doneCaption, makeGraphState(doneCaption, []renderers.WorklistItem{}, currNodes, currEdges), map[string]any{
		"start":         start,
		"distances":     cloneDist(dist),
		"negativeCycle": false,
	})

	return b.Steps()
}
