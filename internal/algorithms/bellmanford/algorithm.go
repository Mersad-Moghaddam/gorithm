package bellmanford

import "math"

// DirectedEdge represents a directed weighted graph edge.
type DirectedEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"`
}

// ShortestPath computes single-source shortest paths using the Bellman-Ford algorithm.
// Returns map of shortest distances and boolean indicating presence of negative cycle.
func ShortestPath(nodes []string, edges []DirectedEdge, start string) (map[string]int, bool) {
	dist := make(map[string]int, len(nodes))
	for _, n := range nodes {
		dist[n] = math.MaxInt
	}
	if _, ok := dist[start]; !ok {
		return dist, false
	}
	dist[start] = 0

	// ponytail: early exit if no edge relaxes in a pass; add SPFA queue if graph is sparse and V > 10k
	for i := 0; i < len(nodes)-1; i++ {
		relaxed := false
		for _, e := range edges {
			if dist[e.From] != math.MaxInt && dist[e.From]+e.Weight < dist[e.To] {
				dist[e.To] = dist[e.From] + e.Weight
				relaxed = true
			}
		}
		if !relaxed {
			break
		}
	}

	for _, e := range edges {
		if dist[e.From] != math.MaxInt && dist[e.From]+e.Weight < dist[e.To] {
			return dist, true
		}
	}

	return dist, false
}
