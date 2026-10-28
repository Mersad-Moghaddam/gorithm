package bellmanford

import (
	"bytes"
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
	Start string                `json:"start,omitempty"`
	Nodes []renderers.GraphNode `json:"nodes,omitempty"`
	Edges []renderers.GraphEdge `json:"edges,omitempty"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "bellmanford",
		Name:     "Bellman-Ford Shortest Path",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Computes single-source shortest paths in a directed weighted graph, correctly handling negative edge weights and detecting reachable negative cycles.",
			TimeComplexity:  "O(V · E)",
			SpaceComplexity: "O(V)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Dynamic Programming on Graphs",
				"Edge Relaxation Passes",
				"Negative Weight Handling",
				"Negative Cycle Detection",
			},
			HowItWorks: "Initializes the distance to the source node as 0 and all other vertices to infinity. Performs |V| - 1 edge relaxation passes; each pass tests whether routing through an edge shortens the tentative distance to its target. Distances stabilize after at most |V| - 1 passes unless a negative-weight cycle is reachable. A final pass over all edges detects any further decrease, confirming the presence of a negative cycle.",
		},
		DefaultInput: `{"start": "A"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func BellmanFord(nodes []string, edges []DirectedEdge, start string) (map[string]int, bool) {"},
		{ID: "L2", Number: 2, Code: "    dist := initDistances(nodes, start) // start = 0, others = ∞"},
		{ID: "L3", Number: 3, Code: "    for i := 1; i < len(nodes); i++ {"},
		{ID: "L4", Number: 4, Code: "        relaxed := false"},
		{ID: "L5", Number: 5, Code: "        for _, edge := range edges {"},
		{ID: "L6", Number: 6, Code: "            if dist[edge.From] != ∞ && dist[edge.From] + edge.Weight < dist[edge.To] {"},
		{ID: "L7", Number: 7, Code: "                dist[edge.To] = dist[edge.From] + edge.Weight"},
		{ID: "L8", Number: 8, Code: "                relaxed = true"},
		{ID: "L9", Number: 9, Code: "            }"},
		{ID: "L10", Number: 10, Code: "        }"},
		{ID: "L11", Number: 11, Code: "        if !relaxed { break }"},
		{ID: "L12", Number: 12, Code: "    }"},
		{ID: "L13", Number: 13, Code: "    for _, edge := range edges {"},
		{ID: "L14", Number: 14, Code: "        if dist[edge.From] != ∞ && dist[edge.From] + edge.Weight < dist[edge.To] {"},
		{ID: "L15", Number: 15, Code: "            return dist, true // negative cycle detected"},
		{ID: "L16", Number: 16, Code: "        }"},
		{ID: "L17", Number: 17, Code: "    }"},
		{ID: "L18", Number: 18, Code: "    return dist, false"},
		{ID: "L19", Number: 19, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 120, Y: 225, State: "unvisited"},
		{ID: "B", Label: "B", X: 280, Y: 120, State: "unvisited"},
		{ID: "C", Label: "C", X: 280, Y: 330, State: "unvisited"},
		{ID: "D", Label: "D", X: 480, Y: 120, State: "unvisited"},
		{ID: "E", Label: "E", X: 640, Y: 225, State: "unvisited"},
	}

	edges := []renderers.GraphEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "C", Weight: 2},
		{From: "B", To: "D", Weight: 2},
		{From: "C", To: "D", Weight: 5},
		{From: "C", To: "E", Weight: 4},
		{From: "D", To: "E", Weight: -3},
	}

	start := "A"

	if len(input) > 0 {
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid graph JSON: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("graph input must be an object")
		}
		for _, key := range []string{"nodes", "edges", "start"} {
			if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s cannot be null", key)
			}
		}
		for _, key := range []string{"nodes", "edges"} {
			if raw, ok := fields[key]; ok {
				var records []map[string]json.RawMessage
				if err := json.Unmarshal(raw, &records); err != nil {
					return nil, fmt.Errorf("%s must contain objects", key)
				}
				for _, record := range records {
					if record == nil {
						return nil, fmt.Errorf("%s cannot contain null", key)
					}
					for field, value := range record {
						if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
							return nil, fmt.Errorf("%s cannot be null", field)
						}
					}
				}
			}
		}
		{
			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			}
			if _, ok := fields["nodes"]; ok && len(custom.Nodes) == 0 {
				return nil, fmt.Errorf("nodes must be nonempty")
			}
			if _, ok := fields["edges"]; ok {
				edges = custom.Edges
			}
			if _, ok := fields["start"]; ok {
				start = custom.Start
			}
		}
	}

	if len(nodes) > 32 || len(edges) > 128 {
		return nil, fmt.Errorf("graph limit: 32 nodes, 128 edges")
	}
	ids := map[string]bool{}
	for _, n := range nodes {
		if n.ID == "" || ids[n.ID] || utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
			return nil, fmt.Errorf("node IDs must be unique and nonempty; IDs and labels limited to 64 code points")
		}
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
			return nil, fmt.Errorf("coordinates must be finite and within ±1000000")
		}
		ids[n.ID] = true
	}
	if !ids[start] {
		return nil, fmt.Errorf("start node does not exist")
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] {
			return nil, fmt.Errorf("edge endpoint does not exist")
		}
		if e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("weight must be within ±10000")
		}
	}
	return Trace(nodes, edges, start), nil
}

func init() {
	catalog.Register(Lesson{})
}
