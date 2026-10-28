package dfs

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
		DefaultInput: `"A"`,
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
	startNode := "A"

	if len(bytes.TrimSpace(input)) > 0 {
		if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
			return nil, fmt.Errorf("DFS input must be a start string or graph object")
		}
		var strInput string
		if err := json.Unmarshal(input, &strInput); err == nil {
			if strInput != "" {
				startNode = strInput
			}
		} else {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(input, &fields); err != nil {
				return nil, fmt.Errorf("invalid DFS object: %w", err)
			}
			for _, key := range []string{"start", "nodes", "edges"} {
				if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
					return nil, fmt.Errorf("%s cannot be null", key)
				}
			}
			var custom struct {
				Start string                `json:"start"`
				Nodes []renderers.GraphNode `json:"nodes"`
				Edges []renderers.GraphEdge `json:"edges"`
			}
			if err := json.Unmarshal(input, &custom); err != nil {
				return nil, fmt.Errorf("invalid DFS input: %w", err)
			} else {
				if custom.Start != "" {
					startNode = custom.Start
				}
				if len(custom.Nodes) > 0 {
					nodes = custom.Nodes
				}
				if len(custom.Edges) > 0 {
					edges = custom.Edges
				}
			}
		}
	}
	if len(nodes) > 32 || len(edges) > 128 {
		return nil, fmt.Errorf("DFS supports at most 32 nodes and 128 edges")
	}
	ids := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if n.ID == "" || utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
			return nil, fmt.Errorf("node IDs must be nonempty and IDs/labels at most 64 characters")
		}
		if ids[n.ID] {
			return nil, fmt.Errorf("duplicate node ID %q", n.ID)
		}
		ids[n.ID] = true
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1e6 || math.Abs(n.Y) > 1e6 {
			return nil, fmt.Errorf("node coordinates must be finite within ±1000000")
		}
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] {
			return nil, fmt.Errorf("edge endpoints must exist")
		}
		if e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("edge weights must be within ±10000")
		}
	}
	if !ids[startNode] {
		return nil, fmt.Errorf("start node %q does not exist", startNode)
	}

	return Trace(nodes, edges, startNode), nil
}

func init() {
	catalog.Register(Lesson{})
}
