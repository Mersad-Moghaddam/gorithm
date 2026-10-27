package kruskal

import (
	"cmp"
	"slices"
)

type WeightedEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight"`
}

// DSU implements Disjoint Set Union with path compression and union-by-rank.
// ponytail: map-based DSU; array-indexed DSU faster for dense integer IDs.
type DSU struct {
	parent map[string]string
	rank   map[string]int
}

func NewDSU(nodes []string) *DSU {
	parent := make(map[string]string, len(nodes))
	rank := make(map[string]int, len(nodes))
	for _, n := range nodes {
		parent[n] = n
	}
	return &DSU{parent: parent, rank: rank}
}

func (d *DSU) Find(x string) string {
	if d.parent[x] == x {
		return x
	}
	d.parent[x] = d.Find(d.parent[x])
	return d.parent[x]
}

func (d *DSU) Union(x, y string) bool {
	rx := d.Find(x)
	ry := d.Find(y)
	if rx == ry {
		return false
	}
	if d.rank[rx] < d.rank[ry] {
		d.parent[rx] = ry
	} else if d.rank[rx] > d.rank[ry] {
		d.parent[ry] = rx
	} else {
		d.parent[ry] = rx
		d.rank[rx]++
	}
	return true
}

func MST(nodes []string, edges []WeightedEdge) ([]WeightedEdge, int) {
	if len(nodes) <= 1 {
		return nil, 0
	}

	sorted := slices.Clone(edges)
	slices.SortFunc(sorted, func(a, b WeightedEdge) int {
		return cmp.Compare(a.Weight, b.Weight)
	})

	dsu := NewDSU(nodes)
	mst := make([]WeightedEdge, 0, len(nodes)-1)
	totalWeight := 0

	for _, e := range sorted {
		if dsu.Union(e.From, e.To) {
			mst = append(mst, e)
			totalWeight += e.Weight
			if len(mst) == len(nodes)-1 {
				break
			}
		}
	}

	return mst, totalWeight
}
