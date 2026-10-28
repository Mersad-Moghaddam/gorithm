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
	rotation, rotationCase string
	rotationPart int
	pivot, newRoot, transfer *Node
	pending bool
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

func (tr *tracer) treeState(states map[*Node]string) renderers.AVLLessonState {
	s:=renderers.AVLLessonState{Rotation:tr.rotation,Case:tr.rotationCase,RotationPart:tr.rotationPart,PivotID:tr.idFor(tr.pivot),NewRootID:tr.idFor(tr.newRoot),TransferID:tr.idFor(tr.transfer),HeightsPending:tr.pending}
	if tr.root == nil {
		return s
	}

	var nodes []renderers.TreeNode
	var count func(*Node) int
	count=func(n *Node)int {if n==nil{return 0};return 1+count(n.Left)+count(n.Right)}
	w:=max(800,88*(count(tr.root)+1));slot:=0
	var layout func(n *Node, x, y, dx float64, parentID string)
	layout = func(n *Node, x, y, dx float64, parentID string) {
		if n == nil {
			return
		}
		id := tr.idFor(n)
		layout(n.Left,0,y+104,0,id)
		slot++; x=float64(w)/2+88*float64(slot)-44*float64(count(tr.root)+1)
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
		s.LeftIDs=append(s.LeftIDs,tr.idFor(n.Left));s.RightIDs=append(s.RightIDs,tr.idFor(n.Right))
		s.Heights=append(s.Heights,n.Height);s.Balances=append(s.Balances,BalanceFactor(n))
		if st=="active"||st=="inserted"||st=="matched" {s.ActiveID=id}
		layout(n.Right,0,y+104,0,id)
	}

	layout(tr.root, 0, 96, 0, "")
	s.TreeState=renderers.TreeState{Nodes:nodes}
	return s
}

func (tr *tracer) traceRightRotate(y *Node, setParent func(*Node)) *Node {
	x := y.Left
	t2 := x.Right
	tr.rotation="right";tr.pivot=y;tr.newRoot=x;tr.transfer=t2

	tr.b.IncOperations()
	tr.b.AddStep("L18", trace.ActionSwap, fmt.Sprintf("Right Rotation around node %d: left child %d becomes new root of this subtree.", y.Val, x.Val),
		tr.treeState(map[*Node]string{y: "active", x: "active"}),
		map[string]any{"rotation": "right", "pivot": y.Val, "newRoot": x.Val})

	x.Right = y
	y.Left = t2
	setParent(x)
	tr.pending=true
	tr.b.AddStep("L19",trace.ActionSet,"Rewire right rotation; stored heights await recomputation.",tr.treeState(map[*Node]string{x:"active"}),nil)

	updateHeight(y)
	updateHeight(x)
	tr.pending=false

	tr.b.AddStep("L20", trace.ActionSet, fmt.Sprintf("Right Rotation done. Subtree root %d height is %d, child %d height is %d.", x.Val, x.Height, y.Val, y.Height),
		tr.treeState(map[*Node]string{x: "matched", y: "normal"}),
		map[string]any{"root": x.Val, "height": x.Height, "bf": BalanceFactor(x), "balanceFactor": BalanceFactor(x)})

	return x
}

func (tr *tracer) traceLeftRotate(x *Node, setParent func(*Node)) *Node {
	y := x.Right
	t2 := y.Left
	tr.rotation="left";tr.pivot=x;tr.newRoot=y;tr.transfer=t2

	tr.b.IncOperations()
	tr.b.AddStep("L22", trace.ActionSwap, fmt.Sprintf("Left Rotation around node %d: right child %d becomes new root of this subtree.", x.Val, y.Val),
		tr.treeState(map[*Node]string{x: "active", y: "active"}),
		map[string]any{"rotation": "left", "pivot": x.Val, "newRoot": y.Val})

	y.Left = x
	x.Right = t2
	setParent(y)
	tr.pending=true
	tr.b.AddStep("L23",trace.ActionSet,"Rewire left rotation; stored heights await recomputation.",tr.treeState(map[*Node]string{y:"active"}),nil)

	updateHeight(x)
	updateHeight(y)
	tr.pending=false

	tr.b.AddStep("L24", trace.ActionSet, fmt.Sprintf("Left Rotation done. Subtree root %d height is %d, child %d height is %d.", y.Val, y.Height, x.Val, x.Height),
		tr.treeState(map[*Node]string{y: "matched", x: "normal"}),
		map[string]any{"root": y.Val, "height": y.Height, "bf": BalanceFactor(y), "balanceFactor": BalanceFactor(y)})

	return y
}

