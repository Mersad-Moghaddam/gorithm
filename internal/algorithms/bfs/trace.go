package bfs

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records forward-edge BFS; weights and Directed do not change traversal.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	if len(nodes) == 0 {
		return nil
	}
	if start == "" {
		start = nodes[0].ID
	}
	s := renderers.BFSLessonState{GraphState: renderers.GraphState{Nodes: slices.Clone(nodes), Edges: slices.Clone(edges)}, Start: start, EdgeIndex: -1, Distance: map[string]int{}, Parent: map[string]string{}, Discovered: map[string]bool{}, Finished: map[string]bool{}}
	adj := map[string][]int{}
	for i, e := range edges {
		adj[e.From] = append(adj[e.From], i)
	}
	b := trace.NewBuilder()
	add := func(line string, action trace.ActionType, explanation string) {
		c := s
		c.Nodes = slices.Clone(s.Nodes)
		c.Edges = slices.Clone(s.Edges)
		for i := range c.Nodes {
			n := &c.Nodes[i]
			n.State = "unvisited"
			if s.Discovered[n.ID] {
				n.State = "queued"
			}
			if n.ID == s.Current {
				n.State = "visiting"
			}
			if s.Finished[n.ID] {
				n.State = "visited"
			}
		}
		for i := range c.Edges {
			c.Edges[i].Active = i == s.EdgeIndex
		}
		c.Queue = slices.Clone(s.Queue)
		c.Order = slices.Clone(s.Order)
		c.Distance = maps.Clone(s.Distance)
		c.Parent = maps.Clone(s.Parent)
		c.Discovered = maps.Clone(s.Discovered)
		c.Finished = maps.Clone(s.Finished)
		b.AddStep(line, action, explanation, c, map[string]any{"start": s.Start, "curr": s.Current, "neighbor": s.Neighbor, "queue": slices.Clone(s.Queue), "order": slices.Clone(s.Order)})
	}
	add("L1", trace.ActionInit, "BFS uses a FIFO queue. Levels count edges, not weighted cost.")
	s.Queue = []string{start}
	s.Discovered[start] = true
	s.Distance[start] = 0
	s.Enqueued = start
	b.IncOperations()
	add("L3", trace.ActionPush, fmt.Sprintf("Discover %s at level 0 and enqueue it.", start))
	s.Enqueued = ""
	add("L5", trace.ActionSet, "Initialize empty visit order. Discovered means queued, not finished.")
	for len(s.Queue) > 0 {
		b.IncComparisons()
		add("L6", trace.ActionCompare, "FIFO is not empty; take its front vertex.")
		s.Current = s.Queue[0]
		s.Queue = s.Queue[1:]
		s.Dequeued = s.Current
		b.IncOperations()
		add("L7", trace.ActionPop, fmt.Sprintf("Dequeue %s; remaining queue %v.", s.Current, s.Queue))
		s.Dequeued = ""
		s.Order = append(s.Order, s.Current)
		b.IncOperations()
		add("L9", trace.ActionVisit, fmt.Sprintf("Append %s to visit order.", s.Current))
		for _, ei := range adj[s.Current] {
			s.EdgeIndex = ei
			s.Neighbor = edges[ei].To
			s.Skip = false
			b.IncComparisons()
			add("L10", trace.ActionHighlight, fmt.Sprintf("Inspect directed edge %s → %s (input edge %d).", s.Current, s.Neighbor, ei))
			b.IncComparisons()
			s.Skip = s.Discovered[s.Neighbor]
			if s.Skip {
				add("L11", trace.ActionCompare, "Skip: neighbor already queued/visited; never enqueue twice.")
				continue
			}
			add("L11", trace.ActionCompare, fmt.Sprintf("%s is undiscovered; it is not in the queue yet.", s.Neighbor))
			s.Discovered[s.Neighbor] = true
			s.Distance[s.Neighbor] = s.Distance[s.Current] + 1
			s.Parent[s.Neighbor] = s.Current
			s.Queue = append(s.Queue, s.Neighbor)
			s.Enqueued = s.Neighbor
			b.IncOperations()
			add("L12", trace.ActionPush, fmt.Sprintf("Discover %s: d[%s] + 1 = %d. Enqueue once.", s.Neighbor, s.Current, s.Distance[s.Neighbor]))
			s.Enqueued = ""
		}
		s.EdgeIndex = -1
		s.Neighbor = ""
		s.Skip = false
		s.Finished[s.Current] = true
		b.IncOperations()
		add("L15", trace.ActionSet, fmt.Sprintf("✓ %s: all outgoing edges inspected. Level stays fixed.", s.Current))
		s.Current = ""
	}
	b.IncComparisons()
	add("L6", trace.ActionCompare, "Queue empty; traversal ends.")
	add("L17", trace.ActionDone, fmt.Sprintf("BFS complete: %v. Unknown levels are unreachable from %s.", s.Order, start))
	steps := b.Steps()
	for i := range steps {
		if steps[i].LineID == "L11" && i+1 < len(steps) && steps[i+1].LineID == "L12" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
