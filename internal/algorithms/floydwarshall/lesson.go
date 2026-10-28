package floydwarshall

import (
	"encoding/json"
	"fmt"
	"math"
	"unicode/utf8"

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
		{ID: "L1", Number: 1, Code: "func FloydWarshallLesson(nodes []string, edges []WeightedEdge) ([][]int64, bool) {"},
		{ID: "L2", Number: 2, Code: "    dist := initDistanceMatrix64(nodes, edges) // int64 distances"},
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
	nodes := []string{"A", "B", "C", "D"}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 3, Directed: true},
		{From: "A", To: "C", Weight: 8, Directed: true},
		{From: "B", To: "C", Weight: 2, Directed: true},
		{From: "B", To: "D", Weight: 5, Directed: true},
		{From: "C", To: "D", Weight: 1, Directed: true},
		{From: "D", To: "A", Weight: 2, Directed: true},
	}

	if len(input) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, fmt.Errorf("invalid graph JSON: %w", err)
		}
		if fields == nil {
			return nil, fmt.Errorf("graph input must be an object")
		}
		for _, key := range []string{"nodes", "edges", "graphNodes", "graphEdges"} {
			if raw, ok := fields[key]; ok {
				var values []json.RawMessage
				if err := json.Unmarshal(raw, &values); err != nil || values == nil {
					return nil, fmt.Errorf("%s must be an array", key)
				}
				for _, value := range values {
					if string(value) == "null" {
						return nil, fmt.Errorf("%s cannot contain null", key)
					}
					if key != "nodes" {
						var object map[string]json.RawMessage
						if err := json.Unmarshal(value, &object); err != nil || object == nil {
							return nil, fmt.Errorf("%s entries must be objects", key)
						}
						for _, field := range []string{"id", "label", "x", "y", "from", "to", "weight", "directed", "active", "state"} {
							if raw, exists := object[field]; exists && string(raw) == "null" {
								return nil, fmt.Errorf("%s cannot be null", field)
							}
						}
					}
				}
			}
		}
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid graph fields: %w", err)
		} else {
			if len(custom.Nodes) > 16 || len(custom.GraphNodes) > 16 || len(custom.Edges) > 128 || len(custom.GraphEdges) > 128 {
				return nil, fmt.Errorf("graph exceeds 16 nodes or 128 edges")
			}
			graphIDs := make(map[string]bool, len(custom.GraphNodes))
			for _, n := range custom.GraphNodes {
				if graphIDs[n.ID] {
					return nil, fmt.Errorf("duplicate graph node ID")
				}
				graphIDs[n.ID] = true
				if utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 || math.IsNaN(n.X) || math.IsInf(n.X, 0) || math.IsNaN(n.Y) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
					return nil, fmt.Errorf("invalid graph node label or coordinate")
				}
			}
			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			} else if len(custom.GraphNodes) > 0 {
				nodes = make([]string, len(custom.GraphNodes))
				for i, gn := range custom.GraphNodes {
					nodes[i] = gn.ID
				}
			}

			if _, ok := fields["edges"]; ok {
				edges = custom.Edges
			} else if _, ok := fields["graphEdges"]; ok {
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
		}
	}

	if len(nodes) > 16 {
		return nil, fmt.Errorf("Floyd–Warshall accepts at most 16 nodes")
	}
	if len(edges) > 128 {
		return nil, fmt.Errorf("graph accepts at most 128 edges")
	}
	ids := make(map[string]bool, len(nodes))
	for _, id := range nodes {
		if utf8.RuneCountInString(id) > 64 || ids[id] {
			return nil, fmt.Errorf("node IDs must be unique and at most 64 Unicode code points")
		}
		ids[id] = true
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] {
			return nil, fmt.Errorf("edge endpoints must exist")
		}
		if e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("edge weights must be within ±10000")
		}
	}
	return Trace(nodes, edges), nil
}

func init() {
	catalog.Register(Lesson{})
}