func (tr *tracer) insertHelper(node *Node, val int, setParent func(*Node)) *Node {
	if node == nil {
		newNode := &Node{Val: val, Height: 1}
		setParent(newNode)
		tr.b.IncOperations()
		tr.b.AddStep("L2", trace.ActionPush, fmt.Sprintf("Empty spot found. Insert node %d with initial height 1.", val),
			tr.treeState(map[*Node]string{newNode: "inserted"}),
			map[string]any{"val": val, "height": 1, "bf": 0, "balanceFactor": 0})
		return newNode
	}

	tr.b.IncComparisons()
	tr.b.AddStep("L1", trace.ActionCompare, fmt.Sprintf("Compare target %d with node %d.", val, node.Val),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"val": val, "curr": node.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})

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
	tr.b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Unwinding to node %d: updated height to %d.", node.Val, node.Height),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"node": node.Val, "height": node.Height, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})

	bf := BalanceFactor(node)
	tr.b.IncComparisons()
	tr.b.AddStep("L7", trace.ActionCompare, fmt.Sprintf("Node %d balance factor = %d (left h=%d, right h=%d).", node.Val, bf, height(node.Left), height(node.Right)),
		tr.treeState(map[*Node]string{node: "active"}),
		map[string]any{"node": node.Val, "bf": bf, "balanceFactor": bf, "height": node.Height})

	// Case 1: Left-Left (LL) -> Right Rotation
	if bf > 1 && val < node.Left.Val {
		tr.rotationCase="LL";tr.rotationPart=1
		tr.b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Imbalance at node %d: bf=%d > 1 and inserted %d < left child %d (Left-Left case). Triggers Right Rotation.", node.Val, bf, val, node.Left.Val),
			tr.treeState(map[*Node]string{node: "active", node.Left: "active"}),
			map[string]any{"node": node.Val, "case": "LL", "bf": bf, "balanceFactor": bf})
		return tr.traceRightRotate(node, setParent)
	}

	// Case 2: Right-Right (RR) -> Left Rotation
	if bf < -1 && val > node.Right.Val {
		tr.rotationCase="RR";tr.rotationPart=1
		tr.b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Imbalance at node %d: bf=%d < -1 and inserted %d > right child %d (Right-Right case). Triggers Left Rotation.", node.Val, bf, val, node.Right.Val),
			tr.treeState(map[*Node]string{node: "active", node.Right: "active"}),
			map[string]any{"node": node.Val, "case": "RR", "bf": bf, "balanceFactor": bf})
		return tr.traceLeftRotate(node, setParent)
	}

	// Case 3: Left-Right (LR) -> Left Rotate Left Child, then Right Rotate Root
	if bf > 1 && val > node.Left.Val {
		tr.rotationCase="LR";tr.rotationPart=1
		tr.b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Imbalance at node %d: bf=%d > 1 and inserted %d > left child %d (Left-Right case). Triggers Double Rotation (Left on %d, then Right on %d).", node.Val, bf, val, node.Left.Val, node.Left.Val, node.Val),
			tr.treeState(map[*Node]string{node: "active", node.Left: "active"}),
			map[string]any{"node": node.Val, "case": "LR", "bf": bf, "balanceFactor": bf})

		tr.b.AddStep("L11", trace.ActionSwap, fmt.Sprintf("LR step 1: Left Rotate left child %d.", node.Left.Val),
			tr.treeState(map[*Node]string{node.Left: "active"}),
			map[string]any{"step": "LR-1", "node": node.Left.Val, "bf": BalanceFactor(node.Left), "balanceFactor": BalanceFactor(node.Left)})
		node.Left = tr.traceLeftRotate(node.Left, func(n *Node) { node.Left = n })
		tr.rotationPart=2

		tr.b.AddStep("L11", trace.ActionSwap, fmt.Sprintf("LR step 2: Right Rotate node %d.", node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"step": "LR-2", "node": node.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})
		return tr.traceRightRotate(node, setParent)
	}

	// Case 4: Right-Left (RL) -> Right Rotate Right Child, then Left Rotate Root
	if bf < -1 && val < node.Right.Val {
		tr.rotationCase="RL";tr.rotationPart=1
		tr.b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Imbalance at node %d: bf=%d < -1 and inserted %d < right child %d (Right-Left case). Triggers Double Rotation (Right on %d, then Left on %d).", node.Val, bf, val, node.Right.Val, node.Right.Val, node.Val),
			tr.treeState(map[*Node]string{node: "active", node.Right: "active"}),
			map[string]any{"node": node.Val, "case": "RL", "bf": bf, "balanceFactor": bf})

		tr.b.AddStep("L14", trace.ActionSwap, fmt.Sprintf("RL step 1: Right Rotate right child %d.", node.Right.Val),
			tr.treeState(map[*Node]string{node.Right: "active"}),
			map[string]any{"step": "RL-1", "node": node.Right.Val, "bf": BalanceFactor(node.Right), "balanceFactor": BalanceFactor(node.Right)})
		node.Right = tr.traceRightRotate(node.Right, func(n *Node) { node.Right = n })
		tr.rotationPart=2

		tr.b.AddStep("L14", trace.ActionSwap, fmt.Sprintf("RL step 2: Left Rotate node %d.", node.Val),
			tr.treeState(map[*Node]string{node: "active"}),
			map[string]any{"step": "RL-2", "node": node.Val, "bf": BalanceFactor(node), "balanceFactor": BalanceFactor(node)})
		return tr.traceLeftRotate(node, setParent)
	}

	tr.b.AddStep("L16", trace.ActionDone, fmt.Sprintf("Node %d is balanced (bf=%d in [-1, 1]). AVL invariant holds.", node.Val, bf),
		tr.treeState(nil),
		map[string]any{"node": node.Val, "bf": bf, "balanceFactor": bf, "height": node.Height})

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
		map[string]any{"values": append([]int(nil),values...)})

	for _, v := range values {
		tr.rotation="";tr.rotationCase="";tr.rotationPart=0;tr.pivot=nil;tr.newRoot=nil;tr.transfer=nil
		tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Begin Insert(%d).", v),
			tr.treeState(nil),
			map[string]any{"val": v, "op": "insert"})
		tr.root = tr.insertHelper(tr.root, v, func(n *Node) { tr.root = n })
	}

	tr.b.AddStep("L16", trace.ActionDone, fmt.Sprintf("AVL tree insertions complete. Height=%d, Root=%d.", height(tr.root), tr.root.Val),
		tr.treeState(nil),
		map[string]any{
			"total":         len(values),
			"height":        height(tr.root),
			"root":          tr.root.Val,
			"bf":            BalanceFactor(tr.root),
			"balanceFactor": BalanceFactor(tr.root),
		})

	steps:=tr.b.Steps()
	for i:=range steps {
		s:=steps[i].State.(renderers.AVLLessonState)
		if v,ok:=steps[i].Variables["val"].(int);ok {s.Target=v} else if i>0 {s.Target=steps[i-1].State.(renderers.AVLLessonState).Target}
		switch steps[i].LineID {
		case "L6","L20","L24":s.Phase="height"
		case "L7","L8","L9","L10","L13":s.Phase="balance"
		case "L18","L19","L22","L23","L11","L14":s.Phase="rotation"
		case "L5","L16":s.Phase="done"
		case "L2":s.Phase="insert"
		default:if steps[i].Action==trace.ActionCompare {s.Phase="compare"} else {s.Phase="insert"}
		}
		if s.ActiveID!="" {byID:=map[string]renderers.TreeNode{};for _,n:=range s.Nodes {byID[n.ID]=n};for id:=s.ActiveID;id!="";id=byID[id].ParentID {s.PathIDs=append([]string{id},s.PathIDs...)}}
		steps[i].State=s
		if (steps[i].LineID=="L19"||steps[i].LineID=="L23") && i>0 {steps[i-1].HoldMS=600;steps[i].HoldMS=1200}
	}
	return steps
}
