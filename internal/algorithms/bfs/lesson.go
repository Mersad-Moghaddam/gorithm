package bfs

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
		ID:       "bfs",
		Name:     "Breadth-First Search (BFS)",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Explores a graph level by level using a queue (FIFO), visiting all immediate neighbors before moving to deeper vertices.",
			TimeComplexity:  "O(V + E)",
			SpaceComplexity: "O(V)",
			KeyConcepts:     []string{"Queue / FIFO", "Level-by-Level", "Shortest Path in Unweighted Graph"},
			HowItWorks:      "Starts at the root/source node and enqueues it. In each step, dequeues the front node, visits it, and inspects its adjacent neighbors. Any unvisited neighbor is marked as queued and appended to the queue. Continues until the queue is empty.",
		},
		DefaultInput: `{"start": "A"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func BFS(adj map[string][]string, start string) []string {"},
		{ID: "L2", Number: 2, Code: "    visited := make(map[string]bool)"},
		{ID: "L3", Number: 3, Code: "    queue := []string{start}; visited[start] = true"},
		{ID: "L4", Number: 4, Code: ""},
		{ID: "L5", Number: 5, Code: "    order := []string{}"},
		{ID: "L6", Number: 6, Code: "    for len(queue) > 0 {"},
		{ID: "L7", Number: 7, Code: "        curr := queue[0]; queue = queue[1:]"},
		{ID: "L8", Number: 8, Code: ""},
		{ID: "L9", Number: 9, Code: "        order = append(order, curr)"},
		{ID: "L10", Number: 10, Code: "        for _, neighbor := range adj[curr] {"},
		{ID: "L11", Number: 11, Code: "            if !visited[neighbor] {"},
		{ID: "L12", Number: 12, Code: "                visited[neighbor] = true; queue = append(queue, neighbor)"},
		{ID: "L13", Number: 13, Code: ""},
		{ID: "L14", Number: 14, Code: "            }"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "    return order"},
		{ID: "L18", Number: 18, Code: "}"},
	}
}

type buildInput struct {
	Nodes []renderers.GraphNode `json:"nodes"`
	Edges []renderers.GraphEdge `json:"edges"`
	Start string                `json:"start"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 150, Y: 225, State: "unvisited"},
		{ID: "B", Label: "B", X: 320, Y: 130, State: "unvisited"},
		{ID: "C", Label: "C", X: 320, Y: 320, State: "unvisited"},
		{ID: "D", Label: "D", X: 500, Y: 130, State: "unvisited"},
		{ID: "E", Label: "E", X: 500, Y: 320, State: "unvisited"},
		{ID: "F", Label: "F", X: 670, Y: 225, State: "unvisited"},
	}

	edges := []renderers.GraphEdge{
		{From: "A", To: "B"},
		{From: "A", To: "C"},
		{From: "B", To: "D"},
		{From: "B", To: "E"},
		{From: "C", To: "E"},
		{From: "D", To: "F"},
		{From: "E", To: "F"},
	}

	start := "A"

	if len(input) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("BFS input must be a JSON object")
		}
		for _, key := range []string{"nodes", "edges", "start"} {
			if bytes.Equal(bytes.TrimSpace(fields[key]), []byte("null")) {
				return nil, fmt.Errorf("%s must not be null", key)
			}
		}
		var custom buildInput
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid BFS input: %w", err)
		} else {
			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			}
			if len(custom.Edges) > 0 {
				edges = custom.Edges
			}
			if custom.Start != "" {
				start = custom.Start
			}
		}
	}
	if len(nodes) > 32 || len(edges) > 128 {
		return nil, fmt.Errorf("BFS supports at most 32 nodes and 128 edges")
	}
	ids := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if n.ID == "" || ids[n.ID] || utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
			return nil, fmt.Errorf("node IDs must be unique, nonempty, and IDs/labels at most 64 characters")
		}
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
			return nil, fmt.Errorf("node coordinates must be finite within ±1000000")
		}
		ids[n.ID] = true
	}
	if !ids[start] {
		return nil, fmt.Errorf("unknown BFS start %q", start)
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] {
			return nil, fmt.Errorf("edge references unknown node")
		}
		if e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("edge weight must be within ±10000")
		}
	}

	return Trace(nodes, edges, start), nil
}

func init() {
	catalog.Register(Lesson{})
}
