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

const worklistLabel = "Operations Queue"

// Trace simulates DSU operations and emits visual execution steps.
// ponytail: circular 2D graph layout; upgrade to hierarchical tree layout if nested depth > 4.
func Trace(n int, ops []Operation) []trace.Step {
	b := trace.NewBuilder()
	if n <= 0 {
		return b.Steps()
	}

	dsu := New(n)
	nodes := initNodes(n)
	edges := buildForestEdges(dsu.parent, nil)

	// Step 1: Initial state
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialized Disjoint Set Union with %d elements (0 to %d). Each node points to itself as root (parent[i] = i).", n, n-1),
		cloneGraphState("Initial state: each node points to itself as root", nodes, edges, makeWorklist(ops, 0)),
		map[string]any{
			"n":          n,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	for opIdx, op := range ops {
		switch op.Op {
		case "union":
			traceUnion(b, dsu, nodes, ops, opIdx, op.I, op.J)
		case "find":
			traceFind(b, dsu, nodes, ops, opIdx, op.I)
		case "connected":
			traceConnected(b, dsu, nodes, ops, opIdx, op.I, op.J)
		}
	}

	// Final Step
	resetNodeStates(nodes)
	edges = buildForestEdges(dsu.parent, nil)
	b.AddStep("L21", trace.ActionDone,
		fmt.Sprintf("Completed all operations. Total remaining disjoint sets: %d.", dsu.Count()),
		cloneGraphState(fmt.Sprintf("Completed all operations: %d disjoint set(s)", dsu.Count()), nodes, edges, makeWorklist(ops, len(ops))),
		map[string]any{
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	return b.Steps()
}

func traceUnion(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, ops []Operation, opIdx int, i, j int) {
	if i < 0 || i >= len(nodes) || j < 0 || j >= len(nodes) {
		return
	}

	worklist := makeWorklist(ops, opIdx)
	caption := fmt.Sprintf("Union(%d, %d)", i, j)

	// L11: Call Union
	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	setNodeState(nodes, j, "visiting")
	edges := buildForestEdges(dsu.parent, nil)

	b.AddStep("L11", trace.ActionHighlight,
		fmt.Sprintf("Union(%d, %d): find representative roots for elements %d and %d.", i, j, i, j),
		cloneGraphState(caption, nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	// L12: Find roots
	b.IncComparisons()
	rootI := dsu.Find(i)
	rootJ := dsu.Find(j)

	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	setNodeState(nodes, j, "visiting")
	setNodeState(nodes, rootI, "queued")
	setNodeState(nodes, rootJ, "queued")

	edges = buildForestEdges(dsu.parent, nil)

	b.AddStep("L12", trace.ActionCompare,
		fmt.Sprintf("Find(%d) = %d, Find(%d) = %d.", i, rootI, j, rootJ),
		cloneGraphState(caption, nodes, edges, worklist),
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
		b.IncComparisons()
		sameSetCaption := fmt.Sprintf("Union(%d, %d): Already in same set", i, j)
		setNodeState(nodes, rootI, "visited")
		b.AddStep("L13", trace.ActionCompare,
			fmt.Sprintf("Roots match (%d == %d). Already in same set.", rootI, rootJ),
			cloneGraphState(sameSetCaption, nodes, edges, worklist),
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
	activeEdge := [2]int{}

	if dsu.rank[rootI] < dsu.rank[rootJ] {
		lineID = "L14"
		dsu.parent[rootI] = rootJ
		activeEdge = [2]int{rootI, rootJ}
		expl = fmt.Sprintf("rank[%d] (%d) < rank[%d] (%d): link smaller rank root %d under larger rank root %d.", rootI, dsu.rank[rootI], rootJ, dsu.rank[rootJ], rootI, rootJ)
	} else if dsu.rank[rootI] > dsu.rank[rootJ] {
		lineID = "L15"
		dsu.parent[rootJ] = rootI
		activeEdge = [2]int{rootJ, rootI}
		expl = fmt.Sprintf("rank[%d] (%d) > rank[%d] (%d): link smaller rank root %d under larger rank root %d.", rootI, dsu.rank[rootI], rootJ, dsu.rank[rootJ], rootJ, rootI)
	} else {
		lineID = "L16"
		dsu.parent[rootJ] = rootI
		dsu.rank[rootI]++
		activeEdge = [2]int{rootJ, rootI}
		expl = fmt.Sprintf("Ranks equal (%d): link root %d under %d and increment rank[%d] to %d.", dsu.rank[rootI]-1, rootJ, rootI, rootI, dsu.rank[rootI])
	}
	dsu.count--

	edges = buildForestEdges(dsu.parent, map[[2]int]bool{activeEdge: true})
	setNodeState(nodes, rootI, "visited")
	setNodeState(nodes, rootJ, "visited")

	b.AddStep(lineID, trace.ActionSet,
		expl,
		cloneGraphState(fmt.Sprintf("Union(%d, %d): linked root %d -> %d", i, j, activeEdge[0], activeEdge[1]), nodes, edges, worklist),
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

	// L17: Union complete
	edges = buildForestEdges(dsu.parent, nil)
	b.AddStep("L17", trace.ActionDone,
		fmt.Sprintf("Union(%d, %d) complete. Components remaining: %d.", i, j, dsu.Count()),
		cloneGraphState(fmt.Sprintf("Union(%d, %d) complete", i, j), nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)
}

func traceFind(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, ops []Operation, opIdx int, i int) {
	if i < 0 || i >= len(nodes) {
		return
	}

	worklist := makeWorklist(ops, opIdx)
	caption := fmt.Sprintf("Find(%d)", i)

	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	edges := buildForestEdges(dsu.parent, nil)

	// L5: Call Find
	b.AddStep("L5", trace.ActionHighlight,
		fmt.Sprintf("Find(%d): follow parent pointers upward to find representative root.", i),
		cloneGraphState(caption, nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	// L6: Traverse parent pointers upward
	b.IncComparisons()
	path := dsu.findPath(i)
	root := path[len(path)-1]

	// Highlight traversed path
	activeMap := make(map[[2]int]bool)
	for idx := 0; idx < len(path)-1; idx++ {
		setNodeState(nodes, path[idx], "visiting")
		activeMap[[2]int{path[idx], path[idx+1]}] = true
	}
	setNodeState(nodes, root, "queued")
	edges = buildForestEdges(dsu.parent, activeMap)

	b.AddStep("L6", trace.ActionCompare,
		fmt.Sprintf("Follow parent pointers: traversed path %v to root %d.", path, root),
		cloneGraphState(fmt.Sprintf("Find(%d): root is %d", i, root), nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"path":       slices.Clone(path),
			"root":       root,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	// L7: Path compression
	compressedNodes := make([]int, 0)
	compressActive := make(map[[2]int]bool)
	for idx := 0; idx < len(path)-1; idx++ {
		node := path[idx]
		if dsu.parent[node] != root {
			dsu.parent[node] = root
			compressedNodes = append(compressedNodes, node)
			compressActive[[2]int{node, root}] = true
		}
	}

	if len(compressedNodes) > 0 {
		b.IncOperations()
		edges = buildForestEdges(dsu.parent, compressActive)
		for _, cn := range compressedNodes {
			setNodeState(nodes, cn, "path")
		}
		b.AddStep("L7", trace.ActionSet,
			fmt.Sprintf("Path compression: pointing nodes directly to root %d.", root),
			cloneGraphState(fmt.Sprintf("Path compression -> root %d", root), nodes, edges, worklist),
			map[string]any{
				"i":          i,
				"root":       root,
				"compressed": true,
				"path":       slices.Clone(path),
				"components": dsu.Count(),
				"parent":     slices.Clone(dsu.parent),
				"rank":       slices.Clone(dsu.rank),
			},
		)
	}

	// L9: Return root
	resetNodeStates(nodes)
	setNodeState(nodes, root, "path")
	edges = buildForestEdges(dsu.parent, nil)

	b.AddStep("L9", trace.ActionDone,
		fmt.Sprintf("Find(%d) returns representative root %d.", i, root),
		cloneGraphState(fmt.Sprintf("Find(%d) = %d", i, root), nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"root":       root,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)
}

func traceConnected(b *trace.Builder, dsu *DSU, nodes []renderers.GraphNode, ops []Operation, opIdx int, i, j int) {
	if i < 0 || i >= len(nodes) || j < 0 || j >= len(nodes) {
		return
	}

	worklist := makeWorklist(ops, opIdx)
	captionQuery := fmt.Sprintf("Connected(%d, %d)?", i, j)

	resetNodeStates(nodes)
	setNodeState(nodes, i, "visiting")
	setNodeState(nodes, j, "visiting")
	edges := buildForestEdges(dsu.parent, nil)

	// L19: Call Connected
	b.AddStep("L19", trace.ActionHighlight,
		fmt.Sprintf("Connected(%d, %d): checking if elements share the same representative root.", i, j),
		cloneGraphState(captionQuery, nodes, edges, worklist),
		map[string]any{
			"i":          i,
			"j":          j,
			"components": dsu.Count(),
			"parent":     slices.Clone(dsu.parent),
			"rank":       slices.Clone(dsu.rank),
		},
	)

	// L20: Return Find(i) == Find(j)
	b.IncComparisons()
	rootI := dsu.Find(i)
	rootJ := dsu.Find(j)
	conn := (rootI == rootJ)

	resetNodeStates(nodes)
	if conn {
		setNodeState(nodes, rootI, "visited")
		setNodeState(nodes, i, "visited")
		setNodeState(nodes, j, "visited")
	} else {
		setNodeState(nodes, rootI, "queued")
		setNodeState(nodes, rootJ, "queued")
		setNodeState(nodes, i, "visiting")
		setNodeState(nodes, j, "visiting")
	}

	edges = buildForestEdges(dsu.parent, nil)
	captionOutcome := fmt.Sprintf("Connected(%d, %d) = %t", i, j, conn)

	expl := fmt.Sprintf("Connected(%d, %d): Find(%d)=%d, Find(%d)=%d -> %t.", i, j, i, rootI, j, rootJ, conn)
	if conn {
		expl += fmt.Sprintf(" Elements %d and %d belong to the same set.", i, j)
	} else {
		expl += fmt.Sprintf(" Elements %d and %d belong to different sets.", i, j)
	}

	b.AddStep("L20", trace.ActionCompare,
		expl,
		cloneGraphState(captionOutcome, nodes, edges, worklist),
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

func buildForestEdges(parent []int, activeEdges map[[2]int]bool) []renderers.GraphEdge {
	edges := make([]renderers.GraphEdge, 0, len(parent))
	for u, p := range parent {
		if u != p {
			active := false
			if activeEdges != nil && activeEdges[[2]int{u, p}] {
				active = true
			}
			edges = append(edges, renderers.GraphEdge{
				From:     fmt.Sprintf("%d", u),
				To:       fmt.Sprintf("%d", p),
				Directed: true,
				Active:   active,
			})
		}
	}
	return edges
}

func cloneGraphState(caption string, nodes []renderers.GraphNode, edges []renderers.GraphEdge, worklist []renderers.WorklistItem) renderers.GraphState {
	clonedNodes := make([]renderers.GraphNode, len(nodes))
	copy(clonedNodes, nodes)

	clonedEdges := make([]renderers.GraphEdge, len(edges))
	copy(clonedEdges, edges)

	clonedWorklist := make([]renderers.WorklistItem, len(worklist))
	copy(clonedWorklist, worklist)

	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: worklistLabel,
		Worklist:      clonedWorklist,
		Nodes:         clonedNodes,
		Edges:         clonedEdges,
	}
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

func opLabel(op Operation) string {
	switch op.Op {
	case "union":
		return fmt.Sprintf("Union(%d, %d)", op.I, op.J)
	case "find":
		return fmt.Sprintf("Find(%d)", op.I)
	case "connected":
		return fmt.Sprintf("Connected(%d, %d)", op.I, op.J)
	default:
		return op.Op
	}
}

func makeWorklist(ops []Operation, fromIdx int) []renderers.WorklistItem {
	if fromIdx >= len(ops) {
		return []renderers.WorklistItem{}
	}
	items := make([]renderers.WorklistItem, 0, len(ops)-fromIdx)
	for i := fromIdx; i < len(ops); i++ {
		items = append(items, renderers.WorklistItem{
			ID:    fmt.Sprintf("op-%d", i),
			Label: opLabel(ops[i]),
		})
	}
	return items
}
