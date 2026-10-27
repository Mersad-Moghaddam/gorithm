package catalog

import "testing"

func TestMatches(t *testing.T) {
	m := Meta{ID: "dijkstra", Name: "Dijkstra's Algorithm", Category: CategoryGraphs,
		Documentation: Documentation{Summary: "Find shortest paths", KeyConcepts: []string{"non-negative weights"}}}
	for _, tc := range []struct {
		query    string
		category Category
		want     bool
	}{
		{"", "", true}, {"  SHORTEST weights ", "", true}, {"dijkstra", "", true},
		{"graphs", "", true}, {"shortest unknown", "", false},
		{"shortest", CategorySorting, false}, {"shortest", CategoryGraphs, true},
	} {
		if got := Matches(m, tc.query, tc.category); got != tc.want {
			t.Errorf("Matches(%q, %q) = %v, want %v", tc.query, tc.category, got, tc.want)
		}
	}
}
