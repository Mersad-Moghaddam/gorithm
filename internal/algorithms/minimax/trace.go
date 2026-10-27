package minimax

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type tracer struct {
	b        *trace.Builder
	allNodes []*Node
	states   map[string]string
	values   map[string]int
	labels   map[string]string
	coords   map[string][2]float64
	parents  map[string]string
}

func newTracer(root *Node) *tracer {
	tr := &tracer{
		b:       trace.NewBuilder(),
		states:  make(map[string]string),
		values:  make(map[string]int),
		labels:  make(map[string]string),
		coords:  make(map[string][2]float64),
		parents: make(map[string]string),
	}
	tr.initTree(root)
	return tr
}

func (tr *tracer) initTree(root *Node) {
	if root == nil {
		return
	}

	var leaves []*Node
	var collect func(n *Node)
	collect = func(n *Node) {
		tr.allNodes = append(tr.allNodes, n)
		tr.states[n.ID] = "normal"
		tr.values[n.ID] = n.Val
		if len(n.Children) == 0 {
			tr.labels[n.ID] = fmt.Sprintf("%d", n.Val)
			leaves = append(leaves, n)
			return
		}
		if n.IsMax {
			tr.labels[n.ID] = "MAX"
		} else {
			tr.labels[n.ID] = "MIN"
		}
		for _, child := range n.Children {
			tr.parents[child.ID] = n.ID
			collect(child)
		}
	}
	collect(root)

	// Layout leaves evenly spaced
	nLeaves := len(leaves)
	for i, leaf := range leaves {
		x := 400.0
		if nLeaves > 1 {
			x = 80.0 + float64(i)*(640.0/float64(nLeaves-1))
		}
		y := 60.0 + float64(leaf.Depth)*85.0
		tr.coords[leaf.ID] = [2]float64{x, y}
	}

	// Layout internal nodes bottom-up
	var layoutInternal func(n *Node) float64
	layoutInternal = func(n *Node) float64 {
		if len(n.Children) == 0 {
			return tr.coords[n.ID][0]
		}
		var firstX, lastX float64
		for i, child := range n.Children {
			cx := layoutInternal(child)
			if i == 0 {
				firstX = cx
			}
			if i == len(n.Children)-1 {
				lastX = cx
			}
		}
		midX := (firstX + lastX) / 2.0
		y := 60.0 + float64(n.Depth)*85.0
		tr.coords[n.ID] = [2]float64{midX, y}
		return midX
	}
	layoutInternal(root)
}

func (tr *tracer) makeTreeState() renderers.TreeState {
	nodes := make([]renderers.TreeNode, len(tr.allNodes))
	for i, n := range tr.allNodes {
		nodes[i] = renderers.TreeNode{
			ID:       n.ID,
			Val:      tr.values[n.ID],
			Label:    tr.labels[n.ID],
			X:        tr.coords[n.ID][0],
			Y:        tr.coords[n.ID][1],
			ParentID: tr.parents[n.ID],
			State:    tr.states[n.ID],
		}
	}
	return renderers.TreeState{Nodes: nodes}
}

func formatBound(v int) string {
	if v <= -1000000 {
		return "-∞"
	}
	if v >= 1000000 {
		return "+∞"
	}
	return fmt.Sprintf("%d", v)
}

