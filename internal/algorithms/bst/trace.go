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
	return renderers.TreeState{Nodes: nodes}
}

func (tr *tracer) traceInsert(val int) {
	tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Begin Insert(%d).", val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "insert"})

	if tr.root == nil {
		tr.b.IncComparisons()
		newNode := &Node{Val: val}
		tr.root = newNode
		tr.b.IncOperations()
		tr.b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Tree empty. Insert %d as root node.", val),
			tr.treeState(map[*Node]string{newNode: "inserted"}),
			map[string]any{"val": val})
		tr.b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Node %d is now root.", val),
			tr.treeState(nil),
			map[string]any{"val": val})
		return
	}

	curr := tr.root
	for {
		tr.b.IncComparisons()
		tr.b.AddStep("L4", trace.ActionCompare, fmt.Sprintf("Inspect node %d for inserting %d.", curr.Val, val),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"val": val, "curr": curr.Val})

		if val < curr.Val {
			tr.b.IncComparisons()
			tr.b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("%d < %d: traverse left.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val})

			if curr.Left == nil {
				newNode := &Node{Val: val}
				curr.Left = newNode
				tr.b.IncOperations()
				tr.b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Left child is nil. Insert %d as left child of %d.", val, curr.Val),
					tr.treeState(map[*Node]string{newNode: "inserted"}),
					map[string]any{"val": val, "parent": curr.Val})
				tr.b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Node %d successfully inserted.", val),
					tr.treeState(nil),
					map[string]any{"val": val})
				return
			}
			curr = curr.Left
		} else if val > curr.Val {
			tr.b.IncComparisons()
			tr.b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("%d > %d: traverse right.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val})

			if curr.Right == nil {
				newNode := &Node{Val: val}
				curr.Right = newNode
				tr.b.IncOperations()
				tr.b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Right child is nil. Insert %d as right child of %d.", val, curr.Val),
					tr.treeState(map[*Node]string{newNode: "inserted"}),
					map[string]any{"val": val, "parent": curr.Val})
				tr.b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Node %d successfully inserted.", val),
					tr.treeState(nil),
					map[string]any{"val": val})
				return
			}
			curr = curr.Right
		} else {
			tr.b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Value %d already exists in BST (duplicates ignored).", val),
				tr.treeState(map[*Node]string{curr: "matched"}),
				map[string]any{"val": val})
			tr.b.AddStep("L8", trace.ActionDone, "Tree unchanged.",
				tr.treeState(nil),
				map[string]any{"val": val})
			return
		}
	}
}

func (tr *tracer) traceSearch(val int) {
	tr.b.AddStep("L10", trace.ActionInit, fmt.Sprintf("Begin Search(%d).", val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "search"})

	curr := tr.root
	for curr != nil {
		tr.b.IncComparisons()
		tr.b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Compare target %d with node %d.", val, curr.Val),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"val": val, "curr": curr.Val})

		if val == curr.Val {
			tr.b.IncComparisons()
			tr.b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Target %d matched at node %d!", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "matched"}),
				map[string]any{"val": val, "found": true})
			tr.b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Search complete: %d is present.", val),
				tr.treeState(nil),
				map[string]any{"val": val, "found": true})
			return
		} else if val < curr.Val {
			tr.b.IncComparisons()
			tr.b.AddStep("L16", trace.ActionCompare, fmt.Sprintf("%d < %d: search left subtree.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val})
			curr = curr.Left
		} else {
			tr.b.IncComparisons()
			tr.b.AddStep("L17", trace.ActionCompare, fmt.Sprintf("%d > %d: search right subtree.", val, curr.Val),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"val": val, "curr": curr.Val})
			curr = curr.Right
		}
	}

	tr.b.AddStep("L14", trace.ActionDone, fmt.Sprintf("Target %d not found in BST (reached nil).", val),
		tr.treeState(nil),
		map[string]any{"val": val, "found": false})
}

func (tr *tracer) traceDelete(val int) {
	tr.b.AddStep("L19", trace.ActionInit, fmt.Sprintf("Begin Delete(%d).", val),
		tr.treeState(nil),
		map[string]any{"val": val, "op": "delete"})

	tr.root = tr.deleteHelper(tr.root, val)
	tr.b.AddStep("L33", trace.ActionDone, fmt.Sprintf("Delete(%d) finished. BST invariant holds.", val),
		tr.treeState(nil),
		map[string]any{"val": val})
}

