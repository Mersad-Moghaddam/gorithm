package prim

import (
	"container/heap"
)

type WeightedEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"`
}

type edgeItem struct {
	from   string
	to     string
	weight int
}

type minEdgeHeap []edgeItem

func (h minEdgeHeap) Len() int           { return len(h) }
func (h minEdgeHeap) Less(i, j int) bool { return h[i].weight < h[j].weight }
func (h minEdgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minEdgeHeap) Push(x any)        { *h = append(*h, x.(edgeItem)) }
func (h *minEdgeHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// ponytail: min-heap on edges O(E log E); upgrade to indexed priority queue with decrease-key O(E log V) if node degree is high.
func MST(nodes []string, edges []WeightedEdge, start string) ([]WeightedEdge, int) {
	if len(nodes) <= 1 {
		return nil, 0
	}

	nodeSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		nodeSet[n] = true
	}

	if start == "" || !nodeSet[start] {
		start = nodes[0]
	}

	adj := make(map[string][]edgeItem, len(nodes))
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], edgeItem{from: e.From, to: e.To, weight: e.Weight})
		adj[e.To] = append(adj[e.To], edgeItem{from: e.To, to: e.From, weight: e.Weight})
	}

	visited := make(map[string]bool, len(nodes))
	visited[start] = true

	pq := &minEdgeHeap{}
	heap.Init(pq)
	for _, e := range adj[start] {
		heap.Push(pq, e)
	}

	mst := make([]WeightedEdge, 0, len(nodes)-1)
	totalWeight := 0

	for pq.Len() > 0 && len(mst) < len(nodes)-1 {
		item := heap.Pop(pq).(edgeItem)
		if visited[item.to] {
			continue
		}

		visited[item.to] = true
		mst = append(mst, WeightedEdge{
			From:   item.from,
			To:     item.to,
			Weight: item.weight,
		})
		totalWeight += item.weight

		for _, next := range adj[item.to] {
			if !visited[next.to] {
				heap.Push(pq, next)
			}
		}
	}

	return mst, totalWeight
}
