package floydwarshall

import (
	"math"
	"math/big"
)

// WeightedEdge represents a weighted graph edge.
type WeightedEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Weight   int    `json:"weight"`
	Directed bool   `json:"directed,omitempty"`
}

// AllPairsShortestPath computes shortest path distances between all pairs of nodes using the Floyd-Warshall algorithm.
// Returns a V x V distance matrix and a boolean indicating whether a negative-weight cycle was detected.
// Unreachable entries use math.MaxInt. Finite results outside the representable
// range [math.MinInt, math.MaxInt-1] are clamped only after exact computation.
// ponytail: O(V^3) all-pairs matrix computation; Johnson's algorithm preferred when graph is sparse (E << V^2).
func AllPairsShortestPath(nodes []string, edges []WeightedEdge) ([][]int, bool) {
	v := len(nodes)
	dist := make([][]*big.Int, v)
	for i := range dist {
		dist[i] = make([]*big.Int, v)
		dist[i][i] = new(big.Int)
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
		weight := big.NewInt(int64(e.Weight))
		if dist[u][to] == nil || weight.Cmp(dist[u][to]) < 0 {
			dist[u][to] = new(big.Int).Set(weight)
		}
		if !e.Directed && (dist[to][u] == nil || weight.Cmp(dist[to][u]) < 0) {
			dist[to][u] = new(big.Int).Set(weight)
		}
	}

	for k := 0; k < v; k++ {
		for i := 0; i < v; i++ {
			for j := 0; j < v; j++ {
				if dist[i][k] != nil && dist[k][j] != nil {
					candidate := new(big.Int).Add(dist[i][k], dist[k][j])
					if dist[i][j] == nil || candidate.Cmp(dist[i][j]) < 0 {
						dist[i][j] = candidate
					}
				}
			}
		}
	}

	result := make([][]int, v)
	negativeCycle := false
	minValue, maxValue := big.NewInt(int64(math.MinInt)), big.NewInt(int64(math.MaxInt-1))
	for i := range dist {
		result[i] = make([]int, v)
		negativeCycle = negativeCycle || dist[i][i].Sign() < 0
		for j, value := range dist[i] {
			switch {
			case value == nil:
				result[i][j] = math.MaxInt
			case value.Cmp(minValue) < 0:
				result[i][j] = math.MinInt
			case value.Cmp(maxValue) > 0:
				result[i][j] = math.MaxInt - 1
			default:
				result[i][j] = int(value.Int64())
			}
		}
	}
	return result, negativeCycle
}
