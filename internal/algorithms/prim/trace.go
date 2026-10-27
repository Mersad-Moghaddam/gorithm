package prim

import (
	"cmp"
	"container/heap"
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

const worklistLabel = "Priority Queue (Crossing Edges: edge, weight)"

func isEdgeInMST(e WeightedEdge, mst []WeightedEdge) bool {
	for _, m := range mst {
		if (e.From == m.From && e.To == m.To) || (e.From == m.To && e.To == m.From) {
			return true
		}
	}
	return false
}

func pqToWorklist(pq *minEdgeHeap) []renderers.WorklistItem {
	if pq == nil || pq.Len() == 0 {
		return []renderers.WorklistItem{}
	}
	items := make([]edgeItem, pq.Len())
	copy(items, *pq)
	slices.SortFunc(items, func(a, b edgeItem) int {
		if a.weight != b.weight {
			return cmp.Compare(a.weight, b.weight)
		}
		if a.from != b.from {
			return cmp.Compare(a.from, b.from)
		}
		return cmp.Compare(a.to, b.to)
	})
	res := make([]renderers.WorklistItem, len(items))
	for i, it := range items {
		res[i] = renderers.WorklistItem{
			ID:    fmt.Sprintf("%s-%s-%d", it.from, it.to, it.weight),
			Label: fmt.Sprintf("%s-%s, %d", it.from, it.to, it.weight),
		}
	}
	return res
}

func makeGraphState(
	nodes []renderers.GraphNode,
	allEdges []WeightedEdge,
	nodeStates map[string]string,
	mstEdges []WeightedEdge,
	candidateEdge *edgeItem,
	worklist []renderers.WorklistItem,
	caption string,
) renderers.GraphState {
	stateNodes := make([]renderers.GraphNode, len(nodes))
	for i, n := range nodes {
		st := "unvisited"
		if s, ok := nodeStates[n.ID]; ok {
			st = s
		}
		stateNodes[i] = renderers.GraphNode{
			ID:    n.ID,
			Label: n.Label,
			X:     n.X,
			Y:     n.Y,
			State: st,
		}
	}

	stateEdges := make([]renderers.GraphEdge, len(allEdges))
	for i, e := range allEdges {
		active := isEdgeInMST(e, mstEdges)
		if !active && candidateEdge != nil {
			if (e.From == candidateEdge.from && e.To == candidateEdge.to) ||
				(e.From == candidateEdge.to && e.To == candidateEdge.from) {
				active = true
			}
		}
		stateEdges[i] = renderers.GraphEdge{
			From:   e.From,
			To:     e.To,
			Weight: e.Weight,
			Active: active,
		}
	}

	clonedWorklist := make([]renderers.WorklistItem, len(worklist))
	copy(clonedWorklist, worklist)

	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: worklistLabel,
		Worklist:      clonedWorklist,
		Nodes:         stateNodes,
		Edges:         stateEdges,
	}
}

