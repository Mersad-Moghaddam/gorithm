package segmenttree

// SegmentTree implements range sum queries and point updates.
// ponytail: array-backed 4n tree; dynamic pointer nodes save memory if array is sparse.
type SegmentTree struct {
	arr  []int
	tree []int
	n    int
}

// BuildTree constructs a new SegmentTree from arr.
func BuildTree(arr []int) *SegmentTree {
	st := &SegmentTree{}
	st.BuildTree(arr)
	return st
}

// BuildTree populates the SegmentTree with elements from arr.
func (st *SegmentTree) BuildTree(arr []int) {
	st.n = len(arr)
	st.arr = make([]int, st.n)
	copy(st.arr, arr)
	if st.n == 0 {
		st.tree = nil
		return
	}
	st.tree = make([]int, 4*st.n)
	st.build(1, 0, st.n-1)
}

func (st *SegmentTree) build(node, l, r int) {
	if l == r {
		st.tree[node] = st.arr[l]
		return
	}
	mid := (l + r) / 2
	st.build(2*node, l, mid)
	st.build(2*node+1, mid+1, r)
	st.tree[node] = st.tree[2*node] + st.tree[2*node+1]
}

// Query returns the sum in range [left, right] inclusive.
func (st *SegmentTree) Query(left, right int) int {
	if st.n == 0 || left > right || right < 0 || left >= st.n {
		return 0
	}
	if left < 0 {
		left = 0
	}
	if right >= st.n {
		right = st.n - 1
	}
	return st.query(1, 0, st.n-1, left, right)
}

func (st *SegmentTree) query(node, l, r, ql, qr int) int {
	if ql <= l && r <= qr {
		return st.tree[node]
	}
	if r < ql || l > qr {
		return 0
	}
	mid := (l + r) / 2
	return st.query(2*node, l, mid, ql, qr) + st.query(2*node+1, mid+1, r, ql, qr)
}

// Update updates the element at index to val.
func (st *SegmentTree) Update(index, val int) {
	if st.n == 0 || index < 0 || index >= st.n {
		return
	}
	st.arr[index] = val
	st.update(1, 0, st.n-1, index, val)
}

func (st *SegmentTree) update(node, l, r, idx, val int) {
	if l == r {
		st.tree[node] = val
		return
	}
	mid := (l + r) / 2
	if idx <= mid {
		st.update(2*node, l, mid, idx, val)
	} else {
		st.update(2*node+1, mid+1, r, idx, val)
	}
	st.tree[node] = st.tree[2*node] + st.tree[2*node+1]
}
