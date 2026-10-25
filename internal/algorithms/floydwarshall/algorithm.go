package floydwarshall

import "math"

// WeightedEdge represents a weighted graph edge.
type WeightedEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Weight   int    `json:"weight"`
	Directed bool   `json:"directed,omitempty"`
}

// AllPairsShortestPath computes shortest path distances between all pairs of nodes using the Floyd-Warshall algorithm.
// Returns a V x V distance matrix and a boolean indicating whether a negative-weight cycle was detected.
// ponytail: O(V^3) all-pairs matrix computation; Johnson's algorithm preferred when graph is sparse (E << V^2).
func AllPairsShortestPath(nodes []string, edges []WeightedEdge) ([][]int, bool) {
	v := len(nodes)
	dist := make([][]int, v)
	for i := range dist {
		dist[i] = make([]int, v)
		for j := range dist[i] {
			if i == j {
				dist[i][j] = 0
			} else {
				dist[i][j] = math.MaxInt
			}
		}
	}

	nodeMap := make(map[string]int, v)
	for i, n := range nodes {
		nodeMap[n] = i
	}

	for _, e := range edges {
		u, uOk := nodeMap[e.From]
		to, toOk := nodeMap[e.To]
		if !uOk || !toOk {
			continue
		}
		if e.Weight < dist[u][to] {
			dist[u][to] = e.Weight
		}
		if !e.Directed && e.Weight < dist[to][u] {
			dist[to][u] = e.Weight
		}
	}

	for k := 0; k < v; k++ {
		for i := 0; i < v; i++ {
			for j := 0; j < v; j++ {
				if dist[i][k] != math.MaxInt && dist[k][j] != math.MaxInt {
					if dist[i][k]+dist[k][j] < dist[i][j] {
						dist[i][j] = dist[i][k] + dist[k][j]
					}
				}
			}
		}
	}

	for i := 0; i < v; i++ {
		if dist[i][i] < 0 {
			return dist, true
		}
	}

	return dist, false
}