func Trace(nodes []renderers.GraphNode, edges []WeightedEdge, start string) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		b.AddStep("L1", trace.ActionDone, "Empty graph; MST is empty.", renderers.GraphState{
			Caption:       "Graph is empty",
			WorklistLabel: worklistLabel,
			Worklist:      []renderers.WorklistItem{},
			Nodes:         []renderers.GraphNode{},
			Edges:         []renderers.GraphEdge{},
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	nodeSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		nodeSet[n.ID] = true
	}

	if start == "" || !nodeSet[start] {
		start = nodes[0].ID
	}

	nodeStates := make(map[string]string, len(nodes))
	for _, n := range nodes {
		nodeStates[n.ID] = "unvisited"
	}

	b.IncOperations()
	caption := fmt.Sprintf("Prim's MST: Start vertex %s | Total weight: 0", start)
	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Prim's algorithm starting from vertex %s.", start),
		makeGraphState(nodes, edges, nodeStates, nil, nil, nil, caption),
		map[string]any{"start": start, "totalNodes": len(nodes)})

	mst := make([]WeightedEdge, 0, len(nodes)-1)
	totalWeight := 0
	b.IncOperations()
	caption = "Growing MST: 0 edges | Total weight: 0"
	b.AddStep("L2", trace.ActionSet, "Initialize empty Minimum Spanning Tree.",
		makeGraphState(nodes, edges, nodeStates, mst, nil, nil, caption),
		map[string]any{"mstCount": 0, "totalWeight": 0})

	visited := make(map[string]bool, len(nodes))
	visited[start] = true
	nodeStates[start] = "visited"
	b.IncOperations()
	caption = fmt.Sprintf("Tree root: vertex %s | Total weight: 0", start)
	b.AddStep("L3", trace.ActionVisit, fmt.Sprintf("Mark start vertex %s as visited.", start),
		makeGraphState(nodes, edges, nodeStates, mst, nil, nil, caption),
		map[string]any{"visitedVertex": start})

	if len(nodes) == 1 {
		nodeStates[start] = "path"
		caption = "MST complete: 1 vertex connected with total weight 0"
		b.AddStep("L18", trace.ActionDone, "Single vertex graph; MST complete with total weight 0.",
			makeGraphState(nodes, edges, nodeStates, mst, nil, nil, caption),
			map[string]any{"totalWeight": 0, "status": "complete"})
		return b.Steps()
	}

	pq := &minEdgeHeap{}
	heap.Init(pq)
	b.IncOperations()
	caption = "Priority queue initialized | Total weight: 0"
	b.AddStep("L4", trace.ActionInit, "Initialize priority queue for candidate cut edges.",
		makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
		map[string]any{"pqSize": 0})

	adj := make(map[string][]edgeItem, len(nodes))
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], edgeItem{from: e.From, to: e.To, weight: e.Weight})
		adj[e.To] = append(adj[e.To], edgeItem{from: e.To, to: e.From, weight: e.Weight})
	}

	for _, e := range adj[start] {
		heap.Push(pq, e)
		if nodeStates[e.to] == "unvisited" {
			nodeStates[e.to] = "queued"
		}
		b.IncOperations()
		caption = fmt.Sprintf("Pushed cut edge %s-%s (%d) | PQ size: %d", e.from, e.to, e.weight, pq.Len())
		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Push cut edge %s-%s (weight %d) into priority queue.", e.from, e.to, e.weight),
			makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
			map[string]any{"from": e.from, "to": e.to, "weight": e.weight, "pqSize": pq.Len()})
	}

	for pq.Len() > 0 && len(mst) < len(nodes)-1 {
		b.IncComparisons()
		caption = fmt.Sprintf("MST: %d/%d edges | Total weight: %d", len(mst), len(nodes)-1, totalWeight)
		b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Check priority queue (%d candidates) and tree edges (%d/%d).", pq.Len(), len(mst), len(nodes)-1),
			makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
			map[string]any{"pqSize": pq.Len(), "mstCount": len(mst)})

		item := heap.Pop(pq).(edgeItem)
		b.IncOperations()

		prevState := nodeStates[item.to]
		if !visited[item.to] {
			nodeStates[item.to] = "visiting"
		}
		caption = fmt.Sprintf("Candidate edge: %s-%s (%d) | MST weight: %d", item.from, item.to, item.weight, totalWeight)
		b.AddStep("L7", trace.ActionPop, fmt.Sprintf("Pop minimum weight edge %s-%s (weight %d) from priority queue.", item.from, item.to, item.weight),
			makeGraphState(nodes, edges, nodeStates, mst, &item, pqToWorklist(pq), caption),
			map[string]any{"from": item.from, "to": item.to, "weight": item.weight})

		b.IncComparisons()
		caption = fmt.Sprintf("Evaluating edge: %s-%s (%d) | MST weight: %d", item.from, item.to, item.weight, totalWeight)
		b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Check if destination vertex %s is already in tree.", item.to),
			makeGraphState(nodes, edges, nodeStates, mst, &item, pqToWorklist(pq), caption),
			map[string]any{"vertex": item.to, "alreadyVisited": visited[item.to]})

		if visited[item.to] {
			nodeStates[item.to] = prevState
			caption = fmt.Sprintf("Cycle skipped: %s-%s | MST weight: %d", item.from, item.to, totalWeight)
			b.AddStep("L9", trace.ActionHighlight, fmt.Sprintf("Skip cycle: both endpoints already in tree (%s, %s); discard edge %s-%s (weight %d).", item.from, item.to, item.from, item.to, item.weight),
				makeGraphState(nodes, edges, nodeStates, mst, &item, pqToWorklist(pq), caption),
				map[string]any{"discarded": fmt.Sprintf("%s-%s", item.from, item.to), "cycle": true})
			continue
		}

		visited[item.to] = true
		nodeStates[item.to] = "visited"
		b.IncOperations()
		caption = fmt.Sprintf("Connected vertex %s to tree | MST weight: %d", item.to, totalWeight)
		b.AddStep("L11", trace.ActionVisit, fmt.Sprintf("Add vertex %s to visited tree set.", item.to),
			makeGraphState(nodes, edges, nodeStates, mst, &item, pqToWorklist(pq), caption),
			map[string]any{"vertex": item.to})

		chosenEdge := WeightedEdge{From: item.from, To: item.to, Weight: item.weight}
		mst = append(mst, chosenEdge)
		b.IncOperations()
		caption = fmt.Sprintf("Added edge %s-%s (%d) | MST: %d/%d edges", item.from, item.to, item.weight, len(mst), len(nodes)-1)
		b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Add edge %s-%s (weight %d) to Minimum Spanning Tree.", item.from, item.to, item.weight),
			makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
			map[string]any{"edge": fmt.Sprintf("%s-%s", item.from, item.to), "weight": item.weight, "mstCount": len(mst)})

		totalWeight += item.weight
		b.IncOperations()
		caption = fmt.Sprintf("MST weight updated to %d | %d/%d edges", totalWeight, len(mst), len(nodes)-1)
		b.AddStep("L13", trace.ActionSet, fmt.Sprintf("Update MST total weight to %d.", totalWeight),
			makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
			map[string]any{"totalWeight": totalWeight})

		caption = fmt.Sprintf("Exploring neighbors of %s | MST weight: %d", item.to, totalWeight)
		b.AddStep("L14", trace.ActionVisit, fmt.Sprintf("Examine outgoing edges from vertex %s.", item.to),
			makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
			map[string]any{"vertex": item.to})

		for _, next := range adj[item.to] {
			b.IncComparisons()
			if !visited[next.to] {
				heap.Push(pq, next)
				if nodeStates[next.to] != "visited" {
					nodeStates[next.to] = "queued"
				}
				b.IncOperations()
				caption = fmt.Sprintf("Pushed cut edge %s-%s (%d) | PQ size: %d", next.from, next.to, next.weight, pq.Len())
				b.AddStep("L15", trace.ActionPush, fmt.Sprintf("Push cut edge %s-%s (weight %d) into priority queue.", next.from, next.to, next.weight),
					makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), caption),
					map[string]any{"from": next.from, "to": next.to, "weight": next.weight, "pqSize": pq.Len()})
			}
		}
	}

	for v := range visited {
		nodeStates[v] = "path"
	}
	var doneCaption string
	var doneMsg string
	if len(visited) == len(nodes) {
		doneCaption = fmt.Sprintf("MST complete: %d vertices connected with total weight %d", len(nodes), totalWeight)
		doneMsg = fmt.Sprintf("Prim's MST complete! All %d vertices connected with %d edges, total weight %d.", len(nodes), len(mst), totalWeight)
	} else {
		doneCaption = fmt.Sprintf("Disconnected graph: %d/%d vertices connected with total weight %d", len(visited), len(nodes), totalWeight)
		doneMsg = fmt.Sprintf("Graph is disconnected; constructed MST for reachable component (%d edges, total weight %d).", len(mst), totalWeight)
	}
	b.AddStep("L18", trace.ActionDone, doneMsg,
		makeGraphState(nodes, edges, nodeStates, mst, nil, pqToWorklist(pq), doneCaption),
		map[string]any{"mstCount": len(mst), "totalWeight": totalWeight, "status": "complete"})

	return b.Steps()
}