func (tr *tracer) vars(node *Node, alpha, beta int, extra map[string]any) map[string]any {
	boundsStr := fmt.Sprintf("[%s, %s]", formatBound(alpha), formatBound(beta))
	player := "MAX"
	if !node.IsMax {
		player = "MIN"
	}
	m := map[string]any{
		"alpha":         formatBound(alpha),
		"beta":          formatBound(beta),
		"bounds":        boundsStr,
		"[alpha, beta]": boundsStr,
		"node":          node.ID,
		"depth":         node.Depth,
		"player":        player,
		"turn":          player,
		"cutoff":        false,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func (tr *tracer) markPruned(n *Node) {
	tr.states[n.ID] = "deleted"
	tr.labels[n.ID] = "✕"
	for _, child := range n.Children {
		tr.markPruned(child)
	}
}

func (tr *tracer) traceMinimax(node *Node, alpha, beta int, isMax bool) int {
	tr.states[node.ID] = "active"
	tr.b.IncOperations()

	player := "MAX"
	if !isMax {
		player = "MIN"
	}

	// Check terminal
	tr.b.AddStep("L2", trace.ActionCompare,
		fmt.Sprintf("Node %s (depth %d, %s turn, bounds [α, β] = [%s, %s]): check if terminal.", node.ID, node.Depth, player, formatBound(alpha), formatBound(beta)),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, nil),
	)

	if len(node.Children) == 0 {
		tr.states[node.ID] = "matched"
		tr.labels[node.ID] = fmt.Sprintf("%d", node.Val)
		tr.b.AddStep("L3", trace.ActionSet,
			fmt.Sprintf("Terminal leaf %s reached with utility score %d [α, β] = [%s, %s].", node.ID, node.Val, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"eval": node.Val}),
		)
		return node.Val
	}

	if isMax {
		tr.b.AddStep("L5", trace.ActionHighlight,
			fmt.Sprintf("MAX node %s (depth %d, bounds [α, β] = [%s, %s]): evaluate children to maximize score.", node.ID, node.Depth, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, nil),
		)

		maxEval := NegInfinity
		tr.b.AddStep("L6", trace.ActionSet,
			fmt.Sprintf("MAX node %s: initialize maxEval = -∞ with bounds [α, β] = [%s, %s].", node.ID, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"maxEval": "-∞"}),
		)

		for i, child := range node.Children {
			tr.b.AddStep("L7", trace.ActionVisit,
				fmt.Sprintf("MAX node %s: traversing child %s (%d of %d) with bounds [α, β] = [%s, %s].", node.ID, child.ID, i+1, len(node.Children), formatBound(alpha), formatBound(beta)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
			)

			// Highlight child as active when descending
			tr.states[child.ID] = "active"
			tr.b.AddStep("L8", trace.ActionVisit,
				fmt.Sprintf("Invoke minimax(%s, depth=%d, α=%s, β=%s, isMax=false). Child %s is active.", child.ID, child.Depth, formatBound(alpha), formatBound(beta), child.ID),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
			)

			eval := tr.traceMinimax(child, alpha, beta, false)
			tr.states[node.ID] = "active"

			if eval > maxEval {
				maxEval = eval
			}
			node.Val = maxEval
			tr.values[node.ID] = maxEval
			tr.labels[node.ID] = fmt.Sprintf("%d", maxEval)
			tr.b.AddStep("L9", trace.ActionSet,
				fmt.Sprintf("MAX node %s: child %s returned score %d. Seeks max(value, child_val) -> maxEval = %d [α, β] = [%s, %s].", node.ID, child.ID, eval, maxEval, formatBound(alpha), formatBound(beta)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"eval": eval, "maxEval": maxEval}),
			)

			if maxEval > alpha {
				alpha = maxEval
			}
			tr.b.AddStep("L10", trace.ActionSet,
				fmt.Sprintf("MAX node %s: update alpha = max(alpha, value) -> α = %s [α, β] = [%s, %s].", node.ID, formatBound(alpha), formatBound(alpha), formatBound(beta)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"alphaVal": alpha}),
			)

			tr.b.IncComparisons()
			tr.b.AddStep("L11", trace.ActionCompare,
				fmt.Sprintf("MAX node %s: check cutoff condition β (%s) <= α (%s) with bounds [α, β] = [%s, %s].", node.ID, formatBound(beta), formatBound(alpha), formatBound(alpha), formatBound(beta)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, nil),
			)

			if beta <= alpha {
				// Prune remaining siblings
				for k := i + 1; k < len(node.Children); k++ {
					tr.markPruned(node.Children[k])
				}
				tr.b.AddStep("L11", trace.ActionHighlight,
					fmt.Sprintf("Alpha-Beta Cutoff: beta <= alpha, branch pruned! β (%s) <= α (%s) under MAX node %s. Pruned %d remaining branch(es).", formatBound(beta), formatBound(alpha), node.ID, len(node.Children)-1-i),
					tr.makeTreeState(),
					tr.vars(node, alpha, beta, map[string]any{"cutoff": true}),
				)
				break
			}
		}

		tr.states[node.ID] = "matched"
		tr.b.AddStep("L13", trace.ActionDone,
			fmt.Sprintf("MAX node %s finished: propagating score %d back up to root/parent [α, β] = [%s, %s].", node.ID, maxEval, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"result": maxEval}),
		)
		return maxEval
	}

	// Minimizer
	tr.b.AddStep("L14", trace.ActionHighlight,
		fmt.Sprintf("MIN node %s (depth %d, bounds [α, β] = [%s, %s]): evaluate children to minimize score.", node.ID, node.Depth, formatBound(alpha), formatBound(beta)),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, nil),
	)

	minEval := Infinity
	tr.b.AddStep("L15", trace.ActionSet,
		fmt.Sprintf("MIN node %s: initialize minEval = +∞ with bounds [α, β] = [%s, %s].", node.ID, formatBound(alpha), formatBound(beta)),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, map[string]any{"minEval": "+∞"}),
	)

	for i, child := range node.Children {
		tr.b.AddStep("L16", trace.ActionVisit,
			fmt.Sprintf("MIN node %s: traversing child %s (%d of %d) with bounds [α, β] = [%s, %s].", node.ID, child.ID, i+1, len(node.Children), formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
		)

		// Highlight child as active when descending
		tr.states[child.ID] = "active"
		tr.b.AddStep("L17", trace.ActionVisit,
			fmt.Sprintf("Invoke minimax(%s, depth=%d, α=%s, β=%s, isMax=true). Child %s is active.", child.ID, child.Depth, formatBound(alpha), formatBound(beta), child.ID),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
		)

		eval := tr.traceMinimax(child, alpha, beta, true)
		tr.states[node.ID] = "active"

		if eval < minEval {
			minEval = eval
		}
		node.Val = minEval
		tr.values[node.ID] = minEval
		tr.labels[node.ID] = fmt.Sprintf("%d", minEval)
		tr.b.AddStep("L18", trace.ActionSet,
			fmt.Sprintf("MIN node %s: child %s returned score %d. Seeks min(value, child_val) -> minEval = %d [α, β] = [%s, %s].", node.ID, child.ID, eval, minEval, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"eval": eval, "minEval": minEval}),
		)

		if minEval < beta {
			beta = minEval
		}
		tr.b.AddStep("L19", trace.ActionSet,
			fmt.Sprintf("MIN node %s: update beta = min(beta, value) -> β = %s [α, β] = [%s, %s].", node.ID, formatBound(beta), formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"betaVal": beta}),
		)

		tr.b.IncComparisons()
		tr.b.AddStep("L20", trace.ActionCompare,
			fmt.Sprintf("MIN node %s: check cutoff condition β (%s) <= α (%s) with bounds [α, β] = [%s, %s].", node.ID, formatBound(beta), formatBound(alpha), formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, nil),
		)

		if beta <= alpha {
			// Prune remaining siblings
			for k := i + 1; k < len(node.Children); k++ {
				tr.markPruned(node.Children[k])
			}
			tr.b.AddStep("L20", trace.ActionHighlight,
				fmt.Sprintf("Alpha-Beta Cutoff: beta <= alpha, branch pruned! β (%s) <= α (%s) under MIN node %s. Pruned %d remaining branch(es).", formatBound(beta), formatBound(alpha), node.ID, len(node.Children)-1-i),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"cutoff": true}),
			)
			break
		}
	}

	tr.states[node.ID] = "matched"
	tr.b.AddStep("L22", trace.ActionDone,
		fmt.Sprintf("MIN node %s finished: propagating score %d back up to parent [α, β] = [%s, %s].", node.ID, minEval, formatBound(alpha), formatBound(beta)),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, map[string]any{"result": minEval}),
	)
	return minEval
}

