package kruskal

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Kruskal's algorithm on a graph.
// ponytail: accepts slice of GraphNode with precalculated positions.
func Trace(nodes []renderers.GraphNode, edges []WeightedEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}

	graphNodes := cloneGraphNodes(nodes)
	graphEdges := make([]renderers.GraphEdge, len(edges))
	for i, e := range edges {
		graphEdges[i] = renderers.GraphEdge{
			From:   e.From,
			To:     e.To,
			Weight: e.Weight,
			Active: false,
		}
	}

	inMST := make(map[string]bool, len(nodes))

	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize Kruskal's algorithm with %d vertices and %d edges.", len(nodes), len(edges)),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"vertices": len(nodes), "edges": len(edges)},
	)

	b.IncOperations()
	sortedEdges := slices.Clone(edges)
	slices.SortFunc(sortedEdges, func(a, b WeightedEdge) int {
		return cmp.Compare(a.Weight, b.Weight)
	})

	b.AddStep("L2", trace.ActionHighlight,
		fmt.Sprintf("Sorted all %d edges by weight in ascending order.", len(sortedEdges)),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"sortedCount": len(sortedEdges)},
	)

	nodeIDs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}
	dsu := NewDSU(nodeIDs)
	mst := make([]WeightedEdge, 0, len(nodes)-1)
	totalWeight := 0

	b.AddStep("L3", trace.ActionSet,
		"Initialize Disjoint Set Union (DSU) with each vertex in its own set.",
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"mstCount": 0, "totalWeight": 0},
	)

	for _, edge := range sortedEdges {
		setNodeStates(graphNodes, edge.From, edge.To, "visiting")
		b.AddStep("L4", trace.ActionHighlight,
			fmt.Sprintf("Inspect edge (%s, %s) with weight %d.", edge.From, edge.To, edge.Weight),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"from": edge.From, "to": edge.To, "weight": edge.Weight},
		)

		b.IncComparisons()
		rootFrom := dsu.Find(edge.From)
		rootTo := dsu.Find(edge.To)

		if rootFrom == rootTo {
			restoreNodeStates(graphNodes, inMST)
			b.AddStep("L5", trace.ActionCompare,
				fmt.Sprintf("Find(%s) == Find(%s) == %q: forms a cycle. Reject edge (%s, %s).",
					edge.From, edge.To, rootFrom, edge.From, edge.To),
				cloneGraphState(graphNodes, graphEdges),
				map[string]any{"cycle": true, "root": rootFrom, "rejected": true},
			)
			continue
		}

		b.AddStep("L5", trace.ActionCompare,
			fmt.Sprintf("Find(%s)=%q != Find(%s)=%q: different components, no cycle.",
				edge.From, rootFrom, edge.To, rootTo),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"cycle": false, "rootFrom": rootFrom, "rootTo": rootTo},
		)

		dsu.Union(edge.From, edge.To)
		b.IncOperations()
		b.AddStep("L6", trace.ActionSet,
			fmt.Sprintf("Union component %q and component %q.", rootFrom, rootTo),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"componentA": rootFrom, "componentB": rootTo},
		)

		mst = append(mst, edge)
		totalWeight += edge.Weight
		inMST[edge.From] = true
		inMST[edge.To] = true
		markEdgeActive(graphEdges, edge.From, edge.To)
		setNodeStates(graphNodes, edge.From, edge.To, "visited")

		b.AddStep("L7", trace.ActionPush,
			fmt.Sprintf("Added edge (%s, %s) [weight %d] to MST. Total weight: %d.",
				edge.From, edge.To, edge.Weight, totalWeight),
			cloneGraphState(graphNodes, graphEdges),
			map[string]any{"mstCount": len(mst), "totalWeight": totalWeight},
		)

		if len(mst) == len(nodes)-1 {
			break
		}
	}

	for i := range graphNodes {
		if inMST[graphNodes[i].ID] {
			graphNodes[i].State = "visited"
		}
	}

	b.AddStep("L10", trace.ActionDone,
		fmt.Sprintf("Kruskal's MST complete! %d edges with total weight %d.", len(mst), totalWeight),
		cloneGraphState(graphNodes, graphEdges),
		map[string]any{"totalWeight": totalWeight, "status": "complete"},
	)

	return b.Steps()
}

func cloneGraphNodes(nodes []renderers.GraphNode) []renderers.GraphNode {
	cloned := make([]renderers.GraphNode, len(nodes))
	copy(cloned, nodes)
	return cloned
}

func cloneGraphState(nodes []renderers.GraphNode, edges []renderers.GraphEdge) renderers.GraphState {
	n := make([]renderers.GraphNode, len(nodes))
	copy(n, nodes)
	e := make([]renderers.GraphEdge, len(edges))
	copy(e, edges)
	return renderers.GraphState{Nodes: n, Edges: e}
}

func setNodeStates(nodes []renderers.GraphNode, u, v, state string) {
	for i := range nodes {
		if nodes[i].ID == u || nodes[i].ID == v {
			nodes[i].State = state
		}
	}
}

func restoreNodeStates(nodes []renderers.GraphNode, inMST map[string]bool) {
	for i := range nodes {
		if inMST[nodes[i].ID] {
			nodes[i].State = "visited"
		} else {
			nodes[i].State = "unvisited"
		}
	}
}

func markEdgeActive(edges []renderers.GraphEdge, u, v string) {
	for i := range edges {
		if (edges[i].From == u && edges[i].To == v) || (edges[i].From == v && edges[i].To == u) {
			edges[i].Active = true
			return
		}
	}
}
