package floydwarshall

import (
	"fmt"
	"math"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func formatDist(d int) string {
	if d == math.MaxInt {
		return "∞"
	}
	return strconv.Itoa(d)
}

func buildTableState(
	nodes []string,
	dist [][]int,
	activeRow, activeCol int,
	sourceRow1, sourceCol1 int,
	sourceRow2, sourceCol2 int,
	updatedRow, updatedCol int,
) renderers.TableState {
	v := len(nodes)
	rowHeaders := make([]string, v)
	colHeaders := make([]string, v)
	copy(rowHeaders, nodes)
	copy(colHeaders, nodes)

	cells := make([]renderers.TableCell, 0, v*v)
	for r := 0; r < v; r++ {
		for c := 0; c < v; c++ {
			state := "normal"
			if r == updatedRow && c == updatedCol {
				state = "computed"
			} else if (r == sourceRow1 && c == sourceCol1) || (r == sourceRow2 && c == sourceCol2) {
				state = "source"
			} else if r == activeRow && c == activeCol {
				state = "active"
			}

			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: formatDist(dist[r][c]),
				State: state,
			})
		}
	}

	return renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace executes Floyd-Warshall recording each step with distance matrix table state.
// ponytail: quadratic matrix table state per step; compress or delta-encode if V > 50.
func Trace(nodes []string, edges []WeightedEdge) []trace.Step {
	b := trace.NewBuilder()
	v := len(nodes)

	dist := make([][]int, v)
	for i := range dist {
		dist[i] = make([]int, v)
		for j := range dist[i] {
			if i == j {
				dist[i][j] = 0
			} else {
				dist[i][j] = math.MaxInt
			}
		}
	}

	nodeMap := make(map[string]int, v)
	for i, n := range nodes {
		nodeMap[n] = i
	}

	for _, e := range edges {
		u, uOk := nodeMap[e.From]
		to, toOk := nodeMap[e.To]
		if !uOk || !toOk {
			continue
		}
		if e.Weight < dist[u][to] {
			dist[u][to] = e.Weight
		}
		if !e.Directed && e.Weight < dist[to][u] {
			dist[to][u] = e.Weight
		}
	}

	b.AddStep(
		"L1",
		trace.ActionInit,
		fmt.Sprintf("Initialize %dx%d distance matrix for %d nodes.", v, v, v),
		buildTableState(nodes, dist, -1, -1, -1, -1, -1, -1, -1, -1),
		map[string]any{"k": -1, "i": -1, "j": -1, "nodes": nodes},
	)

	for k := 0; k < v; k++ {
		b.AddStep(
			"L3",
			trace.ActionVisit,
			fmt.Sprintf("Evaluate node %s (k=%d) as intermediate vertex.", nodes[k], k),
			buildTableState(nodes, dist, -1, -1, -1, -1, -1, -1, -1, -1),
			map[string]any{"k": k, "i": -1, "j": -1, "intermediate": nodes[k]},
		)

		for i := 0; i < v; i++ {
			for j := 0; j < v; j++ {
				b.IncComparisons()
				if dist[i][k] == math.MaxInt || dist[k][j] == math.MaxInt {
					b.AddStep(
						"L6",
						trace.ActionCompare,
						fmt.Sprintf("Checking path %s -> %s via %s: intermediate path unreachable.", nodes[i], nodes[j], nodes[k]),
						buildTableState(nodes, dist, i, j, -1, -1, -1, -1, -1, -1),
						map[string]any{"k": k, "i": i, "j": j, "intermediate": nodes[k], "from": nodes[i], "to": nodes[j]},
					)
					continue
				}

				alt := dist[i][k] + dist[k][j]
				oldDistStr := formatDist(dist[i][j])
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("Checking %s -> %s via %s: dist[%s][%s](%d) + dist[%s][%s](%d) = %d vs current %s.",
						nodes[i], nodes[j], nodes[k], nodes[i], nodes[k], dist[i][k], nodes[k], nodes[j], dist[k][j], alt, oldDistStr),
					buildTableState(nodes, dist, i, j, i, k, k, j, -1, -1),
					map[string]any{
						"k":            k,
						"i":            i,
						"j":            j,
						"intermediate": nodes[k],
						"from":         nodes[i],
						"to":           nodes[j],
						"alt":          alt,
						"current":      oldDistStr,
					},
				)

				if alt < dist[i][j] {
					dist[i][j] = alt
					b.IncOperations()
					b.AddStep(
						"L8",
						trace.ActionSet,
						fmt.Sprintf("Relaxed edge! Updated shortest path %s -> %s via %s to %d.", nodes[i], nodes[j], nodes[k], alt),
						buildTableState(nodes, dist, -1, -1, i, k, k, j, i, j),
						map[string]any{
							"k":            k,
							"i":            i,
							"j":            j,
							"intermediate": nodes[k],
							"from":         nodes[i],
							"to":           nodes[j],
							"newDist":      alt,
						},
					)
				}
			}
		}
	}

	b.AddStep(
		"L14",
		trace.ActionHighlight,
		"Scanning diagonal entries to check for negative cycles.",
		buildTableState(nodes, dist, -1, -1, -1, -1, -1, -1, -1, -1),
		map[string]any{"k": -1, "i": -1, "j": -1},
	)

	hasNegCycle := false
	for i := 0; i < v; i++ {
		b.IncComparisons()
		if dist[i][i] < 0 {
			b.AddStep(
				"L16",
				trace.ActionHighlight,
				fmt.Sprintf("Negative cycle detected! dist[%s][%s] = %d < 0.", nodes[i], nodes[i], dist[i][i]),
				buildTableState(nodes, dist, i, i, -1, -1, -1, -1, -1, -1),
				map[string]any{"k": -1, "i": i, "j": i, "negativeCycle": true},
			)
			hasNegCycle = true
			break
		}
	}

	if !hasNegCycle {
		b.AddStep(
			"L19",
			trace.ActionDone,
			"Floyd-Warshall completed! All-pairs shortest path matrix computed with no negative cycles.",
			buildTableState(nodes, dist, -1, -1, -1, -1, -1, -1, -1, -1),
			map[string]any{"k": -1, "i": -1, "j": -1, "negativeCycle": false},
		)
	}

	return b.Steps()
}
