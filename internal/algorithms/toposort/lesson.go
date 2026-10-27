package toposort

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
		ID:       "toposort",
		Name:     "Topological Sort (DAG)",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Linear ordering of vertices in a Directed Acyclic Graph (DAG) such that every directed edge u -> v comes with u before v, solved via Kahn's algorithm using in-degrees.",
			TimeComplexity:  "O(V + E)",
			SpaceComplexity: "O(V)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Directed Acyclic Graph",
				"In-Degree Zero Queue",
				"Dependency Resolution",
				"Cycle Detection",
			},
			HowItWorks: "Computes the in-degree (number of incoming prerequisites) for each node and enqueues vertices with in-degree 0. While the queue has elements, it dequeues a node, appends it to the sorted list, and decrements the in-degrees of all outgoing neighbors. Whenever a neighbor reaches in-degree 0, it is added to the queue. If all nodes are processed, a valid topological order is achieved; otherwise, any unprocessed nodes participate in a cycle.",
		},
		DefaultInput: `{}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func KahnToposort(nodes []string, edges [][2]string) ([]string, bool) {"},
		{ID: "L2", Number: 2, Code: "    inDegree := calcInDegrees(nodes, edges)"},
		{ID: "L3", Number: 3, Code: "    queue := []string{}"},
		{ID: "L4", Number: 4, Code: "    for _, u := range nodes { if inDegree[u] == 0 { queue = append(queue, u) } }"},
		{ID: "L5", Number: 5, Code: "    order := []string{}"},
		{ID: "L6", Number: 6, Code: "    for len(queue) > 0 {"},
		{ID: "L7", Number: 7, Code: "        curr := queue[0]; queue = queue[1:]"},
		{ID: "L8", Number: 8, Code: "        order = append(order, curr)"},
		{ID: "L9", Number: 9, Code: "        for _, neighbor := range adj[curr] {"},
		{ID: "L10", Number: 10, Code: "            inDegree[neighbor]--"},
		{ID: "L11", Number: 11, Code: "            if inDegree[neighbor] == 0 {"},
		{ID: "L12", Number: 12, Code: "                queue = append(queue, neighbor)"},
		{ID: "L13", Number: 13, Code: "            }"},
		{ID: "L14", Number: 14, Code: "        }"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    return order, len(order) == len(nodes)"},
		{ID: "L17", Number: 17, Code: "}"},
	}
}

type buildInput struct {
	Nodes *[]renderers.GraphNode `json:"nodes"`
	Edges *[]renderers.GraphEdge `json:"edges"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes := []renderers.GraphNode{
		{ID: "clean", Label: "clean", X: 100, Y: 225, State: "unvisited"},
		{ID: "lint", Label: "lint", X: 280, Y: 130, State: "unvisited"},
		{ID: "compile", Label: "compile", X: 280, Y: 320, State: "unvisited"},
		{ID: "test", Label: "test", X: 460, Y: 130, State: "unvisited"},
		{ID: "docs", Label: "docs", X: 460, Y: 320, State: "unvisited"},
		{ID: "package", Label: "package", X: 640, Y: 225, State: "unvisited"},
		{ID: "deploy", Label: "deploy", X: 800, Y: 225, State: "unvisited"},
	}

	edges := []renderers.GraphEdge{
		{From: "clean", To: "lint", Directed: true},
		{From: "clean", To: "compile", Directed: true},
		{From: "compile", To: "test", Directed: true},
		{From: "compile", To: "docs", Directed: true},
		{From: "lint", To: "package", Directed: true},
		{From: "test", To: "package", Directed: true},
		{From: "docs", To: "package", Directed: true},
		{From: "package", To: "deploy", Directed: true},
	}

	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) == 0 {
			return Trace(nodes, edges), nil
		}
		if bytes.Equal(trimmed, []byte("null")) {
			return nil, fmt.Errorf("toposort input cannot be null")
		}
		if trimmed[0] != '{' {
			return nil, fmt.Errorf("toposort input must be a JSON object")
		}
		var custom buildInput
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, err
		}
		if custom.Nodes != nil {
			nodes = *custom.Nodes
			if custom.Edges == nil {
				edges = nil
			}
		}
		if custom.Edges != nil {
			edges = *custom.Edges
		}
		nodeIDs := make(map[string]bool, len(nodes))
		for _, n := range nodes {
			if n.ID == "" {
				return nil, fmt.Errorf("node ID cannot be empty")
			}
			nodeIDs[n.ID] = true
		}
		for i := range edges {
			edges[i].Directed = true
			if edges[i].From == "" || edges[i].To == "" {
				return nil, fmt.Errorf("edge endpoints cannot be empty")
			}
			if !nodeIDs[edges[i].From] || !nodeIDs[edges[i].To] {
				return nil, fmt.Errorf("edge references unknown node (%s -> %s)", edges[i].From, edges[i].To)
			}
		}
	}

	return Trace(nodes, edges), nil
}

func init() {
	catalog.Register(Lesson{})
}
