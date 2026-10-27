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
	coords   map[string][2]float64
	parents  map[string]string
}

func newTracer(root *Node) *tracer {
	tr := &tracer{
		b:       trace.NewBuilder(),
		states:  make(map[string]string),
		values:  make(map[string]int),
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
			leaves = append(leaves, n)
			return
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
	nodes := make([]renderers.TreeNode, 0, len(tr.allNodes))
	for _, n := range tr.allNodes {
		nodes = append(nodes, renderers.TreeNode{
			ID:       n.ID,
			Val:      tr.values[n.ID],
			X:        tr.coords[n.ID][0],
			Y:        tr.coords[n.ID][1],
			ParentID: tr.parents[n.ID],
			State:    tr.states[n.ID],
		})
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
	m := map[string]any{
		"alpha":  formatBound(alpha),
		"beta":   formatBound(beta),
		"node":   node.ID,
		"depth":  node.Depth,
		"cutoff": false,
	}
	if node.IsMax {
		m["player"] = "MAX"
	} else {
		m["player"] = "MIN"
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func (tr *tracer) markPruned(n *Node) {
	tr.states[n.ID] = "deleted"
	for _, child := range n.Children {
		tr.markPruned(child)
	}
}

func (tr *tracer) traceMinimax(node *Node, alpha, beta int, isMax bool) int {
	tr.states[node.ID] = "active"
	tr.b.IncOperations()

	// Check terminal
	tr.b.AddStep("L2", trace.ActionCompare,
		fmt.Sprintf("Node %s (depth %d): check if terminal.", node.ID, node.Depth),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, nil),
	)

	if len(node.Children) == 0 {
		tr.states[node.ID] = "matched"
		tr.b.AddStep("L3", trace.ActionSet,
			fmt.Sprintf("Terminal leaf %s reached with utility score %d.", node.ID, node.Val),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"eval": node.Val}),
		)
		return node.Val
	}

	if isMax {
		tr.b.AddStep("L5", trace.ActionHighlight,
			fmt.Sprintf("MAX node %s: evaluate children to maximize score.", node.ID),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, nil),
		)

		maxEval := NegInfinity
		tr.b.AddStep("L6", trace.ActionSet,
			fmt.Sprintf("MAX node %s: initialize maxEval = -∞.", node.ID),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"maxEval": "-∞"}),
		)

		for i, child := range node.Children {
			tr.b.AddStep("L7", trace.ActionVisit,
				fmt.Sprintf("MAX node %s: traversing child %s (%d of %d).", node.ID, child.ID, i+1, len(node.Children)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
			)

			tr.b.AddStep("L8", trace.ActionVisit,
				fmt.Sprintf("Invoke minimax(%s, depth=%d, α=%s, β=%s, isMax=false).", child.ID, child.Depth, formatBound(alpha), formatBound(beta)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
			)

			eval := tr.traceMinimax(child, alpha, beta, false)
			tr.states[node.ID] = "active"

			if eval > maxEval {
				maxEval = eval
			}
			tr.values[node.ID] = maxEval
			tr.b.AddStep("L9", trace.ActionSet,
				fmt.Sprintf("MAX node %s: child returned %d, maxEval is now %d.", node.ID, eval, maxEval),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"eval": eval, "maxEval": maxEval}),
			)

			if maxEval > alpha {
				alpha = maxEval
			}
			tr.b.AddStep("L10", trace.ActionSet,
				fmt.Sprintf("MAX node %s: α updated to %s.", node.ID, formatBound(alpha)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"alphaVal": alpha}),
			)

			tr.b.IncComparisons()
			tr.b.AddStep("L11", trace.ActionCompare,
				fmt.Sprintf("MAX node %s: check cutoff condition β (%s) <= α (%s).", node.ID, formatBound(beta), formatBound(alpha)),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, nil),
			)

			if beta <= alpha {
				// Prune remaining siblings
				for k := i + 1; k < len(node.Children); k++ {
					tr.markPruned(node.Children[k])
				}
				tr.b.AddStep("L11", trace.ActionHighlight,
					fmt.Sprintf("Beta cutoff! β (%s) <= α (%s). Pruned %d remaining branch(es) under %s.", formatBound(beta), formatBound(alpha), len(node.Children)-1-i, node.ID),
					tr.makeTreeState(),
					tr.vars(node, alpha, beta, map[string]any{"cutoff": true}),
				)
				break
			}
		}

		tr.states[node.ID] = "matched"
		tr.b.AddStep("L13", trace.ActionDone,
			fmt.Sprintf("MAX node %s finished: returning maxEval = %d.", node.ID, maxEval),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"result": maxEval}),
		)
		return maxEval
	}

	// Minimizer
	tr.b.AddStep("L14", trace.ActionHighlight,
		fmt.Sprintf("MIN node %s: evaluate children to minimize score.", node.ID),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, nil),
	)

	minEval := Infinity
	tr.b.AddStep("L15", trace.ActionSet,
		fmt.Sprintf("MIN node %s: initialize minEval = +∞.", node.ID),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, map[string]any{"minEval": "+∞"}),
	)

	for i, child := range node.Children {
		tr.b.AddStep("L16", trace.ActionVisit,
			fmt.Sprintf("MIN node %s: traversing child %s (%d of %d).", node.ID, child.ID, i+1, len(node.Children)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
		)

		tr.b.AddStep("L17", trace.ActionVisit,
			fmt.Sprintf("Invoke minimax(%s, depth=%d, α=%s, β=%s, isMax=true).", child.ID, child.Depth, formatBound(alpha), formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"child": child.ID}),
		)

		eval := tr.traceMinimax(child, alpha, beta, true)
		tr.states[node.ID] = "active"

		if eval < minEval {
			minEval = eval
		}
		tr.values[node.ID] = minEval
		tr.b.AddStep("L18", trace.ActionSet,
			fmt.Sprintf("MIN node %s: child returned %d, minEval is now %d.", node.ID, eval, minEval),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"eval": eval, "minEval": minEval}),
		)

		if minEval < beta {
			beta = minEval
		}
		tr.b.AddStep("L19", trace.ActionSet,
			fmt.Sprintf("MIN node %s: β updated to %s.", node.ID, formatBound(beta)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"betaVal": beta}),
		)

		tr.b.IncComparisons()
		tr.b.AddStep("L20", trace.ActionCompare,
			fmt.Sprintf("MIN node %s: check cutoff condition β (%s) <= α (%s).", node.ID, formatBound(beta), formatBound(alpha)),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, nil),
		)

		if beta <= alpha {
			// Prune remaining siblings
			for k := i + 1; k < len(node.Children); k++ {
				tr.markPruned(node.Children[k])
			}
			tr.b.AddStep("L20", trace.ActionHighlight,
				fmt.Sprintf("Alpha cutoff! β (%s) <= α (%s). Pruned %d remaining branch(es) under %s.", formatBound(beta), formatBound(alpha), len(node.Children)-1-i, node.ID),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"cutoff": true}),
			)
			break
		}
	}

	tr.states[node.ID] = "matched"
	tr.b.AddStep("L22", trace.ActionDone,
		fmt.Sprintf("MIN node %s finished: returning minEval = %d.", node.ID, minEval),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, map[string]any{"result": minEval}),
	)
	return minEval
}

// Trace executes Minimax with alpha-beta pruning on the tree and returns the execution trace.
func Trace(scores []int, branching int) []trace.Step {
	root := BuildTree(scores, branching)
	tr := newTracer(root)

	tr.b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize Minimax with Alpha-Beta Pruning. Root is MAX player (α=%s, β=%s).", formatBound(NegInfinity), formatBound(Infinity)),
		tr.makeTreeState(),
		tr.vars(root, NegInfinity, Infinity, nil),
	)

	bestVal := tr.traceMinimax(root, NegInfinity, Infinity, root.IsMax)

	// Determine optimal move from root
	var bestChild *Node
	for _, child := range root.Children {
		if child.Val == bestVal {
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
