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
	caption string,
	activeRow, activeCol int,
	sourceRow1, sourceCol1 int,
	sourceRow2, sourceCol2 int,
	updatedRow, updatedCol int,
	deps []renderers.TableDependency,
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
			} else if r == activeRow && c == activeCol {
				state = "active"
			} else if (r == sourceRow1 && c == sourceCol1) || (r == sourceRow2 && c == sourceCol2) {
				state = "source"
			}

			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: formatDist(dist[r][c]),
				State: state,
			})
		}
	}

	var dependencies []renderers.TableDependency
	if len(deps) > 0 {
		dependencies = make([]renderers.TableDependency, len(deps))
		copy(dependencies, deps)
	}

	return renderers.TableState{
		Caption:      caption,
		Dependencies: dependencies,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
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
		"L2",
		trace.ActionInit,
		fmt.Sprintf("Initialize %dx%d distance matrix for %d nodes.", v, v, v),
		buildTableState(nodes, dist, fmt.Sprintf("Initialize %dx%d distance matrix", v, v), -1, -1, -1, -1, -1, -1, -1, -1, nil),
		map[string]any{"k": -1, "i": -1, "j": -1, "nodes": nodes},
	)

	for k := 0; k < v; k++ {
		b.AddStep(
			"L3",
			trace.ActionVisit,
			fmt.Sprintf("Evaluate node %s (k=%d) as intermediate vertex.", nodes[k], k),
			buildTableState(nodes, dist, fmt.Sprintf("Intermediate vertex k=%s", nodes[k]), -1, -1, -1, -1, -1, -1, -1, -1, nil),
			map[string]any{"k": k, "i": -1, "j": -1, "intermediate": nodes[k]},
		)

		for i := 0; i < v; i++ {
			for j := 0; j < v; j++ {
				pairDeps := []renderers.TableDependency{
					{FromRow: i, FromCol: k, ToRow: i, ToCol: j},
					{FromRow: k, FromCol: j, ToRow: i, ToCol: j},
				}

				b.IncComparisons()
				if dist[i][k] == math.MaxInt || dist[k][j] == math.MaxInt {
					unreachDesc := fmt.Sprintf("Checking path %s -> %s via %s: intermediate path unreachable.", nodes[i], nodes[j], nodes[k])
					unreachCaption := fmt.Sprintf("Intermediate k=%s: path %s -> %s via %s unreachable", nodes[k], nodes[i], nodes[j], nodes[k])
					b.AddStep(
						"L6",
						trace.ActionCompare,
						unreachDesc,
						buildTableState(nodes, dist, unreachCaption, i, j, i, k, k, j, -1, -1, pairDeps),
						map[string]any{"k": k, "i": i, "j": j, "intermediate": nodes[k], "from": nodes[i], "to": nodes[j], "unreachable": true},
					)
					continue
				}

				alt := dist[i][k] + dist[k][j]
				oldDist := dist[i][j]
				oldDistStr := formatDist(oldDist)
				compareDesc := fmt.Sprintf("Checking %s -> %s via %s: dist[%s][%s](%d) + dist[%s][%s](%d) = %d vs current %s.",
					nodes[i], nodes[j], nodes[k], nodes[i], nodes[k], dist[i][k], nodes[k], nodes[j], dist[k][j], alt, oldDistStr)
				compareCaption := fmt.Sprintf("Intermediate k=%s: checking %s -> %s: dist[%s][%s](%d) + dist[%s][%s](%d) = %d vs current %s",
					nodes[k], nodes[i], nodes[j], nodes[i], nodes[k], dist[i][k], nodes[k], nodes[j], dist[k][j], alt, oldDistStr)

				b.AddStep(
					"L7",
					trace.ActionCompare,
					compareDesc,
					buildTableState(nodes, dist, compareCaption, i, j, i, k, k, j, -1, -1, pairDeps),
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
					relaxCaption := fmt.Sprintf("Path through intermediate k=%s is shorter: %d + %d = %d < %d", nodes[k], dist[i][k], dist[k][j], alt, oldDist)
					b.AddStep(
						"L8",
						trace.ActionSet,
						relaxCaption,
						buildTableState(nodes, dist, relaxCaption, -1, -1, i, k, k, j, i, j, pairDeps),
						map[string]any{
							"k":            k,
							"i":            i,
							"j":            j,
							"intermediate": nodes[k],
							"from":         nodes[i],
							"to":           nodes[j],
							"newDist":      alt,
							"oldDist":      oldDist,
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
		buildTableState(nodes, dist, "Scanning diagonal for negative cycles (dist[i][i] < 0)", -1, -1, -1, -1, -1, -1, -1, -1, nil),
		map[string]any{"k": -1, "i": -1, "j": -1},
	)

	hasNegCycle := false
	for i := 0; i < v; i++ {
		b.IncComparisons()
		if dist[i][i] < 0 {
			diagCaption := fmt.Sprintf("Negative cycle detected! dist[%s][%s] = %d < 0.", nodes[i], nodes[i], dist[i][i])
			b.AddStep(
				"L16",
				trace.ActionHighlight,
				diagCaption,
				buildTableState(nodes, dist, fmt.Sprintf("Negative cycle detected: dist[%s][%s] = %d < 0", nodes[i], nodes[i], dist[i][i]), i, i, -1, -1, -1, -1, -1, -1, nil),
				map[string]any{"k": -1, "i": i, "j": i, "negativeCycle": true, "cycleVertex": nodes[i]},
			)
			hasNegCycle = true
			break
		}
	}

	if !hasNegCycle {
		doneCaption := "Floyd-Warshall completed! All-pairs shortest path matrix computed with no negative cycles."
		b.AddStep(
			"L19",
			trace.ActionDone,
			doneCaption,
			buildTableState(nodes, dist, doneCaption, -1, -1, -1, -1, -1, -1, -1, -1, nil),
			map[string]any{"k": -1, "i": -1, "j": -1, "negativeCycle": false},
		)
	}

	return b.Steps()
}
