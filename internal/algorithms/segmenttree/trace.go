package segmenttree

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type nodeInfo struct {
	id       string
	parentID string
	l, r     int
	x, y     float64
}

type tracer struct {
	b         *trace.Builder
	arr       []int
	tree      []int
	n         int
	nodeMap   map[int]nodeInfo
	nodeOrder []int
}

func newTracer(arr []int) *tracer {
	n := len(arr)
	tr := &tracer{
		b:       trace.NewBuilder(),
		arr:     make([]int, n),
		tree:    make([]int, 4*n+1),
		n:       n,
		nodeMap: make(map[int]nodeInfo),
	}
	copy(tr.arr, arr)

	if n > 0 {
		var recurseLayout func(node, l, r int, x, y, dx float64, parentID string)
		recurseLayout = func(node, l, r int, x, y, dx float64, parentID string) {
			id := fmt.Sprintf("node-%d", node)
			tr.nodeMap[node] = nodeInfo{
				id:       id,
				parentID: parentID,
				l:        l,
				r:        r,
				x:        x,
				y:        y,
			}
			tr.nodeOrder = append(tr.nodeOrder, node)
			if l == r {
				return
			}
			mid := (l + r) / 2
			nextDx := dx / 2
			if nextDx < 20 {
				nextDx = 20
			}
			recurseLayout(2*node, l, mid, x-dx, y+70, nextDx, id)
			recurseLayout(2*node+1, mid+1, r, x+dx, y+70, nextDx, id)
		}
		recurseLayout(1, 0, n-1, 400, 60, 190, "")
	}

	return tr
}

func (tr *tracer) treeState(nodeStates map[int]string) renderers.TreeState {
	treeNodes := make([]renderers.TreeNode, 0, len(tr.nodeOrder))
	for _, nodeIdx := range tr.nodeOrder {
		info := tr.nodeMap[nodeIdx]
		st := "normal"
		if nodeStates != nil {
			if s, ok := nodeStates[nodeIdx]; ok {
				st = s
			}
		}
		treeNodes = append(treeNodes, renderers.TreeNode{
			ID:       info.id,
			Val:      tr.tree[nodeIdx],
			X:        info.x,
			Y:        info.y,
			ParentID: info.parentID,
			State:    st,
		})
	}
	return renderers.TreeState{Nodes: treeNodes}
}

func (tr *tracer) traceBuild(node, l, r int) {
	tr.b.AddStep("L4", trace.ActionVisit, fmt.Sprintf("Inspect node %d for range [%d, %d].", node, l, r),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"node": node, "l": l, "r": r})

	if l == r {
		tr.tree[node] = tr.arr[l]
		tr.b.IncOperations()
		tr.b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Leaf node %d covers index %d: value = %d.", node, l, tr.arr[l]),
			tr.treeState(map[int]string{node: "inserted"}),
			map[string]any{"node": node, "l": l, "val": tr.arr[l]})
		return
	}

	mid := (l + r) / 2
	tr.b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Split range [%d, %d] at mid = %d.", l, r, mid),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"node": node, "l": l, "r": r, "mid": mid})

	tr.b.AddStep("L7", trace.ActionVisit, fmt.Sprintf("Recurse left child %d for range [%d, %d].", 2*node, l, mid),
		tr.treeState(map[int]string{node: "active", 2 * node: "active"}),
		map[string]any{"node": 2 * node, "l": l, "r": mid})
	tr.traceBuild(2*node, l, mid)

	tr.b.AddStep("L8", trace.ActionVisit, fmt.Sprintf("Recurse right child %d for range [%d, %d].", 2*node+1, mid+1, r),
		tr.treeState(map[int]string{node: "active", 2*node + 1: "active"}),
		map[string]any{"node": 2*node + 1, "l": mid + 1, "r": r})
	tr.traceBuild(2*node+1, mid+1, r)

	tr.tree[node] = tr.tree[2*node] + tr.tree[2*node+1]
	tr.b.IncOperations()
	tr.b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Combine children sums for node %d: %d + %d = %d.", node, tr.tree[2*node], tr.tree[2*node+1], tr.tree[node]),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"node": node, "sum": tr.tree[node]})

	tr.b.AddStep("L10", trace.ActionDone, fmt.Sprintf("Subtree at node %d built with sum %d.", node, tr.tree[node]),
		tr.treeState(nil),
		map[string]any{"node": node, "sum": tr.tree[node]})
}

