package avl

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type tracer struct {
	b         *trace.Builder
	nodeIDs   map[*Node]string
	nodeIDSeq int
	root      *Node
}

func newTracer() *tracer {
	return &tracer{
		b:       trace.NewBuilder(),
		nodeIDs: make(map[*Node]string),
	}
}

func (tr *tracer) idFor(n *Node) string {
	if n == nil {
		return ""
	}
	if id, ok := tr.nodeIDs[n]; ok {
		return id
	}
	tr.nodeIDSeq++
	id := fmt.Sprintf("node-%d", tr.nodeIDSeq)
	tr.nodeIDs[n] = id
	return id
}

func (tr *tracer) treeState(states map[*Node]string) renderers.TreeState {
	if tr.root == nil {
		return renderers.TreeState{Nodes: make([]renderers.TreeNode, 0)}
	}

	var nodes []renderers.TreeNode
	var layout func(n *Node, x, y, dx float64, parentID string)
	layout = func(n *Node, x, y, dx float64, parentID string) {
		if n == nil {
			return
		}
		id := tr.idFor(n)
		st := "normal"
		if states != nil {
			if s, ok := states[n]; ok {
				st = s
			}
		}
		nodes = append(nodes, renderers.TreeNode{
			ID:       id,
			Val:      n.Val,
			X:        x,
			Y:        y,
			ParentID: parentID,
			State:    st,
		})

		nextDx := dx / 2
		if nextDx < 16 {
			nextDx = 16
		}

		if n.Left != nil {
			lx := x - dx
			if lx < 30 {
				lx = 30
			}
			layout(n.Left, lx, y+70, nextDx, id)
		}
		if n.Right != nil {
			rx := x + dx
			if rx > 770 {
				rx = 770
			}
			layout(n.Right, rx, y+70, nextDx, id)
		}
	}

	layout(tr.root, 400, 60, 180, "")
	cloned := make([]renderers.TreeNode, len(nodes))
	copy(cloned, nodes)
	return renderers.TreeState{Nodes: cloned}
}

func (tr *tracer) traceRightRotate(y *Node, setParent func(*Node)) *Node {
	x := y.Left
	t2 := x.Right

	tr.b.IncOperations()
	tr.b.AddStep("L19", trace.ActionSwap,
		fmt.Sprintf("Right Rotation (LL): pivot node %d rises to subtree root; root %d descends right. Subtree between them attaches as left child of %d, restoring balance while preserving BST order (%d < %d).", x.Val, y.Val, y.Val, x.Val, y.Val),
		tr.treeState(map[*Node]string{y: "active", x: "active"}),
		map[string]any{
			"rotation":      "right",
			"case":          "LL",
			"pivot":         x.Val,
			"descending":    y.Val,
			"criticalNode":  y.Val,
			"newRoot":       x.Val,
			"bf":            BalanceFactor(y),
			"balanceFactor": BalanceFactor(y),
		})

	x.Right = y
	y.Left = t2

	updateHeight(y)
	updateHeight(x)
	setParent(x)

	tr.b.AddStep("L20", trace.ActionSet,
		fmt.Sprintf("Right Rotation complete: restored balance with root %d (bf=%d: left.Height(%d) - right.Height(%d)), right child %d (bf=%d). BST invariant preserved.", x.Val, BalanceFactor(x), height(x.Left), height(x.Right), y.Val, BalanceFactor(y)),
		tr.treeState(map[*Node]string{x: "matched", y: "normal"}),
		map[string]any{
			"root":          x.Val,
			"height":        x.Height,
			"bf":            BalanceFactor(x),
			"balanceFactor": BalanceFactor(x),
			"pivot":         x.Val,
			"descended":     y.Val,
		})

	return x
}

func (tr *tracer) traceLeftRotate(x *Node, setParent func(*Node)) *Node {
	y := x.Right
	t2 := y.Left

	tr.b.IncOperations()
	tr.b.AddStep("L23", trace.ActionSwap,
		fmt.Sprintf("Left Rotation (RR): pivot node %d rises to subtree root; root %d descends left. Subtree between them attaches as right child of %d, restoring balance while preserving BST order (%d < %d).", y.Val, x.Val, x.Val, x.Val, y.Val),
		tr.treeState(map[*Node]string{x: "active", y: "active"}),
		map[string]any{
			"rotation":      "left",
			"case":          "RR",
			"pivot":         y.Val,
			"descending":    x.Val,
			"criticalNode":  x.Val,
			"newRoot":       y.Val,
			"bf":            BalanceFactor(x),
			"balanceFactor": BalanceFactor(x),
		})

	y.Left = x
	x.Right = t2

	updateHeight(x)
	updateHeight(y)
	setParent(y)

	tr.b.AddStep("L24", trace.ActionSet,
		fmt.Sprintf("Left Rotation complete: restored balance with root %d (bf=%d: left.Height(%d) - right.Height(%d)), left child %d (bf=%d). BST invariant preserved.", y.Val, BalanceFactor(y), height(y.Left), height(y.Right), x.Val, BalanceFactor(x)),
		tr.treeState(map[*Node]string{y: "matched", x: "normal"}),
		map[string]any{
			"root":          y.Val,
			"height":        y.Height,
			"bf":            BalanceFactor(y),
			"balanceFactor": BalanceFactor(y),
			"pivot":         y.Val,
			"descended":     x.Val,
		})

	return y
}

