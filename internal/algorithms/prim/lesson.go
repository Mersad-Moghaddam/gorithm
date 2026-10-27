package prim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Lesson struct{}

type LessonInput struct {
	Nodes []renderers.GraphNode `json:"nodes,omitempty"`
	Edges []WeightedEdge        `json:"edges,omitempty"`
	Start *string               `json:"start,omitempty"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "prim",
		Name:     "Prim's Minimum Spanning Tree",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "A greedy algorithm that finds a Minimum Spanning Tree for a weighted undirected graph by growing a single tree from an arbitrary starting vertex.",
			TimeComplexity:  "O(E log V)",
			SpaceComplexity: "O(V)",
			KeyConcepts:     []string{"Greedy", "Cut Property", "Growing Tree", "Priority Queue"},
			HowItWorks:      "Prim's algorithm starts with an arbitrary node and grows a minimum spanning tree vertex by vertex. At each step, it selects the minimum-weight edge connecting a vertex inside the tree to a vertex outside the tree (the cut property) using a priority queue, adding the edge and destination vertex until all vertices are visited.",
		},
		DefaultInput: `{"start":"A","edges":[{"from":"A","to":"B","weight":4},{"from":"A","to":"C","weight":2},{"from":"B","to":"C","weight":1},{"from":"B","to":"D","weight":5},{"from":"C","to":"D","weight":8},{"from":"C","to":"E","weight":10},{"from":"D","to":"E","weight":2},{"from":"D","to":"F","weight":6},{"from":"E","to":"F","weight":3}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func PrimMST(nodes []string, edges []WeightedEdge, start string) ([]WeightedEdge, int) {"},
		{ID: "L2", Number: 2, Code: "    mst := []WeightedEdge{}"},
		{ID: "L3", Number: 3, Code: "    visited := map[string]bool{start: true}"},
		{ID: "L4", Number: 4, Code: "    pq := newMinHeap()"},
		{ID: "L5", Number: 5, Code: "    for _, e := range adj[start] { pq.Push(e) }"},
		{ID: "L6", Number: 6, Code: "    for !pq.Empty() && len(mst) < len(nodes)-1 {"},
		{ID: "L7", Number: 7, Code: "        edge := pq.PopMin()"},
		{ID: "L8", Number: 8, Code: "        if visited[edge.To] {"},
		{ID: "L9", Number: 9, Code: "            continue // Skip cycle"},
		{ID: "L10", Number: 10, Code: "        }"},
		{ID: "L11", Number: 11, Code: "        visited[edge.To] = true"},
		{ID: "L12", Number: 12, Code: "        mst = append(mst, edge)"},
		{ID: "L13", Number: 13, Code: "        totalWeight += edge.Weight"},
		{ID: "L14", Number: 14, Code: "        for _, next := range adj[edge.To] {"},
		{ID: "L15", Number: 15, Code: "            if !visited[next.To] { pq.Push(next) }"},
		{ID: "L16", Number: 16, Code: "        }"},
		{ID: "L17", Number: 17, Code: "    }"},
		{ID: "L18", Number: 18, Code: "    return mst, totalWeight"},
		{ID: "L19", Number: 19, Code: "}"},
	}
}

func DefaultGraph() ([]renderers.GraphNode, []WeightedEdge, string) {
	nodes := []renderers.GraphNode{
		{ID: "A", Label: "A", X: 150, Y: 225, State: "unvisited"},
		{ID: "B", Label: "B", X: 300, Y: 100, State: "unvisited"},
		{ID: "C", Label: "C", X: 300, Y: 350, State: "unvisited"},
		{ID: "D", Label: "D", X: 500, Y: 100, State: "unvisited"},
		{ID: "E", Label: "E", X: 500, Y: 350, State: "unvisited"},
		{ID: "F", Label: "F", X: 650, Y: 225, State: "unvisited"},
	}
	edges := []WeightedEdge{
		{From: "A", To: "B", Weight: 4},
		{From: "A", To: "C", Weight: 2},
		{From: "B", To: "C", Weight: 1},
		{From: "B", To: "D", Weight: 5},
		{From: "C", To: "D", Weight: 8},
		{From: "C", To: "E", Weight: 10},
		{From: "D", To: "E", Weight: 2},
		{From: "D", To: "F", Weight: 6},
		{From: "E", To: "F", Weight: 3},
	}
	return nodes, edges, "A"
}

func layoutNodesFromEdges(edges []WeightedEdge) []renderers.GraphNode {
	seen := make(map[string]bool)
	var order []string
	for _, e := range edges {
		if !seen[e.From] {
			seen[e.From] = true
			order = append(order, e.From)
		}
		if !seen[e.To] {
			seen[e.To] = true
			order = append(order, e.To)
		}
	}
	n := len(order)
	nodes := make([]renderers.GraphNode, n)
	if n == 0 {
		return nodes
	}
	for i, id := range order {
		angle := (float64(i) / float64(n)) * 2 * math.Pi
		x := 400 + 250*math.Cos(angle)
		y := 225 + 140*math.Sin(angle)
		nodes[i] = renderers.GraphNode{
			ID:    id,
			Label: id,
			X:     math.Round(x),
			Y:     math.Round(y),
			State: "unvisited",
		}
	}
	return nodes
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges, start := DefaultGraph()
	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) > 0 {
			if bytes.Equal(trimmed, []byte("null")) {
				return nil, fmt.Errorf("input cannot be null")
			}

			var rawMap map[string]json.RawMessage
			if err := json.Unmarshal(trimmed, &rawMap); err != nil {
				return nil, fmt.Errorf("malformed JSON input: %w", err)
			}

			if rawStart, ok := rawMap["start"]; ok && bytes.Equal(bytes.TrimSpace(rawStart), []byte("null")) {
				return nil, fmt.Errorf("start node cannot be null")
			}

			var custom LessonInput
			if err := json.Unmarshal(trimmed, &custom); err != nil {
				return nil, fmt.Errorf("malformed JSON input: %w", err)
			}

			if len(custom.Edges) > 0 {
				edges = custom.Edges
				if len(custom.Nodes) > 0 {
					nodes = custom.Nodes
				} else {
					nodes = layoutNodesFromEdges(edges)
				}
			} else if len(custom.Nodes) > 0 {
				nodes = custom.Nodes
			}

			if custom.Start != nil {
				start = *custom.Start
			} else if len(custom.Edges) > 0 && len(nodes) > 0 {
				found := false
				for _, n := range nodes {
					if n.ID == start {
						found = true
						break
					}
				}
				if !found {
					start = nodes[0].ID
				}
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
