package unionfind

// DSU implements Disjoint Set Union with path compression and union by rank.
// ponytail: array-backed dense [0..n-1] keys; use map-backed indices for arbitrary IDs.
type DSU struct {
	parent []int
	rank   []int
	count  int
}

// New initializes DSU with n elements (0 to n-1).
func New(n int) *DSU {
	if n < 0 {
		n = 0
	}
	parent := make([]int, n)
	rank := make([]int, n)
	for i := 0; i < n; i++ {
		parent[i] = i
	}
	return &DSU{
		parent: parent,
		rank:   rank,
		count:  n,
	}
}

// Find returns representative root of element i with path compression.
func (d *DSU) Find(i int) int {
	if i < 0 || i >= len(d.parent) {
		return -1
	}
	if d.parent[i] != i {
		d.parent[i] = d.Find(d.parent[i])
	}
	return d.parent[i]
}

// Union merges sets containing elements i and j using rank heuristic.
func (d *DSU) Union(i, j int) bool {
	rootI := d.Find(i)
	rootJ := d.Find(j)
	if rootI == -1 || rootJ == -1 || rootI == rootJ {
		return false
	}

	if d.rank[rootI] < d.rank[rootJ] {
		d.parent[rootI] = rootJ
	} else if d.rank[rootI] > d.rank[rootJ] {
		d.parent[rootJ] = rootI
	} else {
		d.parent[rootJ] = rootI
		d.rank[rootI]++
	}
	d.count--
	return true
}

// Connected reports whether elements i and j share the same root.
func (d *DSU) Connected(i, j int) bool {
	if i < 0 || i >= len(d.parent) || j < 0 || j >= len(d.parent) {
		return false
	}
	return d.Find(i) == d.Find(j)
}

// Count returns number of disjoint sets.
func (d *DSU) Count() int {
	return d.count
}
