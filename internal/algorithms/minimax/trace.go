package minimax

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type tracer struct {
	b         *minimaxBuilder
	allNodes  []*Node
	states    map[string]string
	values    map[string]int
	coords    map[string][2]float64
	parents   map[string]string
	stack     []string
	evaluated map[string]bool
	bestChild map[string]string
}

func newTracer(root *Node) *tracer {
	tr := &tracer{
		states:    make(map[string]string),
		values:    make(map[string]int),
		coords:    make(map[string][2]float64),
		parents:   make(map[string]string),
		evaluated: make(map[string]bool),
		bestChild: make(map[string]string),
	}
	tr.b = &minimaxBuilder{Builder: trace.NewBuilder(), tr: tr}
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
	if v == NegInfinity {
		return "-∞"
	}
	if v == Infinity {
		return "+∞"
	}
	return fmt.Sprintf("%d", v)
}

func (tr *tracer) vars(node *Node, alpha, beta int, extra map[string]any) map[string]any {
	m := map[string]any{
		"alpha":      formatBound(alpha),
		"beta":       formatBound(beta),
		"node":       node.ID,
		"depth":      node.Depth,
		"cutoff":     false,
		"alphaValue": alpha,
		"betaValue":  beta,
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

// minimaxBuilder captures typed provenance at the same instant as the source event.
type minimaxBuilder struct {
	*trace.Builder
	tr *tracer
}

func (b *minimaxBuilder) AddStep(line string, action trace.ActionType, explanation string, state any, vars map[string]any) {
	tr := b.tr
	s := renderers.MinimaxLessonState{TreeState: state.(renderers.TreeState), Stack: append([]string(nil), tr.stack...)}
	s.ActiveID, _ = vars["node"].(string)
	if len(s.Stack) == 0 {
		s.Stack = []string{s.ActiveID}
	}
	s.ChildID, _ = vars["child"].(string)
	s.ReturnedFromID, _ = vars["returnedFrom"].(string)
	s.Alpha, _ = vars["alphaValue"].(int)
	s.Beta, _ = vars["betaValue"].(int)
	s.Cutoff, _ = vars["cutoff"].(bool)
	s.Eval, s.EvalValid = vars["eval"].(int)
	s.BestChildID = tr.bestChild[s.ActiveID]
	s.BestValue = tr.values[s.ActiveID]
	s.BestValueValid = tr.evaluated[s.ActiveID]
	s.Phase = "decision"
	switch line {
	case "L1":
		s.Phase = "init"
	case "L2":
		s.Phase = "terminal-check"
	case "L3":
		s.Phase = "leaf-return"
	case "L6", "L15":
		s.Phase = "initialize-best"
	case "L7", "L8", "L16", "L17":
		s.Phase = "descend"
	case "L9", "L18":
		s.Phase = "child-return"
	case "L10", "L19":
		s.Phase = "bound-update"
	case "L11", "L20":
		s.Phase = "cutoff-check"
	case "L13", "L22":
		s.Phase = "return"
	case "L24":
		s.Phase = "done"
	}
	if s.Cutoff {
		s.Phase = "prune"
	}
	for _, n := range tr.allNodes {
		pruned := tr.states[n.ID] == "deleted"
		s.Nodes = append(s.Nodes, renderers.MinimaxNodeView{ID: n.ID, ParentID: tr.parents[n.ID], Depth: n.Depth, Value: tr.values[n.ID], IsMax: n.IsMax, Terminal: len(n.Children) == 0, Evaluated: tr.evaluated[n.ID], Pruned: pruned})
		if pruned {
			s.PrunedIDs = append(s.PrunedIDs, n.ID)
		}
	}
	b.Builder.AddStep(line, action, explanation, s, vars)
}

func (tr *tracer) markPruned(n *Node) {
	tr.states[n.ID] = "deleted"
	for _, child := range n.Children {
		tr.markPruned(child)
	}
}

func (tr *tracer) traceMinimax(node *Node, alpha, beta int, isMax bool) int {
	tr.stack = append(tr.stack, node.ID)
	defer func() { tr.stack = tr.stack[:len(tr.stack)-1] }()
	tr.states[node.ID] = "active"
	tr.b.IncOperations()

	// Check terminal
	tr.b.AddStep("L2", trace.ActionCompare,
		fmt.Sprintf("Node %s (depth %d): check if terminal.", node.ID, node.Depth),
		tr.makeTreeState(),
		tr.vars(node, alpha, beta, nil),
	)

	if len(node.Children) == 0 {
		tr.evaluated[node.ID] = true
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
				tr.bestChild[node.ID] = child.ID
			}
			tr.evaluated[node.ID] = true
			tr.values[node.ID] = maxEval
			tr.b.AddStep("L9", trace.ActionSet,
				fmt.Sprintf("MAX node %s: child returned %d, maxEval is now %d.", node.ID, eval, maxEval),
				tr.makeTreeState(),
				tr.vars(node, alpha, beta, map[string]any{"eval": eval, "maxEval": maxEval, "returnedFrom": child.ID}),
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
			tr.bestChild[node.ID] = child.ID
		}
		tr.evaluated[node.ID] = true
		tr.values[node.ID] = minEval
		tr.b.AddStep("L18", trace.ActionSet,
			fmt.Sprintf("MIN node %s: child returned %d, minEval is now %d.", node.ID, eval, minEval),
			tr.makeTreeState(),
			tr.vars(node, alpha, beta, map[string]any{"eval": eval, "minEval": minEval, "returnedFrom": child.ID}),
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
	if err := validateMinimaxInput(scores, branching); err != nil && len(scores) != 0 {
		b := trace.NewBuilder()
		b.AddStep("L1", trace.ActionDone, err.Error(), renderers.MinimaxLessonState{Phase: "invalid-input"}, map[string]any{"error": err.Error()})
		return b.Steps()
	}
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
		if tr.evaluated[child.ID] && tr.values[child.ID] == bestVal {
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
		tr.vars(root, bestVal, Infinity, map[string]any{
			"optimalMove": optID,
			"bestValue":   bestVal,
		}),
	)

	steps := tr.b.Steps()
	if len(scores) == 0 {
		steps[len(steps)-1].Explanation = "Empty input: terminal root utility 0; no game move."
	}
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L2" && steps[i+1].LineID == "L3" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
