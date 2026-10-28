package bellmanford

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"math"
)

func cloneDist(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Trace records in-place relaxation; cycle detection never changes distances.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.BellmanFordLessonState{GraphState: renderers.GraphState{Nodes: append([]renderers.GraphNode(nil), nodes...), Edges: append([]renderers.GraphEdge(nil), edges...)}, Start: start, Distance: map[string]int{}, EdgeIndex: -1, WitnessEdge: -1, MaxPasses: len(nodes) - 1}
	for _, n := range nodes {
		s.Distance[n.ID] = math.MaxInt
	}
	if s.MaxPasses < 0 {
		s.MaxPasses = 0
	}
	s.PassStartDistance = cloneDist(s.Distance)
	add := func(line string, action trace.ActionType, explanation string, hold int) {
		snap := s
		snap.Nodes = append([]renderers.GraphNode(nil), s.Nodes...)
		snap.Edges = append([]renderers.GraphEdge(nil), s.Edges...)
		snap.Distance = cloneDist(s.Distance)
		snap.PassStartDistance = cloneDist(s.PassStartDistance)
		b.AddStep(line, action, explanation, snap, map[string]any{"start": start, "pass": s.Pass, "from": s.From, "to": s.To, "distances": cloneDist(s.Distance), "relaxed": s.ChangedThisPass, "negativeCycle": s.NegativeCycle})
		steps := b.Steps()
		steps[len(steps)-1].HoldMS = hold
	}
	if _, ok := s.Distance[start]; !ok {
		s.InvalidStart = true
		add("L1", trace.ActionInit, "Start node not found.", 0)
		add("L18", trace.ActionDone, "Invalid start: no shortest-path result.", 0)
		return b.Steps()
	}
	add("L1", trace.ActionInit, "Initialize all distances to infinity.", 0)
	s.Distance[start] = 0
	b.IncOperations()
	add("L2", trace.ActionSet, "Source distance is zero; other vertices remain unreachable.", 0)
	focus := func(i int) {
		s.EdgeIndex = i
		s.From = edges[i].From
		s.To = edges[i].To
		s.Weight = edges[i].Weight
		s.OldDistance = s.Distance[s.To]
		s.HasCandidate = false
		s.Improved = false
		for j := range s.Edges {
			s.Edges[j].Active = j == i
			s.Edges[j].Directed = true
		}
		u, okU := s.Distance[s.From]
		_, okV := s.Distance[s.To]
		w := s.Weight
		if okU && okV && u != math.MaxInt && !(w > 0 && u > math.MaxInt-w) && !(w < 0 && u < math.MinInt-w) {
			s.Candidate = u + w
			s.HasCandidate = true
			s.Improved = s.Candidate < s.OldDistance
		}
	}
	for pass := 1; pass <= s.MaxPasses; pass++ {
		s.Pass = pass
		s.PassStartDistance = cloneDist(s.Distance)
		s.EdgeIndex = -1
		s.From = ""
		s.To = ""
		s.HasCandidate = false
		s.Improved = false
		add("L3", trace.ActionHighlight, fmt.Sprintf("Begin pass %d of %d using current distances.", pass, s.MaxPasses), 0)
		s.ChangedThisPass = false
		add("L4", trace.ActionSet, "Reset changed flag for this pass.", 350)
		for i := range edges {
			focus(i)
			add("L5", trace.ActionVisit, fmt.Sprintf("Inspect edge %d: %s → %s, weight %d.", i+1, s.From, s.To, s.Weight), 350)
			b.IncComparisons()
			hold := 0
			if s.Improved {
				hold = 600
			}
			explanation := "Source unreachable: skip arithmetic."
			if s.HasCandidate {
				explanation = fmt.Sprintf("Candidate %d; previous target %s. Improvement: %t.", s.Candidate, renderers.BellmanFordDistanceLabel(s.OldDistance), s.Improved)
			}
			add("L6", trace.ActionCompare, explanation, hold)
			if s.Improved {
				s.Distance[s.To] = s.Candidate
				b.IncOperations()
				add("L7", trace.ActionSet, fmt.Sprintf("Updated %s to %d; not yet final.", s.To, s.Candidate), 1200)
				s.ChangedThisPass = true
				add("L8", trace.ActionSet, "At least one distance changed this pass.", 350)
			}
		}
		for i := range s.Edges {
			s.Edges[i].Active = false
		}
		s.EdgeIndex = -1
		s.From = ""
		s.To = ""
		s.HasCandidate = false
		s.Improved = false
		add("L11", trace.ActionCompare, fmt.Sprintf("Pass %d complete; changed: %t. Stop early if unchanged.", pass, s.ChangedThisPass), 0)
		if !s.ChangedThisPass {
			break
		}
	}
	s.CheckingCycle = true
	add("L13", trace.ActionHighlight, "Can any reachable edge still improve?", 0)
	for i := range edges {
		focus(i)
		b.IncComparisons()
		add("L14", trace.ActionCompare, "Test reachable source and candidate against stored target distance.", 0)
		if s.Improved {
			s.NegativeCycle = true
			s.WitnessEdge = i
			add("L15", trace.ActionDone, "Reachable negative cycle: this edge can still improve. No finite optimum.", 0)
			return b.Steps()
		}
	}
	s.EdgeIndex = -1
	s.From = ""
	s.To = ""
	s.HasCandidate = false
	s.Improved = false
	for i := range s.Edges {
		s.Edges[i].Active = false
	}
	add("L18", trace.ActionDone, "No reachable negative cycle. Finite distances are shortest; infinity remains unreachable.", 0)
	return b.Steps()
}
