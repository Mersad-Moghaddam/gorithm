package astar

import (
	"encoding/json"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	Nodes []AStarNode `json:"nodes"`
	Edges []AStarEdge `json:"edges"`
	Start string      `json:"start"`
	Goal  string      `json:"goal"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "astar",
		Name:     "A* Pathfinding",
		Category: catalog.CategoryGraphs,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "An informed search algorithm that finds the shortest path between a starting node and a goal node by combining known path cost g(n) with heuristic estimate h(n).",
			TimeComplexity:  "O(E)",
			SpaceComplexity: "O(V)",
			KeyConcepts:     []string{"Heuristic Search", "f(n)=g(n)+h(n)", "Admissible Heuristic", "Optimal Path"},
			HowItWorks:      "Maintains an open set min-priority queue ordered by estimated total cost f(n) = g(n) + h(n), where g(n) is the exact cost from start to n and h(n) is an admissible heuristic estimate from n to goal. Repeatedly expands the node with the lowest f(n), updates neighbor tentative costs, and tracks predecessor pointers to reconstruct the optimal path once the goal is reached.",
		},
		DefaultInput: `{"nodes":[{"id":"A","label":"A","x":140,"y":220},{"id":"B","label":"B","x":300,"y":120},{"id":"C","label":"C","x":300,"y":320},{"id":"D","label":"D","x":500,"y":120},{"id":"E","label":"E","x":500,"y":320},{"id":"F","label":"F","x":660,"y":220}],"edges":[{"from":"A","to":"B","weight":190,"undirected":true},{"from":"A","to":"C","weight":200,"undirected":true},{"from":"B","to":"D","weight":200,"undirected":true},{"from":"C","to":"E","weight":200,"undirected":true},{"from":"B","to":"E","weight":280,"undirected":true},{"from":"C","to":"D","weight":280,"undirected":true},{"from":"D","to":"F","weight":190,"undirected":true},{"from":"E","to":"F","weight":200,"undirected":true}],"start":"A","goal":"F"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func AStar(start, goal Node) []Node {"},
		{ID: "L2", Number: 2, Code: "    openSet := NewPriorityQueue()"},
		{ID: "L3", Number: 3, Code: "    openSet.Push(start, fScore[start])"},
		{ID: "L4", Number: 4, Code: "    for !openSet.IsEmpty() {"},
		{ID: "L5", Number: 5, Code: "        current := openSet.PopMin()"},
		{ID: "L6", Number: 6, Code: "        if current == goal {"},
		{ID: "L7", Number: 7, Code: "            return reconstructPath(cameFrom, current)"},
		{ID: "L8", Number: 8, Code: "        }"},
		{ID: "L9", Number: 9, Code: "        for _, edge := range neighbors(current) {"},
		{ID: "L10", Number: 10, Code: "            tentativeG := gScore[current] + edge.weight"},
		{ID: "L11", Number: 11, Code: "            if tentativeG < gScore[edge.to] {"},
		{ID: "L12", Number: 12, Code: "                cameFrom[edge.to] = current"},
		{ID: "L13", Number: 13, Code: "                gScore[edge.to] = tentativeG"},
		{ID: "L14", Number: 14, Code: "                fScore[edge.to] = tentativeG + h(edge.to, goal)"},
		{ID: "L15", Number: 15, Code: "                openSet.PushOrUpdate(edge.to, fScore[edge.to])"},
		{ID: "L16", Number: 16, Code: "            }"},
		{ID: "L17", Number: 17, Code: "        }"},
		{ID: "L18", Number: 18, Code: "    }"},
		{ID: "L19", Number: 19, Code: "    return nil // Path not found"},
		{ID: "L20", Number: 20, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	nodes, edges, start, goal := defaultGraph()
	if len(input) > 0 {
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err == nil && len(custom.Nodes) > 0 {
			nodes = custom.Nodes
			edges = custom.Edges
			if custom.Start != "" {
				start = custom.Start
			} else {
				start = nodes[0].ID
			}
			if custom.Goal != "" {
				goal = custom.Goal
			} else {
				goal = nodes[len(nodes)-1].ID
			}

			for i := range nodes {
				if nodes[i].Label == "" {
					nodes[i].Label = nodes[i].ID
				}
				if nodes[i].X == 0 && nodes[i].Y == 0 {
					angle := 2 * math.Pi * float64(i) / float64(len(nodes))
					nodes[i].X = 400 + 180*math.Cos(angle)
					nodes[i].Y = 225 + 130*math.Sin(angle)
				}
			}
		}
	}
	return Trace(nodes, edges, start, goal), nil
}

func defaultGraph() ([]AStarNode, []AStarEdge, string, string) {
	nodes := []AStarNode{
		{ID: "A", Label: "A", X: 140, Y: 220},
		{ID: "B", Label: "B", X: 300, Y: 120},
		{ID: "C", Label: "C", X: 300, Y: 320},
		{ID: "D", Label: "D", X: 500, Y: 120},
		{ID: "E", Label: "E", X: 500, Y: 320},
		{ID: "F", Label: "F", X: 660, Y: 220},
	}
	edges := []AStarEdge{
		{From: "A", To: "B", Weight: 190, Undirected: true},
		{From: "A", To: "C", Weight: 200, Undirected: true},
		{From: "B", To: "D", Weight: 200, Undirected: true},
		{From: "C", To: "E", Weight: 200, Undirected: true},
		{From: "B", To: "E", Weight: 280, Undirected: true},
		{From: "C", To: "D", Weight: 280, Undirected: true},
		{From: "D", To: "F", Weight: 190, Undirected: true},
		{From: "E", To: "F", Weight: 200, Undirected: true},
	}
	return nodes, edges, "A", "F"
}

func init() {
	catalog.Register(Lesson{})
}
