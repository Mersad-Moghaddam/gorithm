package bst

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Operation struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type tracer struct {
	b         *trace.Builder
	root      *Node
	nodeIDs   map[*Node]string
	nodeIDSeq int
	s         renderers.BSTLessonState
}

func (tr *tracer) idFor(n *Node) string {
	if n == nil {
		return ""
	}
	if id := tr.nodeIDs[n]; id != "" {
		return id
	}
	tr.nodeIDSeq++
	id := fmt.Sprintf("node-%d", tr.nodeIDSeq)
	tr.nodeIDs[n] = id
	return id
}

func (tr *tracer) emit(line, phase string, action trace.ActionType, explanation string) {
	s := tr.s
	s.Phase = phase
	s.Nodes = nil
	s.LeftIDs = nil
	s.RightIDs = nil
	s.PathIDs = append([]string(nil), tr.s.PathIDs...)
	var walk func(*Node, string, int)
	walk = func(n *Node, parent string, depth int) {
		if n == nil {
			return
		}
		id := tr.idFor(n)
		state := "normal"
		if id == s.ActiveID {
			state = "active"
		}
		if id == s.ActiveID && (phase == "done" || phase == "duplicate") && s.Found {
			state = "matched"
		}
		if id == s.ActiveID && phase == "insert" {
			state = "inserted"
		}
		if id == s.ActiveID && phase == "successor" {
			state = "deleted"
		}
		s.Nodes = append(s.Nodes, renderers.TreeNode{ID: id, Val: n.Val, ParentID: parent, State: state, Y: 88 + 76*float64(depth)})
		s.LeftIDs = append(s.LeftIDs, tr.idFor(n.Left))
		s.RightIDs = append(s.RightIDs, tr.idFor(n.Right))
		walk(n.Left, id, depth+1)
		walk(n.Right, id, depth+1)
	}
	walk(tr.root, "", 0)
	// In-order slots use topology, including temporary equal-valued successor copies.
	byID := make(map[string]int, len(s.Nodes))
	for i, n := range s.Nodes {
		byID[n.ID] = i
	}
	w := 800
	if x := 64 * (len(s.Nodes) + 1); x > w {
		w = x
	}
	rank := 0
	var place func(string)
	place = func(id string) {
		if id == "" {
			return
		}
		i := byID[id]
		place(s.LeftIDs[i])
		s.Nodes[i].X = float64(64*(rank+1) + (w-64*(len(s.Nodes)+1))/2)
		rank++
		place(s.RightIDs[i])
	}
	if tr.root != nil {
		place(tr.idFor(tr.root))
	}
	tr.b.AddStep(line, action, explanation, s, map[string]any{"op": s.Operation, "target": s.Target, "found": s.Found, "resultKnown": s.ResultKnown})
}

func (tr *tracer) begin(op string, val int, line string) {
	tr.s = renderers.BSTLessonState{Operation: op, Target: val}
	tr.emit(line, "init", trace.ActionInit, fmt.Sprintf("Begin %s(%d).", op, val))
}

func (tr *tracer) visit(n *Node, line string) {
	tr.s.ActiveID = tr.idFor(n)
	tr.s.PathIDs = append(tr.s.PathIDs, tr.s.ActiveID)
	tr.b.IncComparisons()
	tr.emit(line, "compare", trace.ActionCompare, fmt.Sprintf("Compare target %d with node %d.", tr.s.Target, n.Val))
}

func (tr *tracer) descend(n *Node, left bool, line string) {
	tr.s.ParentID = tr.idFor(n)
	tr.b.IncComparisons()
	side := "right"
	if left {
		side = "left"
	}
	tr.emit(line, "descend", trace.ActionCompare, fmt.Sprintf("Target %d: take %s child of %d.", tr.s.Target, side, n.Val))
}

func (tr *tracer) traceInsert(val int) {
	tr.begin("insert", val, "L1")
	link := &tr.root
	for *link != nil {
		n := *link
		tr.visit(n, "L4")
		if val == n.Val {
			tr.s.Found = true
			tr.s.ResultKnown = true
			tr.emit("L8", "duplicate", trace.ActionDone, fmt.Sprintf("%d already exists; tree unchanged.", val))
			return
		}
		if val < n.Val {
			tr.descend(n, true, "L6")
			link = &n.Left
		} else {
			tr.descend(n, false, "L7")
			link = &n.Right
		}
	}
	if tr.root == nil {
		tr.b.IncComparisons()
	}
	n := &Node{Val: val}
	*link = n
	tr.s.ActiveID = tr.idFor(n)
	tr.b.IncOperations()
	tr.s.Found = true
	tr.s.ResultKnown = true
	tr.emit("L5", "insert", trace.ActionPush, fmt.Sprintf("Insert %d at nil child; new node identity.", val))
	tr.emit("L8", "done", trace.ActionDone, fmt.Sprintf("Insert %d complete.", val))
}

