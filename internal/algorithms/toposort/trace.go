package toposort

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Kahn's topological sort algorithm.
// ponytail: map-based adjacency and in-degree tracking; dense indexed slices faster when node count exceeds 10k.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		b.AddStep("L16", trace.ActionDone,
			"Empty graph: topological sort complete.",
			renderers.GraphState{
				Caption:       "Topological order: []",
				WorklistLabel: "Ready Queue (In-Degree 0)",
				Worklist:      []renderers.WorklistItem{},
				Nodes:         []renderers.GraphNode{},
				Edges:         []renderers.GraphEdge{},
			},
			map[string]any{
				"inDegrees": map[string]int{},
				"order":     []string{},
				"processed": []string{},
				"queue":     []string{},
				"status":    "complete",
				"valid":     true,
			},
		)
		return b.Steps()
	}

	nodeIDs := make([]string, len(nodes))
	nodeStates := make(map[string]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
		nodeStates[n.ID] = "unvisited"
	}

	adj := make(map[string][]string, len(nodes))
	inDegree := make(map[string]int, len(nodes))
	for _, id := range nodeIDs {
		inDegree[id] = 0
	}
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
		inDegree[e.To]++
	}

	// L1: Init
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize Kahn's topological sort for DAG with %d tasks and %d dependencies.", len(nodes), len(edges)),
		cloneGraphState(
			fmt.Sprintf("Initialize Kahn's topological sort (%d tasks, %d dependencies)", len(nodes), len(edges)),
			nodes, nodeStates, edges, "", "", nil,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     []string{},
			"processed": []string{},
			"queue":     []string{},
		},
	)

	// L2: Calculate in-degrees
	b.IncOperations()
	b.AddStep("L2", trace.ActionSet,
		fmt.Sprintf("Compute in-degree for each vertex by counting incoming dependencies: %s.", formatInDegrees(nodeIDs, inDegree)),
		cloneGraphState(
			fmt.Sprintf("Calculated in-degrees: %s", formatInDegrees(nodeIDs, inDegree)),
			nodes, nodeStates, edges, "", "", nil,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     []string{},
			"processed": []string{},
			"queue":     []string{},
		},
	)

	// L3: Initialize queue
	b.AddStep("L3", trace.ActionSet,
		"Initialize ready queue to hold vertices that have no prerequisites (in-degree 0).",
		cloneGraphState(
			"Initialize ready queue for in-degree 0 vertices",
			nodes, nodeStates, edges, "", "", nil,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     []string{},
			"processed": []string{},
			"queue":     []string{},
		},
	)

	// L4: Find all nodes with in-degree 0
	var queue []string
	for _, id := range nodeIDs {
		b.IncComparisons()
		if inDegree[id] == 0 {
			queue = append(queue, id)
			nodeStates[id] = "queued"
			b.IncOperations()
		}
	}

	l4Caption := fmt.Sprintf("Enqueued in-degree 0 vertices into ready queue: %v", queue)
	if len(queue) == 0 {
		l4Caption = "No vertices with in-degree 0 found"
	}
	b.AddStep("L4", trace.ActionPush,
		fmt.Sprintf("Find vertices with in-degree 0 (no prerequisites) and enqueue: %v.", queue),
		cloneGraphState(
			l4Caption,
			nodes, nodeStates, edges, "", "", queue,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     []string{},
			"processed": []string{},
			"queue":     cloneSlice(queue),
		},
	)

	order := make([]string, 0, len(nodes))
	b.AddStep("L5", trace.ActionSet,
		"Initialize empty list to store topologically sorted order.",
		cloneGraphState(
			"Initialize empty topological order list",
			nodes, nodeStates, edges, "", "", queue,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     cloneSlice(order),
			"processed": cloneSlice(order),
			"queue":     cloneSlice(queue),
		},
	)

	for len(queue) > 0 {
		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare,
			fmt.Sprintf("Ready queue contains %d ready vertex/vertices: %v. Proceed with next task.", len(queue), queue),
			cloneGraphState(
				fmt.Sprintf("Ready queue contains %d task(s): %v", len(queue), queue),
				nodes, nodeStates, edges, "", "", queue,
			),
			map[string]any{
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"queue":     cloneSlice(queue),
			},
		)

		curr := queue[0]
		queue = queue[1:]
		nodeStates[curr] = "visiting"
		b.IncOperations()

		b.AddStep("L7", trace.ActionPop,
			fmt.Sprintf("Dequeued vertex %q from ready queue with in-degree 0. Marked visiting.", curr),
			cloneGraphState(
				fmt.Sprintf("Dequeued %q (in-degree 0); marked visiting", curr),
				nodes, nodeStates, edges, "", "", queue,
			),
			map[string]any{
				"curr":      curr,
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"queue":     cloneSlice(queue),
			},
		)

		order = append(order, curr)
		nodeStates[curr] = "visited"
		b.IncOperations()

		b.AddStep("L8", trace.ActionVisit,
			fmt.Sprintf("Add vertex %q to topological order: %v.", curr, order),
			cloneGraphState(
				fmt.Sprintf("Appended %q to topological order: %v", curr, order),
				nodes, nodeStates, edges, "", "", queue,
			),
			map[string]any{
				"curr":      curr,
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"queue":     cloneSlice(queue),
			},
		)

		for _, neighbor := range adj[curr] {
			b.IncComparisons()
			b.AddStep("L9", trace.ActionHighlight,
				fmt.Sprintf("Inspect dependency edge %q -> %q.", curr, neighbor),
				cloneGraphState(
					fmt.Sprintf("Inspect dependency edge %s -> %s", curr, neighbor),
					nodes, nodeStates, edges, curr, neighbor, queue,
				),
				map[string]any{
					"curr":      curr,
					"neighbor":  neighbor,
					"inDegrees": cloneInDegrees(inDegree),
					"order":     cloneSlice(order),
					"processed": cloneSlice(order),
					"queue":     cloneSlice(queue),
				},
			)

			inDegree[neighbor]--
			b.IncOperations()
			b.AddStep("L10", trace.ActionSet,
				fmt.Sprintf("Decrement in-degree of %q to %d (dependency on %q resolved).", neighbor, inDegree[neighbor], curr),
				cloneGraphState(
					fmt.Sprintf("Decremented in-degree of %s: now %d", neighbor, inDegree[neighbor]),
					nodes, nodeStates, edges, curr, neighbor, queue,
				),
				map[string]any{
					"curr":      curr,
					"neighbor":  neighbor,
					"inDegrees": cloneInDegrees(inDegree),
					"order":     cloneSlice(order),
					"processed": cloneSlice(order),
					"queue":     cloneSlice(queue),
				},
			)

			b.IncComparisons()
			if inDegree[neighbor] == 0 {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("All prerequisites for %q resolved (in-degree 0).", neighbor),
					cloneGraphState(
						fmt.Sprintf("In-degree of %s reached 0", neighbor),
						nodes, nodeStates, edges, "", "", queue,
					),
					map[string]any{
						"curr":      curr,
						"neighbor":  neighbor,
						"inDegrees": cloneInDegrees(inDegree),
						"order":     cloneSlice(order),
						"processed": cloneSlice(order),
						"queue":     cloneSlice(queue),
					},
				)

				queue = append(queue, neighbor)
				nodeStates[neighbor] = "queued"
				b.IncOperations()

				b.AddStep("L12", trace.ActionPush,
					fmt.Sprintf("In-degree reached 0; added to ready queue (%s).", neighbor),
					cloneGraphState(
						"In-degree reached 0; added to ready queue",
						nodes, nodeStates, edges, "", "", queue,
					),
					map[string]any{
						"curr":      curr,
						"neighbor":  neighbor,
						"inDegrees": cloneInDegrees(inDegree),
						"order":     cloneSlice(order),
						"processed": cloneSlice(order),
						"queue":     cloneSlice(queue),
					},
				)
			} else {
				b.AddStep("L11", trace.ActionCompare,
					fmt.Sprintf("Vertex %q still has %d unresolved dependency(ies).", neighbor, inDegree[neighbor]),
					cloneGraphState(
						fmt.Sprintf("Node %s still has %d unresolved prerequisite(s)", neighbor, inDegree[neighbor]),
						nodes, nodeStates, edges, "", "", queue,
					),
					map[string]any{
						"curr":      curr,
						"neighbor":  neighbor,
						"inDegrees": cloneInDegrees(inDegree),
						"order":     cloneSlice(order),
						"processed": cloneSlice(order),
						"queue":     cloneSlice(queue),
					},
				)
			}
		}
	}

	b.IncComparisons()
	b.AddStep("L6", trace.ActionCompare,
		"Ready queue is empty. Topological processing finished.",
		cloneGraphState(
			"Ready queue is empty",
			nodes, nodeStates, edges, "", "", nil,
		),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     cloneSlice(order),
			"processed": cloneSlice(order),
			"queue":     []string{},
		},
	)

	isDAG := len(order) == len(nodes)
	if isDAG {
		orderStr := "[]"
		if len(order) > 0 {
			orderStr = strings.Join(order, " -> ")
		}
		b.AddStep("L16", trace.ActionDone,
			fmt.Sprintf("Topological sort complete! Valid DAG with order: %v.", order),
			cloneGraphState(
				fmt.Sprintf("Topological order: %s", orderStr),
				nodes, nodeStates, edges, "", "", nil,
			),
			map[string]any{
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"queue":     []string{},
				"status":    "complete",
				"valid":     true,
			},
		)
	} else {
		b.AddStep("L16", trace.ActionDone,
			fmt.Sprintf("Cycle detected! Remaining vertices cannot be resolved. Only %d of %d resolved (%v).", len(order), len(nodes), order),
			cloneGraphState(
				"Cycle detected! Remaining vertices cannot be resolved",
				nodes, nodeStates, edges, "", "", nil,
			),
			map[string]any{
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"queue":     []string{},
				"status":    "cycle_detected",
				"valid":     false,
			},
		)
	}

	return b.Steps()
}

