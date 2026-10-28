package dijkstra

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Same distance-only comparison as the native heap; serial records identity, not priority.
type dijkstraTraceHeap []renderers.DijkstraQueueEntry

func (h dijkstraTraceHeap) Len() int           { return len(h) }
func (h dijkstraTraceHeap) Less(i, j int) bool { return h[i].Distance < h[j].Distance }
func (h dijkstraTraceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *dijkstraTraceHeap) Push(x any)        { *h = append(*h, x.(renderers.DijkstraQueueEntry)) }
func (h *dijkstraTraceHeap) Pop() any          { n := len(*h) - 1; x := (*h)[n]; *h = (*h)[:n]; return x }

func findNodeIndex(nodes []renderers.GraphNode, id string) int {
	for i, n := range nodes {
		if n.ID == id {
			return i
		}
	}
	return -1
}

// Trace follows edge direction and stops when the target is settled.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start, target string) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.DijkstraLessonState{GraphState: renderers.GraphState{Nodes: append([]renderers.GraphNode(nil), nodes...), Edges: append([]renderers.GraphEdge(nil), edges...)}, Start: start, Target: target, EdgeIndex: -1, PoppedSerial: -1, Distance: map[string]int{}, Previous: map[string]string{}, Settled: map[string]bool{}}
	for _, n := range nodes {
		s.Distance[n.ID] = math.MaxInt
	}
	pq := &dijkstraTraceHeap{}
	prevEdge := map[string]int{}
	serial := 0
	add := func(line string, action trace.ActionType, text string) {
		c := s
		c.Nodes = append([]renderers.GraphNode(nil), s.Nodes...)
		c.Edges = append([]renderers.GraphEdge(nil), s.Edges...)
		c.Distance = map[string]int{}
		for k, v := range s.Distance {
			c.Distance[k] = v
		}
		c.Previous = map[string]string{}
		for k, v := range s.Previous {
			c.Previous[k] = v
		}
		c.Settled = map[string]bool{}
		for k, v := range s.Settled {
			c.Settled[k] = v
		}
		c.Queue = append([]renderers.DijkstraQueueEntry(nil), (*pq)...)
		c.Path = append([]string(nil), s.Path...)
		vars := map[string]any{"distances": c.Distance, "current": s.Current, "target": target}
		// Variables must not alias typed state, either.
		d := map[string]int{}
		for k, v := range c.Distance {
			d[k] = v
		}
		vars["distances"] = d
		if s.HasCandidate {
			vars["alt"] = s.Candidate
			vars["neighbor"] = s.Neighbor
		}
		if s.Found {
			vars["path"] = append([]string(nil), s.Path...)
			vars["totalDistance"] = s.Distance[target]
		}
		b.AddStep(line, action, text, c, vars)
	}
	if _, ok := s.Distance[start]; !ok {
		s.InvalidStart = true
		add("L1", trace.ActionInit, fmt.Sprintf("Start %q is absent.", start))
		add("L17", trace.ActionDone, "Invalid start; no path.")
		return b.Steps()
	}
	add("L1", trace.ActionInit, "Initialize distances and predecessor ledger.")
	s.Distance[start] = 0
	add("L2", trace.ActionSet, fmt.Sprintf("Distance to %s = 0; others = ∞.", start))
	heap.Push(pq, renderers.DijkstraQueueEntry{Node: start, Distance: 0, Serial: serial})
	serial++
	s.Nodes[findNodeIndex(s.Nodes, start)].State = "queued"
	b.IncOperations()
	add("L4", trace.ActionPush, "Push start into the minimum heap.")
	for pq.Len() > 0 {
		s.Current = ""
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.HasCandidate = false
		s.Improved = false
		s.Stale = false
		s.PoppedSerial = -1
		add("L5", trace.ActionHighlight, "Heap is nonempty; extract its root.")
		item := heap.Pop(pq).(renderers.DijkstraQueueEntry)
		u := item.Node
		s.Current = u
		s.PoppedDistance = item.Distance
		s.PoppedSerial = item.Serial
		s.Stale = s.Settled[u]
		b.IncOperations()
		if s.Stale {
			add("L6", trace.ActionPop, fmt.Sprintf("Skip %s:%d: already settled.", u, item.Distance))
			continue
		}
		s.Settled[u] = true
		s.Nodes[findNodeIndex(s.Nodes, u)].State = "visiting"
		add("L6", trace.ActionPop, fmt.Sprintf("Extract %s:%d; minimum distance is now final.", u, item.Distance))
		b.IncComparisons()
		add("L7", trace.ActionCompare, fmt.Sprintf("Check %s against target %s.", u, target))
		if target != "" && u == target {
			add("L7", trace.ActionHighlight, "Target settled; stop before exploring its neighbors.")
			break
		}
		for i, e := range s.Edges {
			v := ""
			if e.From == u {
				v = e.To
			} else if !e.Directed && e.To == u {
				v = e.From
			} else {
				continue
			}
			old, exists := s.Distance[v]
			if !exists || s.Settled[v] {
				continue
			}
			s.Neighbor = v
			s.EdgeIndex = i
			s.HasCandidate = false
			s.Improved = false
			s.Edges[i].Active = true
			add("L8", trace.ActionVisit, fmt.Sprintf("Inspect %s to %s, weight %d.", u, v, e.Weight))
			s.OldDistance = old
			s.Candidate = s.Distance[u] + e.Weight
			s.HasCandidate = true
			add("L9", trace.ActionSet, fmt.Sprintf("Candidate = %d + %d = %d.", s.Distance[u], e.Weight, s.Candidate))
			b.IncComparisons()
			s.Improved = s.Candidate < old
			add("L10", trace.ActionCompare, fmt.Sprintf("Compare candidate %d with old distance %s.", s.Candidate, dijkstraDistance(old)))
			if s.Improved {
				s.Distance[v] = s.Candidate
				add("L11", trace.ActionSet, fmt.Sprintf("Commit distance %s = %d.", v, s.Candidate))
				s.Previous[v] = u
				prevEdge[v] = i
				add("L12", trace.ActionSet, fmt.Sprintf("Set predecessor of %s to %s.", v, u))
				heap.Push(pq, renderers.DijkstraQueueEntry{Node: v, Distance: s.Candidate, Serial: serial})
				serial++
				s.Nodes[findNodeIndex(s.Nodes, v)].State = "queued"
				b.IncOperations()
				add("L13", trace.ActionPush, fmt.Sprintf("Push %s:%d; keep older heap entries.", v, s.Candidate))
			}
			s.Edges[i].Active = false
		}
		s.Nodes[findNodeIndex(s.Nodes, u)].State = "visited"
		s.Neighbor = ""
		s.EdgeIndex = -1
		s.HasCandidate = false
		add("L15", trace.ActionHighlight, fmt.Sprintf("Finished neighbors of %s.", u))
	}
	s.Neighbor = ""
	s.EdgeIndex = -1
	s.HasCandidate = false
	s.Stale = false
	if d, ok := s.Distance[target]; target != "" && ok && d != math.MaxInt {
		for v := target; v != ""; v = s.Previous[v] {
			s.Path = append([]string{v}, s.Path...)
			if v == start {
				break
			}
		}
		s.Found = len(s.Path) > 0 && s.Path[0] == start
	}
	if !s.Found {
		s.Path = nil
	}
	for i := range s.Edges {
		s.Edges[i].Active = false
	}
	if s.Found {
		for _, v := range s.Path {
			s.Nodes[findNodeIndex(s.Nodes, v)].State = "path"
			if i, ok := prevEdge[v]; ok {
				s.Edges[i].Active = true
			}
		}
		add("L17", trace.ActionDone, fmt.Sprintf("Shortest path %v; cost %d.", s.Path, s.Distance[target]))
	} else if target != "" {
		add("L17", trace.ActionDone, "Target absent or unreachable; no path (∞).")
	} else {
		add("L17", trace.ActionDone, "All reachable distances settled.")
	}
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L10" && steps[i+1].LineID == "L11" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}

func dijkstraDistance(d int) string {
	if d == math.MaxInt {
		return "∞"
	}
	return fmt.Sprint(d)
}
