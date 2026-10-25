package toposort

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Kahn's topological sort algorithm.
// ponytail: map-based adjacency and in-degree tracking; dense indexed slices faster when node count exceeds 10k.
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
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
		cloneGraphState(nodes, nodeStates, edges, "", ""),
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
		fmt.Sprintf("Compute in-degree for each node by counting incoming dependencies: %v.", inDegree),
		cloneGraphState(nodes, nodeStates, edges, "", ""),
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

	b.AddStep("L4", trace.ActionPush,
		fmt.Sprintf("Find nodes with in-degree 0 (no prerequisites) and enqueue: %v.", queue),
		cloneGraphState(nodes, nodeStates, edges, "", ""),
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
		cloneGraphState(nodes, nodeStates, edges, "", ""),
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
			fmt.Sprintf("Queue contains %d ready node(s): %v. Proceed with next task.", len(queue), queue),
			cloneGraphState(nodes, nodeStates, edges, "", ""),
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
			fmt.Sprintf("Dequeue node %q (in-degree 0). Begin resolving downstream dependents.", curr),
			cloneGraphState(nodes, nodeStates, edges, "", ""),
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
			fmt.Sprintf("Add node %q to topological order: %v.", curr, order),
			cloneGraphState(nodes, nodeStates, edges, "", ""),
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
				cloneGraphState(nodes, nodeStates, edges, curr, neighbor),
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
				cloneGraphState(nodes, nodeStates, edges, curr, neighbor),
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
				queue = append(queue, neighbor)
				nodeStates[neighbor] = "queued"
				b.IncOperations()

				b.AddStep("L12", trace.ActionPush,
					fmt.Sprintf("All dependencies for %q resolved (in-degree 0). Enqueue %q.", neighbor, neighbor),
					cloneGraphState(nodes, nodeStates, edges, "", ""),
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
					fmt.Sprintf("Node %q still has %d unresolved dependency(ies).", neighbor, inDegree[neighbor]),
					cloneGraphState(nodes, nodeStates, edges, "", ""),
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
		"Queue is empty. Topological processing finished.",
		cloneGraphState(nodes, nodeStates, edges, "", ""),
		map[string]any{
			"inDegrees": cloneInDegrees(inDegree),
			"order":     cloneSlice(order),
			"processed": cloneSlice(order),
			"queue":     []string{},
		},
	)

	isDAG := len(order) == len(nodes)
	if isDAG {
		b.AddStep("L16", trace.ActionDone,
			fmt.Sprintf("Topological sort complete! Valid DAG with order: %v.", order),
			cloneGraphState(nodes, nodeStates, edges, "", ""),
			map[string]any{
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"status":    "complete",
				"valid":     true,
			},
		)
	} else {
		b.AddStep("L16", trace.ActionDone,
			fmt.Sprintf("Cycle detected! Only %d of %d nodes resolved (%v). Graph contains cyclic dependencies.", len(order), len(nodes), order),
			cloneGraphState(nodes, nodeStates, edges, "", ""),
			map[string]any{
				"inDegrees": cloneInDegrees(inDegree),
				"order":     cloneSlice(order),
				"processed": cloneSlice(order),
				"status":    "cycle_detected",
				"valid":     false,
			},
		)
	}

	return b.Steps()
}

func cloneGraphState(nodes []renderers.GraphNode, nodeStates map[string]string, edges []renderers.GraphEdge, activeFrom, activeTo string) renderers.GraphState {
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
		if activeFrom != "" && activeTo != "" && e.From == activeFrom && e.To == activeTo {
			clonedEdges[i].Active = true
		} else {
			clonedEdges[i].Active = false
		}
	}

	return renderers.GraphState{
		Nodes: clonedNodes,
		Edges: clonedEdges,
	}
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
