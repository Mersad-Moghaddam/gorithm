package toposort

import (
	"encoding/json"

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
	Nodes []renderers.GraphNode `json:"nodes"`
	Edges []renderers.GraphEdge `json:"edges"`
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
		{From: "clean", To: "lint"},
		{From: "clean", To: "compile"},
		{From: "compile", To: "test"},
		{From: "compile", To: "docs"},
		{From: "lint", To: "package"},
		{From: "test", To: "package"},
		{From: "docs", To: "package"},
		{From: "package", To: "deploy"},
	}

	if len(input) > 0 {
		var custom buildInput
		if err := json.Unmarshal(input, &custom); err == nil {
			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			}
			if len(custom.Edges) > 0 {
				edges = custom.Edges
			}
		}
	}

	return Trace(nodes, edges), nil
}

func init() {
	catalog.Register(Lesson{})
}