func (tr *tracer) deleteHelper(node *Node, val int) *Node {
	if node == nil {
		tr.b.AddStep("L23", trace.ActionDone, fmt.Sprintf("Node %d not found in tree.", val),
			tr.treeState(nil),
			map[string]any{"val": val})
		return nil
	}

	tr.b.IncComparisons()
	tr.b.AddStep("L22", trace.ActionCompare, fmt.Sprintf("Inspect node %d for deletion.", node.Val),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"val": val, "curr": node.Val})

	if val < node.Val {
		tr.b.IncComparisons()
		tr.b.AddStep("L24", trace.ActionCompare, fmt.Sprintf("%d < %d: descend left.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val})
		node.Left = tr.deleteHelper(node.Left, val)
		return node
	} else if val > node.Val {
		tr.b.IncComparisons()
		tr.b.AddStep("L25", trace.ActionCompare, fmt.Sprintf("%d > %d: descend right.", val, node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"val": val, "curr": node.Val})
		node.Right = tr.deleteHelper(node.Right, val)
		return node
	}

	// Found node to delete
	tr.b.AddStep("L26", trace.ActionCompare, fmt.Sprintf("Found target node %d to delete.", node.Val),
		tr.treeState(map[*Node]string{node: "deleted"}),
		map[string]any{"val": val})

	if node.Left == nil {
		tr.b.IncOperations()
		tr.b.AddStep("L27", trace.ActionPop, fmt.Sprintf("Node %d has no left child. Replace with right child.", node.Val),
			tr.treeState(map[*Node]string{node: "deleted"}),
			map[string]any{"val": node.Val})
		return node.Right
	}
	if node.Right == nil {
		tr.b.IncOperations()
		tr.b.AddStep("L28", trace.ActionPop, fmt.Sprintf("Node %d has no right child. Replace with left child.", node.Val),
			tr.treeState(map[*Node]string{node: "deleted"}),
			map[string]any{"val": node.Val})
		return node.Left
	}

	// Node with two children
	tr.b.AddStep("L29", trace.ActionVisit, fmt.Sprintf("Node %d has two children. Find in-order successor in right subtree.", node.Val),
		tr.treeState(map[*Node]string{node: "deleted"}),
		map[string]any{"val": node.Val})

	minN := node.Right
	for minN.Left != nil {
		tr.b.IncComparisons()
		tr.b.AddStep("L29", trace.ActionVisit, fmt.Sprintf("Traverse right subtree to find minimum: visiting %d.", minN.Val),
			tr.treeState(map[*Node]string{node: "deleted", minN: "active"}),
			map[string]any{"val": node.Val, "min": minN.Val})
		minN = minN.Left
	}

	minVal := minN.Val
	tr.b.AddStep("L29", trace.ActionVisit, fmt.Sprintf("In-order successor is %d.", minVal),
		tr.treeState(map[*Node]string{node: "deleted", minN: "matched"}),
		map[string]any{"val": node.Val, "successor": minVal})

	tr.b.IncSwaps()
	node.Val = minVal
	tr.b.AddStep("L30", trace.ActionSwap, fmt.Sprintf("Copy successor value %d into target node.", minVal),
		tr.treeState(map[*Node]string{node: "active", minN: "deleted"}),
		map[string]any{"newVal": minVal})

	tr.b.AddStep("L31", trace.ActionCompare, fmt.Sprintf("Delete successor %d from right subtree.", minVal),
		tr.treeState(map[*Node]string{node: "active", minN: "deleted"}),
		map[string]any{"succ": minVal})
	node.Right = tr.deleteHelper(node.Right, minVal)

	return node
}

// Trace executes BST operations and produces execution steps.
// ponytail: pointer tree with recursive descent; iterative traversal eliminates call stack overhead for depth > 1000.
func Trace(initial []int, ops []Operation) []trace.Step {
	tr := newTracer()

	if len(initial) == 0 && len(ops) == 0 {
		initial = []int{50, 30, 70, 20, 40, 60, 80}
		ops = []Operation{
			{Type: "insert", Value: 35},
			{Type: "search", Value: 40},
			{Type: "delete", Value: 30},
		}
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

	tr.b.AddStep("L1", trace.ActionInit, initExpl,
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
