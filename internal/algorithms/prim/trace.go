package prim

import (
	"container/heap"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Equal weights retain the native heap's deterministic insertion/heap ordering.
type primTraceHeap []renderers.PrimCandidate

func (h primTraceHeap) Len() int           { return len(h) }
func (h primTraceHeap) Less(i, j int) bool { return h[i].Weight < h[j].Weight }
func (h primTraceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *primTraceHeap) Push(v any)        { *h = append(*h, v.(renderers.PrimCandidate)) }
func (h *primTraceHeap) Pop() any          { old := *h; v := old[len(old)-1]; *h = old[:len(old)-1]; return v }

func Trace(nodes []renderers.GraphNode, edges []WeightedEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	members := map[string]bool{}
	valid := map[string]bool{}
	for _, n := range nodes {
		valid[n.ID] = true
	}
	if !valid[start] && len(nodes) > 0 {
		start = nodes[0].ID
	}
	pq := &primTraceHeap{}
	selected := []int{}
	current := ""
	active, total, candidate, serial := -1, 0, 0, 0
	activeSerial := -1
	discarded := false
	snapshot := func() renderers.PrimLessonState {
		s := renderers.PrimLessonState{Start: start, Current: current, InTree: map[string]bool{}, Candidates: append([]renderers.PrimCandidate{}, (*pq)...), SelectedEdges: append([]int{}, selected...), ActiveEdge: active, ActiveSerial: activeSerial, TotalWeight: total, CandidateWeight: candidate, Discarded: discarded, Connected: len(members) == len(nodes)}
		for k, v := range members {
			s.InTree[k] = v
		}
		s.Nodes = append([]renderers.GraphNode{}, nodes...)
		for i := range s.Nodes {
			s.Nodes[i].State = "unvisited"
			if members[s.Nodes[i].ID] {
				s.Nodes[i].State = "visited"
			}
			if s.Connected {
				s.Nodes[i].State = "path"
			}
		}
		for i, e := range edges {
			chosen := false
			for _, id := range selected {
				if id == i {
					chosen = true
				}
			}
			s.Edges = append(s.Edges, renderers.GraphEdge{From: e.From, To: e.To, Weight: e.Weight, Active: chosen})
			if members[e.From] != members[e.To] {
				s.CutEdges = append(s.CutEdges, i)
			}
		}
		return s
	}
	add := func(line string, action trace.ActionType, explanation string) {
		b.AddStep(line, action, explanation, snapshot(), map[string]any{"totalWeight": total, "mstCount": len(selected), "pqSize": pq.Len(), "activeEdge": active})
	}
	if len(nodes) == 0 {
		add("L1", trace.ActionDone, "Empty graph; tree is empty.")
		return b.Steps()
	}
	b.IncOperations()
	add("L1", trace.ActionInit, fmt.Sprintf("Start Prim at %s; grow only its connected component.", start))
	b.IncOperations()
	add("L2", trace.ActionSet, "Initialize empty tree and zero total.")
	members[start] = true
	current = start
	b.IncOperations()
	add("L3", trace.ActionVisit, "First vertex joins for free; no edge selected.")
	if len(nodes) == 1 {
		add("L18", trace.ActionDone, "Single vertex: complete MST, 0 edges, total 0.")
		return b.Steps()
	}
	heap.Init(pq)
	b.IncOperations()
	add("L4", trace.ActionInit, "Initialize candidate min-heap; stale internal edges may remain later.")
	adj := map[string][]renderers.PrimCandidate{}
	for i, e := range edges {
		if valid[e.From] && valid[e.To] {
			adj[e.From] = append(adj[e.From], renderers.PrimCandidate{From: e.From, To: e.To, Weight: e.Weight, EdgeIndex: i})
			adj[e.To] = append(adj[e.To], renderers.PrimCandidate{From: e.To, To: e.From, Weight: e.Weight, EdgeIndex: i})
		}
	}
	push := func(c renderers.PrimCandidate, line string) {
		c.Serial = serial
		serial++
		heap.Push(pq, c)
		b.IncOperations()
		add(line, trace.ActionPush, fmt.Sprintf("Push %s–%s, weight %d, original edge #%d.", c.From, c.To, c.Weight, c.EdgeIndex))
	}
	for _, e := range adj[start] {
		push(e, "L5")
	}
	for pq.Len() > 0 && len(selected) < len(nodes)-1 {
		active = -1
		candidate = 0
		discarded = false
		b.IncComparisons()
		add("L6", trace.ActionCompare, "Check heap and selected edge count.")
		e := heap.Pop(pq).(renderers.PrimCandidate)
		active = e.EdgeIndex
		activeSerial = e.Serial
		candidate = e.Weight
		current = e.To
		b.IncOperations()
		add("L7", trace.ActionPop, fmt.Sprintf("Probe %s–%s (%d). Not accepted yet.", e.From, e.To, e.Weight))
		b.IncComparisons()
		add("L8", trace.ActionCompare, fmt.Sprintf("Is %s already in tree? %t.", e.To, members[e.To]))
		if members[e.To] {
			discarded = true
			add("L9", trace.ActionHighlight, "Discard internal edge: both endpoints already in tree.")
			continue
		}
		members[e.To] = true
		b.IncOperations()
		add("L11", trace.ActionVisit, fmt.Sprintf("Add %s to tree membership; edge and sum unchanged.", e.To))
		selected = append(selected, e.EdgeIndex)
		b.IncOperations()
		add("L12", trace.ActionSet, fmt.Sprintf("Accept original edge #%d, %s–%s.", e.EdgeIndex, e.From, e.To))
		old := total
		total += e.Weight
		b.IncOperations()
		add("L13", trace.ActionSet, fmt.Sprintf("Update total: %d + (%d) = %d.", old, e.Weight, total))
		add("L14", trace.ActionVisit, fmt.Sprintf("Inspect incident edges of %s.", e.To))
		for _, next := range adj[e.To] {
			b.IncComparisons()
			if !members[next.To] {
				push(next, "L15")
			}
		}
	}
	active = -1
	candidate = 0
	discarded = false
	result := fmt.Sprintf("Complete MST: %d edges, total %d; spanning all vertices.", len(selected), total)
	if len(members) != len(nodes) {
		result = fmt.Sprintf("Partial start-component tree: %d of %d vertices, %d edges, total %d; graph disconnected.", len(members), len(nodes), len(selected), total)
	}
	add("L18", trace.ActionDone, result)
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L8" && (steps[i+1].LineID == "L9" || steps[i+1].LineID == "L11") {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
