package dijkstra

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

const worklistLabel = "Priority Queue (Min-Heap: node, dist)"

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

func buildWorklist(pq *minHeap) []renderers.WorklistItem {
	if pq == nil || pq.Len() == 0 {
		return []renderers.WorklistItem{}
	}
	items := make([]pqItem, pq.Len())
	copy(items, *pq)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].dist != items[j].dist {
			return items[i].dist < items[j].dist
		}
		return items[i].node < items[j].node
	})
	res := make([]renderers.WorklistItem, len(items))
	for i, it := range items {
		res[i] = renderers.WorklistItem{
			ID:    fmt.Sprintf("%s-%d", it.node, it.dist),
			Label: fmt.Sprintf("%s (d=%d)", it.node, it.dist),
		}
	}
	return res
}

func makeGraphState(nodes []renderers.GraphNode, edges []renderers.GraphEdge, worklist []renderers.WorklistItem, caption string) renderers.GraphState {
	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: worklistLabel,
		Worklist:      cloneWorklist(worklist),
		Nodes:         cloneNodes(nodes),
		Edges:         cloneEdges(edges),
	}
}

func findNodeIndex(nodes []renderers.GraphNode, id string) int {
	for i, n := range nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

func findEdgeIndex(edges []renderers.GraphEdge, u, v string) int {
	for i, e := range edges {
		if e.From == u && e.To == v {
			return i
		}
		if !e.Directed && e.From == v && e.To == u {
			return i
		}
	}
	return -1
}

// Trace executes Dijkstra's algorithm and records each step for visualization.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start, target string) []trace.Step {
	b := trace.NewBuilder()
	currNodes := cloneNodes(nodes)
	currEdges := cloneEdges(edges)

	dist := make(map[string]int, len(nodes))
	prev := make(map[string]string, len(nodes))
	for _, n := range nodes {
		dist[n.ID] = math.MaxInt
	}

	startIdx := findNodeIndex(currNodes, start)
	if startIdx < 0 {
		capErr := fmt.Sprintf("Start node %q not found in graph.", start)
		b.AddStep("L1", trace.ActionInit, capErr, makeGraphState(currNodes, currEdges, nil, capErr), map[string]any{"distances": cloneDist(dist)})
		b.AddStep("L17", trace.ActionDone, "Search terminated: invalid start node.", makeGraphState(currNodes, currEdges, nil, "Search terminated: invalid start node."), map[string]any{"distances": cloneDist(dist)})
		return b.Steps()
	}

	dist[start] = 0

	initCap := fmt.Sprintf("Initialize Dijkstra: find shortest path from %s to %s", start, target)
	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Dijkstra's Shortest Path from start node %s to target %s.", start, target),
		makeGraphState(currNodes, currEdges, nil, initCap),
		map[string]any{"current": "", "target": target, "distances": cloneDist(dist)})

	distCap := fmt.Sprintf("Set initial distance dist[%s] = 0, all other nodes = ∞", start)
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Set initial distance to start node %s = 0, all other nodes = ∞.", start),
		makeGraphState(currNodes, currEdges, nil, distCap),
		map[string]any{"current": start, "target": target, "distances": cloneDist(dist)})

	pqCap := "Initialize priority queue (min-heap) for tentative distances"
	b.AddStep("L3", trace.ActionInit, "Initialize empty min-priority queue.",
		makeGraphState(currNodes, currEdges, nil, pqCap),
		map[string]any{"distances": cloneDist(dist)})

	currNodes[startIdx].State = "queued"
	pq := &minHeap{}
	heap.Init(pq)
	heap.Push(pq, pqItem{node: start, dist: 0})
	b.IncOperations()

	wl := buildWorklist(pq)
	pushStartCap := fmt.Sprintf("Push start node %s (distance 0) into priority queue", start)
	b.AddStep("L4", trace.ActionPush, fmt.Sprintf("Push start node %s (distance 0) into priority queue.", start),
		makeGraphState(currNodes, currEdges, wl, pushStartCap),
		map[string]any{"current": start, "target": target, "distances": cloneDist(dist)})

	visited := make(map[string]bool, len(nodes))

	for pq.Len() > 0 {
		wl = buildWorklist(pq)
		loopCap := fmt.Sprintf("Priority queue has %d candidate(s); extract vertex with minimum distance", pq.Len())
		b.AddStep("L5", trace.ActionHighlight, "Priority queue is non-empty: selecting next candidate vertex.",
			makeGraphState(currNodes, currEdges, wl, loopCap),
			map[string]any{"distances": cloneDist(dist)})

		item := heap.Pop(pq).(pqItem)
		u := item.node
		d := item.dist
		b.IncOperations()

		wl = buildWorklist(pq)

		if visited[u] {
			skipCap := fmt.Sprintf("Extracted %s (d=%d): node already visited, discard outdated queue entry", u, d)
			b.AddStep("L6", trace.ActionPop, fmt.Sprintf("Extracted vertex %s (distance %d). Node already visited, skipping outdated queue entry.", u, d),
				makeGraphState(currNodes, currEdges, wl, skipCap),
				map[string]any{"current": u, "distances": cloneDist(dist)})
			continue
		}

		uIdx := findNodeIndex(currNodes, u)
		if uIdx >= 0 {
			currNodes[uIdx].State = "visiting"
		}

		popCap := fmt.Sprintf("Extracted node %s with minimum distance %d from priority queue", u, d)
		b.AddStep("L6", trace.ActionPop, fmt.Sprintf("Extracted vertex %s with minimum tentative distance %d from priority queue.", u, d),
			makeGraphState(currNodes, currEdges, wl, popCap),
			map[string]any{"current": u, "distances": cloneDist(dist)})

		b.IncComparisons()
		cmpTargetCap := fmt.Sprintf("Check if current vertex %s matches target %s", u, target)
		b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("Check if current vertex %s matches target %s.", u, target),
			makeGraphState(currNodes, currEdges, wl, cmpTargetCap),
			map[string]any{"current": u, "target": target, "distances": cloneDist(dist)})

		if target != "" && u == target {
			targetReachedCap := fmt.Sprintf("Target vertex %s reached with shortest distance %d!", target, dist[target])
			b.AddStep("L7", trace.ActionHighlight, targetReachedCap,
				makeGraphState(currNodes, currEdges, wl, targetReachedCap),
				map[string]any{"current": u, "target": target, "distances": cloneDist(dist)})
			break
		}

		// Inspect all incident edges
		for eIdx := range currEdges {
			edge := currEdges[eIdx]
			var v string
			if edge.From == u {
				v = edge.To
			} else if !edge.Directed && edge.To == u {
				v = edge.From
			} else {
				continue
			}

			if visited[v] {
				continue
			}

			currEdges[eIdx].Active = true

			inspectCap := fmt.Sprintf("Inspect neighbor %s from %s via edge weight %d", v, u, edge.Weight)
			b.AddStep("L8", trace.ActionVisit, fmt.Sprintf("Inspecting neighbor %s via edge (%s-%s, weight %d).", v, edge.From, edge.To, edge.Weight),
				makeGraphState(currNodes, currEdges, wl, inspectCap),
				map[string]any{"current": u, "neighbor": v, "edgeWeight": edge.Weight, "distances": cloneDist(dist)})

			alt := dist[u] + edge.Weight
			altCap := fmt.Sprintf("Calculate candidate distance: dist[%s] (%d) + weight (%d) = %d", u, dist[u], edge.Weight, alt)
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Calculate alternative distance to %s: dist[%s] (%d) + weight (%d) = %d.", v, u, dist[u], edge.Weight, alt),
				makeGraphState(currNodes, currEdges, wl, altCap),
				map[string]any{"current": u, "neighbor": v, "alt": alt, "distances": cloneDist(dist)})

			b.IncComparisons()
			if alt < dist[v] {
				oldDistStr := "∞"
				if dist[v] != math.MaxInt {
					oldDistStr = fmt.Sprintf("%d", dist[v])
				}
				relaxCap := fmt.Sprintf("Relax edge (%s, %s): alt (%d) < dist[%s] (%s) — shorter path found, update dist[%s] = %d", u, v, alt, v, oldDistStr, v, alt)
				b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Relaxation check: alt (%d) < dist[%s] (%s): condition holds.", alt, v, oldDistStr),
					makeGraphState(currNodes, currEdges, wl, relaxCap),
					map[string]any{"current": u, "neighbor": v, "alt": alt, "relaxed": true, "distances": cloneDist(dist)})

				dist[v] = alt
				updateDistCap := fmt.Sprintf("Update distance: dist[%s] = %d (reduced from %s)", v, alt, oldDistStr)
				b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Update shortest distance to %s = %d.", v, alt),
					makeGraphState(currNodes, currEdges, wl, updateDistCap),
					map[string]any{"current": u, "neighbor": v, "distances": cloneDist(dist)})

				prev[v] = u
				predCap := fmt.Sprintf("Update predecessor: prev[%s] = %s", v, u)
				b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Set predecessor of %s to %s.", v, u),
					makeGraphState(currNodes, currEdges, wl, predCap),
					map[string]any{"current": u, "neighbor": v, "predecessor": u, "distances": cloneDist(dist)})

				heap.Push(pq, pqItem{node: v, dist: alt})
				b.IncOperations()

				vIdx := findNodeIndex(currNodes, v)
				if vIdx >= 0 && currNodes[vIdx].State != "visiting" && currNodes[vIdx].State != "visited" {
					currNodes[vIdx].State = "queued"
				}

				wl = buildWorklist(pq)
				pushCap := fmt.Sprintf("Push %s (d=%d) into priority queue", v, alt)
				b.AddStep("L13", trace.ActionPush, fmt.Sprintf("Push %s with updated distance %d into priority queue.", v, alt),
					makeGraphState(currNodes, currEdges, wl, pushCap),
					map[string]any{"current": u, "neighbor": v, "distances": cloneDist(dist)})
			} else {
				noRelaxCap := fmt.Sprintf("Relaxation check: alt (%d) ≥ dist[%s] (%d) — keep existing shorter path", alt, v, dist[v])
				b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Relaxation check: alt (%d) >= dist[%s] (%d): existing distance is shorter or equal.", alt, v, dist[v]),
					makeGraphState(currNodes, currEdges, wl, noRelaxCap),
					map[string]any{"current": u, "neighbor": v, "alt": alt, "relaxed": false, "distances": cloneDist(dist)})
			}

			currEdges[eIdx].Active = false
		}

		visited[u] = true
		if uIdx >= 0 {
			currNodes[uIdx].State = "visited"
		}
		wl = buildWorklist(pq)
		settledCap := fmt.Sprintf("Settled node %s with shortest distance %d; marked as visited", u, dist[u])
		b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Finished exploring all neighbors for vertex %s. Node marked as visited.", u),
			makeGraphState(currNodes, currEdges, wl, settledCap),
			map[string]any{"current": u, "distances": cloneDist(dist)})
	}

	// Reconstruct path
	var path []string
	if target != "" && dist[target] != math.MaxInt {
		curr := target
		for curr != "" {
			path = append([]string{curr}, path...)
			if curr == start {
				break
			}
			curr = prev[curr]
		}
	}

	finalWorklist := buildWorklist(pq)

	if len(path) > 0 && path[0] == start {
		pathSet := make(map[string]bool, len(path))
		for _, p := range path {
			pathSet[p] = true
		}
		for i := range currNodes {
			if pathSet[currNodes[i].ID] {
				currNodes[i].State = "path"
			}
		}

		for i := range currEdges {
			currEdges[i].Active = false
		}
		for i := 0; i < len(path)-1; i++ {
			eIdx := findEdgeIndex(currEdges, path[i], path[i+1])
			if eIdx >= 0 {
				currEdges[eIdx].Active = true
			}
		}

		finalCap := fmt.Sprintf("Shortest path from %s to %s: %s (final cost: %d)", start, target, strings.Join(path, " → "), dist[target])
		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Shortest path from %s to %s is %v with total distance %d.", start, target, path, dist[target]),
			makeGraphState(currNodes, currEdges, finalWorklist, finalCap),
			map[string]any{"distances": cloneDist(dist), "path": path, "target": target, "totalDistance": dist[target]})
	} else if target != "" {
		for i := range currEdges {
			currEdges[i].Active = false
		}
		unreachCap := fmt.Sprintf("Target %s is unreachable from %s", target, start)
		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Target %s is unreachable from %s.", target, start),
			makeGraphState(currNodes, currEdges, finalWorklist, unreachCap),
			map[string]any{"distances": cloneDist(dist), "path": []string{}})
	} else {
		for i := range currEdges {
			currEdges[i].Active = false
		}
		allNodesCap := fmt.Sprintf("Dijkstra completed: computed shortest paths from %s to all reachable nodes", start)
		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Computed shortest distances from %s to all reachable nodes.", start),
			makeGraphState(currNodes, currEdges, finalWorklist, allNodesCap),
			map[string]any{"distances": cloneDist(dist)})
	}

	return b.Steps()
}
