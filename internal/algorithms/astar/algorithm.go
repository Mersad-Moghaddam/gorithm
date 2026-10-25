package astar

import (
	"container/heap"
	"math"
)

type AStarNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label,omitempty"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type AStarEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Weight     int    `json:"weight"`
	Undirected bool   `json:"undirected,omitempty"`
}

func EuclideanDistance(a, b AStarNode) int {
	return int(math.Round(math.Hypot(a.X-b.X, a.Y-b.Y)))
}

func ManhattanDistance(a, b AStarNode) int {
	return int(math.Round(math.Abs(a.X-b.X) + math.Abs(a.Y-b.Y)))
}

type pqItem struct {
	node   string
	fScore int
	hScore int
	index  int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].fScore == pq[j].fScore {
		return pq[i].hScore < pq[j].hScore
	}
	return pq[i].fScore < pq[j].fScore
}
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	it := x.(*pqItem)
	it.index = n
	*pq = append(*pq, it)
}
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	it := old[n-1]
	old[n-1] = nil
	it.index = -1
	*pq = old[0 : n-1]
	return it
}

// ponytail: min-heap A* with adjacency list; bidirectional A* or contraction hierarchies faster on large graphs.
func FindPath(nodes []AStarNode, edges []AStarEdge, start, goal string) (int, []string) {
	return FindPathWithHeuristic(nodes, edges, start, goal, EuclideanDistance)
}

func FindPathWithHeuristic(nodes []AStarNode, edges []AStarEdge, start, goal string, h func(a, b AStarNode) int) (int, []string) {
	nodeMap := make(map[string]AStarNode, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	startNode, hasStart := nodeMap[start]
	goalNode, hasGoal := nodeMap[goal]
	if !hasStart || !hasGoal {
		return -1, nil
	}

	if start == goal {
		return 0, []string{start}
	}

	type neighbor struct {
		to     string
		weight int
	}
	adj := make(map[string][]neighbor)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], neighbor{to: e.To, weight: e.Weight})
		if e.Undirected {
			adj[e.To] = append(adj[e.To], neighbor{to: e.From, weight: e.Weight})
		}
	}

	gScore := make(map[string]int)
	cameFrom := make(map[string]string)
	visited := make(map[string]bool)

	pq := &priorityQueue{}
	heap.Init(pq)
	itemMap := make(map[string]*pqItem)

	hStart := h(startNode, goalNode)
	gScore[start] = 0

	startItem := &pqItem{node: start, fScore: hStart, hScore: hStart}
	heap.Push(pq, startItem)
	itemMap[start] = startItem

	for pq.Len() > 0 {
		curr := heap.Pop(pq).(*pqItem)
		delete(itemMap, curr.node)
		currID := curr.node

		if currID == goal {
			path := []string{currID}
			for p, ok := cameFrom[currID]; ok; p, ok = cameFrom[p] {
				path = append([]string{p}, path...)
			}
			return gScore[goal], path
		}

		visited[currID] = true

		for _, edge := range adj[currID] {
			nbrID := edge.to
			if visited[nbrID] {
				continue
			}

			tentativeG := gScore[currID] + edge.weight
			prevG, exists := gScore[nbrID]
			if !exists || tentativeG < prevG {
				cameFrom[nbrID] = currID
				gScore[nbrID] = tentativeG
				nbrNode := nodeMap[nbrID]
				hVal := h(nbrNode, goalNode)
				fVal := tentativeG + hVal

				if it, ok := itemMap[nbrID]; ok {
					it.fScore = fVal
					it.hScore = hVal
					heap.Fix(pq, it.index)
				} else {
					it := &pqItem{node: nbrID, fScore: fVal, hScore: hVal}
					heap.Push(pq, it)
					itemMap[nbrID] = it
				}
			}
		}
	}

	return -1, nil
}
