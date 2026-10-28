package kruskal

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace preserves original edge identities, including parallel edges.
func Trace(nodes []renderers.GraphNode, edges []WeightedEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}
	s := renderers.KruskalLessonState{GraphState: renderers.GraphState{Nodes: slices.Clone(nodes)}, Cursor: -1, ActiveEdge: -1}
	ids := make([]string, len(nodes))
	for i, n := range nodes {
		ids[i] = n.ID
	}
	for i, e := range edges {
		s.Edges = append(s.Edges, renderers.GraphEdge{From: e.From, To: e.To, Weight: e.Weight})
		s.SortedEdgeIndices = append(s.SortedEdgeIndices, i)
	}
	var dsu *DSU
	add := func(line string, action trace.ActionType, explanation string, vars map[string]any) {
		snapshot := s
		snapshot.Nodes = slices.Clone(s.Nodes)
		snapshot.Edges = slices.Clone(s.Edges)
		snapshot.SortedEdgeIndices = slices.Clone(s.SortedEdgeIndices)
		snapshot.SelectedEdges = slices.Clone(s.SelectedEdges)
		snapshot.RejectedEdges = slices.Clone(s.RejectedEdges)
		if dsu != nil {
			snapshot.Parent = maps.Clone(dsu.parent)
			snapshot.Rank = maps.Clone(dsu.rank)
			snapshot.Components = nil
			group := map[string]int{}
			for _, id := range ids {
				root := id
				for dsu.parent[root] != root {
					root = dsu.parent[root]
				}
				index, ok := group[root]
				if !ok {
					index = len(snapshot.Components)
					group[root] = index
					snapshot.Components = append(snapshot.Components, nil)
				}
				snapshot.Components[index] = append(snapshot.Components[index], id)
			}
			snapshot.Connected = len(snapshot.Components) == 1
		}
		b.AddStep(line, action, explanation, snapshot, vars)
	}
	add("L1", trace.ActionInit, fmt.Sprintf("Start with %d vertices and %d undirected edges.", len(nodes), len(edges)), map[string]any{"vertices": len(nodes), "edges": len(edges)})
	slices.SortFunc(s.SortedEdgeIndices, func(a, b int) int { return cmp.Compare(edges[a].Weight, edges[b].Weight) })
	b.IncOperations()
	add("L2", trace.ActionHighlight, "Sort edges by ascending weight; original edge numbers stay attached.", nil)
	dsu = NewDSU(ids)
	add("L3", trace.ActionSet, "Each vertex starts in its own component.", map[string]any{"mstCount": 0, "totalWeight": 0})
	for cursor, index := range s.SortedEdgeIndices {
		edge := edges[index]
		s.Cursor = cursor
		s.ActiveEdge = index
		s.Cycle = false
		s.RootFrom = ""
		s.RootTo = ""
		add("L4", trace.ActionHighlight, fmt.Sprintf("Inspect edge #%d (%s, %s), weight %d.", index, edge.From, edge.To, edge.Weight), map[string]any{"from": edge.From, "to": edge.To, "weight": edge.Weight})
		b.IncComparisons()
		s.RootFrom = dsu.Find(edge.From)
		s.RootTo = dsu.Find(edge.To)
		s.Cycle = s.RootFrom == s.RootTo
		if s.Cycle {
			s.RejectedEdges = append(s.RejectedEdges, index)
			add("L5", trace.ActionCompare, fmt.Sprintf("Both roots are %s. Reject edge #%d: same component, hence a cycle.", s.RootFrom, index), map[string]any{"cycle": true, "rejected": true, "root": s.RootFrom})
			continue
		}
		add("L5", trace.ActionCompare, fmt.Sprintf("Roots %s and %s differ. Union is safe.", s.RootFrom, s.RootTo), map[string]any{"cycle": false, "rootFrom": s.RootFrom, "rootTo": s.RootTo})
		dsu.Union(edge.From, edge.To)
		b.IncOperations()
		add("L6", trace.ActionSet, fmt.Sprintf("Merge components rooted at %s and %s. Edge not accepted yet.", s.RootFrom, s.RootTo), map[string]any{"componentA": s.RootFrom, "componentB": s.RootTo})
		s.SelectedEdges = append(s.SelectedEdges, index)
		s.TotalWeight += edge.Weight
		s.Edges[index].Active = true
		add("L7", trace.ActionPush, fmt.Sprintf("Accept edge #%d. Cost becomes %d.", index, s.TotalWeight), map[string]any{"mstCount": len(s.SelectedEdges), "totalWeight": s.TotalWeight})
		if len(s.SelectedEdges) == len(nodes)-1 {
			add("L8", trace.ActionHighlight, "The spanning tree has V−1 edges; stop scanning.", nil)
			break
		}
	}
	s.ActiveEdge = -1
	kind := "forest"
	if len(s.SelectedEdges) == len(nodes)-1 {
		kind = "tree"
	}
	add("L10", trace.ActionDone, fmt.Sprintf("Minimum spanning %s: %d edges, total weight %d.", kind, len(s.SelectedEdges), s.TotalWeight), map[string]any{"totalWeight": s.TotalWeight, "status": "complete"})
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L5" && steps[i+1].LineID == "L6" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