// Trace executes Minimax with alpha-beta pruning on the tree and returns the execution trace.
// ponytail: alpha-beta pruning with tree state snapshots; transposition tables and iterative deepening needed for deep game trees.
func Trace(scores []int, branching int) []trace.Step {
	root := BuildTree(scores, branching)
	tr := newTracer(root)

	tr.b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize Minimax with Alpha-Beta Pruning. Root is MAX player with bounds [α, β] = [%s, %s].", formatBound(NegInfinity), formatBound(Infinity)),
		tr.makeTreeState(),
		tr.vars(root, NegInfinity, Infinity, nil),
	)

	bestVal := tr.traceMinimax(root, NegInfinity, Infinity, root.IsMax)

	// Determine optimal move from root
	var bestChild *Node
	for _, child := range root.Children {
		if tr.values[child.ID] == bestVal {
			bestChild = child
			break
		}
	}

	tr.states[root.ID] = "inserted"
	if bestChild != nil {
		tr.states[bestChild.ID] = "inserted"
	}

	optID := ""
	if bestChild != nil {
		optID = bestChild.ID
	}

	tr.b.AddStep("L24", trace.ActionDone,
		fmt.Sprintf("Minimax search completed! Optimal root score is %d (chosen move: %s).", bestVal, optID),
		tr.makeTreeState(),
		tr.vars(root, NegInfinity, Infinity, map[string]any{
			"optimalMove": optID,
			"bestValue":   bestVal,
		}),
	)

	return tr.b.Steps()
}
