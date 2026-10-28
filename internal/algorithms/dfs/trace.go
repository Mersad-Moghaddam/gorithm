package dfs

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace follows outgoing edges in input order and records actual recursive calls.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}
	if start == "" {
		start = nodes[0].ID
	}
	index := make(map[string]int, len(nodes))
	for i, n := range nodes {
		index[n.ID] = i
	}
	if _, ok := index[start]; !ok {
		return b.Steps()
	}
	s := renderers.DFSLessonState{GraphState: renderers.GraphState{Nodes: slices.Clone(nodes), Edges: slices.Clone(edges)}, Start: start, EdgeIndex: -1, Parent: map[string]string{}, Discovered: map[string]bool{}, Finished: map[string]bool{}}
	for i := range s.Nodes {
		s.Nodes[i].State = "unvisited"
	}
	adj := map[string][]int{}
	for i := range s.Edges {
		s.Edges[i].Active = false
		adj[s.Edges[i].From] = append(adj[s.Edges[i].From], i)
	}
	add := func(line string, action trace.ActionType, message string, vars map[string]any) {
		sn := s
		sn.Nodes = slices.Clone(s.Nodes)
		sn.Edges = slices.Clone(s.Edges)
		sn.Stack = slices.Clone(s.Stack)
		sn.Order = slices.Clone(s.Order)
		sn.FinishOrder = slices.Clone(s.FinishOrder)
		sn.Parent = maps.Clone(s.Parent)
		sn.Discovered = maps.Clone(s.Discovered)
		sn.Finished = maps.Clone(s.Finished)
		b.AddStep(line, action, message, sn, vars)
	}
	add("L1", trace.ActionInit, fmt.Sprintf("Start DFS at %s; follow outgoing edges in input order.", start), map[string]any{"start": start})
	add("L2", trace.ActionSet, "No vertices discovered yet.", nil)
	add("L3", trace.ActionSet, "Preorder and finish order begin empty.", nil)
	s.Stack = append(s.Stack, renderers.DFSCallFrame{Node: start})
	s.Current = start
	add("L14", trace.ActionPush, fmt.Sprintf("Push root call dfs(%s).", start), map[string]any{"target": start})
	var visit func(string)
	visit = func(u string) {
		s.Current = u
		s.Neighbor = ""
		s.ReturningFrom = ""
		s.EdgeIndex = -1
		s.Skip = false
		s.BackEdge = false
		s.Discovered[u] = true
		s.Nodes[index[u]].State = "visiting"
		b.IncOperations()
		add("L6", trace.ActionVisit, fmt.Sprintf("Discover %s; its call remains active until all outgoing edges finish.", u), map[string]any{"u": u})
		s.Order = append(s.Order, u)
		add("L7", trace.ActionPush, fmt.Sprintf("Append %s to preorder: %v.", u, s.Order), map[string]any{"order": slices.Clone(s.Order)})
		add("L8", trace.ActionHighlight, fmt.Sprintf("Inspect outgoing neighbors of %s in input order.", u), map[string]any{"u": u})
		for cursor, ei := range adj[u] {
			s.Stack[len(s.Stack)-1].NextNeighbor = cursor + 1
			v := s.Edges[ei].To
			if _, ok := index[v]; !ok {
				continue
			}
			s.Current = u
			s.Neighbor = v
			s.ReturningFrom = ""
			s.EdgeIndex = ei
			s.Skip = s.Discovered[v]
			s.BackEdge = s.Skip && !s.Finished[v]
			s.Edges[ei].Active = true
			b.IncComparisons()
			reason := fmt.Sprintf("%s is unseen; call it next.", v)
			if s.BackEdge {
				reason = fmt.Sprintf("%s is already active: back-edge to an ancestor; skip.", v)
			} else if s.Skip {
				reason = fmt.Sprintf("%s is already finished; skip, not a back-edge.", v)
			}
			add("L9", trace.ActionCompare, reason, map[string]any{"u": u, "v": v, "visited": s.Skip})
			if !s.Skip {
				s.Parent[v] = u
				s.Stack = append(s.Stack, renderers.DFSCallFrame{Node: v, Caller: u})
				s.Current = v
				add("L10", trace.ActionPush, fmt.Sprintf("Push dfs(%s), called by %s; newest call runs first.", v, u), map[string]any{"from": u, "to": v})
				visit(v)
				s.Stack = s.Stack[:len(s.Stack)-1]
				s.Current = u
				s.Neighbor = ""
				s.ReturningFrom = v
				s.EdgeIndex = ei
				s.Skip = false
				s.BackEdge = false
				s.Edges[ei].Active = false
				add("L11", trace.ActionPop, fmt.Sprintf("Return from %s to %s; parent call resumes.", v, u), map[string]any{"u": u, "backFrom": v})
			}
			s.Edges[ei].Active = false
		}
		s.Current = u
		s.Neighbor = ""
		s.ReturningFrom = ""
		s.EdgeIndex = -1
		s.Skip = false
		s.BackEdge = false
		s.Nodes[index[u]].State = "visited"
		s.Finished[u] = true
		s.FinishOrder = append(s.FinishOrder, u)
		b.IncOperations()
		add("L12", trace.ActionSet, fmt.Sprintf("All neighbors of %s processed. Finish %s, then return.", u, u), map[string]any{"u": u, "state": "visited"})
	}
	visit(start)
	s.Stack = nil
	s.Current = ""
	s.Neighbor = ""
	s.ReturningFrom = start
	add("L15", trace.ActionDone, fmt.Sprintf("DFS complete. Preorder %v; finish order %v.", s.Order, s.FinishOrder), map[string]any{"order": slices.Clone(s.Order), "status": "complete"})
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L9" && steps[i+1].LineID == "L10" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
