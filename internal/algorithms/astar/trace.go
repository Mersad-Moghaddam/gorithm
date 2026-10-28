package astar

import (
	"container/heap"
	"fmt"
	"maps"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records the same closed-set, f-then-h heap search as FindPath.
func Trace(nodes []AStarNode, edges []AStarEdge, start, goal string) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}
	s := renderers.AStarLessonState{Start: start, Goal: goal, EdgeIndex: -1, Cost: -1, G: map[string]int{}, H: map[string]int{}, F: map[string]int{}, CameFrom: map[string]string{}, Closed: map[string]bool{}, HeuristicConsistent: true}
	nodeMap := map[string]AStarNode{}
	for _, n := range nodes {
		nodeMap[n.ID] = n
		label := n.Label
		if label == "" {
			label = n.ID
		}
		s.Nodes = append(s.Nodes, renderers.GraphNode{ID: n.ID, Label: label, X: n.X, Y: n.Y, State: "unvisited"})
	}
	for _, e := range edges {
		s.Edges = append(s.Edges, renderers.GraphEdge{From: e.From, To: e.To, Weight: e.Weight, Directed: !e.Undirected})
		s.EdgeUndirected = append(s.EdgeUndirected, e.Undirected)
	}
	_, hasStart := nodeMap[start]
	goalNode, hasGoal := nodeMap[goal]
	s.InvalidEndpoints = !hasStart || !hasGoal
	if hasGoal {
		for _, n := range nodes {
			s.H[n.ID] = EuclideanDistance(n, goalNode)
		}
	}
	for _, e := range edges {
		_, a := nodeMap[e.From]
		_, z := nodeMap[e.To]
		if !a || !z || e.Weight < 0 || s.H[e.From] > e.Weight+s.H[e.To] || (e.Undirected && s.H[e.To] > e.Weight+s.H[e.From]) {
			s.HeuristicConsistent = false
		}
	}
	pq := &priorityQueue{}
	heap.Init(pq)
	items := map[string]*pqItem{}
	record := func(line string, action trace.ActionType, explanation string) {
		out := s
		out.Nodes = slices.Clone(s.Nodes)
		out.Edges = slices.Clone(s.Edges)
		out.EdgeUndirected = slices.Clone(s.EdgeUndirected)
		out.G = maps.Clone(s.G)
		out.H = maps.Clone(s.H)
		out.F = maps.Clone(s.F)
		out.CameFrom = maps.Clone(s.CameFrom)
		out.Closed = maps.Clone(s.Closed)
		out.Path = slices.Clone(s.Path)
		out.Open = make([]renderers.AStarOpenEntry, pq.Len())
		for i, it := range *pq {
			out.Open[i] = renderers.AStarOpenEntry{Node: it.node, G: s.G[it.node], H: it.hScore, F: it.fScore, HeapIndex: i}
		}
		vars := map[string]any{"start": start, "goal": goal, "current": s.Current, "neighbor": s.Neighbor, "found": s.Found, "cost": s.Cost, "openSetCount": pq.Len()}
		if s.Current != "" {
			vars["g"] = s.G[s.Current]
			vars["h"] = s.H[s.Current]
			vars["f"] = s.F[s.Current]
		}
		if s.Found {
			vars["path"] = slices.Clone(s.Path)
		}
		b.AddStep(line, action, explanation, out, vars)
	}
	record("L1", trace.ActionInit, "Initialize A*: g is known cost; h is rounded Euclidean estimate using input coordinates.")
	if s.InvalidEndpoints {
		record("L19", trace.ActionDone, "Invalid start or goal; no path.")
		return b.Steps()
	}
	record("L2", trace.ActionSet, "Create open heap, scores, parents, and closed set.")
	s.G[start] = 0
	s.F[start] = s.H[start]
	it := &pqItem{node: start, fScore: s.H[start], hScore: s.H[start]}
	heap.Push(pq, it)
	items[start] = it
	setNodeState(s.Nodes, start, "queued")
	record("L3", trace.ActionPush, fmt.Sprintf("Push %s: f = 0 + %d.", start, s.H[start]))
	for pq.Len() > 0 {
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.HasTentative = false
		s.Improved = false
		record("L4", trace.ActionHighlight, fmt.Sprintf("Open heap has %d candidates; minimum f wins, then lower h.", pq.Len()))
		curr := heap.Pop(pq).(*pqItem)
		delete(items, curr.node)
		s.Current = curr.node
		setNodeState(s.Nodes, s.Current, "visiting")
		b.IncOperations()
		record("L5", trace.ActionPop, fmt.Sprintf("Pop %s: g=%d, h=%d, f=%d.", s.Current, s.G[s.Current], curr.hScore, curr.fScore))
		b.IncComparisons()
		record("L6", trace.ActionCompare, fmt.Sprintf("Goal check: %s == %s is %t.", s.Current, goal, s.Current == goal))
		if s.Current == goal {
			s.Found = true
			s.Cost = s.G[goal]
			s.Path = []string{goal}
			for p, ok := s.CameFrom[goal]; ok; p, ok = s.CameFrom[p] {
				s.Path = append([]string{p}, s.Path...)
			}
			for _, id := range s.Path {
				setNodeState(s.Nodes, id, "path")
			}
			for i, e := range edges {
				for j := 1; j < len(s.Path); j++ {
					if (e.From == s.Path[j-1] && e.To == s.Path[j]) || (e.Undirected && e.To == s.Path[j-1] && e.From == s.Path[j]) {
						s.Edges[i].Active = true
					}
				}
			}
			record("L7", trace.ActionDone, fmt.Sprintf("Path found: %v; cost %d. Optimality requires consistent heuristic for this closed-set search.", s.Path, s.Cost))
			return astarHolds(b.Steps())
		}
		s.Closed[s.Current] = true
		setNodeState(s.Nodes, s.Current, "visited")
		for i, e := range edges {
			neighbor := ""
			if e.From == s.Current {
				neighbor = e.To
			} else if e.Undirected && e.To == s.Current {
				neighbor = e.From
			}
			if neighbor == "" {
				continue
			}
			s.Neighbor = neighbor
			s.EdgeIndex = i
			s.HasTentative = false
			s.Improved = false
			s.OldG, s.KnownOldG = s.G[neighbor]
			s.TentativeG = 0
			s.Edges[i].Active = true
			if s.Closed[neighbor] {
				record("L9", trace.ActionHighlight, fmt.Sprintf("%s is closed; skip before computing tentative g.", neighbor))
				s.Edges[i].Active = false
				continue
			}
			record("L9", trace.ActionHighlight, fmt.Sprintf("Examine edge %s to %s, weight %d; neighbor is not closed.", s.Current, neighbor, e.Weight))
			s.TentativeG = s.G[s.Current] + e.Weight
			s.HasTentative = true
			record("L10", trace.ActionSet, fmt.Sprintf("Tentative g = %d + %d = %d.", s.G[s.Current], e.Weight, s.TentativeG))
			b.IncComparisons()
			s.Improved = !s.KnownOldG || s.TentativeG < s.OldG
			old := "∞"
			if s.KnownOldG {
				old = fmt.Sprint(s.OldG)
			}
			record("L11", trace.ActionCompare, fmt.Sprintf("Compare tentative %d with known %s: improved = %t.", s.TentativeG, old, s.Improved))
			if s.Improved {
				s.CameFrom[neighbor] = s.Current
				record("L12", trace.ActionSet, fmt.Sprintf("Parent of %s is now %s.", neighbor, s.Current))
				s.G[neighbor] = s.TentativeG
				record("L13", trace.ActionSet, fmt.Sprintf("Commit g(%s) = %d.", neighbor, s.TentativeG))
				f := s.TentativeG + s.H[neighbor]
				s.F[neighbor] = f
				record("L14", trace.ActionSet, fmt.Sprintf("f(%s) = g %d + h %d = %d.", neighbor, s.TentativeG, s.H[neighbor], f))
				if it, ok := items[neighbor]; ok {
					it.fScore = f
					it.hScore = s.H[neighbor]
					heap.Fix(pq, it.index)
				} else {
					it := &pqItem{node: neighbor, fScore: f, hScore: s.H[neighbor]}
					heap.Push(pq, it)
					items[neighbor] = it
				}
				setNodeState(s.Nodes, neighbor, "queued")
				record("L15", trace.ActionPush, fmt.Sprintf("Push/update %s once in actual heap; f=%d.", neighbor, f))
			}
			s.Edges[i].Active = false
		}
	}
	s.Neighbor = ""
	s.EdgeIndex = -1
	s.HasTentative = false
	record("L19", trace.ActionDone, "Open heap exhausted; no path. Unknown scores remain ∞.")
	return astarHolds(b.Steps())
}

func astarHolds(steps []trace.Step) []trace.Step {
	for i := range steps {
		if i+1 < len(steps) && steps[i].LineID == "L11" && steps[i+1].LineID == "L12" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}

func setNodeState(nodes []renderers.GraphNode, id, state string) {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].State = state
			return
		}
	}
}
