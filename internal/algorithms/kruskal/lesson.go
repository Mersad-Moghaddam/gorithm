package kruskal

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
	Nodes []renderers.GraphNode `json:"nodes"`
	Edges []WeightedEdge        `json:"edges"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "kruskal",
		Name:     "Kruskal's Minimum Spanning Tree",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Finds a minimum spanning tree for a connected weighted undirected graph by greedily adding the lowest-weight edges that do not form a cycle.",
			TimeComplexity:  "O(E log E)",
			SpaceComplexity: "O(V)",
			KeyConcepts:     []string{"Greedy", "Edge Sorting", "Disjoint Set Union", "Cycle Detection"},
			HowItWorks:      "Sorts all edges in ascending order of weight. Iterates through the sorted edges, using Disjoint Set Union (DSU / Union-Find) with path compression to check if endpoints belong to different connected components. If no cycle is formed, the edge is added to the spanning tree and the components are merged.",
		},
		DefaultInput: `{"nodes":[{"id":"A","label":"A","x":160,"y":140},{"id":"B","label":"B","x":400,"y":100},{"id":"C","label":"C","x":640,"y":140},{"id":"D","label":"D","x":160,"y":320},{"id":"E","label":"E","x":400,"y":350},{"id":"F","label":"F","x":640,"y":320}],"edges":[{"from":"A","to":"B","weight":4},{"from":"A","to":"D","weight":2},{"from":"B","to":"C","weight":5},{"from":"B","to":"D","weight":1},{"from":"B","to":"E","weight":3},{"from":"C","to":"E","weight":6},{"from":"C","to":"F","weight":2},{"from":"D","to":"E","weight":7},{"from":"E","to":"F","weight":4}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func KruskalMST(nodes []string, edges []WeightedEdge) []WeightedEdge {"},
		{ID: "L2", Number: 2, Code: "    sort.Slice(edges, func(i, j int) bool { return edges[i].Weight < edges[j].Weight })"},
		{ID: "L3", Number: 3, Code: "    dsu := NewDSU(nodes); mst := []WeightedEdge{}"},
		{ID: "L4", Number: 4, Code: "    for _, edge := range edges {"},
		{ID: "L5", Number: 5, Code: "        if dsu.Find(edge.From) != dsu.Find(edge.To) {"},
		{ID: "L6", Number: 6, Code: "            dsu.Union(edge.From, edge.To)"},
		{ID: "L7", Number: 7, Code: "            mst = append(mst, edge)"},
		{ID: "L8", Number: 8, Code: "            if len(mst) == len(nodes)-1 { break }"},
		{ID: "L12", Number: 9, Code: "        }"},
		{ID: "L9", Number: 10, Code: "    }"},
		{ID: "L10", Number: 11, Code: "    return mst"},
		{ID: "L11", Number: 12, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges := defaultGraph()
	if len(input) > 0 {
		var custom LessonInput
		if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
			return nil, fmt.Errorf("graph must be an object")
		}
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid graph: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		for _, name := range []string{"nodes", "edges"} {
			if raw, present := fields[name]; present && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s must be an array, not null", name)
			}
			var rows []map[string]json.RawMessage
			if err := json.Unmarshal(fields[name], &rows); len(fields[name]) > 0 && err != nil {
				return nil, fmt.Errorf("invalid %s: %w", name, err)
			}
			for _, row := range rows {
				if row == nil {
					return nil, fmt.Errorf("%s entries must be objects", name)
				}
				for key, raw := range row {
					if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
						return nil, fmt.Errorf("%s.%s must not be null", name, key)
					}
				}
			}
		}
		if custom.Nodes != nil || custom.Edges != nil {
			if len(custom.Nodes) == 0 || len(custom.Nodes) > 32 || len(custom.Edges) > 128 {
				return nil, fmt.Errorf("graph requires 1–32 nodes and at most 128 edges")
			}
			seen := map[string]bool{}
			for _, n := range custom.Nodes {
				if n.ID == "" || seen[n.ID] || utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
					return nil, fmt.Errorf("node IDs must be unique and nonempty; IDs and labels at most 64 characters")
				}
				if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
					return nil, fmt.Errorf("coordinates must be finite within ±1000000")
				}
				seen[n.ID] = true
			}
			for _, e := range custom.Edges {
				if !seen[e.From] || !seen[e.To] || e.Weight < -10000 || e.Weight > 10000 {
					return nil, fmt.Errorf("edges require known endpoints and weights within ±10000")
				}
			}
			nodes = custom.Nodes
			edges = custom.Edges
			for i := range nodes {
				if nodes[i].Label == "" {
					nodes[i].Label = nodes[i].ID
				}
				if nodes[i].State == "" {
					nodes[i].State = "unvisited"
				}
				if nodes[i].X == 0 && nodes[i].Y == 0 {
					angle := 2 * math.Pi * float64(i) / float64(len(nodes))
					nodes[i].X = 400 + 180*math.Cos(angle)
					nodes[i].Y = 225 + 130*math.Sin(angle)
				}
			}
		}
	}
	return Trace(nodes, edges), nil
}

func defaultGraph() ([]renderers.GraphNode, []WeightedEdge) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 160, Y: 140, State: "unvisited"},
		{ID: "B", Label: "B", X: 400, Y: 100, State: "unvisited"},
		{ID: "C", Label: "C", X: 640, Y: 140, State: "unvisited"},
		{ID: "D", Label: "D", X: 160, Y: 320, State: "unvisited"},
		{ID: "E", Label: "E", X: 400, Y: 350, State: "unvisited"},
		{ID: "F", Label: "F", X: 640, Y: 320, State: "unvisited"},
	}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "D", Weight: 2},
		{From: "B", To: "C", Weight: 5},
		{From: "B", To: "D", Weight: 1},
		{From: "B", To: "E", Weight: 3},
		{From: "C", To: "E", Weight: 6},
		{From: "C", To: "F", Weight: 2},
		{From: "D", To: "E", Weight: 7},
		{From: "E", To: "F", Weight: 4},
	}
	return nodes, edges
}

func init() {
	catalog.Register(Lesson{})
}
