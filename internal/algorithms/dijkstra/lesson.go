package dijkstra

import (
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Lesson struct{}

type LessonInput struct {
	Start  string `json:"start"`
	Target string `json:"target"`
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

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
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

	start := "A"
	target := "E"

	if len(input) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		if fields == nil {
			return nil, fmt.Errorf("input must be an object")
		}
		for key, dst := range map[string]*string{"start": &start, "target": &target} {
			if raw, ok := fields[key]; ok {
				var value *string
				if err := json.Unmarshal(raw, &value); err != nil || value == nil {
					return nil, fmt.Errorf("%s must be a string", key)
				}
				if *value == "" {
					continue
				} // Preserve existing omitted/empty defaults.
				if findNodeIndex(nodes, *value) < 0 {
					return nil, fmt.Errorf("%s must name an existing node", key)
				}
				*dst = *value
			}
		}
	}

	return Trace(nodes, edges, start, target), nil
}

func init() {
	catalog.Register(Lesson{})
}
