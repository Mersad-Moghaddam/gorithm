package segmenttree

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "segmenttree",
		Name:     "Segment Tree (Range Queries)",
		Category: catalog.CategoryTrees,
		Renderer: "tree",
		Documentation: catalog.Documentation{
			Summary:         "A tree data structure for storing intervals or segments, allowing fast range sum queries and point updates in logarithmic time.",
			TimeComplexity:  "O(log n) query/update, O(n) build",
			SpaceComplexity: "O(4n)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Range Decomposition",
				"Tree Representation of Array Intervals",
				"Logarithmic Point Updates",
			},
			HowItWorks: "A Segment Tree is a full binary tree where each node represents an aggregate (such as sum) over a contiguous subarray segment [L, R].\n\n1. Build: Recursively splits array into halves until single-element leaf nodes [i, i] are reached, then computes parent values by combining children (O(n) time, O(4n) space).\n2. Range Query: Traverses tree nodes that partition [queryL, queryR]. If a node segment is completely inside the query range, its precomputed sum is used immediately; disjoint nodes are skipped, and partial overlaps recurse into both halves (at most O(log n) nodes visited).\n3. Point Update: Navigates down from root to target leaf, updates its value, and recalculates aggregate sums along return path to root in O(log n) time.",
		},
		DefaultInput: `{"arr": [1, 3, 5, 7, 9, 11], "query": [1, 4], "update": {"index": 2, "value": 6}}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func (st *SegmentTree) Build(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    st.build(1, 0, len(arr)-1)"},
		{ID: "L3", Number: 3, Code: "}"},
		{ID: "L4", Number: 4, Code: "func (st *SegmentTree) build(node, l, r int) {"},
		{ID: "L5", Number: 5, Code: "    if l == r { st.tree[node] = st.arr[l]; return }"},
		{ID: "L6", Number: 6, Code: "    mid := (l + r) / 2"},
		{ID: "L7", Number: 7, Code: "    st.build(2*node, l, mid)"},
		{ID: "L8", Number: 8, Code: "    st.build(2*node+1, mid+1, r)"},
		{ID: "L9", Number: 9, Code: "    st.tree[node] = st.tree[2*node] + st.tree[2*node+1]"},
		{ID: "L10", Number: 10, Code: "}"},
		{ID: "L11", Number: 11, Code: "func (st *SegmentTree) Query(ql, qr int) int {"},
		{ID: "L12", Number: 12, Code: "    return st.query(1, 0, st.n-1, ql, qr)"},
		{ID: "L13", Number: 13, Code: "}"},
		{ID: "L14", Number: 14, Code: "func (st *SegmentTree) query(node, l, r, ql, qr int) int {"},
		{ID: "L15", Number: 15, Code: "    if ql <= l && r <= qr { return st.tree[node] }"},
		{ID: "L16", Number: 16, Code: "    if r < ql || l > qr { return 0 }"},
		{ID: "L17", Number: 17, Code: "    mid := (l + r) / 2"},
		{ID: "L18", Number: 18, Code: "    return st.query(2*node, l, mid, ql, qr) +"},
		{ID: "L19", Number: 19, Code: "           st.query(2*node+1, mid+1, r, ql, qr)"},
		{ID: "L20", Number: 20, Code: "}"},
		{ID: "L21", Number: 21, Code: "func (st *SegmentTree) Update(idx, val int) {"},
		{ID: "L22", Number: 22, Code: "    st.update(1, 0, st.n-1, idx, val)"},
		{ID: "L23", Number: 23, Code: "}"},
		{ID: "L24", Number: 24, Code: "func (st *SegmentTree) update(node, l, r, idx, val int) {"},
		{ID: "L25", Number: 25, Code: "    if l == r { st.tree[node] = val; return }"},
		{ID: "L26", Number: 26, Code: "    mid := (l + r) / 2"},
		{ID: "L27", Number: 27, Code: "    if idx <= mid { st.update(2*node, l, mid, idx, val) }"},
		{ID: "L28", Number: 28, Code: "    else { st.update(2*node+1, mid+1, r, idx, val) }"},
		{ID: "L29", Number: 29, Code: "    st.tree[node] = st.tree[2*node] + st.tree[2*node+1]"},
		{ID: "L30", Number: 30, Code: "}"},
	}
}

type updateOp struct {
	Index int `json:"index"`
	Value int `json:"value"`
}

type buildInput struct {
	Arr    []int     `json:"arr"`
	Query  []int     `json:"query"`
	Update *updateOp `json:"update"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{1, 3, 5, 7, 9, 11}
	ql, qr := 1, 4
	updIdx, updVal := 2, 6

	if len(input) > 0 {
		var custom buildInput
		if err := json.Unmarshal(input, &custom); err == nil && len(custom.Arr) > 0 {
			arr = custom.Arr
			if len(custom.Query) == 2 {
				ql, qr = custom.Query[0], custom.Query[1]
			} else {
				ql, qr = 0, len(arr)-1
			}
			if custom.Update != nil {
				updIdx = custom.Update.Index
				updVal = custom.Update.Value
			}
		} else {
			var customArr []int
			if err := json.Unmarshal(input, &customArr); err == nil && len(customArr) > 0 {
				arr = customArr
				ql, qr = 0, len(arr)-1
			}
		}
	}

	return Trace(arr, ql, qr, updIdx, updVal), nil
}

func init() {
	catalog.Register(Lesson{})
}
