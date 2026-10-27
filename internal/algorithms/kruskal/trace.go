package kruskal

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

const worklistTitle = "Sorted Edges by Weight"

// Trace generates visual execution steps for Kruskal's algorithm on a graph.
// ponytail: accepts slice of GraphNode with precalculated positions.
func Trace(nodes []renderers.GraphNode, edges []WeightedEdge) []trace.Step {
	b := trace.NewBuilder()
	if len(nodes) == 0 {
		return b.Steps()
	}

	if len(nodes) == 1 {
		nodeDisplay := syncNodeDisplay(nodes, nil, nil, "", "")
		b.AddStep("L11", trace.ActionDone,
			"Single vertex graph; MST contains 0 edges.",
			cloneGraphState("MST complete: total weight = 0, edges: []", worklistTitle, nodeDisplay, nil, nil),
			map[string]any{"status": "complete", "totalWeight": 0, "mstCount": 0},
		)
		return b.Steps()
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

	inMST := make(map[string]bool, len(nodes))

	// L1: Initialize
	initDisplay := syncNodeDisplay(nodes, nil, inMST, "", "")
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Initialize Kruskal's algorithm with %d vertices and %d edges.", len(nodes), len(edges)),
		cloneGraphState("Initialize Kruskal's algorithm: sort edges by weight", worklistTitle, initDisplay, graphEdges, nil),
		map[string]any{"vertices": len(nodes), "edges": len(edges)},
	)

	b.IncOperations()
	sortedEdges := slices.Clone(edges)
	slices.SortFunc(sortedEdges, func(a, b WeightedEdge) int {
		if c := cmp.Compare(a.Weight, b.Weight); c != 0 {
			return c
		}
		if c := cmp.Compare(a.From, b.From); c != 0 {
			return c
		}
		return cmp.Compare(a.To, b.To)
	})

	worklist := make([]renderers.WorklistItem, len(sortedEdges))
	for i, e := range sortedEdges {
		worklist[i] = renderers.WorklistItem{
			ID:    fmt.Sprintf("%s-%s", e.From, e.To),
			Label: fmt.Sprintf("%s-%s:%d", e.From, e.To, e.Weight),
		}
	}

	// L2: Sort edges
	b.AddStep("L2", trace.ActionHighlight,
		fmt.Sprintf("Sorted all %d edges by weight in ascending order.", len(sortedEdges)),
		cloneGraphState("Edges sorted in ascending order of weight", worklistTitle, initDisplay, graphEdges, worklist),
		map[string]any{"sortedCount": len(sortedEdges)},
	)

	nodeIDs := make([]string, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}
	dsu := NewDSU(nodeIDs)
	mst := make([]WeightedEdge, 0, len(nodes)-1)
	totalWeight := 0

	// L3: Init DSU
	dsuDisplay := syncNodeDisplay(nodes, dsu, inMST, "", "")
	b.AddStep("L3", trace.ActionSet,
		"Initialize Disjoint Set Union (DSU) with each vertex in its own component.",
		cloneGraphState("DSU initialized: each vertex represents its own component", worklistTitle, dsuDisplay, graphEdges, worklist),
		map[string]any{
			"mstCount":    0,
			"totalWeight": 0,
			"components":  getComponentsMap(nodes, dsu),
		},
	)

	for i, edge := range sortedEdges {
		worklist[i].Label = fmt.Sprintf("▶ %s-%s:%d", edge.From, edge.To, edge.Weight)
		visitingDisplay := syncNodeDisplay(nodes, dsu, inMST, edge.From, edge.To)

		// L4: Inspect edge
		b.AddStep("L4", trace.ActionHighlight,
			fmt.Sprintf("Inspect edge (%s, %s) with weight %d.", edge.From, edge.To, edge.Weight),
			cloneGraphState(
				fmt.Sprintf("Inspecting edge (%s, %s, weight %d)", edge.From, edge.To, edge.Weight),
				worklistTitle,
				visitingDisplay,
				graphEdges,
				worklist,
			),
			map[string]any{
				"from":       edge.From,
				"to":         edge.To,
				"weight":     edge.Weight,
				"components": getComponentsMap(nodes, dsu),
			},
		)

		b.IncComparisons()
		rootFrom := dsu.Find(edge.From)
		rootTo := dsu.Find(edge.To)

		if rootFrom == rootTo {
			// L5: Cycle condition met
			b.AddStep("L5", trace.ActionCompare,
				fmt.Sprintf("Find(%s) == Find(%s) == %q: both endpoints in same component.", edge.From, edge.To, rootFrom),
				cloneGraphState(
					"Cycle detected: both endpoints in same component; reject edge",
					worklistTitle,
					visitingDisplay,
					graphEdges,
					worklist,
				),
				map[string]any{
					"cycle":      true,
					"rejected":   true,
					"root":       rootFrom,
					"from":       edge.From,
					"to":         edge.To,
					"weight":     edge.Weight,
					"components": getComponentsMap(nodes, dsu),
				},
			)

			// L6: Cycle rejection
			worklist[i].Label = fmt.Sprintf("✗ %s-%s:%d", edge.From, edge.To, edge.Weight)
			normalDisplay := syncNodeDisplay(nodes, dsu, inMST, "", "")
			b.AddStep("L6", trace.ActionHighlight,
				fmt.Sprintf("Cycle detected: both endpoints %s and %s in component %q; reject edge (%s, %s).",
					edge.From, edge.To, rootFrom, edge.From, edge.To),
				cloneGraphState(
					"Cycle detected: both endpoints in same component; reject edge",
					worklistTitle,
					normalDisplay,
					graphEdges,
					worklist,
				),
				map[string]any{
					"cycle":      true,
					"rejected":   true,
					"root":       rootFrom,
					"from":       edge.From,
					"to":         edge.To,
					"weight":     edge.Weight,
					"components": getComponentsMap(nodes, dsu),
				},
			)
			continue
		}

		// L5: Endpoints in different components
		b.AddStep("L5", trace.ActionCompare,
			fmt.Sprintf("Find(%s)=%q != Find(%s)=%q: different components, no cycle.", edge.From, rootFrom, edge.To, rootTo),
			cloneGraphState(
				fmt.Sprintf("Different components: %s != %s (no cycle)", rootFrom, rootTo),
				worklistTitle,
				visitingDisplay,
				graphEdges,
				worklist,
			),
			map[string]any{
				"cycle":      false,
				"rootFrom":   rootFrom,
				"rootTo":     rootTo,
				"from":       edge.From,
				"to":         edge.To,
				"weight":     edge.Weight,
				"components": getComponentsMap(nodes, dsu),
			},
		)

		// L8: DSU Union
		b.IncOperations()
		rankFrom := dsu.Rank(edge.From)
		rankTo := dsu.Rank(edge.To)
		dsu.Union(edge.From, edge.To)
		newRoot := dsu.Find(edge.From)
		unionDisplay := syncNodeDisplay(nodes, dsu, inMST, edge.From, edge.To)

		b.AddStep("L8", trace.ActionSet,
			fmt.Sprintf("Union component %q (rank %d) and component %q (rank %d): merged into [%s]. Endpoints had different roots, so no cycle forms.",
				rootFrom, rankFrom, rootTo, rankTo, newRoot),
			cloneGraphState(
				fmt.Sprintf("DSU merge: components %s and %s merged into [%s]", rootFrom, rootTo, newRoot),
				worklistTitle,
				unionDisplay,
				graphEdges,
				worklist,
			),
			map[string]any{
				"componentA": rootFrom,
				"componentB": rootTo,
				"newRoot":    newRoot,
				"components": getComponentsMap(nodes, dsu),
			},
		)

		// L9: Add edge to MST
		mst = append(mst, edge)
		totalWeight += edge.Weight
		inMST[edge.From] = true
		inMST[edge.To] = true
		markEdgeActive(graphEdges, edge.From, edge.To)
		worklist[i].Label = fmt.Sprintf("✓ %s-%s:%d", edge.From, edge.To, edge.Weight)
		acceptedDisplay := syncNodeDisplay(nodes, dsu, inMST, "", "")

		b.AddStep("L9", trace.ActionPush,
			fmt.Sprintf("Added edge (%s, %s) [weight %d] to MST. Total weight: %d.",
				edge.From, edge.To, edge.Weight, totalWeight),
			cloneGraphState(
				fmt.Sprintf("Accepted (%s-%s): weight %d added to MST", edge.From, edge.To, edge.Weight),
				worklistTitle,
				acceptedDisplay,
				graphEdges,
				worklist,
			),
			map[string]any{
				"mstCount":    len(mst),
				"totalWeight": totalWeight,
				"from":        edge.From,
				"to":          edge.To,
				"weight":      edge.Weight,
				"components":  getComponentsMap(nodes, dsu),
			},
		)

		if len(mst) == len(nodes)-1 {
			break
		}
	}

	finalDisplay := syncNodeDisplay(nodes, dsu, inMST, "", "")
	mstEdgeNames := make([]string, len(mst))
	for j, e := range mst {
		mstEdgeNames[j] = fmt.Sprintf("%s-%s", e.From, e.To)
	}

	// L11: Done
	b.AddStep("L11", trace.ActionDone,
		fmt.Sprintf("Kruskal's MST complete! %d edges with total weight %d: %s.",
			len(mst), totalWeight, strings.Join(mstEdgeNames, ", ")),
		cloneGraphState(
			fmt.Sprintf("MST complete: total weight = %d, edges: [%s]", totalWeight, strings.Join(mstEdgeNames, ", ")),
			worklistTitle,
			finalDisplay,
			graphEdges,
			worklist,
		),
		map[string]any{
			"totalWeight": totalWeight,
			"status":      "complete",
			"mstCount":    len(mst),
			"mstEdges":    mstEdgeNames,
			"components":  getComponentsMap(nodes, dsu),
		},
	)

	return b.Steps()
}

