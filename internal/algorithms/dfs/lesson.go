package dfs

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "dfs",
		Name:     "Depth-First Search (DFS)",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Traverses graph vertices by exploring as deep as possible along each branch before backtracking.",
			TimeComplexity:  "O(V + E)",
			SpaceComplexity: "O(V)",
			KeyConcepts: []string{
				"Recursion / LIFO Stack",
				"Deep Exploration",
				"Backtracking",
				"Cycle Detection",
			},
			HowItWorks: "Starts at a chosen root vertex and marks it as visiting. Traverses outgoing edges to unvisited adjacent vertices recursively. Upon reaching a vertex with no unvisited neighbors, backtracks to explore remaining branches until all reachable vertices are processed and marked visited.",
		},
		DefaultInput: `{"start": "A"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Traverse(adj map[string][]string, start string) []string {"},
		{ID: "L2", Number: 2, Code: "    visited := make(map[string]bool)"},
		{ID: "L3", Number: 3, Code: "    order := make([]string, 0)"},
		{ID: "L4", Number: 4, Code: "    var dfs func(u string)"},
		{ID: "L5", Number: 5, Code: "    dfs = func(u string) {"},
		{ID: "L6", Number: 6, Code: "        visited[u] = true"},
		{ID: "L7", Number: 7, Code: "        order = append(order, u)"},
		{ID: "L8", Number: 8, Code: "        for _, v := range adj[u] {"},
		{ID: "L9", Number: 9, Code: "            if !visited[v] {"},
		{ID: "L10", Number: 10, Code: "                dfs(v)"},
		{ID: "L11", Number: 11, Code: "            }"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "    }"},
		{ID: "L14", Number: 14, Code: "    dfs(start)"},
		{ID: "L15", Number: 15, Code: "    return order"},
		{ID: "L16", Number: 16, Code: "}"},
	}
}

type buildInput struct {
	Nodes []renderers.GraphNode `json:"nodes"`
	Edges []renderers.GraphEdge `json:"edges"`
	Start *string               `json:"start"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 400, Y: 80, State: "unvisited"},
		{ID: "B", Label: "B", X: 240, Y: 200, State: "unvisited"},
		{ID: "C", Label: "C", X: 560, Y: 200, State: "unvisited"},
		{ID: "D", Label: "D", X: 160, Y: 340, State: "unvisited"},
		{ID: "E", Label: "E", X: 360, Y: 340, State: "unvisited"},
		{ID: "F", Label: "F", X: 640, Y: 340, State: "unvisited"},
	}
	edges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "A", To: "C"},
		{From: "B", To: "D"},
		{From: "B", To: "E"},
		{From: "C", To: "E"},
		{From: "C", To: "F"},
	}
	start := "A"

	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) > 0 {
			if bytes.Equal(trimmed, []byte("null")) {
				return nil, fmt.Errorf("input cannot be null")
			}

			var rawMap map[string]json.RawMessage
			if err := json.Unmarshal(trimmed, &rawMap); err == nil {
				if rawStart, ok := rawMap["start"]; ok && bytes.Equal(bytes.TrimSpace(rawStart), []byte("null")) {
					return nil, fmt.Errorf("start node cannot be null")
				}
			}

			var custom buildInput
			if err := json.Unmarshal(trimmed, &custom); err != nil {
				var startStr string
				if strErr := json.Unmarshal(trimmed, &startStr); strErr == nil {
					custom.Start = &startStr
				} else {
					return nil, fmt.Errorf("malformed JSON input: %w", err)
				}
			}

			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
				if custom.Start == nil {
					start = nodes[0].ID
				}
			}
			if len(custom.Edges) > 0 {
				edges = custom.Edges
			}
			if custom.Start != nil {
				if *custom.Start == "" {
					return nil, fmt.Errorf("start node cannot be empty")
				}
				start = *custom.Start
			}
		}
	}

	startFound := false
	for _, n := range nodes {
		if n.ID == start {
			startFound = true
			break
		}
	}
	if !startFound {
		return nil, fmt.Errorf("unknown start node: %q", start)
	}

	return Trace(nodes, edges, start), nil
}

func init() {
	catalog.Register(Lesson{})
}
