package toposort

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
			HowItWorks: "Computes incoming prerequisites and enqueues zero-degree vertices. Each dequeued vertex enters the order before its outgoing dependencies resolve. Neighbors reaching zero join the FIFO. If processing stops early, a cycle exists; blocked vertices may also be downstream of that cycle.",
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
		if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
			return nil, fmt.Errorf("expected graph object, not null")
		}
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid graph: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		for _, key := range []string{"nodes", "edges"} {
			if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s must be an array", key)
			}
		}
		{
			if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			}
			if _, ok := fields["edges"]; ok {
				edges = custom.Edges
			}
		}
	}

	if len(nodes) > 32 || len(edges) > 128 {
		return nil, fmt.Errorf("graph limit: 32 nodes, 128 edges")
	}
	ids := map[string]bool{}
	for _, n := range nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("node ID must not be empty")
		}
		if ids[n.ID] {
			return nil, fmt.Errorf("duplicate node ID %q", n.ID)
		}
		ids[n.ID] = true
		if utf8.RuneCountInString(n.ID) > 64 || utf8.RuneCountInString(n.Label) > 64 {
			return nil, fmt.Errorf("node ID and label limit: 64 characters")
		}
		if math.IsNaN(n.X) || math.IsInf(n.X, 0) || math.IsNaN(n.Y) || math.IsInf(n.Y, 0) || math.Abs(n.X) > 1000000 || math.Abs(n.Y) > 1000000 {
			return nil, fmt.Errorf("coordinates must be finite within ±1000000")
		}
	}
	for _, e := range edges {
		if !ids[e.From] || !ids[e.To] {
			return nil, fmt.Errorf("dependency endpoint missing: %q to %q", e.From, e.To)
		}
		if e.Weight < -10000 || e.Weight > 10000 {
			return nil, fmt.Errorf("weight limit: ±10000")
		}
	}

	return Trace(nodes, edges), nil
}

func init() {
	catalog.Register(Lesson{})
}