func (tr *tracer) traceQuery(node, l, r, ql, qr int) int {
	tr.b.IncComparisons()
	tr.b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Query inspecting node %d [%d, %d] for range [%d, %d].", node, l, r, ql, qr),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"L": ql, "R": qr, "node": node, "l": l, "r": r})

	if ql <= l && r <= qr {
		tr.b.IncComparisons()
		tr.b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Node %d [%d, %d] fully inside [%d, %d]: return sum %d.", node, l, r, ql, qr, tr.tree[node]),
			tr.treeState(map[int]string{node: "matched"}),
			map[string]any{"L": ql, "R": qr, "node": node, "sum": tr.tree[node]})
		return tr.tree[node]
	}

	if r < ql || l > qr {
		tr.b.IncComparisons()
		tr.b.AddStep("L16", trace.ActionDone, fmt.Sprintf("Node %d [%d, %d] disjoint from [%d, %d]: return 0.", node, l, r, ql, qr),
			tr.treeState(nil),
			map[string]any{"L": ql, "R": qr, "node": node})
		return 0
	}

	mid := (l + r) / 2
	tr.b.AddStep("L17", trace.ActionSet, fmt.Sprintf("Partial overlap at node %d [%d, %d]. Split at mid = %d.", node, l, r, mid),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"L": ql, "R": qr, "mid": mid})

	tr.b.AddStep("L18", trace.ActionVisit, fmt.Sprintf("Query left child %d [%d, %d].", 2*node, l, mid),
		tr.treeState(map[int]string{node: "active", 2 * node: "active"}),
		map[string]any{"L": ql, "R": qr, "child": 2 * node})
	leftSum := tr.traceQuery(2*node, l, mid, ql, qr)

	tr.b.AddStep("L19", trace.ActionVisit, fmt.Sprintf("Query right child %d [%d, %d].", 2*node+1, mid+1, r),
		tr.treeState(map[int]string{node: "active", 2*node + 1: "active"}),
		map[string]any{"L": ql, "R": qr, "child": 2*node + 1})
	rightSum := tr.traceQuery(2*node+1, mid+1, r, ql, qr)

	total := leftSum + rightSum
	tr.b.IncOperations()
	tr.b.AddStep("L20", trace.ActionDone, fmt.Sprintf("Combine subqueries for node %d: %d + %d = %d.", node, leftSum, rightSum, total),
		tr.treeState(map[int]string{node: "matched"}),
		map[string]any{"L": ql, "R": qr, "sum": total})
	return total
}

