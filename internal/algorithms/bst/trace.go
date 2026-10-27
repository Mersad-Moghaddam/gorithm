package bst

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Operation struct {
	Type  string `json:"type"` // "insert", "search", "delete"
	Value int    `json:"value"`
}

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

// cloneTreeState deep copies the tree state and every node inside it.
func cloneTreeState(ts renderers.TreeState) renderers.TreeState {
	if ts.Nodes == nil {
		return renderers.TreeState{Nodes: []renderers.TreeNode{}}
	}
	cloned := make([]renderers.TreeNode, len(ts.Nodes))
	for i, n := range ts.Nodes {
		cloned[i] = renderers.TreeNode{
			ID:       n.ID,
			Val:      n.Val,
			Label:    n.Label,
			X:        n.X,
			Y:        n.Y,
			ParentID: n.ParentID,
			State:    n.State,
		}
	}
	return renderers.TreeState{Nodes: cloned}
}

func (tr *tracer) addStep(lineID string, action trace.ActionType, expl string, state renderers.TreeState, vars map[string]any) {
	clonedVars := make(map[string]any, len(vars))
	for k, v := range vars {
		clonedVars[k] = v
	}
	tr.b.AddStep(lineID, action, expl, cloneTreeState(state), clonedVars)
}

func (tr *tracer) treeState(states map[*Node]string) renderers.TreeState {
	if tr.root == nil {
		return renderers.TreeState{Nodes: []renderers.TreeNode{}}
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
	return cloneTreeState(renderers.TreeState{Nodes: nodes})
}

func (tr *tracer) traceInsert(val int) {
	tr.addStep("L1", trace.ActionInit, fmt.Sprintf("Begin Insert(%d): preserve BST invariant (left < root < right).", val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "insert"})

	tr.addStep("L2", trace.ActionInit, fmt.Sprintf("Call insert(root, %d) starting traversal at root.", val),
		tr.treeState(nil),
		map[string]any{"val": val})

	if tr.root == nil {
		tr.b.IncComparisons()
		newNode := &Node{Val: val}
		tr.root = newNode
		tr.b.IncOperations()
		tr.addStep("L5", trace.ActionPush, fmt.Sprintf("Tree empty: allocate new Node(%d) as root.", val),
			tr.treeState(map[*Node]string{newNode: "inserted"}),
			map[string]any{"val": val, "inserted": val})
		tr.addStep("L8", trace.ActionDone, fmt.Sprintf("Insert(%d) complete: %d is now root of BST.", val, val),
			tr.treeState(map[*Node]string{newNode: "inserted"}),
			map[string]any{"val": val})
		return
	}

	curr := tr.root
	for {
		tr.b.IncComparisons()
		tr.addStep("L4", trace.ActionCompare, fmt.Sprintf("Inspect node %d: comparing target %d with node key %d.", curr.Val, val, curr.Val),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"val": val, "curr": curr.Val})

		if val < curr.Val {
			tr.b.IncComparisons()
			tr.addStep("L6", trace.ActionCompare, fmt.Sprintf("%d < %d: BST property dictates smaller values must go to left subtree; traverse left.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val, "branch": "left"})

			if curr.Left == nil {
				newNode := &Node{Val: val}
				curr.Left = newNode
				tr.b.IncOperations()
				tr.addStep("L5", trace.ActionPush, fmt.Sprintf("Left child of %d is nil: insert %d as newly placed leaf node.", curr.Val, val),
					tr.treeState(map[*Node]string{newNode: "inserted", curr: "active"}),
					map[string]any{"val": val, "parent": curr.Val, "inserted": val})
				tr.addStep("L8", trace.ActionDone, fmt.Sprintf("Insert(%d) complete: node %d added to BST.", val, val),
					tr.treeState(map[*Node]string{newNode: "inserted"}),
					map[string]any{"val": val})
				return
			}
			curr = curr.Left
		} else if val > curr.Val {
			tr.b.IncComparisons()
			tr.addStep("L7", trace.ActionCompare, fmt.Sprintf("%d > %d: BST property dictates larger values must go to right subtree; traverse right.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val, "branch": "right"})

			if curr.Right == nil {
				newNode := &Node{Val: val}
				curr.Right = newNode
				tr.b.IncOperations()
				tr.addStep("L5", trace.ActionPush, fmt.Sprintf("Right child of %d is nil: insert %d as newly placed leaf node.", curr.Val, val),
					tr.treeState(map[*Node]string{newNode: "inserted", curr: "active"}),
					map[string]any{"val": val, "parent": curr.Val, "inserted": val})
				tr.addStep("L8", trace.ActionDone, fmt.Sprintf("Insert(%d) complete: node %d added to BST.", val, val),
					tr.treeState(map[*Node]string{newNode: "inserted"}),
					map[string]any{"val": val})
				return
			}
			curr = curr.Right
		} else {
			tr.addStep("L8", trace.ActionDone, fmt.Sprintf("Target %d == %d: duplicate key already exists; BST keeps existing node without inserting.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "matched"}),
				map[string]any{"val": val, "duplicate": true})
			return
		}
	}
}

func (tr *tracer) traceSearch(val int) {
	tr.addStep("L10", trace.ActionInit, fmt.Sprintf("Begin Search(%d): search BST for key %d.", val, val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "search"})

	tr.addStep("L11", trace.ActionInit, fmt.Sprintf("Call search(root, %d) starting traversal at root.", val),
		tr.treeState(nil),
		map[string]any{"val": val})

	curr := tr.root
	for curr != nil {
		tr.b.IncComparisons()
		tr.addStep("L13", trace.ActionCompare, fmt.Sprintf("Inspect node %d: comparing target %d with node key %d.", curr.Val, val, curr.Val),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"val": val, "curr": curr.Val})

		if val == curr.Val {
			tr.b.IncComparisons()
			tr.addStep("L15", trace.ActionDone, fmt.Sprintf("Target %d == node %d: key matched!", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "matched"}),
				map[string]any{"val": val, "found": true, "curr": curr.Val})
			return
		} else if val < curr.Val {
			tr.b.IncComparisons()
			tr.addStep("L16", trace.ActionCompare, fmt.Sprintf("%d < %d: BST property dictates smaller values must be in left subtree; traverse left.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val, "branch": "left"})
			curr = curr.Left
		} else {
			tr.b.IncComparisons()
			tr.addStep("L17", trace.ActionCompare, fmt.Sprintf("%d > %d: BST property dictates larger values must be in right subtree; traverse right.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val, "branch": "right"})
			curr = curr.Right
		}
	}

	tr.b.IncComparisons()
	tr.addStep("L14", trace.ActionDone, fmt.Sprintf("Reached nil: target key %d not found in BST.", val),
		tr.treeState(nil),
		map[string]any{"val": val, "found": false})
}

func (tr *tracer) traceDelete(val int) {
	tr.addStep("L19", trace.ActionInit, fmt.Sprintf("Begin Delete(%d): remove key %d from BST while preserving BST invariant.", val, val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "delete"})

	tr.addStep("L20", trace.ActionInit, fmt.Sprintf("Call deleteNode(root, %d) starting traversal at root.", val),
		tr.treeState(nil),
		map[string]any{"val": val})

	tr.root = tr.deleteHelper(tr.root, val)
	tr.addStep("L33", trace.ActionDone, fmt.Sprintf("Delete(%d) finished: BST invariant holds.", val),
		tr.treeState(nil),
		map[string]any{"val": val})
}

func (tr *tracer) deleteHelper(node *Node, val int) *Node {
	if node == nil {
		tr.addStep("L23", trace.ActionDone, fmt.Sprintf("Reached nil: key %d not found in tree.", val),
			tr.treeState(nil),
			map[string]any{"val": val, "found": false})
		return nil
	}

	tr.b.IncComparisons()
	tr.addStep("L22", trace.ActionCompare, fmt.Sprintf("Inspect node %d: comparing target %d with node key %d.", node.Val, val, node.Val),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"val": val, "curr": node.Val})

	if val < node.Val {
		tr.b.IncComparisons()
		tr.addStep("L24", trace.ActionCompare, fmt.Sprintf("%d < %d: BST property dictates smaller values must be in left subtree; descend left.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val, "branch": "left"})
		node.Left = tr.deleteHelper(node.Left, val)
		return node
	} else if val > node.Val {
		tr.b.IncComparisons()
		tr.addStep("L25", trace.ActionCompare, fmt.Sprintf("%d > %d: BST property dictates larger values must be in right subtree; descend right.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val, "branch": "right"})
		node.Right = tr.deleteHelper(node.Right, val)
		return node
	}

	// Found target node to delete: val == node.Val
	tr.addStep("L26", trace.ActionCompare, fmt.Sprintf("Target matched: found node %d to delete.", node.Val),
		tr.treeState(map[*Node]string{node: "deleted"}),
		map[string]any{"val": val, "target": node.Val})

	if node.Left == nil && node.Right == nil {
		// Case 1: Leaf
		tr.b.IncOperations()
		tr.addStep("L27", trace.ActionPop, fmt.Sprintf("Case 1 (Leaf): node %d has no children (left is nil, right is nil). Delete leaf by returning nil.", node.Val),
			tr.treeState(map[*Node]string{node: "deleted"}),
			map[string]any{"val": node.Val, "case": "leaf"})
		return nil
	}

	if node.Left == nil {
		// Case 2: One child (right child)
		tr.b.IncOperations()
		tr.addStep("L27", trace.ActionPop, fmt.Sprintf("Case 2 (One Child): node %d has no left child. Bypass node %d and promote right child %d.", node.Val, node.Val, node.Right.Val),
			tr.treeState(map[*Node]string{node: "deleted", node.Right: "active"}),
			map[string]any{"val": node.Val, "case": "one_child", "child": node.Right.Val})
		return node.Right
	}

	if node.Right == nil {
		// Case 2: One child (left child)
		tr.b.IncOperations()
		tr.addStep("L28", trace.ActionPop, fmt.Sprintf("Case 2 (One Child): node %d has no right child. Bypass node %d and promote left child %d.", node.Val, node.Val, node.Left.Val),
			tr.treeState(map[*Node]string{node: "deleted", node.Left: "active"}),
			map[string]any{"val": node.Val, "case": "one_child", "child": node.Left.Val})
		return node.Left
	}

	// Case 3: Two children
	tr.addStep("L29", trace.ActionVisit, fmt.Sprintf("Case 3 (Two Children): node %d has two children. Find in-order successor (minimum in right subtree) to replace it.", node.Val),
		tr.treeState(map[*Node]string{node: "deleted", node.Right: "active"}),
		map[string]any{"val": node.Val, "case": "two_children"})

	minN := node.Right
	for minN.Left != nil {
		tr.b.IncComparisons()
		tr.addStep("L29", trace.ActionVisit, fmt.Sprintf("Traverse left in right subtree to find minimum successor: visiting %d.", minN.Val),
			tr.treeState(map[*Node]string{node: "deleted", minN: "active"}),
			map[string]any{"val": node.Val, "min": minN.Val})
		minN = minN.Left
	}

	minVal := minN.Val
	tr.addStep("L29", trace.ActionVisit, fmt.Sprintf("In-order successor is %d (smallest key in right subtree).", minVal),
		tr.treeState(map[*Node]string{node: "deleted", minN: "matched"}),
		map[string]any{"val": node.Val, "successor": minVal})

	tr.b.IncSwaps()
	node.Val = minVal
	tr.addStep("L30", trace.ActionSwap, fmt.Sprintf("Copy successor value %d into target node (formerly %d).", minVal, val),
		tr.treeState(map[*Node]string{node: "active", minN: "deleted"}),
		map[string]any{"newVal": minVal, "oldVal": val})

	tr.addStep("L31", trace.ActionCompare, fmt.Sprintf("Recursively delete successor %d from right subtree to remove duplicate.", minVal),
		tr.treeState(map[*Node]string{node: "active", minN: "deleted"}),
		map[string]any{"successor": minVal})

	node.Right = tr.deleteHelper(node.Right, minVal)
	return node
}

// Trace executes BST operations and produces execution steps.
// ponytail: pointer tree with recursive descent; iterative traversal eliminates call stack overhead for depth > 1000.
func Trace(initial []int, ops []Operation) []trace.Step {
	tr := newTracer()

	if len(initial) == 0 && len(ops) == 0 {
		initial = defaultInitial()
		ops = defaultOps()
	}

	for _, v := range initial {
		tr.root = insert(tr.root, v)
	}

	var initExpl string
	if len(initial) > 0 {
		initExpl = fmt.Sprintf("Binary Search Tree initialized with %v.", initial)
	} else {
		initExpl = "Binary Search Tree initialized (empty)."
	}

	tr.addStep("L1", trace.ActionInit, initExpl,
		tr.treeState(nil),
		map[string]any{"initial": initial})

	for _, op := range ops {
		switch op.Type {
		case "insert":
			tr.traceInsert(op.Value)
		case "search":
			tr.traceSearch(op.Value)
		case "delete":
			tr.traceDelete(op.Value)
		}
	}

	return tr.b.Steps()
}