func (tr *tracer) insertHelper(node *Node, val int, setParent func(*Node)) *Node {
	if node == nil {
		newNode := &Node{Val: val, Height: 1}
		setParent(newNode)
		tr.b.IncOperations()
		tr.b.AddStep("L2", trace.ActionPush, fmt.Sprintf("Empty spot found. Insert node %d with initial height 1 (bf = 0: left.Height(0) - right.Height(0)).", val),
			tr.treeState(map[*Node]string{newNode: "inserted"}),
			map[string]any{"val": val, "height": 1, "bf": 0, "balanceFactor": 0, "leftHeight": 0, "rightHeight": 0})
		return newNode
	}

	tr.b.IncComparisons()
	tr.b.AddStep("L1", trace.ActionCompare, fmt.Sprintf("Compare target %d with node %d (bf=%d: left.Height(%d) - right.Height(%d)).", val, node.Val, BalanceFactor(node), height(node.Left), height(node.Right)),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"val": val, "curr": node.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node), "height": node.Height})

	if val < node.Val {
		tr.b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("%d < %d: traverse into left subtree.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val, "direction": "left"})
		node.Left = tr.insertHelper(node.Left, val, func(n *Node) { node.Left = n })
	} else if val > node.Val {
		tr.b.AddStep("L4", trace.ActionCompare, fmt.Sprintf("%d > %d: traverse into right subtree.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val, "direction": "right"})
		node.Right = tr.insertHelper(node.Right, val, func(n *Node) { node.Right = n })
	} else {
		tr.b.AddStep("L5", trace.ActionDone, fmt.Sprintf("Value %d already exists. Skip duplicate insert.", val),
			tr.treeState(map[*Node]string{node: "matched"}),
			map[string]any{"val": val, "curr": node.Val})
		return node
	}

	updateHeight(node)
	tr.b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Unwinding to node %d: updated height to %d (1 + max(left:%d, right:%d)).", node.Val, node.Height, height(node.Left), height(node.Right)),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"node": node.Val, "height": node.Height, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node), "leftHeight": height(node.Left), "rightHeight": height(node.Right)})

	bf := BalanceFactor(node)
	tr.b.IncComparisons()
	tr.b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("Node %d balance factor = %d (bf = left.Height(%d) - right.Height(%d)).", node.Val, bf, height(node.Left), height(node.Right)),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"node": node.Val, "bf": bf, "balanceFactor": bf, "height": node.Height, "leftHeight": height(node.Left), "rightHeight": height(node.Right)})

	// Case 1: Left-Left (LL) -> Right Rotation
	if bf > 1 && val < node.Left.Val {
		tr.b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Imbalance at critical node %d: |bf|=%d > 1 (bf = left.Height(%d) - right.Height(%d)). Inserted %d < left child %d indicates Left-Left (LL) case. Triggers Right Rotation.", node.Val, bf, height(node.Left), height(node.Right), val, node.Left.Val),
			tr.treeState(map[*Node]string{node: "active", node.Left: "active"}),
			map[string]any{"node": node.Val, "case": "LL", "criticalNode": node.Val, "pivot": node.Left.Val, "bf": bf, "balanceFactor": bf})
		return tr.traceRightRotate(node, setParent)
	}

	// Case 2: Right-Right (RR) -> Left Rotation
	if bf < -1 && val > node.Right.Val {
		tr.b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Imbalance at critical node %d: |bf|=%d > 1 (bf = left.Height(%d) - right.Height(%d)). Inserted %d > right child %d indicates Right-Right (RR) case. Triggers Left Rotation.", node.Val, -bf, height(node.Left), height(node.Right), val, node.Right.Val),
			tr.treeState(map[*Node]string{node: "active", node.Right: "active"}),
			map[string]any{"node": node.Val, "case": "RR", "criticalNode": node.Val, "pivot": node.Right.Val, "bf": bf, "balanceFactor": bf})
		return tr.traceLeftRotate(node, setParent)
	}

	// Case 3: Left-Right (LR) -> Left Rotate Left Child, then Right Rotate Root
	if bf > 1 && val > node.Left.Val {
		tr.b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Imbalance at critical node %d: |bf|=%d > 1 (bf = left.Height(%d) - right.Height(%d)). Inserted %d > left child %d indicates Left-Right (LR) case. Triggers Double Rotation (child %d rotates Left first, then root %d rotates Right).", node.Val, bf, height(node.Left), height(node.Right), val, node.Left.Val, node.Left.Val, node.Val),
			tr.treeState(map[*Node]string{node: "active", node.Left: "active"}),
			map[string]any{"node": node.Val, "case": "LR", "criticalNode": node.Val, "child": node.Left.Val, "bf": bf, "balanceFactor": bf})

		tr.b.AddStep("L11", trace.ActionSwap, fmt.Sprintf("LR Double Rotation (Step 1/2): Left-rotate left child %d first to straighten zig-zag into LL line while preserving BST order.", node.Left.Val),
			tr.treeState(map[*Node]string{node.Left: "active"}),
			map[string]any{"case": "LR", "step": "LR-1", "child": node.Left.Val, "criticalNode": node.Val, "bf": BalanceFactor(node.Left), "balanceFactor": BalanceFactor(node.Left)})
		node.Left = tr.traceLeftRotate(node.Left, func(n *Node) { node.Left = n })

		tr.b.AddStep("L11", trace.ActionSwap, fmt.Sprintf("LR Double Rotation (Step 2/2): Now in LL form, Right-rotate critical root %d. Pivot rises to subtree root and restores balance while preserving BST order.", node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"case": "LR", "step": "LR-2", "criticalNode": node.Val, "pivot": node.Left.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})
		return tr.traceRightRotate(node, setParent)
	}

	// Case 4: Right-Left (RL) -> Right Rotate Right Child, then Left Rotate Root
	if bf < -1 && val < node.Right.Val {
		tr.b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Imbalance at critical node %d: |bf|=%d > 1 (bf = left.Height(%d) - right.Height(%d)). Inserted %d < right child %d indicates Right-Left (RL) case. Triggers Double Rotation (child %d rotates Right first, then root %d rotates Left).", node.Val, -bf, height(node.Left), height(node.Right), val, node.Right.Val, node.Right.Val, node.Val),
			tr.treeState(map[*Node]string{node: "active", node.Right: "active"}),
			map[string]any{"node": node.Val, "case": "RL", "criticalNode": node.Val, "child": node.Right.Val, "bf": bf, "balanceFactor": bf})

		tr.b.AddStep("L14", trace.ActionSwap, fmt.Sprintf("RL Double Rotation (Step 1/2): Right-rotate right child %d first to straighten zig-zag into RR line while preserving BST order.", node.Right.Val),
			tr.treeState(map[*Node]string{node.Right: "active"}),
			map[string]any{"case": "RL", "step": "RL-1", "child": node.Right.Val, "criticalNode": node.Val, "bf": BalanceFactor(node.Right), "balanceFactor": BalanceFactor(node.Right)})
		node.Right = tr.traceRightRotate(node.Right, func(n *Node) { node.Right = n })

		tr.b.AddStep("L14", trace.ActionSwap, fmt.Sprintf("RL Double Rotation (Step 2/2): Now in RR form, Left-rotate critical root %d. Pivot rises to subtree root and restores balance while preserving BST order.", node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"case": "RL", "step": "RL-2", "criticalNode": node.Val, "pivot": node.Right.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})
		return tr.traceLeftRotate(node, setParent)
	}

	tr.b.AddStep("L16", trace.ActionDone, fmt.Sprintf("Node %d is balanced (bf=%d in [-1, 1], bf = left.Height(%d) - right.Height(%d)). AVL invariant holds.", node.Val, bf, height(node.Left), height(node.Right)),
		tr.treeState(nil),
		map[string]any{"node": node.Val, "bf": bf, "balanceFactor": bf, "height": node.Height, "leftHeight": height(node.Left), "rightHeight": height(node.Right)})

	return node
}

