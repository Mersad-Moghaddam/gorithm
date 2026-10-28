package kosaraju

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records actual recursive calls; unknown endpoints are ignored by the direct API.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}
	s := renderers.KosarajuLessonState{GraphState: renderers.GraphState{Nodes: slices.Clone(nodes), Edges: slices.Clone(edges)}, Pass: 1, OrderIndex: -1, EdgeIndex: -1, OriginalEdges: slices.Clone(edges), ComponentOf: map[string]int{}, VisitedForward: map[string]bool{}, VisitedReverse: map[string]bool{}}
	idx := map[string]int{}
	for i, n := range s.Nodes {
		idx[n.ID] = i
		s.Nodes[i].State = "unvisited"
	}
	adj, rev := map[string][]int{}, map[string][]int{}
	for i, e := range edges {
		if _, ok := idx[e.From]; !ok {
			continue
		}
		if _, ok := idx[e.To]; !ok {
			continue
		}
		adj[e.From] = append(adj[e.From], i)
		rev[e.To] = append(rev[e.To], i)
	}
	add := func(line string, action trace.ActionType, text string) {
		out := s
		out.Nodes = slices.Clone(s.Nodes)
		out.Edges = slices.Clone(s.Edges)
		out.OriginalEdges = slices.Clone(s.OriginalEdges)
		out.Stack = slices.Clone(s.Stack)
		out.FinishOrder = slices.Clone(s.FinishOrder)
		out.CurrentComponent = slices.Clone(s.CurrentComponent)
		out.Components = cloneComponents(s.Components)
		out.ComponentOf = maps.Clone(s.ComponentOf)
		out.VisitedForward = maps.Clone(s.VisitedForward)
		out.VisitedReverse = maps.Clone(s.VisitedReverse)
		b.AddStep(line, action, text, out, map[string]any{"components": cloneComponents(s.Components), "order": slices.Clone(s.FinishOrder), "pass": s.Pass, "u": s.Current, "order_index": s.OrderIndex, "current_component": slices.Clone(s.CurrentComponent)})
	}
	add("L1", trace.ActionInit, "Find strongly connected components: every pair must have paths in both directions.")
	add("L2", trace.ActionSet, "Build forward and reverse adjacency lists; retain original edge indices.")
	add("L3", trace.ActionSet, "Start forward DFS with empty visited map and finish stack.")
	var dfs func(string, bool)
	dfs = func(u string, reverse bool) {
		s.Stack = append(s.Stack, u)
		s.Current = u
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.FinishedVertex = ""
		s.ReturningFrom = ""
		line := "L5"
		links := adj[u]
		seen := s.VisitedForward
		if reverse {
			line = "L12"
			links = rev[u]
			seen = s.VisitedReverse
			s.CurrentComponent = append(s.CurrentComponent, u)
			s.ComponentOf[u] = len(s.Components)
		}
		seen[u] = true
		s.Nodes[idx[u]].State = "visiting"
		b.IncOperations()
		add(line, trace.ActionVisit, fmt.Sprintf("Pass %d: enter %s; call stack %v.", s.Pass, u, s.Stack))
		for _, ei := range links {
			v := edges[ei].To
			if reverse {
				v = edges[ei].From
			}
			s.Current = u
			s.Neighbor = v
			s.EdgeIndex = ei
			s.ReturningFrom = ""
			s.FinishedVertex = ""
			s.Edges[ei].Active = true
			b.IncComparisons()
			add(line, trace.ActionHighlight, fmt.Sprintf("Probe edge #%d: %s → %s; visited: %t.", ei, u, v, seen[v]))
			if !seen[v] {
				dfs(v, reverse)
				s.Current = u
				s.Neighbor = v
				s.EdgeIndex = ei
				s.ReturningFrom = v
				s.FinishedVertex = ""
				add(line, trace.ActionSet, fmt.Sprintf("Return from %s to %s.", v, u))
			}
			s.Edges[ei].Active = false
		}
		s.Current = u
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.ReturningFrom = ""
		s.Stack = s.Stack[:len(s.Stack)-1]
		s.Nodes[idx[u]].State = "visited"
		if !reverse {
			s.FinishOrder = append(s.FinishOrder, u)
			s.FinishedVertex = u
			b.IncOperations()
			add(line, trace.ActionPush, fmt.Sprintf("Finish %s; push onto finish stack.", u))
		} else {
			add(line, trace.ActionSet, fmt.Sprintf("Reverse call for %s complete.", u))
		}
	}
	for _, n := range nodes {
		s.Current = n.ID
		s.FinishedVertex = ""
		s.ReturningFrom = ""
		b.IncComparisons()
		add("L4", trace.ActionCompare, fmt.Sprintf("Check root %s; visited: %t.", n.ID, s.VisitedForward[n.ID]))
		if !s.VisitedForward[n.ID] {
			dfs(n.ID, false)
		}
	}
	s.Current = ""
	s.FinishedVertex = ""
	add("L6", trace.ActionSet, "Forward pass complete; finish order is not discovery order.")
	s.Pass = 2
	s.Transposed = true
	for i := range s.Nodes {
		s.Nodes[i].State = "unvisited"
	}
	for i := range s.Edges {
		s.Edges[i].From, s.Edges[i].To = s.Edges[i].To, s.Edges[i].From
		s.Edges[i].Active = false
	}
	add("L7", trace.ActionSet, "Switch display to transpose already built at L2; reset reverse visited map.")
	for i := len(s.FinishOrder) - 1; i >= 0; i-- {
		u := s.FinishOrder[i]
		s.Current = u
		s.OrderIndex = i
		s.ReturningFrom = ""
		b.IncComparisons()
		add("L8", trace.ActionCompare, fmt.Sprintf("Inspect finish index %d: %s; assigned: %t.", i, u, s.VisitedReverse[u]))
		if !s.VisitedReverse[u] {
			s.CurrentComponent = nil
			add("L11", trace.ActionSet, fmt.Sprintf("Start empty component seeded by %s.", u))
			dfs(u, true)
			s.Components = append(s.Components, slices.Clone(s.CurrentComponent))
			add("L13", trace.ActionSet, fmt.Sprintf("Commit SCC %d: %v. Paths exist both ways.", len(s.Components), s.CurrentComponent))
		}
	}
	s.Transposed = false
	s.Current = ""
	s.Neighbor = ""
	s.OrderIndex = -1
	s.EdgeIndex = -1
	s.CurrentComponent = nil
	s.ReturningFrom = ""
	s.Edges = slices.Clone(s.OriginalEdges)
	for i := range s.Edges {
		s.Edges[i].Active = false
	}
	add("L16", trace.ActionDone, fmt.Sprintf("Restore original graph: %d SCCs. One-way bridges do not merge components.", len(s.Components)))
	return b.Steps()
}

func cloneComponents(c [][]string) [][]string {
	out := make([][]string, len(c))
	for i := range c {
		out[i] = slices.Clone(c[i])
	}
	return out
}