func (tr *tracer) traceUpdate(node, l, r, idx, val int) {
	tr.b.AddStep("L24", trace.ActionVisit, fmt.Sprintf("Update inspecting node %d [%d, %d] for target index %d.", node, l, r, idx),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"node": node, "idx": idx, "val": val})

	if l == r {
		tr.tree[node] = val
		tr.arr[idx] = val
		tr.b.IncOperations()
		tr.b.AddStep("L25", trace.ActionSet, fmt.Sprintf("Reached leaf node %d (index %d). Set value to %d.", node, idx, val),
			tr.treeState(map[int]string{node: "inserted"}),
			map[string]any{"node": node, "idx": idx, "val": val})
		return
	}

	mid := (l + r) / 2
	tr.b.IncComparisons()
	tr.b.AddStep("L26", trace.ActionCompare, fmt.Sprintf("Check index %d against mid = %d.", idx, mid),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"idx": idx, "mid": mid})

	if idx <= mid {
		tr.b.AddStep("L27", trace.ActionVisit, fmt.Sprintf("Index %d <= mid %d: traverse left child %d.", idx, mid, 2*node),
			tr.treeState(map[int]string{node: "active", 2 * node: "active"}),
			map[string]any{"child": 2 * node})
		tr.traceUpdate(2*node, l, mid, idx, val)
	} else {
		tr.b.AddStep("L28", trace.ActionVisit, fmt.Sprintf("Index %d > mid %d: traverse right child %d.", idx, mid, 2*node+1),
			tr.treeState(map[int]string{node: "active", 2*node + 1: "active"}),
			map[string]any{"child": 2*node + 1})
		tr.traceUpdate(2*node+1, mid+1, r, idx, val)
	}

	tr.tree[node] = tr.tree[2*node] + tr.tree[2*node+1]
	tr.b.IncOperations()
	tr.b.AddStep("L29", trace.ActionSet, fmt.Sprintf("Recalculate ancestor node %d sum: %d + %d = %d.", node, tr.tree[2*node], tr.tree[2*node+1], tr.tree[node]),
		tr.treeState(map[int]string{node: "active"}),
		map[string]any{"node": node, "newSum": tr.tree[node]})

	tr.b.AddStep("L30", trace.ActionDone, fmt.Sprintf("Node %d updated with new sum %d.", node, tr.tree[node]),
		tr.treeState(nil),
		map[string]any{"node": node, "newSum": tr.tree[node]})
}

// Trace executes Segment Tree build, range query, and point update, emitting execution steps.
// ponytail: array-backed 4n tree; dynamic pointer nodes save memory if array is sparse.
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

	tr := newTracer(arr)

	// Phase 1: Build
	tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Build segment tree from array %v.", arr),
		tr.treeState(nil),
		map[string]any{"n": n})
	tr.b.AddStep("L2", trace.ActionVisit, fmt.Sprintf("Call build on root node 1 covering range [0, %d].", n-1),
		tr.treeState(map[int]string{1: "active"}),
		map[string]any{"node": 1, "l": 0, "r": n - 1})
	tr.traceBuild(1, 0, n-1)
	tr.b.AddStep("L3", trace.ActionDone, fmt.Sprintf("Segment tree built. Root total sum = %d.", tr.tree[1]),
		tr.treeState(nil),
		map[string]any{"totalSum": tr.tree[1]})

	// Phase 2: Range Query
	tr.b.AddStep("L11", trace.ActionInit, fmt.Sprintf("Begin Query for range [%d, %d].", ql, qr),
		tr.treeState(nil),
		map[string]any{"L": ql, "R": qr, "op": "query"})
	tr.b.AddStep("L12", trace.ActionVisit, fmt.Sprintf("Query root node 1 [0, %d] for range [%d, %d].", n-1, ql, qr),
		tr.treeState(map[int]string{1: "active"}),
		map[string]any{"L": ql, "R": qr, "node": 1})
	ans := tr.traceQuery(1, 0, n-1, ql, qr)
	tr.b.AddStep("L13", trace.ActionDone, fmt.Sprintf("Query for range [%d, %d] completed with sum = %d.", ql, qr, ans),
		tr.treeState(nil),
		map[string]any{"L": ql, "R": qr, "result": ans})

	// Phase 3: Point Update
	tr.b.AddStep("L21", trace.ActionInit, fmt.Sprintf("Begin Update(index=%d, val=%d).", updIdx, updVal),
		tr.treeState(nil),
		map[string]any{"idx": updIdx, "val": updVal, "op": "update"})
	tr.b.AddStep("L22", trace.ActionVisit, fmt.Sprintf("Start update at root node 1 covering range [0, %d].", n-1),
		tr.treeState(map[int]string{1: "active"}),
		map[string]any{"node": 1, "idx": updIdx, "val": updVal})
	tr.traceUpdate(1, 0, n-1, updIdx, updVal)
	tr.b.AddStep("L23", trace.ActionDone, fmt.Sprintf("Update complete: array[%d] is now %d. New root sum = %d.", updIdx, updVal, tr.tree[1]),
		tr.treeState(nil),
		map[string]any{"idx": updIdx, "val": updVal, "rootSum": tr.tree[1]})

	return tr.b.Steps()
}
