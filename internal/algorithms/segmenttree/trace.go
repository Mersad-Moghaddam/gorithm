package segmenttree

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type tracer struct {
	b     *trace.Builder
	s     renderers.SegmentTreeLessonState
	nodes map[int]int
}

func (tr *tracer) emit(line string, action trace.ActionType, node int, explanation string) {
	s := tr.s
	s.ActiveID = ""
	if node > 0 {
		s.ActiveID = fmt.Sprintf("node-%d", node)
	}
	s.Nodes = append([]renderers.SegmentTreeLessonNode(nil), s.Nodes...)
	s.Values = append([]int(nil), s.Values...)
	s.AcceptedIDs = append([]string(nil), s.AcceptedIDs...)
	s.PathIDs = append([]string(nil), s.PathIDs...)
	tr.b.AddStep(line, action, explanation, s, map[string]any{"L": s.QL, "R": s.QR, "node": node, "idx": s.UpdateIndex, "val": s.UpdateValue})
}

func (tr *tracer) clearResult() {
	tr.s.Overlap = ""
	tr.s.Result, tr.s.LeftResult, tr.s.RightResult = 0, 0, 0
	tr.s.ResultKnown, tr.s.LeftKnown, tr.s.RightKnown = false, false, false
}

func (tr *tracer) build(node int) {
	i := tr.nodes[node]
	n := tr.s.Nodes[i]
	tr.clearResult()
	tr.emit("L4", trace.ActionVisit, node, fmt.Sprintf("Inspect interval [%d, %d]; its sum is not built yet.", n.L, n.R))
	if n.L == n.R {
		tr.s.Nodes[i].Sum = tr.s.Values[n.L]
		tr.s.Nodes[i].Built = true
		tr.b.IncOperations()
		tr.emit("L5", trace.ActionSet, node, fmt.Sprintf("Leaf [%d, %d] stores %d.", n.L, n.R, tr.s.Values[n.L]))
		return
	}
	tr.emit("L6", trace.ActionSet, node, fmt.Sprintf("Split [%d, %d] at %d.", n.L, n.R, (n.L+n.R)/2))
	tr.emit("L7", trace.ActionVisit, node, "Build left interval first.")
	tr.build(2 * node)
	tr.emit("L8", trace.ActionVisit, node, "Build right interval next.")
	tr.build(2*node + 1)
	tr.combine(node)
	tr.emit("L9", trace.ActionSet, node, fmt.Sprintf("Combine child sums: %d + %d = %d.", tr.s.LeftResult, tr.s.RightResult, tr.s.Result))
	tr.emit("L10", trace.ActionDone, node, "This interval is built; return to its parent.")
}

func (tr *tracer) combine(node int) {
	tr.clearResult()
	tr.s.LeftResult = tr.s.Nodes[tr.nodes[2*node]].Sum
	tr.s.RightResult = tr.s.Nodes[tr.nodes[2*node+1]].Sum
	tr.s.Result = tr.s.LeftResult + tr.s.RightResult
	tr.s.LeftKnown, tr.s.RightKnown, tr.s.ResultKnown = true, true, true
	i := tr.nodes[node]
	tr.s.Nodes[i].Sum = tr.s.Result
	tr.s.Nodes[i].Built = true
	tr.b.IncOperations()
}

func (tr *tracer) query(node int) int {
	n := tr.s.Nodes[tr.nodes[node]]
	tr.clearResult()
	tr.b.IncComparisons()
	tr.emit("L14", trace.ActionCompare, node, fmt.Sprintf("Compare interval [%d, %d] with query [%d, %d].", n.L, n.R, tr.s.QL, tr.s.QR))
	if tr.s.QL <= n.L && n.R <= tr.s.QR {
		tr.s.Overlap = "full"
		tr.s.Result = n.Sum
		tr.s.ResultKnown = true
		tr.s.AcceptedIDs = append(tr.s.AcceptedIDs, n.ID)
		tr.b.IncComparisons()
		tr.emit("L15", trace.ActionHighlight, node, fmt.Sprintf("Full overlap: accept [%d, %d] once, contributing %d.", n.L, n.R, n.Sum))
		return n.Sum
	}
	if n.R < tr.s.QL || n.L > tr.s.QR {
		tr.s.Overlap = "none"
		tr.s.ResultKnown = true
		tr.b.IncComparisons()
		tr.emit("L16", trace.ActionDone, node, "Disjoint interval contributes +0; do not add it to accepted cover.")
		return 0
	}
	tr.s.Overlap = "partial"
	tr.emit("L17", trace.ActionSet, node, "Partial overlap: split query across both children.")
	tr.emit("L18", trace.ActionVisit, node, "Query left child; its subtotal is pending.")
	left := tr.query(2 * node)
	tr.clearResult()
	tr.s.Overlap = "partial"
	tr.s.LeftResult = left
	tr.s.LeftKnown = true
	tr.emit("L19", trace.ActionVisit, node, fmt.Sprintf("Left returned %d; query right child next.", left))
	right := tr.query(2*node + 1)
	tr.clearResult()
	tr.s.Overlap = "partial"
	tr.s.LeftResult, tr.s.RightResult, tr.s.Result = left, right, left+right
	tr.s.LeftKnown, tr.s.RightKnown, tr.s.ResultKnown = true, true, true
	tr.b.IncOperations()
	tr.emit("L20", trace.ActionDone, node, fmt.Sprintf("Return %d + %d = %d. Accepted cover is unchanged.", left, right, left+right))
	return left + right
}

