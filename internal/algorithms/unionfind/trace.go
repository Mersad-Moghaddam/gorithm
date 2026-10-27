package unionfind

import (
	"fmt"
	"math"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Operation represents an action to execute in union-find trace.
type Operation struct {
	Op string `json:"op"` // "union", "find", "connected"
	I  int    `json:"i"`
	J  int    `json:"j,omitempty"`
}

// Trace simulates DSU operations and emits visual execution steps.
// ponytail: circular 2D graph layout; upgrade to hierarchical tree layout if nested depth > 4.
func Trace(n int, ops []Operation) []trace.Step {
	b := trace.NewBuilder()
	if n <= 0 {
		return b.Steps()
	}

	dsu := New(n)
	nodes := initNodes(n)
	edges := make([]renderers.GraphEdge, 0, n)

	// Step 1: Initial state
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialized Disjoint Set Union with %d elements (0 to %d). Each element is its own representative root.", n, n-1),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"n":          n,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	for _, op := range ops {
		switch op.Op {
		case "union":
			traceUnion(b, dsu, nodes, &edges, op.I, op.J)
		case "find":
			traceFind(b, dsu, nodes, edges, op.I)
		case "connected":
			traceConnected(b, dsu, nodes, edges, op.I, op.J)
		}
	}

	// Final Step
	resetNodeStates(nodes)
	b.AddStep("L17", trace.ActionDone,
		fmt.Sprintf("Completed all operations. Total remaining disjoint components: %d.", dsu.Count()),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	return b.Steps()
}

func traceUnion(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, edges *[]renderers.GraphEdge, i, j int) {
	if i < 0 || i >= len(nodes) || j < 0 || j >= len(nodes) {
		return
	}

	// L11: Call Union
	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	setNodeState(nodes, j, "visiting")
	b.AddStep("L11", trace.ActionHighlight,
		fmt.Sprintf("Union(%d, %d): finding roots for %d and %d.", i, j, i, j),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(*edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	b.IncComparisons()
	rootI := dsu.Find(i)
	rootJ := dsu.Find(j)

	setNodeState(nodes, rootI, "queued")
	setNodeState(nodes, rootJ, "queued")

	// L12: Roots resolved
	b.AddStep("L12", trace.ActionCompare,
		fmt.Sprintf("Find(%d) = %d, Find(%d) = %d.", i, rootI, j, rootJ),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(*edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"root_i":     rootI,
			"root_j":     rootJ,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	// L13: Check equality
	if rootI == rootJ {
		b.AddStep("L13", trace.ActionCompare,
			fmt.Sprintf("Roots match (%d == %d). Elements %d and %d are already in the same set.", rootI, rootJ, i, j),
			renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(*edges)},
			map[string]any{
				"i":          i,
				"j":          j,
				"root_i":     rootI,
				"root_j":     rootJ,
				"redundant":  true,
				"components": dsu.Count(),
				"parent":     slices.Clone(dsu.parent),
				"rank":       slices.Clone(dsu.rank),
			},
		)
		return
	}

	b.IncOperations()
	var lineID string
	var expl string

	if dsu.rank[rootI] < dsu.rank[rootJ] {
		lineID = "L14"
		dsu.parent[rootI] = rootJ
		expl = fmt.Sprintf("rank[%d] (%d) < rank[%d] (%d): attach tree %d under root %d.", rootI, dsu.rank[rootI], rootJ, dsu.rank[rootJ], rootI, rootJ)
	} else if dsu.rank[rootI] > dsu.rank[rootJ] {
		lineID = "L15"
		dsu.parent[rootJ] = rootI
		expl = fmt.Sprintf("rank[%d] (%d) > rank[%d] (%d): attach tree %d under root %d.", rootI, dsu.rank[rootI], rootJ, dsu.rank[rootJ], rootJ, rootI)
	} else {
		lineID = "L16"
		dsu.parent[rootJ] = rootI
		dsu.rank[rootI]++
		expl = fmt.Sprintf("Ranks equal (%d): attach tree %d under %d and increment rank[%d] to %d.", dsu.rank[rootI]-1, rootJ, rootI, rootI, dsu.rank[rootI])
	}
	dsu.count--

	addEdge(edges, i, j)
	setNodeState(nodes, i, "visited")
	setNodeState(nodes, j, "visited")
	setNodeState(nodes, rootI, "visited")
	setNodeState(nodes, rootJ, "visited")

	b.AddStep(lineID, trace.ActionSet,
		expl,
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(*edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"root_i":     rootI,
			"root_j":     rootJ,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	b.AddStep("L17", trace.ActionDone,
		fmt.Sprintf("Union(%d, %d) complete. Components remaining: %d.", i, j, dsu.Count()),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(*edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)
}

func traceFind(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, edges []renderers.GraphEdge, i int) {
	if i < 0 || i >= len(nodes) {
		return
	}

	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")

	b.AddStep("L5", trace.ActionHighlight,
		fmt.Sprintf("Find(%d): traversing parent pointers to determine representative root.", i),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"i":          i,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	b.IncComparisons()
	origParent := dsu.parent[i]
	root := dsu.Find(i)

	if origParent != root {
		b.IncOperations()
		b.AddStep("L7", trace.ActionSet,
			fmt.Sprintf("Path compression: updated parent[%d] from %d directly to root %d.", i, origParent, root),
			renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
			map[string]any{
				"i":          i,
				"root":       root,
				"compressed": true,
				"components": dsu.Count(),
				"parent":     slices.Clone(dsu.parent),
				"rank":       slices.Clone(dsu.rank),
			},
		)
	}

	setNodeState(nodes, root, "path")
	b.AddStep("L9", trace.ActionDone,
		fmt.Sprintf("Find(%d) returns root %d.", i, root),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"i":          i,
			"root":       root,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)
}

func traceConnected(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, edges []renderers.GraphEdge, i, j int) {
	if i < 0 || i >= len(nodes) || j < 0 || j >= len(nodes) {
		return
	}

	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	setNodeState(nodes, j, "visiting")

	b.AddStep("L19", trace.ActionHighlight,
		fmt.Sprintf("Connected(%d, %d): checking if elements share the same root.", i, j),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	b.IncComparisons()
	rootI := dsu.Find(i)
	rootJ := dsu.Find(j)
	conn := (rootI == rootJ)

	state := "visited"
	if !conn {
		state = "queued"
	}
	setNodeState(nodes, rootI, state)
	setNodeState(nodes, rootJ, state)

	b.AddStep("L20", trace.ActionCompare,
		fmt.Sprintf("Connected(%d, %d): Find(%d)=%d, Find(%d)=%d -> %t.", i, j, i, rootI, j, rootJ, conn),
		renderers.GraphState{Nodes: cloneNodes(nodes), Edges: cloneEdges(edges)},
		map[string]any{
			"i":          i,
			"j":          j,
			"root_i":     rootI,
			"root_j":     rootJ,
			"connected":  conn,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)
}

func initNodes(n int) []renderers.GraphNode {
	nodes := make([]renderers.GraphNode, n)
	for i := 0; i < n; i++ {
		angle := 2*math.Pi*float64(i)/float64(n) - math.Pi/2
		x := 400.0 + 220.0*math.Cos(angle)
		y := 220.0 + 120.0*math.Sin(angle)
		nodes[i] = renderers.GraphNode{
			ID:    fmt.Sprintf("%d", i),
			Label: fmt.Sprintf("%d", i),
			X:     math.Round(x*10) / 10,
			Y:     math.Round(y*10) / 10,
			State: "unvisited",
		}
	}
	return nodes
}

func cloneNodes(nodes []renderers.GraphNode) []renderers.GraphNode {
	cp := make([]renderers.GraphNode, len(nodes))
	copy(cp, nodes)
	return cp
}

func cloneEdges(edges []renderers.GraphEdge) []renderers.GraphEdge {
	cp := make([]renderers.GraphEdge, len(edges))
	copy(cp, edges)
	return cp
}

func setNodeState(nodes []renderers.GraphNode, idx int, state string) {
	if idx >= 0 && idx < len(nodes) {
		nodes[idx].State = state
	}
}

func resetNodeStates(nodes []renderers.GraphNode) {
	for i := range nodes {
		nodes[i].State = "unvisited"
	}
}

func addEdge(edges *[]renderers.GraphEdge, u, v int) {
	uStr := fmt.Sprintf("%d", u)
	vStr := fmt.Sprintf("%d", v)
	for _, e := range *edges {
		if (e.From == uStr && e.To == vStr) || (e.From == vStr && e.To == uStr) {
			return
		}
	}
	*edges = append(*edges, renderers.GraphEdge{
		From:   uStr,
		To:     vStr,
		Active: false,
	})
}
