package dijkstra

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Lesson struct{}

type LessonInput struct {
	Start  string                `json:"start"`
	Target string                `json:"target"`
	Nodes  []renderers.GraphNode `json:"nodes,omitempty"`
	Edges  []renderers.GraphEdge `json:"edges,omitempty"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "dijkstra",
		Name:     "Dijkstra's Shortest Path",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Finds the shortest paths between nodes in a weighted graph with non-negative edge weights using a priority queue.",
			TimeComplexity:  "O((V + E) log V)",
			SpaceComplexity: "O(V)",
			KeyConcepts:     []string{"Greedy", "Priority Queue", "Edge Relaxation", "Non-negative Weights"},
			HowItWorks:      "Maintains a min-priority queue of vertices ordered by tentative distance from the start node. At each step, it extracts the unvisited vertex with the lowest distance and relaxes all outgoing edges. If a shorter path to a neighbor is discovered, its distance is updated and queued. Because all edge weights are non-negative, once a node is visited its shortest path is guaranteed to be optimal.",
		},
		DefaultInput: `{"start": "A", "target": "E"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Dijkstra(graph Graph, start, target string) (map[string]int, []string) {"},
		{ID: "L2", Number: 2, Code: "    dist := initDistances(graph, start) // start = 0, others = ∞"},
		{ID: "L3", Number: 3, Code: "    pq := priorityqueue.New()"},
		{ID: "L4", Number: 4, Code: "    pq.Push(start, 0)"},
		{ID: "L5", Number: 5, Code: "    for !pq.IsEmpty() {"},
		{ID: "L6", Number: 6, Code: "        curr := pq.Pop() // vertex with min distance"},
		{ID: "L7", Number: 7, Code: "        if curr == target { break }"},
		{ID: "L8", Number: 8, Code: "        for _, edge := range graph.Neighbors(curr) {"},
		{ID: "L9", Number: 9, Code: "            alt := dist[curr] + edge.Weight"},
		{ID: "L10", Number: 10, Code: "            if alt < dist[edge.To] {"},
		{ID: "L11", Number: 11, Code: "                dist[edge.To] = alt"},
		{ID: "L12", Number: 12, Code: "                prev[edge.To] = curr"},
		{ID: "L13", Number: 13, Code: "                pq.Push(edge.To, alt)"},
		{ID: "L14", Number: 14, Code: "            }"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "    return dist, reconstructPath(prev, start, target)"},
		{ID: "L18", Number: 18, Code: "}"},
	}
}

func defaultGraph() ([]renderers.GraphNode, []renderers.GraphEdge) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 120, Y: 225, State: "unvisited"},
		{ID: "B", Label: "B", X: 320, Y: 120, State: "unvisited"},
		{ID: "C", Label: "C", X: 320, Y: 330, State: "unvisited"},
		{ID: "D", Label: "D", X: 520, Y: 120, State: "unvisited"},
		{ID: "E", Label: "E", X: 680, Y: 225, State: "unvisited"},
	}

	edges := []renderers.GraphEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "C", Weight: 2},
		{From: "B", To: "C", Weight: 1},
		{From: "B", To: "D", Weight: 5},
		{From: "C", To: "D", Weight: 8},
		{From: "C", To: "E", Weight: 10},
		{From: "D", To: "E", Weight: 2},
	}
	return nodes, edges
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges := defaultGraph()
	start := "A"
	target := "E"

	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if string(trimmed) == "null" {
			return nil, fmt.Errorf("input cannot be null")
		}
		if len(trimmed) == 0 || trimmed[0] != '{' {
			return nil, fmt.Errorf("malformed JSON input: expected JSON object")
		}

		var custom LessonInput
		if err := json.Unmarshal(trimmed, &custom); err != nil {
			return nil, fmt.Errorf("malformed JSON input: %w", err)
		}

		if len(custom.Nodes) > 0 {
			nodes = custom.Nodes
		}
		if len(custom.Edges) > 0 {
			edges = custom.Edges
		}
		if custom.Start != "" {
			start = custom.Start
		}
		if custom.Target != "" {
			target = custom.Target
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("graph must contain at least one node")
	}

	nodeMap := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("node ID cannot be empty")
		}
		nodeMap[n.ID] = true
	}

	if !nodeMap[start] {
		return nil, fmt.Errorf("invalid start node %q: node does not exist in graph", start)
	}
	if target != "" && !nodeMap[target] {
		return nil, fmt.Errorf("invalid target node %q: node does not exist in graph", target)
	}

	for _, e := range edges {
		if e.Weight < 0 {
			return nil, fmt.Errorf("negative edge weight %d between %s and %s is not permitted in Dijkstra", e.Weight, e.From, e.To)
		}
		if !nodeMap[e.From] || !nodeMap[e.To] {
			return nil, fmt.Errorf("edge connects unknown node(s): %s -> %s", e.From, e.To)
		}
	}

	return Trace(nodes, edges, start, target), nil
}

func init() {
	catalog.Register(Lesson{})
}
