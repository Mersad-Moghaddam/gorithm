package kosaraju

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"

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
				"Graph Transposition",
				"Backward DFS Decomposition",
			},
			HowItWorks: "Phase 1 performs a forward DFS on the original graph, pushing vertices onto a finish stack in post-order. Phase 2 transposes the graph by reversing all directed edges; in the component condensation DAG, this reverses cross-component edges so traversal cannot escape an SCC. Phase 3 repeatedly pops vertices from the finish stack and runs backward DFS on the transposed graph, discovering one complete strongly connected component per unvisited vertex.",
		},
		DefaultInput: `{"nodes":[{"id":"A","label":"A","x":140,"y":150},{"id":"B","label":"B","x":260,"y":150},{"id":"C","label":"C","x":200,"y":300},{"id":"D","label":"D","x":380,"y":220},{"id":"E","label":"E","x":480,"y":300},{"id":"F","label":"F","x":600,"y":150},{"id":"G","label":"G","x":700,"y":260}],"edges":[{"from":"A","to":"B","directed":true},{"from":"B","to":"C","directed":true},{"from":"C","to":"A","directed":true},{"from":"C","to":"D","directed":true},{"from":"D","to":"E","directed":true},{"from":"E","to":"D","directed":true},{"from":"E","to":"F","directed":true},{"from":"F","to":"G","directed":true},{"from":"G","to":"F","directed":true}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func FindSCCs(nodes []string, edges [][2]string) [][]string {"},
		{ID: "L2", Number: 2, Code: "    adj, revAdj := buildAdjacency(nodes, edges)"},
		{ID: "L3", Number: 3, Code: "    visited := make(map[string]bool); finishStack := []string{}"},
		{ID: "L4", Number: 4, Code: "    for _, u := range nodes {"},
		{ID: "L5", Number: 5, Code: "        if !visited[u] { dfsForward(u) } // push to finishStack on return"},
		{ID: "L6", Number: 6, Code: "    }"},
		{ID: "L7", Number: 7, Code: "    transpose(graph) // reverse all edges to trap backward DFS in SCCs"},
		{ID: "L8", Number: 8, Code: "    visitedRev := make(map[string]bool); sccs := [][]string{}"},
		{ID: "L9", Number: 9, Code: "    for len(finishStack) > 0 {"},
		{ID: "L10", Number: 10, Code: "        u := pop(&finishStack) // inspect vertex with latest finish time"},
		{ID: "L11", Number: 11, Code: "        if !visitedRev[u] {"},
		{ID: "L12", Number: 12, Code: "            var comp []string"},
		{ID: "L13", Number: 13, Code: "            dfsReverse(u, &comp) // backward DFS in transposed graph"},
		{ID: "L14", Number: 14, Code: "            sccs = append(sccs, comp)"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "    return sccs"},
		{ID: "L18", Number: 18, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges := defaultGraph()
	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) == 0 {
			return Trace(nodes, edges), nil
		}
		if string(trimmed) == "null" {
			return nil, errors.New("input cannot be null")
		}
		var custom LessonInput
		if err := json.Unmarshal(trimmed, &custom); err != nil {
			return nil, fmt.Errorf("malformed JSON input: %w", err)
		}
		if len(custom.Nodes) == 0 {
			return nil, errors.New("graph must contain at least one node")
		}

		nodeSet := make(map[string]bool, len(custom.Nodes))
		for i := range custom.Nodes {
			id := custom.Nodes[i].ID
			if id == "" {
				return nil, errors.New("node ID cannot be empty")
			}
			if nodeSet[id] {
				return nil, fmt.Errorf("duplicate node ID: %s", id)
			}
			nodeSet[id] = true

			if custom.Nodes[i].Label == "" {
				custom.Nodes[i].Label = id
			}
			if custom.Nodes[i].State == "" {
				custom.Nodes[i].State = "unvisited"
			}
			if custom.Nodes[i].X == 0 && custom.Nodes[i].Y == 0 {
				angle := 2 * math.Pi * float64(i) / float64(len(custom.Nodes))
				custom.Nodes[i].X = 400 + 180*math.Cos(angle)
				custom.Nodes[i].Y = 225 + 130*math.Sin(angle)
			}
		}

		for i := range custom.Edges {
			if !nodeSet[custom.Edges[i].From] || !nodeSet[custom.Edges[i].To] {
				return nil, fmt.Errorf("edge from %q to %q references unknown node", custom.Edges[i].From, custom.Edges[i].To)
			}
			custom.Edges[i].Directed = true
		}

		nodes = custom.Nodes
		edges = custom.Edges
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
		{From: "A", To: "B", Directed: true},
		{From: "B", To: "C", Directed: true},
		{From: "C", To: "A", Directed: true},
		{From: "C", To: "D", Directed: true},
		{From: "D", To: "E", Directed: true},
		{From: "E", To: "D", Directed: true},
		{From: "E", To: "F", Directed: true},
		{From: "F", To: "G", Directed: true},
		{From: "G", To: "F", Directed: true},
	}
	return nodes, edges
}

func init() {
	catalog.Register(Lesson{})
}