func cloneGraphState(
	caption, worklistLabel string,
	nodes []renderers.GraphNode,
	edges []renderers.GraphEdge,
	worklist []renderers.WorklistItem,
) renderers.GraphState {
	var clonedNodes []renderers.GraphNode
	if nodes != nil {
		clonedNodes = make([]renderers.GraphNode, len(nodes))
		copy(clonedNodes, nodes)
	}

	var clonedEdges []renderers.GraphEdge
	if edges != nil {
		clonedEdges = make([]renderers.GraphEdge, len(edges))
		copy(clonedEdges, edges)
	}

	var clonedWorklist []renderers.WorklistItem
	if worklist != nil {
		clonedWorklist = make([]renderers.WorklistItem, len(worklist))
		copy(clonedWorklist, worklist)
	}

	return renderers.GraphState{
		Caption:       caption,
		WorklistLabel: worklistLabel,
		Worklist:      clonedWorklist,
		Nodes:         clonedNodes,
		Edges:         clonedEdges,
	}
}

func syncNodeDisplay(base []renderers.GraphNode, dsu *DSU, inMST map[string]bool, visitingFrom, visitingTo string) []renderers.GraphNode {
	nodes := make([]renderers.GraphNode, len(base))
	for i, n := range base {
		st := "unvisited"
		if inMST != nil && inMST[n.ID] {
			st = "visited"
		}
		if n.ID == visitingFrom || n.ID == visitingTo {
			st = "visiting"
		}
		label := n.ID
		if dsu != nil {
			label = fmt.Sprintf("%s[%s]", n.ID, dsu.Find(n.ID))
		}
		nodes[i] = renderers.GraphNode{
			ID:    n.ID,
			Label: label,
			X:     n.X,
			Y:     n.Y,
			State: st,
		}
	}
	return nodes
}

func getComponentsMap(nodes []renderers.GraphNode, dsu *DSU) map[string]string {
	if dsu == nil {
		return nil
	}
	m := make(map[string]string, len(nodes))
	for _, n := range nodes {
		m[n.ID] = dsu.Find(n.ID)
	}
	return m
}

func markEdgeActive(edges []renderers.GraphEdge, u, v string) {
	for i := range edges {
		if (edges[i].From == u && edges[i].To == v) || (edges[i].From == v && edges[i].To == u) {
			edges[i].Active = true
			return
		}
	}
}
