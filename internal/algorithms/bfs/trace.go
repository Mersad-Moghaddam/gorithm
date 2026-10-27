package bfs

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Breadth-First Search traversal.
// ponytail: adj derived from slice of edges; adjacency list cache faster for dynamic updates.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step {
	if len(nodes) == 0 {
		return nil
	}

	if start == "" {
		start = nodes[0].ID
	}

	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
	}

	nodeStates := make(map[string]string, len(nodes))
	for _, n := range nodes {
		nodeStates[n.ID] = "unvisited"
	}

	b := trace.NewBuilder()

	// Initial step
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize BFS traversal starting at node %s. All nodes unvisited.", start),
		cloneGraphState(nodes, nodeStates, edges, "", "",
			fmt.Sprintf("Start BFS from node %s; all nodes unvisited", start), nil),
		map[string]any{"start": start})

	visited := make(map[string]bool)
	b.AddStep("L2", trace.ActionSet,
		"Initialize visited tracking set.",
		cloneGraphState(nodes, nodeStates, edges, "", "",
			"Initialize visited tracking set", nil),
		map[string]any{"visited": []string{}})

	queue := []string{start}
	nodeStates[start] = "queued"
	b.IncOperations()
	b.AddStep("L3", trace.ActionPush,
		fmt.Sprintf("Enqueue start node %s into FIFO queue and mark it queued.", start),
		cloneGraphState(nodes, nodeStates, edges, "", "",
			fmt.Sprintf("Enqueue start node %s into FIFO queue", start), queue),
		map[string]any{"start": start, "queue": cloneSlice(queue)})

	visited[start] = true
	b.IncOperations()
	b.AddStep("L4", trace.ActionSet,
		fmt.Sprintf("Mark start node %s as discovered in visited set.", start),
		cloneGraphState(nodes, nodeStates, edges, "", "",
			fmt.Sprintf("Mark start node %s as discovered", start), queue),
		map[string]any{"start": start, "visited": []string{start}})

	var order []string
	b.AddStep("L5", trace.ActionSet,
		"Initialize empty traversal order list.",
		cloneGraphState(nodes, nodeStates, edges, "", "",
			"Initialize traversal order list", queue),
		map[string]any{"queue": cloneSlice(queue), "order": cloneSlice(order)})

	for len(queue) > 0 {
		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare,
			fmt.Sprintf("Queue is not empty (%v). Continue BFS loop.", queue),
			cloneGraphState(nodes, nodeStates, edges, "", "",
				fmt.Sprintf("Queue has %d node(s): [%s]", len(queue), strings.Join(queue, ", ")), queue),
			map[string]any{"queue": cloneSlice(queue), "order": cloneSlice(order)})

		curr := queue[0]
		b.AddStep("L7", trace.ActionPop,
			fmt.Sprintf("Inspect front node %s from queue.", curr),
			cloneGraphState(nodes, nodeStates, edges, "", "",
				fmt.Sprintf("Read front node %s from queue", curr), queue),
			map[string]any{"curr": curr, "queue": cloneSlice(queue)})

		queue = queue[1:]
		nodeStates[curr] = "visiting"
		b.IncOperations()

		unvisited := unvisitedNeighbors(adj, curr, visited)
		captionL8 := fmt.Sprintf("Dequeue front node %s; inspect unvisited neighbors %s", curr, formatUnvisited(unvisited))
		b.AddStep("L8", trace.ActionPop,
			fmt.Sprintf("Dequeue node %s from front of queue to explore neighbors. Mark %s as visiting.", curr, curr),
			cloneGraphState(nodes, nodeStates, edges, "", "", captionL8, queue),
			map[string]any{"curr": curr, "queue": cloneSlice(queue)})

		order = append(order, curr)
		b.IncOperations()
		b.AddStep("L9", trace.ActionVisit,
			fmt.Sprintf("Visit node %s and append it to traversal order.", curr),
			cloneGraphState(nodes, nodeStates, edges, "", "",
				fmt.Sprintf("Visit node %s; append to visit order [%s]", curr, strings.Join(order, ", ")), queue),
			map[string]any{"curr": curr, "order": cloneSlice(order)})

		for _, neighbor := range adj[curr] {
			b.IncComparisons()
			b.AddStep("L10", trace.ActionHighlight,
				fmt.Sprintf("Inspect outgoing edge %s -> %s.", curr, neighbor),
				cloneGraphState(nodes, nodeStates, edges, curr, neighbor,
					fmt.Sprintf("Inspect edge %s -> %s from current node %s", curr, neighbor, curr), queue),
				map[string]any{"curr": curr, "neighbor": neighbor})

			b.IncComparisons()
			if !visited[neighbor] {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("Neighbor node %s is unvisited. Prepare to enqueue.", neighbor),
					cloneGraphState(nodes, nodeStates, edges, curr, neighbor,
						fmt.Sprintf("Neighbor %s is unvisited; prepare to enqueue", neighbor), queue),
					map[string]any{"curr": curr, "neighbor": neighbor})

				visited[neighbor] = true
				b.IncOperations()
				b.AddStep("L12", trace.ActionSet,
					fmt.Sprintf("Mark %s as discovered in visited set.", neighbor),
					cloneGraphState(nodes, nodeStates, edges, curr, neighbor,
						fmt.Sprintf("Mark neighbor %s as discovered", neighbor), queue),
					map[string]any{"curr": curr, "neighbor": neighbor, "visited": true})

				queue = append(queue, neighbor)
				nodeStates[neighbor] = "queued"
				b.IncOperations()
				b.AddStep("L13", trace.ActionPush,
					fmt.Sprintf("Mark %s as queued and enqueue to back of queue: %v.", neighbor, queue),
					cloneGraphState(nodes, nodeStates, edges, curr, neighbor,
						fmt.Sprintf("Enqueue %s to back of FIFO queue; queue: [%s]", neighbor, strings.Join(queue, ", ")), queue),
					map[string]any{"curr": curr, "neighbor": neighbor, "queue": cloneSlice(queue)})
			} else {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("Neighbor node %s is already visited or queued. Skip.", neighbor),
					cloneGraphState(nodes, nodeStates, edges, curr, neighbor,
						fmt.Sprintf("Neighbor %s already discovered; skip", neighbor), queue),
					map[string]any{"curr": curr, "neighbor": neighbor})
			}
		}

		nodeStates[curr] = "visited"
		b.IncOperations()
		b.AddStep("L15", trace.ActionSet,
			fmt.Sprintf("Finished inspecting neighbors of %s. Mark %s as visited.", curr, curr),
			cloneGraphState(nodes, nodeStates, edges, "", "",
				fmt.Sprintf("Finished all neighbors of %s; mark %s as visited", curr, curr), queue),
			map[string]any{"curr": curr, "queue": cloneSlice(queue), "order": cloneSlice(order)})
	}

	b.IncComparisons()
	b.AddStep("L6", trace.ActionCompare,
		"Queue is empty. Traversal loop terminates.",
		cloneGraphState(nodes, nodeStates, edges, "", "",
			"Queue is empty; traversal terminates", []string{}),
		map[string]any{"queue": []string{}, "order": cloneSlice(order)})

	b.AddStep("L17", trace.ActionDone,
		fmt.Sprintf("BFS traversal complete. Traversal order: %v.", order),
		cloneGraphState(nodes, nodeStates, edges, "", "",
			fmt.Sprintf("BFS traversal complete; visit order: [%s]", strings.Join(order, ", ")), []string{}),
		map[string]any{"order": cloneSlice(order), "status": "complete"})

	return b.Steps()
}