func (tr *tracer) traceSearch(val int) {
	tr.begin("search", val, "L10")
	n := tr.root
	for n != nil {
		tr.visit(n, "L13")
		if val == n.Val {
			tr.b.IncComparisons()
			tr.s.Found = true
			tr.s.ResultKnown = true
			tr.emit("L15", "done", trace.ActionDone, fmt.Sprintf("Found %d.", val))
			return
		}
		if val < n.Val {
			tr.descend(n, true, "L16")
			n = n.Left
		} else {
			tr.descend(n, false, "L17")
			n = n.Right
		}
	}
	tr.s.ActiveID = ""
	tr.s.ResultKnown = true
	tr.emit("L14", "done", trace.ActionDone, fmt.Sprintf("%d not found: reached nil.", val))
}

// deleteLink commits the owning link before recording a removal snapshot.
func (tr *tracer) deleteLink(link **Node, val int) {
	n := *link
	if n == nil {
		tr.s.ActiveID = ""
		tr.s.ResultKnown = true
		tr.emit("L23", "done", trace.ActionDone, fmt.Sprintf("%d absent: reached nil.", val))
		return
	}
	tr.visit(n, "L22")
	if val < n.Val {
		tr.descend(n, true, "L24")
		tr.deleteLink(&n.Left, val)
		return
	}
	if val > n.Val {
		tr.descend(n, false, "L25")
		tr.deleteLink(&n.Right, val)
		return
	}
	tr.s.Found = true
	tr.s.ResultKnown = true
	tr.emit("L26", "compare", trace.ActionCompare, fmt.Sprintf("Found node %d to delete.", val))
	if n.Left == nil || n.Right == nil {
		replacement := n.Right
		line := "L27"
		if n.Left != nil {
			replacement = n.Left
			line = "L28"
		}
		tr.s.RemovedID = tr.idFor(n)
		tr.s.ReplacementID = tr.idFor(replacement)
		*link = replacement
		tr.s.ActiveID = tr.idFor(replacement)
		tr.b.IncOperations()
		tr.emit(line, "remove", trace.ActionPop, fmt.Sprintf("Remove %d; reconnect parent to replacement (or nil).", val))
		return
	}
	min := n.Right
	for {
		tr.s.ActiveID = tr.idFor(n)
		tr.s.SuccessorID = tr.idFor(min)
		tr.emit("L29", "successor", trace.ActionVisit, fmt.Sprintf("Inspect %d for minimum in right subtree.", min.Val))
		if min.Left == nil {
			break
		}
		tr.b.IncComparisons()
		min = min.Left
	}
	tr.s.OldValue = n.Val
	tr.s.HasOldValue = true
	n.Val = min.Val
	tr.b.IncSwaps()
	tr.emit("L30", "copy", trace.ActionSwap, fmt.Sprintf("Copy %d into target; keep target identity. Successor still exists.", min.Val))
	tr.emit("L31", "descend", trace.ActionCompare, fmt.Sprintf("Delete original successor %d from right subtree.", min.Val))
	tr.deleteLink(&n.Right, min.Val)
}

// Trace preserves direct empty-input defaults and ignores unknown operations.
func Trace(initial []int, ops []Operation) []trace.Step {
	tr := &tracer{b: trace.NewBuilder(), nodeIDs: make(map[*Node]string)}
	if len(initial) == 0 && len(ops) == 0 {
		initial = []int{50, 30, 70, 20, 40, 60, 80}
		ops = []Operation{{"insert", 35}, {"search", 40}, {"delete", 30}}
	}
	for _, v := range initial {
		tr.root = insert(tr.root, v)
	}
	tr.emit("L1", "init", trace.ActionInit, fmt.Sprintf("Initial tree prebuilt from %v; initialization inserts are not traced.", initial))
	for _, op := range ops {
		switch op.Type {
		case "insert":
			tr.traceInsert(op.Value)
		case "search":
			tr.traceSearch(op.Value)
		case "delete":
			tr.begin("delete", op.Value, "L19")
			tr.deleteLink(&tr.root, op.Value)
			tr.emit("L33", "done", trace.ActionDone, fmt.Sprintf("Delete(%d) complete; links committed.", op.Value))
		}
	}
	steps := tr.b.Steps()
	for i := 1; i < len(steps); i++ {
		phase := steps[i].State.(renderers.BSTLessonState).Phase
		if phase == "copy" || phase == "insert" || phase == "remove" {
			steps[i-1].HoldMS = 600
			steps[i].HoldMS = 1200
		}
	}
	return steps
}