func cloneGraphState(
	caption string,
	nodes []renderers.GraphNode,
	nodeStates map[string]string,
	edges []renderers.GraphEdge,
	activeFrom, activeTo string,
	queue []string,
) renderers.GraphState {
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
		clonedEdges[i].Directed = true
		if activeFrom != "" && activeTo != "" && e.From == activeFrom && e.To == activeTo {
			clonedEdges[i].Active = true
		} else {
			clonedEdges[i].Active = false
		}
	}

	worklist := make([]renderers.WorklistItem, len(queue))
	for i, id := range queue {
		worklist[i] = renderers.WorklistItem{
			ID:    id,
			Label: id,
		}
	}

	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: "Ready Queue (In-Degree 0)",
		Worklist:      worklist,
		Nodes:         clonedNodes,
		Edges:         clonedEdges,
	}
}

func formatInDegrees(nodes []string, inDegree map[string]int) string {
	parts := make([]string, len(nodes))
	for i, id := range nodes {
		parts[i] = fmt.Sprintf("%s: %d", id, inDegree[id])
	}
	return strings.Join(parts, ", ")
}

func cloneSlice(s []string) []string {
	if s == nil {
		return nil
	}
	res := make([]string, len(s))
	copy(res, s)
	return res
}

func cloneInDegrees(m map[string]int) map[string]int {
	res := make(map[string]int, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
