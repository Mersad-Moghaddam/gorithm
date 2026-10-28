package toposort

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records Kahn's FIFO operations; empty graphs retain the direct API's zero-step guard.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}
	s := renderers.ToposortLessonState{GraphState: renderers.GraphState{Nodes: nodes, Edges: edges}, InDegree: map[string]int{}, Processed: map[string]bool{}, EdgeIndex: -1, OldInDegree: -1, NewInDegree: -1}
	adj := map[string][]int{}
	for _, n := range nodes {
		s.InDegree[n.ID] = 0
	}
	for i, e := range edges {
		adj[e.From] = append(adj[e.From], i)
		s.InDegree[e.To]++
	}
	add := func(line string, action trace.ActionType, explanation string) {
		snap := s
		snap.Nodes = append([]renderers.GraphNode(nil), nodes...)
		snap.Edges = append([]renderers.GraphEdge(nil), edges...)
		snap.InDegree = cloneInDegrees(s.InDegree)
		snap.Processed = map[string]bool{}
		for k, v := range s.Processed {
			snap.Processed[k] = v
		}
		snap.Queue = cloneSlice(s.Queue)
		snap.Order = cloneSlice(s.Order)
		snap.Blocked = cloneSlice(s.Blocked)
		snap.ResolvedEdges = append([]int(nil), s.ResolvedEdges...)
		vars := map[string]any{"curr": s.Current, "neighbor": s.Neighbor, "inDegrees": cloneInDegrees(s.InDegree), "order": cloneSlice(s.Order), "processed": cloneSlice(s.Order), "queue": cloneSlice(s.Queue)}
		if s.Terminal {
			vars["valid"] = s.Valid
			if s.Valid {
				vars["status"] = "complete"
			} else {
				vars["status"] = "cycle_detected"
			}
		}
		b.AddStep(line, action, explanation, snap, vars)
	}
	add("L1", trace.ActionInit, fmt.Sprintf("Count prerequisites for %d tasks and %d dependencies.", len(nodes), len(edges)))
	b.IncOperations()
	add("L2", trace.ActionSet, "Each incoming dependency contributes one prerequisite, including parallel edges.")
	for _, n := range nodes {
		b.IncComparisons()
		if s.InDegree[n.ID] == 0 {
			s.Queue = append(s.Queue, n.ID)
			b.IncOperations()
		}
	}
	add("L4", trace.ActionPush, fmt.Sprintf("Ready FIFO starts with %v, in input order.", s.Queue))
	add("L5", trace.ActionSet, "Start with an empty output order.")
	for len(s.Queue) > 0 {
		s.Current = ""
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.Decremented = false
		s.Ready = false
		s.OldInDegree = -1
		s.NewInDegree = -1
		b.IncComparisons()
		add("L6", trace.ActionCompare, fmt.Sprintf("%d ready tasks remain.", len(s.Queue)))
		s.Current = s.Queue[0]
		s.Queue = s.Queue[1:]
		b.IncOperations()
		add("L7", trace.ActionPop, fmt.Sprintf("Dequeue %q; all its prerequisites are resolved.", s.Current))
		s.Order = append(s.Order, s.Current)
		s.Processed[s.Current] = true
		b.IncOperations()
		add("L8", trace.ActionVisit, fmt.Sprintf("Append %q to output. Its outgoing dependencies are not yet resolved.", s.Current))
		for _, i := range adj[s.Current] {
			e := edges[i]
			s.Neighbor = e.To
			s.EdgeIndex = i
			s.Decremented = false
			s.Ready = false
			s.OldInDegree = s.InDegree[e.To]
			s.NewInDegree = s.OldInDegree
			b.IncComparisons()
			add("L9", trace.ActionHighlight, fmt.Sprintf("Inspect dependency %q to %q; %d prerequisites remain.", e.From, e.To, s.OldInDegree))
			s.InDegree[e.To]--
			s.NewInDegree = s.InDegree[e.To]
			s.Decremented = true
			s.ResolvedEdges = append(s.ResolvedEdges, i)
			b.IncOperations()
			add("L10", trace.ActionSet, fmt.Sprintf("%s: %d − 1 = %d remaining prerequisites.", e.To, s.OldInDegree, s.NewInDegree))
			s.Decremented = false
			b.IncComparisons()
			if s.NewInDegree == 0 {
				s.Queue = append(s.Queue, e.To)
				s.Ready = true
				b.IncOperations()
				add("L12", trace.ActionPush, fmt.Sprintf("%q reaches zero; append it to ready FIFO.", e.To))
			} else {
				add("L11", trace.ActionCompare, fmt.Sprintf("%q still waits for %d prerequisites.", e.To, s.NewInDegree))
			}
		}
	}
	s.Current = ""
	s.Neighbor = ""
	s.EdgeIndex = -1
	s.OldInDegree = -1
	s.NewInDegree = -1
	s.Ready = false
	s.Decremented = false
	b.IncComparisons()
	add("L6", trace.ActionCompare, "Ready FIFO is empty.")
	s.Terminal = true
	s.Valid = len(s.Order) == len(nodes)
	for _, n := range nodes {
		if !s.Processed[n.ID] {
			s.Blocked = append(s.Blocked, n.ID)
		}
	}
	if s.Valid {
		add("L16", trace.ActionDone, fmt.Sprintf("Valid topological order: %v.", s.Order))
	} else {
		add("L16", trace.ActionDone, fmt.Sprintf("Cycle prevents completion. Blocked tasks %v may include downstream tasks, not only cycle members.", s.Blocked))
	}
	steps := b.Steps()
	for i := range steps {
		if steps[i].LineID == "L9" {
			steps[i].HoldMS = 600
		}
		if steps[i].LineID == "L10" {
			steps[i].HoldMS = 1200
		}
	}
	return steps
}

func cloneSlice(s []string) []string { return append([]string(nil), s...) }
func cloneInDegrees(m map[string]int) map[string]int {
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