func (tr *tracer) update(node int) {
	i := tr.nodes[node]
	n := tr.s.Nodes[i]
	tr.s.PathIDs = append(tr.s.PathIDs, n.ID)
	tr.clearResult()
	tr.emit("L24", trace.ActionVisit, node, fmt.Sprintf("Interval [%d, %d] covers update index %d.", n.L, n.R, tr.s.UpdateIndex))
	if n.L == n.R {
		tr.s.Nodes[i].Sum = tr.s.UpdateValue
		tr.s.Values[n.L] = tr.s.UpdateValue
		tr.b.IncOperations()
		tr.emit("L25", trace.ActionSet, node, fmt.Sprintf("Set array[%d] from %d to %d.", n.L, n.Sum, tr.s.UpdateValue))
		return
	}
	mid := (n.L + n.R) / 2
	tr.b.IncComparisons()
	tr.emit("L26", trace.ActionCompare, node, fmt.Sprintf("Compare update index %d with midpoint %d.", tr.s.UpdateIndex, mid))
	if tr.s.UpdateIndex <= mid {
		tr.emit("L27", trace.ActionVisit, node, "Target is in left interval; other child stays unchanged.")
		tr.update(2 * node)
	} else {
		tr.emit("L28", trace.ActionVisit, node, "Target is in right interval; other child stays unchanged.")
		tr.update(2*node + 1)
	}
	tr.combine(node)
	tr.emit("L29", trace.ActionSet, node, fmt.Sprintf("Recompute ancestor: %d + %d = %d.", tr.s.LeftResult, tr.s.RightResult, tr.s.Result))
	tr.emit("L30", trace.ActionDone, node, "Ancestor sum settled; return upward.")
}

// Trace preserves direct-call defaults and normalization; Lesson.Build validates browser input.
func Trace(arr []int, ql, qr, updIdx, updVal int) []trace.Step {
	if len(arr) == 0 {
		arr = []int{1, 3, 5, 7, 9, 11}
		ql, qr = 1, 4
		updIdx, updVal = 2, 6
	}
	n := len(arr)
	if ql < 0 {
		ql = 0
	}
	if qr >= n {
		qr = n - 1
	}
	if ql > qr {
		ql, qr = 0, n-1
	}
	if updIdx < 0 || updIdx >= n {
		updIdx = 0
	}
	tr := &tracer{b: trace.NewBuilder(), nodes: make(map[int]int), s: renderers.SegmentTreeLessonState{Values: append([]int(nil), arr...), Phase: "build", QL: ql, QR: qr, UpdateIndex: updIdx, UpdateValue: updVal}}
	var layout func(int, int, int, string)
	layout = func(node, l, r int, parent string) {
		id := fmt.Sprintf("node-%d", node)
		tr.nodes[node] = len(tr.s.Nodes)
		tr.s.Nodes = append(tr.s.Nodes, renderers.SegmentTreeLessonNode{ID: id, ParentID: parent, L: l, R: r})
		if l < r {
			mid := (l + r) / 2
			layout(2*node, l, mid, id)
			layout(2*node+1, mid+1, r, id)
		}
	}
	layout(1, 0, n-1, "")
	tr.emit("L1", trace.ActionInit, 0, fmt.Sprintf("Build intervals for %v. Query [%d, %d]; update index %d to %d (normalized parameters).", arr, ql, qr, updIdx, updVal))
	tr.emit("L2", trace.ActionVisit, 1, "Build root interval.")
	tr.build(1)
	tr.clearResult()
	tr.s.Result = tr.s.Nodes[0].Sum
	tr.s.ResultKnown = true
	tr.emit("L3", trace.ActionDone, 1, fmt.Sprintf("Build complete: root sum %d.", tr.s.Result))
	tr.s.Phase = "query"
	tr.clearResult()
	tr.emit("L11", trace.ActionInit, 0, fmt.Sprintf("Query inclusive range [%d, %d].", ql, qr))
	tr.emit("L12", trace.ActionVisit, 1, "Query root interval.")
	ans := tr.query(1)
	tr.clearResult()
	tr.s.Result = ans
	tr.s.ResultKnown = true
	tr.emit("L13", trace.ActionDone, 1, fmt.Sprintf("Query complete: disjoint accepted intervals sum to %d.", ans))
	tr.s.Phase = "update"
	tr.s.AcceptedIDs = nil
	tr.clearResult()
	tr.emit("L21", trace.ActionInit, 0, fmt.Sprintf("Update index %d to %d; only covering ancestors change.", updIdx, updVal))
	tr.emit("L22", trace.ActionVisit, 1, "Start update at root.")
	tr.update(1)
	tr.clearResult()
	tr.s.Result = tr.s.Nodes[0].Sum
	tr.s.ResultKnown = true
	tr.emit("L23", trace.ActionDone, 1, fmt.Sprintf("Update complete: root sum %d. No second query was executed.", tr.s.Result))
	return tr.b.Steps()
}
