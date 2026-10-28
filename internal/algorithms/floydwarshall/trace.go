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

	var current renderers.FloydWarshallLessonState
	current.I, current.J, current.K, current.DiagonalIndex = -1, -1, -1, -1
	snapshot := func() renderers.FloydWarshallLessonState {
		s := current
		s.Nodes = append([]string{}, nodes...)
		s.Edges = make([]renderers.GraphEdge, len(edges))
		for i, e := range edges {
			s.Edges[i] = renderers.GraphEdge{From: e.From, To: e.To, Weight: e.Weight, Directed: e.Directed}
		}
		s.Distance = make([][]int, v)
		for i := range dist {
			s.Distance[i] = append([]int{}, dist[i]...)
		}
		s.TableState = buildTableState(nodes, dist, s.I, s.J, s.I, s.K, s.K, s.J, -1, -1)
		return s
	}

	b.AddStep(
		"L1",
		trace.ActionInit,
		fmt.Sprintf("Initialize %dx%d distance matrix for %d nodes.", v, v, v),
		snapshot(),
		map[string]any{"k": -1, "i": -1, "j": -1, "nodes": append([]string{}, nodes...)},
	)

	for k := 0; k < v; k++ {
		current.K, current.I, current.J = k, -1, -1
		current.HasCandidate, current.Improved = false, false
		b.AddStep(
			"L3",
			trace.ActionVisit,
			fmt.Sprintf("Evaluate node %s (k=%d) as intermediate vertex.", nodes[k], k),
			snapshot(),
			map[string]any{"k": k, "i": -1, "j": -1, "intermediate": nodes[k]},
		)

		for i := 0; i < v; i++ {
			for j := 0; j < v; j++ {
				current.I, current.J = i, j
				current.LeftDistance, current.RightDistance, current.OldDistance = dist[i][k], dist[k][j], dist[i][j]
				current.Candidate, current.HasCandidate, current.Improved = 0, false, false
				b.IncComparisons()
				if dist[i][k] == math.MaxInt || dist[k][j] == math.MaxInt {
					b.AddStep(
						"L6",
						trace.ActionCompare,
						fmt.Sprintf("Checking path %s -> %s via %s: intermediate path unreachable.", nodes[i], nodes[j], nodes[k]),
						snapshot(),
						map[string]any{"k": k, "i": i, "j": j, "intermediate": nodes[k], "from": nodes[i], "to": nodes[j]},
					)
					continue
				}

				alt := dist[i][k] + dist[k][j]
				current.Candidate, current.HasCandidate = alt, true
				oldDistStr := formatDist(dist[i][j])
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("Checking %s -> %s via %s: dist[%s][%s](%d) + dist[%s][%s](%d) = %d vs current %s.",
						nodes[i], nodes[j], nodes[k], nodes[i], nodes[k], dist[i][k], nodes[k], nodes[j], dist[k][j], alt, oldDistStr),
					snapshot(),
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
					current.Improved = true
					b.IncOperations()
					b.AddStep(
						"L8",
						trace.ActionSet,
						fmt.Sprintf("Relaxed edge! Updated shortest path %s -> %s via %s to %d.", nodes[i], nodes[j], nodes[k], alt),
						snapshot(),
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

	current.I, current.J, current.K = -1, -1, -1
	current.HasCandidate, current.Improved, current.CheckingDiagonal = false, false, true
	b.AddStep(
		"L14",
		trace.ActionHighlight,
		"Scanning diagonal entries to check for negative cycles.",
		snapshot(),
		map[string]any{"k": -1, "i": -1, "j": -1},
	)

	hasNegCycle := false
	for i := 0; i < v; i++ {
		b.IncComparisons()
		if dist[i][i] < 0 {
			current.I, current.J, current.DiagonalIndex, current.NegativeCycle = i, i, i, true
			b.AddStep(
				"L16",
				trace.ActionHighlight,
				fmt.Sprintf("Negative cycle detected! dist[%s][%s] = %d < 0.", nodes[i], nodes[i], dist[i][i]),
				snapshot(),
				map[string]any{"k": -1, "i": i, "j": i, "negativeCycle": true},
			)
			hasNegCycle = true
			break
		}
	}

	if !hasNegCycle {
		current.CheckingDiagonal = false
		b.AddStep(
			"L19",
			trace.ActionDone,
			"Floyd-Warshall completed! All-pairs shortest path matrix computed with no negative cycles.",
			snapshot(),
			map[string]any{"k": -1, "i": -1, "j": -1, "negativeCycle": false},
		)
	}

	steps := b.Steps()
	for i := 1; i < len(steps); i++ {
		if steps[i].LineID == "L8" && steps[i-1].LineID == "L7" {
			steps[i-1].HoldMS = 600
			steps[i].HoldMS = 1200
		}
	}
	return steps
}
