package astar

import (
	"container/heap"
	"fmt"
	"sort"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for A* pathfinding on graph.
// ponytail: Euclidean heuristic with adjacency-map; grid-specialized jump-point search faster for tilemaps.
func Trace(nodes []AStarNode, edges []AStarEdge, start, goal string) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}

	nodeMap := make(map[string]AStarNode, len(nodes))
	graphNodes := make([]renderers.GraphNode, len(nodes))
	for i, n := range nodes {
		nodeMap[n.ID] = n
		label := n.Label
		if label == "" {
			label = n.ID
		}
		graphNodes[i] = renderers.GraphNode{
			ID:    n.ID,
			Label: label,
			X:     n.X,
			Y:     n.Y,
			State: "unvisited",
		}
	}

	graphEdges := make([]renderers.GraphEdge, len(edges))
	for i, e := range edges {
		graphEdges[i] = renderers.GraphEdge{
			From:     e.From,
			To:       e.To,
			Weight:   e.Weight,
			Active:   false,
			Directed: !e.Undirected,
		}
	}

	startNode, hasStart := nodeMap[start]
	goalNode, hasGoal := nodeMap[goal]
	if !hasStart || !hasGoal {
		caption1 := fmt.Sprintf("Invalid search parameters: start %q or goal %q not in graph.", start, goal)
		b.AddStep("L1", trace.ActionInit,
			caption1,
			cloneGraphState(caption1, graphNodes, graphEdges, nil),
			map[string]any{"start": start, "goal": goal},
		)
		caption2 := "Cannot proceed with search. Returning nil."
		b.AddStep("L19", trace.ActionDone,
			caption2,
			cloneGraphState(caption2, graphNodes, graphEdges, nil),
			map[string]any{"start": start, "goal": goal, "found": false, "cost": -1},
		)
		return b.Steps()
	}

	type neighbor struct {
		to     string
		weight int
	}
	adj := make(map[string][]neighbor)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], neighbor{to: e.To, weight: e.Weight})
		if e.Undirected {
			adj[e.To] = append(adj[e.To], neighbor{to: e.From, weight: e.Weight})
		}
	}

	captionInit := fmt.Sprintf("Initialize A* search from %s to %s with heuristic h(n) = Euclidean distance", start, goal)
	b.AddStep("L1", trace.ActionInit,
		captionInit,
		cloneGraphState(captionInit, graphNodes, graphEdges, nil),
		map[string]any{"start": start, "goal": goal, "vertices": len(nodes), "edges": len(edges)},
	)

	captionPQ := "Initialize open set min-heap priority queue, gScore, and fScore maps"
	b.AddStep("L2", trace.ActionSet,
		captionPQ,
		cloneGraphState(captionPQ, graphNodes, graphEdges, nil),
		map[string]any{"openSetCount": 0, "start": start, "goal": goal},
	)

	gScore := make(map[string]int)
	fScore := make(map[string]int)
	cameFrom := make(map[string]string)
	visited := make(map[string]bool)

	pq := &priorityQueue{}
	heap.Init(pq)
	itemMap := make(map[string]*pqItem)

	hStart := EuclideanDistance(startNode, goalNode)
	gScore[start] = 0
	fScore[start] = hStart

	startItem := &pqItem{node: start, fScore: hStart, hScore: hStart}
	heap.Push(pq, startItem)
	itemMap[start] = startItem
	setNodeState(graphNodes, start, "queued")

	wl := buildWorklist(*pq, gScore)
	captionPush := fmt.Sprintf("Push start node %s into open set with f(%s) = g(%d) + h(%d) = %d", start, start, 0, hStart, hStart)
	b.AddStep("L3", trace.ActionPush,
		captionPush,
		cloneGraphState(captionPush, graphNodes, graphEdges, wl),
		map[string]any{"node": start, "g": 0, "h": hStart, "f": hStart, "openSetCount": pq.Len()},
	)

	for pq.Len() > 0 {
		wl = buildWorklist(*pq, gScore)
		captionLoop := fmt.Sprintf("Open set has %d candidate node(s); evaluating next with lowest f(n)", pq.Len())
		b.AddStep("L4", trace.ActionHighlight,
			captionLoop,
			cloneGraphState(captionLoop, graphNodes, graphEdges, wl),
			map[string]any{"openSetCount": pq.Len()},
		)

		curr := heap.Pop(pq).(*pqItem)
		delete(itemMap, curr.node)
		currID := curr.node
		setNodeState(graphNodes, currID, "visiting")
		b.IncOperations()

		wl = buildWorklist(*pq, gScore)
		captionPop := fmt.Sprintf("Pop node %s with lowest total estimated cost f(n)=g+h (%d = %d + %d)", currID, curr.fScore, gScore[currID], curr.hScore)
		b.AddStep("L5", trace.ActionPop,
			captionPop,
			cloneGraphState(captionPop, graphNodes, graphEdges, wl),
			map[string]any{"current": currID, "g": gScore[currID], "h": curr.hScore, "f": curr.fScore, "openSetCount": pq.Len()},
		)

		b.IncComparisons()
		if currID == goal {
			captionGoal := fmt.Sprintf("Goal node %s reached with lowest estimated cost f=%d; optimal path guaranteed", currID, curr.fScore)
			b.AddStep("L6", trace.ActionCompare,
				captionGoal,
				cloneGraphState(captionGoal, graphNodes, graphEdges, wl),
				map[string]any{"current": currID, "goal": goal, "reached": true, "cost": gScore[goal]},
			)

			// Reconstruct path
			path := []string{currID}
			for p, ok := cameFrom[currID]; ok; p, ok = cameFrom[p] {
				path = append([]string{p}, path...)
			}

			// Clear all active edges first
			for i := range graphEdges {
				graphEdges[i].Active = false
			}

			// Highlight path nodes and path edges
			for _, nodeID := range path {
				setNodeState(graphNodes, nodeID, "path")
			}
			for i := 0; i < len(path)-1; i++ {
				setEdgeActive(graphEdges, path[i], path[i+1], true)
			}

			captionPath := fmt.Sprintf("Reconstructed optimal path in purple: %v with total cost %d", path, gScore[goal])
			b.AddStep("L7", trace.ActionDone,
				captionPath,
				cloneGraphState(captionPath, graphNodes, graphEdges, wl),
				map[string]any{
					"path":  path,
					"cost":  gScore[goal],
					"g":     gScore[goal],
					"h":     0,
					"f":     gScore[goal],
					"found": true,
				},
			)
			return b.Steps()
		}

		captionNotGoal := fmt.Sprintf("Node %s is not goal %s; exploring outgoing neighbors", currID, goal)
		b.AddStep("L6", trace.ActionCompare,
			captionNotGoal,
			cloneGraphState(captionNotGoal, graphNodes, graphEdges, wl),
			map[string]any{"current": currID, "goal": goal, "reached": false, "g": gScore[currID], "h": curr.hScore, "f": curr.fScore},
		)

		visited[currID] = true
		setNodeState(graphNodes, currID, "visited")

		for _, edge := range adj[currID] {
			nbrID := edge.to
			setEdgeActive(graphEdges, currID, nbrID, true)

			captionEdge := fmt.Sprintf("Examine edge (%s -> %s) with weight %d", currID, nbrID, edge.weight)
			b.AddStep("L9", trace.ActionHighlight,
				captionEdge,
				cloneGraphState(captionEdge, graphNodes, graphEdges, wl),
				map[string]any{"from": currID, "to": nbrID, "weight": edge.weight, "current": currID},
			)

			if visited[nbrID] {
				setEdgeActive(graphEdges, currID, nbrID, false)
				captionClosed := fmt.Sprintf("Neighbor %s already in closed set (visited); skip", nbrID)
				b.AddStep("L9", trace.ActionHighlight,
					captionClosed,
					cloneGraphState(captionClosed, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "closed": true},
				)
				continue
			}

			tentativeG := gScore[currID] + edge.weight
			captionTentative := fmt.Sprintf("Compute tentative cost to %s: tentative_g = g(%s) + %d = %d", nbrID, currID, edge.weight, tentativeG)
			b.AddStep("L10", trace.ActionSet,
				captionTentative,
				cloneGraphState(captionTentative, graphNodes, graphEdges, wl),
				map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "current": currID},
			)

			b.IncComparisons()
			prevG, exists := gScore[nbrID]
			if !exists || tentativeG < prevG {
				prevGStr := "∞"
				if exists {
					prevGStr = fmt.Sprintf("%d", prevG)
				}
				captionImproved := fmt.Sprintf("Update neighbor %s: tentative g is lower (%d < %s)", nbrID, tentativeG, prevGStr)
				b.AddStep("L11", trace.ActionCompare,
					captionImproved,
					cloneGraphState(captionImproved, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "improved": true, "prevG": prevGStr},
				)

				cameFrom[nbrID] = currID
				captionCameFrom := fmt.Sprintf("Record predecessor: cameFrom[%s] = %s", nbrID, currID)
				b.AddStep("L12", trace.ActionSet,
					captionCameFrom,
					cloneGraphState(captionCameFrom, graphNodes, graphEdges, wl),
					map[string]any{"cameFrom": currID, "neighbor": nbrID},
				)

				gScore[nbrID] = tentativeG
				captionGScore := fmt.Sprintf("Update gScore[%s] = %d (known distance from start)", nbrID, tentativeG)
				b.AddStep("L13", trace.ActionSet,
					captionGScore,
					cloneGraphState(captionGScore, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "g": tentativeG},
				)

				nbrNode := nodeMap[nbrID]
				hVal := EuclideanDistance(nbrNode, goalNode)
				fVal := tentativeG + hVal
				fScore[nbrID] = fVal
				captionFScore := fmt.Sprintf("Calculate fScore[%s] = g(%d) + h(%d) = %d", nbrID, tentativeG, hVal, fVal)
				b.AddStep("L14", trace.ActionSet,
					captionFScore,
					cloneGraphState(captionFScore, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "g": tentativeG, "h": hVal, "f": fVal},
				)

				if it, ok := itemMap[nbrID]; ok {
					it.fScore = fVal
					it.hScore = hVal
					heap.Fix(pq, it.index)
				} else {
					it := &pqItem{node: nbrID, fScore: fVal, hScore: hVal}
					heap.Push(pq, it)
					itemMap[nbrID] = it
				}
				setNodeState(graphNodes, nbrID, "queued")

				wl = buildWorklist(*pq, gScore)
				captionQueue := fmt.Sprintf("Queue neighbor %s in open set with f(%d) = g(%d) + h(%d)", nbrID, fVal, tentativeG, hVal)
				b.AddStep("L15", trace.ActionPush,
					captionQueue,
					cloneGraphState(captionQueue, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "g": tentativeG, "h": hVal, "f": fVal, "openSetCount": pq.Len()},
				)
			} else {
				captionSkip := fmt.Sprintf("Skip neighbor %s: tentative g(%d) >= known g(%d)", nbrID, tentativeG, prevG)
				b.AddStep("L11", trace.ActionCompare,
					captionSkip,
					cloneGraphState(captionSkip, graphNodes, graphEdges, wl),
					map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "improved": false, "prevG": prevG},
				)
			}

			setEdgeActive(graphEdges, currID, nbrID, false)
		}
	}

	captionExhausted := fmt.Sprintf("Open set exhausted; no path found between %s and %s", start, goal)
	b.AddStep("L19", trace.ActionDone,
		captionExhausted,
		cloneGraphState(captionExhausted, graphNodes, graphEdges, nil),
		map[string]any{"start": start, "goal": goal, "found": false, "cost": -1},
	)

	return b.Steps()
}

