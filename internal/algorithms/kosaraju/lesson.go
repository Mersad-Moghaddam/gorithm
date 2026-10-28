package kosaraju

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
	Edges []renderers.GraphEdge `json:"edges"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "kosaraju",
		Name:     "Kosaraju SCC (Components)",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Finds all strongly connected components in a directed graph using a two-pass depth-first search with graph transposition.",
			TimeComplexity:  "O(V + E)",
			SpaceComplexity: "O(V)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Strong Connectivity",
				"Forward DFS Finish Order",
				"Graph Transpose / Reversal",
				"Backward DFS Decomposition",
			},
			HowItWorks: "First, a forward DFS traversal computes the finishing times of all vertices, pushing finished vertices onto a stack. Next, all directed edges are reversed to construct the transposed graph. Finally, a second DFS pass iterates through vertices in decreasing finish order; each DFS tree discovered in the transposed graph forms an independent strongly connected component.",
		},
		DefaultInput: `{"nodes":[{"id":"A","label":"A","x":140,"y":150},{"id":"B","label":"B","x":260,"y":150},{"id":"C","label":"C","x":200,"y":300},{"id":"D","label":"D","x":380,"y":220},{"id":"E","label":"E","x":480,"y":300},{"id":"F","label":"F","x":600,"y":150},{"id":"G","label":"G","x":700,"y":260}],"edges":[{"from":"A","to":"B"},{"from":"B","to":"C"},{"from":"C","to":"A"},{"from":"C","to":"D"},{"from":"D","to":"E"},{"from":"E","to":"D"},{"from":"E","to":"F"},{"from":"F","to":"G"},{"from":"G","to":"F"}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func FindSCCs(nodes []string, edges [][2]string) [][]string {"},
		{ID: "L2", Number: 2, Code: "    adj, revAdj := buildAdjacency(nodes, edges)"},
		{ID: "L3", Number: 3, Code: "    visited := make(map[string]bool); order := []string{}"},
		{ID: "L4", Number: 4, Code: "    for _, u := range nodes {"},
		{ID: "L5", Number: 5, Code: "        if !visited[u] { dfsForward(u) } // Record finish times"},
		{ID: "L6", Number: 6, Code: "    }"},
		{ID: "L7", Number: 7, Code: "    visitedRev := make(map[string]bool); sccs := [][]string{}"},
		{ID: "L8", Number: 8, Code: "    for i := len(order) - 1; i >= 0; i-- {"},
		{ID: "L9", Number: 9, Code: "        u := order[i]"},
		{ID: "L10", Number: 10, Code: "        if !visitedRev[u] {"},
		{ID: "L11", Number: 11, Code: "            var comp []string"},
		{ID: "L12", Number: 12, Code: "            dfsReverse(u, &comp) // Explore transposed graph"},
		{ID: "L13", Number: 13, Code: "            sccs = append(sccs, comp)"},
		{ID: "L14", Number: 14, Code: "        }"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    return sccs"},
		{ID: "L17", Number: 17, Code: "}"},
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
			return nil, fmt.Errorf("graph must be an object: %w", err)
		}
		for _, key := range []string{"nodes", "edges"} {
			if value, present := fields[key]; present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return nil, fmt.Errorf("%s must be an array, not null", key)
			}
		}
		if custom.Nodes != nil && len(custom.Nodes) == 0 {
			return nil, fmt.Errorf("graph needs at least one node")
		}
		if len(custom.Nodes) > 0 {
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
	if len(nodes) > 32 || len(edges) > 128 {
		return nil, fmt.Errorf("graph limit: 32 nodes, 128 edges")
	}
	ids := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if n.ID == "" || ids[n.ID] || utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
			return nil, fmt.Errorf("node IDs must be unique, nonempty; IDs and labels limited to 64 characters")
		}
		if math.IsNaN(n.X) || math.IsInf(n.X, 0) || math.IsNaN(n.Y) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
			return nil, fmt.Errorf("node coordinates must be finite within ±1000000")
		}
		ids[n.ID] = true
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] || e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("edges need known endpoints and weights within ±10000")
		}
	}
	return Trace(nodes, edges), nil
}

func defaultGraph() ([]renderers.GraphNode, []renderers.GraphEdge) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 140, Y: 150, State: "unvisited"},
		{ID: "B", Label: "B", X: 260, Y: 150, State: "unvisited"},
		{ID: "C", Label: "C", X: 200, Y: 300, State: "unvisited"},
		{ID: "D", Label: "D", X: 380, Y: 220, State: "unvisited"},
		{ID: "E", Label: "E", X: 480, Y: 300, State: "unvisited"},
		{ID: "F", Label: "F", X: 600, Y: 150, State: "unvisited"},
		{ID: "G", Label: "G", X: 700, Y: 260, State: "unvisited"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "B", To: "C"},
		{From: "C", To: "A"},
		{From: "C", To: "D"},
		{From: "D", To: "E"},
		{From: "E", To: "D"},
		{From: "E", To: "F"},
		{From: "F", To: "G"},
		{From: "G", To: "F"},
	}
	return nodes, edges
}

func init() {
	catalog.Register(Lesson{})
}