func cloneGraphState(nodes []renderers.GraphNode, nodeStates map[string]string, edges []renderers.GraphEdge, activeFrom, activeTo, caption string, queue []string) renderers.GraphState {
	clonedNodes := make([]renderers.GraphNode, len(nodes))
	for i, n := range nodes {
		clonedNodes[i] = n
		if st, ok := nodeStates[n.ID]; ok {
			clonedNodes[i].State = st
		}
	}

	clonedEdges := make([]renderers.GraphEdge, len(edges))
	for i, e := range edges {
		clonedEdges[i] = e
		if activeFrom != "" && activeTo != "" &&
			((e.From == activeFrom && e.To == activeTo) || (!e.Directed && e.From == activeTo && e.To == activeFrom)) {
			clonedEdges[i].Active = true
		} else {
			clonedEdges[i].Active = false
		}
	}

	worklist := make([]renderers.WorklistItem, len(queue))
	for i, item := range queue {
		worklist[i] = renderers.WorklistItem{
			ID:    item,
			Label: item,
		}
	}

	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: "Queue (FIFO)",
		Worklist:      worklist,
		Nodes:         clonedNodes,
		Edges:         clonedEdges,
	}
}

func unvisitedNeighbors(adj map[string][]string, curr string, visited map[string]bool) []string {
	var res []string
	for _, n := range adj[curr] {
		if !visited[n] {
			res = append(res, n)
		}
	}
	return res
}

func formatUnvisited(nodes []string) string {
	if len(nodes) == 0 {
		return "none"
	}
	return strings.Join(nodes, ", ")
}

func cloneSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	res := make([]string, len(s))
	copy(res, s)
	return res
}
