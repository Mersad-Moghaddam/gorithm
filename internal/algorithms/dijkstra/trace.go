package dijkstra

import (
	"container/heap"
	"fmt"
	"math"

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

func findEdgeIndex(edges []renderers.GraphEdge, u, v string) int {
	for i, e := range edges {
		if (e.From == u && e.To == v) || (e.From == v && e.To == u) {
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
		b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Start node %q not found in graph.", start), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})
		b.AddStep("L17", trace.ActionDone, "Search terminated: invalid start node.", renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})
		return b.Steps()
	}

	dist[start] = 0

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Dijkstra's Shortest Path from start node %s to target %s.", start, target), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"current": "", "target": target, "distances": cloneDist(dist)})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Set initial distance to start node %s = 0, all other nodes = ∞.", start), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"current": start, "target": target, "distances": cloneDist(dist)})

	currNodes[startIdx].State = "queued"
	b.IncOperations()
	b.AddStep("L4", trace.ActionPush, fmt.Sprintf("Push start node %s (distance 0) into priority queue.", start), renderers.GraphState{
		Nodes: cloneNodes(currNodes),
		Edges: cloneEdges(currEdges),
	}, map[string]any{"current": start, "target": target, "distances": cloneDist(dist)})

	pq := &minHeap{}
	heap.Init(pq)
	heap.Push(pq, pqItem{node: start, dist: 0})

	visited := make(map[string]bool, len(nodes))

	for pq.Len() > 0 {
		b.AddStep("L5", trace.ActionHighlight, "Priority queue is non-empty: selecting next candidate vertex.", renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})

		item := heap.Pop(pq).(pqItem)
		u := item.node
		d := item.dist
		b.IncOperations()

		if visited[u] {
			b.AddStep("L6", trace.ActionPop, fmt.Sprintf("Extracted vertex %s (distance %d). Node already visited, skipping outdated queue entry.", u, d), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"current": u, "distances": cloneDist(dist)})
			continue
		}

		uIdx := findNodeIndex(currNodes, u)
		if uIdx >= 0 {
			currNodes[uIdx].State = "visiting"
		}

		b.AddStep("L6", trace.ActionPop, fmt.Sprintf("Extracted vertex %s with minimum tentative distance %d from priority queue.", u, d), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"current": u, "distances": cloneDist(dist)})

		b.IncComparisons()
		b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("Check if current vertex %s matches target %s.", u, target), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"current": u, "target": target, "distances": cloneDist(dist)})

		if target != "" && u == target {
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Target vertex %s reached with shortest distance %d!", target, dist[target]), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"current": u, "target": target, "distances": cloneDist(dist)})
			break
		}

		// Inspect all incident edges
		for eIdx := range currEdges {
			edge := currEdges[eIdx]
			var v string
			if edge.From == u {
				v = edge.To
			} else if edge.To == u {
				v = edge.From
			} else {
				continue
			}

			if visited[v] {
				continue
			}

			currEdges[eIdx].Active = true
			b.AddStep("L8", trace.ActionVisit, fmt.Sprintf("Inspecting neighbor %s via edge (%s-%s, weight %d).", v, edge.From, edge.To, edge.Weight), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"current": u, "neighbor": v, "edgeWeight": edge.Weight, "distances": cloneDist(dist)})

			alt := dist[u] + edge.Weight
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Calculate alternative distance to %s: dist[%s] (%d) + weight (%d) = %d.", v, u, dist[u], edge.Weight, alt), renderers.GraphState{
				Nodes: cloneNodes(currNodes),
				Edges: cloneEdges(currEdges),
			}, map[string]any{"current": u, "neighbor": v, "alt": alt, "distances": cloneDist(dist)})

			b.IncComparisons()
			if alt < dist[v] {
				oldDistStr := "∞"
				if dist[v] != math.MaxInt {
					oldDistStr = fmt.Sprintf("%d", dist[v])
				}
				b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Relaxation check: alt (%d) < dist[%s] (%s): condition holds.", alt, v, oldDistStr), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"current": u, "neighbor": v, "alt": alt, "relaxed": true, "distances": cloneDist(dist)})

				dist[v] = alt
				b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Update shortest distance to %s = %d.", v, alt), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"current": u, "neighbor": v, "distances": cloneDist(dist)})

				prev[v] = u
				b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Set predecessor of %s to %s.", v, u), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"current": u, "neighbor": v, "predecessor": u, "distances": cloneDist(dist)})

				heap.Push(pq, pqItem{node: v, dist: alt})
				b.IncOperations()

				vIdx := findNodeIndex(currNodes, v)
				if vIdx >= 0 && currNodes[vIdx].State != "visiting" && currNodes[vIdx].State != "visited" {
					currNodes[vIdx].State = "queued"
				}

				b.AddStep("L13", trace.ActionPush, fmt.Sprintf("Push %s with updated distance %d into priority queue.", v, alt), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"current": u, "neighbor": v, "distances": cloneDist(dist)})
			} else {
				b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Relaxation check: alt (%d) >= dist[%s] (%d): existing distance is shorter or equal.", alt, v, dist[v]), renderers.GraphState{
					Nodes: cloneNodes(currNodes),
					Edges: cloneEdges(currEdges),
				}, map[string]any{"current": u, "neighbor": v, "alt": alt, "relaxed": false, "distances": cloneDist(dist)})
			}

			currEdges[eIdx].Active = false
		}

		visited[u] = true
		if uIdx >= 0 {
			currNodes[uIdx].State = "visited"
		}
		b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Finished exploring all neighbors for vertex %s. Node marked as visited.", u), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"current": u, "distances": cloneDist(dist)})
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

		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Shortest path from %s to %s is %v with total distance %d.", start, target, path, dist[target]), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist), "path": path, "target": target, "totalDistance": dist[target]})
	} else if target != "" {
		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Target %s is unreachable from %s.", target, start), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist), "path": []string{}})
	} else {
		b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Dijkstra completed! Computed shortest distances from %s to all reachable nodes.", start), renderers.GraphState{
			Nodes: cloneNodes(currNodes),
			Edges: cloneEdges(currEdges),
		}, map[string]any{"distances": cloneDist(dist)})
	}

	return b.Steps()
}
