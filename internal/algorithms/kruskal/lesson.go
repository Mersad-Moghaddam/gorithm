package kruskal

import (
	"encoding/json"
	"math"

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
		{ID: "L8", Number: 8, Code: "        }"},
		{ID: "L9", Number: 9, Code: "    }"},
		{ID: "L10", Number: 10, Code: "    return mst"},
		{ID: "L11", Number: 11, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges := defaultGraph()
	if len(input) > 0 {
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err == nil && len(custom.Nodes) > 0 && len(custom.Edges) > 0 {
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
