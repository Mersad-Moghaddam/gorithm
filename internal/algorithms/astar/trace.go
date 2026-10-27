package astar

import (
	"container/heap"
	"fmt"

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
			From:   e.From,
			To:     e.To,
			Weight: e.Weight,
			Active: false,
		}
	}

	startNode, hasStart := nodeMap[start]
	goalNode, hasGoal := nodeMap[goal]
	if !hasStart || !hasGoal {
		b.AddStep("L1", trace.ActionInit,
			fmt.Sprintf("Invalid search parameters: start %q or goal %q not in graph.", start, goal),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"start": start, "goal": goal},
		)
		b.AddStep("L19", trace.ActionDone,
			"Cannot proceed with search. Returning nil.",
			cloneGraphState(graphNodes, graphEdges),
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

	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Start A* pathfinding from %s to %s with %d vertices and %d edges.", start, goal, len(nodes), len(edges)),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"start": start, "goal": goal, "vertices": len(nodes), "edges": len(edges)},
	)

	b.AddStep("L2", trace.ActionSet,
		"Initialize priority queue (open set), gScore, and fScore maps.",
		cloneGraphState(graphNodes, graphEdges),
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

	b.AddStep("L3", trace.ActionPush,
		fmt.Sprintf("Pushed start node %s into open set (g=0, h=%d, f=%d).", start, hStart, hStart),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"node": start, "g": 0, "h": hStart, "f": hStart, "openSetCount": pq.Len()},
	)

	for pq.Len() > 0 {
		b.AddStep("L4", trace.ActionHighlight,
			fmt.Sprintf("Open set has %d candidate node(s).", pq.Len()),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"openSetCount": pq.Len()},
		)

		curr := heap.Pop(pq).(*pqItem)
		delete(itemMap, curr.node)
		currID := curr.node
		setNodeState(graphNodes, currID, "visiting")
		b.IncOperations()

		b.AddStep("L5", trace.ActionPop,
			fmt.Sprintf("Selected node %s with lowest f(n) = %d (g=%d, h=%d).", currID, curr.fScore, gScore[currID], curr.hScore),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"current": currID, "g": gScore[currID], "h": curr.hScore, "f": curr.fScore},
		)

		b.IncComparisons()
		if currID == goal {
			b.AddStep("L6", trace.ActionCompare,
				fmt.Sprintf("Node %s is goal %s. Path found!", currID, goal),
				cloneGraphState(graphNodes, graphEdges),
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

			b.AddStep("L7", trace.ActionDone,
				fmt.Sprintf("Shortest path found: %v with total cost %d.", path, gScore[goal]),
				cloneGraphState(graphNodes, graphEdges),
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

		b.AddStep("L6", trace.ActionCompare,
			fmt.Sprintf("Node %s is not goal %s. Exploring neighbors.", currID, goal),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"current": currID, "goal": goal, "reached": false, "g": gScore[currID], "h": curr.hScore, "f": curr.fScore},
		)

		visited[currID] = true
		setNodeState(graphNodes, currID, "visited")

		for _, edge := range adj[currID] {
			nbrID := edge.to
			setEdgeActive(graphEdges, currID, nbrID, true)

			b.AddStep("L9", trace.ActionHighlight,
				fmt.Sprintf("Examine edge (%s -> %s) with weight %d.", currID, nbrID, edge.weight),
				cloneGraphState(graphNodes, graphEdges),
				map[string]any{"from": currID, "to": nbrID, "weight": edge.weight, "current": currID},
			)

			if visited[nbrID] {
				setEdgeActive(graphEdges, currID, nbrID, false)
				b.AddStep("L9", trace.ActionHighlight,
					fmt.Sprintf("Neighbor %s already in closed set. Skip.", nbrID),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"neighbor": nbrID, "closed": true},
				)
				continue
			}

			tentativeG := gScore[currID] + edge.weight
			b.AddStep("L10", trace.ActionSet,
				fmt.Sprintf("Tentative g(%s) = g(%s) + %d = %d.", nbrID, currID, edge.weight, tentativeG),
				cloneGraphState(graphNodes, graphEdges),
				map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "current": currID},
			)

			b.IncComparisons()
			prevG, exists := gScore[nbrID]
			if !exists || tentativeG < prevG {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("Tentative g(%d) is shorter than known g(%v) for %s.", tentativeG, prevG, nbrID),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "improved": true},
				)

				cameFrom[nbrID] = currID
				b.AddStep("L12", trace.ActionSet,
					fmt.Sprintf("cameFrom[%s] = %s", nbrID, currID),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"cameFrom": currID, "neighbor": nbrID},
				)

				gScore[nbrID] = tentativeG
				b.AddStep("L13", trace.ActionSet,
					fmt.Sprintf("gScore[%s] = %d", nbrID, tentativeG),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"neighbor": nbrID, "g": tentativeG},
				)

				nbrNode := nodeMap[nbrID]
				hVal := EuclideanDistance(nbrNode, goalNode)
				fVal := tentativeG + hVal
				fScore[nbrID] = fVal
				b.AddStep("L14", trace.ActionSet,
					fmt.Sprintf("fScore[%s] = g(%d) + h(%d) = %d", nbrID, tentativeG, hVal, fVal),
					cloneGraphState(graphNodes, graphEdges),
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

				b.AddStep("L15", trace.ActionPush,
					fmt.Sprintf("Enqueued neighbor %s in open set (f=%d).", nbrID, fVal),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"neighbor": nbrID, "g": tentativeG, "h": hVal, "f": fVal, "openSetCount": pq.Len()},
				)
			} else {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("Tentative g(%d) >= known g(%d). Skip updating %s.", tentativeG, prevG, nbrID),
					cloneGraphState(graphNodes, graphEdges),
					map[string]any{"neighbor": nbrID, "tentativeG": tentativeG, "improved": false},
				)
			}

			setEdgeActive(graphEdges, currID, nbrID, false)
		}
	}

	b.AddStep("L19", trace.ActionDone,
		fmt.Sprintf("Open set exhausted. No path found between %s and %s.", start, goal),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"start": start, "goal": goal, "found": false, "cost": -1},
	)

	return b.Steps()
}

func cloneGraphState(nodes []renderers.GraphNode, edges []renderers.GraphEdge) renderers.GraphState {
	n := make([]renderers.GraphNode, len(nodes))
	copy(n, nodes)
	e := make([]renderers.GraphEdge, len(edges))
	copy(e, edges)
	return renderers.GraphState{Nodes: n, Edges: e}
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
