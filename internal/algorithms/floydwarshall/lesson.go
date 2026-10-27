package floydwarshall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Lesson struct{}

type LessonInput struct {
	Nodes      []string              `json:"nodes,omitempty"`
	Edges      []WeightedEdge        `json:"edges,omitempty"`
	GraphNodes []renderers.GraphNode `json:"graphNodes,omitempty"`
	GraphEdges []renderers.GraphEdge `json:"graphEdges,omitempty"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "floydwarshall",
		Name:     "Floyd-Warshall All-Pairs Path",
		Category: catalog.CategoryGraphs,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Computes shortest paths between all pairs of vertices in a weighted graph using dynamic programming across intermediate vertices, correctly handling negative edge weights and detecting negative cycles.",
			TimeComplexity:  "O(V³)",
			SpaceComplexity: "O(V²)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"All-Pairs Distance",
				"Intermediate Vertex Optimization",
				"Adjacency Matrix DP",
			},
			HowItWorks: "Maintains a |V| x |V| distance matrix initialized with direct edge weights and 0 on the diagonal. It iteratively considers each vertex k as an intermediate candidate. For every pair (i, j), if routing through k offers a shorter path (dist[i][k] + dist[k][j] < dist[i][j]), the matrix entry is relaxed. After considering all intermediate vertices, dist[i][j] holds the all-pairs shortest path distance, with negative values along the diagonal indicating negative cycles.",
		},
		DefaultInput: `{"nodes": ["A", "B", "C", "D"], "edges": [{"from": "A", "to": "B", "weight": 3, "directed": true}, {"from": "A", "to": "C", "weight": 8, "directed": true}, {"from": "B", "to": "C", "weight": 2, "directed": true}, {"from": "B", "to": "D", "weight": 5, "directed": true}, {"from": "C", "to": "D", "weight": 1, "directed": true}, {"from": "D", "to": "A", "weight": 2, "directed": true}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func FloydWarshall(nodes []string, edges []WeightedEdge) ([][]int, bool) {"},
		{ID: "L2", Number: 2, Code: "    dist := initDistanceMatrix(nodes, edges)"},
		{ID: "L3", Number: 3, Code: "    for k := 0; k < len(nodes); k++ {"},
		{ID: "L4", Number: 4, Code: "        for i := 0; i < len(nodes); i++ {"},
		{ID: "L5", Number: 5, Code: "            for j := 0; j < len(nodes); j++ {"},
		{ID: "L6", Number: 6, Code: "                if dist[i][k] != ∞ && dist[k][j] != ∞ {"},
		{ID: "L7", Number: 7, Code: "                    if dist[i][k] + dist[k][j] < dist[i][j] {"},
		{ID: "L8", Number: 8, Code: "                        dist[i][j] = dist[i][k] + dist[k][j]"},
		{ID: "L9", Number: 9, Code: "                    }"},
		{ID: "L10", Number: 10, Code: "                }"},
		{ID: "L11", Number: 11, Code: "            }"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "    }"},
		{ID: "L14", Number: 14, Code: "    for i := 0; i < len(nodes); i++ {"},
		{ID: "L15", Number: 15, Code: "        if dist[i][i] < 0 {"},
		{ID: "L16", Number: 16, Code: "            return dist, true // negative cycle detected"},
		{ID: "L17", Number: 17, Code: "        }"},
		{ID: "L18", Number: 18, Code: "    }"},
		{ID: "L19", Number: 19, Code: "    return dist, false"},
		{ID: "L20", Number: 20, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	defaultNodes := []string{"A", "B", "C", "D"}
	defaultEdges := []WeightedEdge{
		{From: "A", To: "B", Weight: 3, Directed: true},
		{From: "A", To: "C", Weight: 8, Directed: true},
		{From: "B", To: "C", Weight: 2, Directed: true},
		{From: "B", To: "D", Weight: 5, Directed: true},
		{From: "C", To: "D", Weight: 1, Directed: true},
		{From: "D", To: "A", Weight: 2, Directed: true},
	}

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(defaultNodes, defaultEdges), nil
	}
	if string(trimmed) == "null" {
		return nil, fmt.Errorf("input cannot be null")
	}

	var custom LessonInput
	if err := json.Unmarshal(trimmed, &custom); err != nil {
		return nil, fmt.Errorf("malformed JSON: %w", err)
	}

	var nodes []string
	if len(custom.Nodes) > 0 {
		nodes = custom.Nodes
	} else if len(custom.GraphNodes) > 0 {
		nodes = make([]string, len(custom.GraphNodes))
		for i, gn := range custom.GraphNodes {
			nodes[i] = gn.ID
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodes cannot be empty")
	}

	for _, n := range nodes {
		if strings.TrimSpace(n) == "" {
			return nil, fmt.Errorf("node identifier cannot be empty")
		}
	}

	var edges []WeightedEdge
	if len(custom.Edges) > 0 {
		edges = custom.Edges
	} else if len(custom.GraphEdges) > 0 {
		edges = make([]WeightedEdge, len(custom.GraphEdges))
		for i, ge := range custom.GraphEdges {
			edges[i] = WeightedEdge{
				From:     ge.From,
				To:       ge.To,
				Weight:   ge.Weight,
				Directed: true,
			}
		}
	}

	return Trace(nodes, edges), nil
}

func init() {
	catalog.Register(Lesson{})
}
