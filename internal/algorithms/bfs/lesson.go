package bfs

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
		{ID: "L3", Number: 3, Code: "    queue := []string{start}"},
		{ID: "L4", Number: 4, Code: "    visited[start] = true"},
		{ID: "L5", Number: 5, Code: "    order := []string{}"},
		{ID: "L6", Number: 6, Code: "    for len(queue) > 0 {"},
		{ID: "L7", Number: 7, Code: "        curr := queue[0]"},
		{ID: "L8", Number: 8, Code: "        queue = queue[1:]"},
		{ID: "L9", Number: 9, Code: "        order = append(order, curr)"},
		{ID: "L10", Number: 10, Code: "        for _, neighbor := range adj[curr] {"},
		{ID: "L11", Number: 11, Code: "            if !visited[neighbor] {"},
		{ID: "L12", Number: 12, Code: "                visited[neighbor] = true"},
		{ID: "L13", Number: 13, Code: "                queue = append(queue, neighbor)"},
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
	Start *string               `json:"start"`
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
