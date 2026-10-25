package dijkstra

import (
	"container/heap"
	"math"
)

type WeightedEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Weight   int    `json:"weight"`
	Directed bool   `json:"directed,omitempty"`
}

type pqItem struct {
	node string
	dist int
}

type minHeap []pqItem

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(pqItem)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

type neighbor struct {
	to     string
	weight int
}

// ShortestPath computes the single-source shortest path using Dijkstra's algorithm.
func ShortestPath(nodes []string, edges []WeightedEdge, start, target string) (map[string]int, []string) {
	adj := make(map[string][]neighbor, len(nodes))
	for _, n := range nodes {
		adj[n] = nil
	}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], neighbor{to: e.To, weight: e.Weight})
		if !e.Directed {
			adj[e.To] = append(adj[e.To], neighbor{to: e.From, weight: e.Weight})
		}
	}

	dist := make(map[string]int, len(nodes))
	prev := make(map[string]string, len(nodes))
	for _, n := range nodes {
		dist[n] = math.MaxInt
	}
	if _, exists := dist[start]; !exists {
		return dist, nil
	}
	dist[start] = 0

	pq := &minHeap{}
	heap.Init(pq)
	heap.Push(pq, pqItem{node: start, dist: 0})

	visited := make(map[string]bool, len(nodes))

	// ponytail: runs until PQ empty so dist map contains all shortest paths; add early exit when graphs exceed 100k nodes
	for pq.Len() > 0 {
		curr := heap.Pop(pq).(pqItem)
		if visited[curr.node] {
			continue
		}
		visited[curr.node] = true

		for _, edge := range adj[curr.node] {
			if visited[edge.to] {
				continue
			}
			newDist := dist[curr.node] + edge.weight
			if newDist < dist[edge.to] {
				dist[edge.to] = newDist
				prev[edge.to] = curr.node
				heap.Push(pq, pqItem{node: edge.to, dist: newDist})
			}
		}
	}

	if target == "" || dist[target] == math.MaxInt {
		return dist, nil
	}

	var path []string
	curr := target
	for curr != "" {
		path = append([]string{curr}, path...)
		if curr == start {
			break
		}
		curr = prev[curr]
	}
	if len(path) == 0 || path[0] != start {
		return dist, nil
	}

	return dist, path
}