func buildWorklist(pq priorityQueue, gScore map[string]int) []renderers.WorklistItem {
	items := make([]*pqItem, len(pq))
	copy(items, pq)
	sort.Slice(items, func(i, j int) bool {
		if items[i].fScore != items[j].fScore {
			return items[i].fScore < items[j].fScore
		}
		if items[i].hScore != items[j].hScore {
			return items[i].hScore < items[j].hScore
		}
		return items[i].node < items[j].node
	})
	res := make([]renderers.WorklistItem, len(items))
	for i, it := range items {
		res[i] = renderers.WorklistItem{
			ID:    it.node,
			Label: fmt.Sprintf("%s: f=%d (g=%d, h=%d)", it.node, it.fScore, gScore[it.node], it.hScore),
		}
	}
	return res
}

func cloneGraphState(caption string, nodes []renderers.GraphNode, edges []renderers.GraphEdge, worklist []renderers.WorklistItem) renderers.GraphState {
	n := make([]renderers.GraphNode, len(nodes))
	copy(n, nodes)
	e := make([]renderers.GraphEdge, len(edges))
	copy(e, edges)
	var w []renderers.WorklistItem
	if len(worklist) > 0 {
		w = make([]renderers.WorklistItem, len(worklist))
		copy(w, worklist)
	} else {
		w = []renderers.WorklistItem{}
	}
	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: "Open Set (Min-Heap: node, f = g + h)",
		Worklist:      w,
		Nodes:         n,
		Edges:         e,
	}
}

func setNodeState(nodes []renderers.GraphNode, id, state string) {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].State = state
			return
		}
	}
}

func setEdgeActive(edges []renderers.GraphEdge, u, v string, active bool) {
	for i := range edges {
		if (edges[i].From == u && edges[i].To == v) || (edges[i].From == v && edges[i].To == u) {
			edges[i].Active = active
			return
		}
	}
}