// Trace executes AVL inserts with visualization steps.
// ponytail: recursive trace generator; flat iterator needed if trace step count exceeds memory for n > 500.
func Trace(values []int) []trace.Step {
	tr := newTracer()
	if len(values) == 0 {
		values = []int{10, 20, 30, 40, 50, 25}
	}

	tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("AVL Tree initialized. Inserting values: %v.", values),
		tr.treeState(nil),
		map[string]any{"values": values})

	for _, v := range values {
		tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Begin Insert(%d).", v),
			tr.treeState(nil),
			map[string]any{"val": v, "op": "insert"})
		tr.root = tr.insertHelper(tr.root, v, func(n *Node) { tr.root = n })
	}

	tr.b.AddStep("L16", trace.ActionDone, fmt.Sprintf("AVL tree insertions complete. Height=%d, Root=%d (bf=%d: left.Height(%d) - right.Height(%d)).", height(tr.root), tr.root.Val, BalanceFactor(tr.root), height(tr.root.Left), height(tr.root.Right)),
		tr.treeState(nil),
		map[string]any{
			"total":         len(values),
			"height":        height(tr.root),
			"root":          tr.root.Val,
			"bf":            BalanceFactor(tr.root),
			"balanceFactor": BalanceFactor(tr.root),
			"leftHeight":    height(tr.root.Left),
			"rightHeight":   height(tr.root.Right),
		})

	return tr.b.Steps()
}
