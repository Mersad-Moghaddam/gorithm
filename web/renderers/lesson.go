package renderers

import (
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func StepHoldMS(step trace.Step, speed float64) int {
	if speed <= 0 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		speed = 1
	}
	hold := step.HoldMS
	switch hold {
	case 350, 600, 1200, 1800:
	default:
		hold = 1800
	}
	ms := math.Round(float64(hold) / speed)
	if ms < 100 {
		return 100
	}
	// Saturate before conversion: tiny positive speeds can overflow an int.
	max := int(^uint(0) >> 1)
	if ms >= float64(max) {
		return max
	}
	return int(ms)
}

// RenderLessonSVG dispatches only to dedicated renderers that understand each
// lesson's concrete typed trace state. A generic fallback would silently drop
// the teaching information carried by that state.
func RenderLessonSVG(id string, step trace.Step, width, height int) (string, bool) {
	switch id {
	case "astar":
		return RenderAStarLessonSVG(step, width, height), true
	case "avl":
		return RenderAVLLessonSVG(step, width, height), true
	case "bellmanford":
		return RenderBellmanFordLessonSVG(step, width, height), true
	case "bfs":
		return RenderBFSLessonSVG(step, width, height), true
	case "binarysearch":
		return RenderBinarySearchLessonSVG(step, width, height), true
	case "bst":
		return RenderBSTLessonSVG(step, width, height), true
	case "bubble":
		return RenderBubbleLessonSVG(step, width, height), true
	case "bucketsort":
		return RenderBucketsortLessonSVG(step, width, height), true
	case "coinchange":
		return RenderCoinChangeLessonSVG(step, width, height), true
	case "countingsort":
		return RenderCountingsortLessonSVG(step, width, height), true
	case "dfs":
		return RenderDFSLessonSVG(step, width, height), true
	case "dijkstra":
		return RenderDijkstraLessonSVG(step, width, height), true
	case "fibonacci":
		return RenderFibonacciLessonSVG(step, width, height), true
	case "floydwarshall":
		return RenderFloydWarshallLessonSVG(step, width, height), true
	case "genetic":
		return RenderGeneticLessonSVG(step, width, height), true
	case "gradientdescent":
		return RenderGradientDescentLessonSVG(step, width, height), true
	case "hanoitower":
		return RenderHanoiTowerLessonSVG(step, width, height), true
	case "heap":
		return RenderHeapLessonSVG(step, width, height), true
	case "insertion":
		return RenderInsertionLessonSVG(step, width, height), true
	case "kmeans":
		return RenderKMeansLessonSVG(step, width, height), true
	case "kmp":
		return RenderKMPLessonSVG(step, width, height), true
	case "knapsack":
		return RenderKnapsackLessonSVG(step, width, height), true
	case "kosaraju":
		return RenderKosarajuLessonSVG(step, width, height), true
	case "kruskal":
		return RenderKruskalLessonSVG(step, width, height), true
	case "lcs":
		return RenderLCSLessonSVG(step, width, height), true
	case "linearsearch":
		return RenderLinearSearchLessonSVG(step, width, height), true
	case "merge":
		return RenderMergeLessonSVG(step, width, height), true
	case "minimax":
		return RenderMinimaxLessonSVG(step, width, height), true
	case "nqueens":
		return RenderNQueensLessonSVG(step, width, height), true
	case "perceptron":
		return RenderPerceptronLessonSVG(step, width, height), true
	case "prim":
		return RenderPrimLessonSVG(step, width, height), true
	case "qlearning":
		return RenderQLearningLessonSVG(step, width, height), true
	case "quick":
		return RenderQuickLessonSVG(step, width, height), true
	case "radixsort":
		return RenderRadixsortLessonSVG(step, width, height), true
	case "segmenttree":
		return RenderSegmentTreeLessonSVG(step, width, height), true
	case "selection":
		return RenderSelectionLessonSVG(step, width, height), true
	case "sudoku":
		return RenderSudokuLessonSVG(step, width, height), true
	case "toposort":
		return RenderToposortLessonSVG(step, width, height), true
	case "trie":
		return RenderTrieLessonSVG(step, width, height), true
	case "unionfind":
		return RenderUnionFindLessonSVG(step, width, height), true
	default:
		return "", false
	}
}

type LessonLegendEntry struct{ Label, Color string }

// LessonLegend describes the meaning of the dedicated scene's colors. Keep
// entries short enough to wrap on a phone; the same semantics are also stated
// in scene labels and text so color is never the only cue.
func LessonLegend(id string) []LessonLegendEntry {
	switch id {
	case "astar":
		return []LessonLegendEntry{
			{Label: "Item identity fill (stable per node)", Color: "#c45d75"},
			{Label: "Current / open node", Color: "#ffe29a"},
			{Label: "f = g + h estimate (not an edge)", Color: "#8bd5ff"},
			{Label: "Closed / path found ✓", Color: "#9bf0bb"},
			{Label: "No path / invalid input", Color: "#ff9b9b"},
		}
	case "avl":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Active comparison / balance check", Color: "#ffe29a"},
			{Label: "Rotation transfer or height label", Color: "#8bd5ff"},
			{Label: "Balanced / complete ✓", Color: "#9bf0bb"},
		}
	case "bellmanford":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Relaxed edge / improved distance", Color: "#ffe29a"},
			{Label: "Current source or scan cursor", Color: "#8bd5ff"},
			{Label: "Final distance / complete ✓", Color: "#9bf0bb"},
			{Label: "Negative cycle detected", Color: "#ff9b9b"},
		}
	case "bfs":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Neighbor discovery test", Color: "#ffe29a"},
			{Label: "Current dequeued node", Color: "#8bd5ff"},
			{Label: "Finished ✓ (distance final)", Color: "#9bf0bb"},
			{Label: "? unknown level", Color: "#bbcab9"},
		}
	case "binarysearch":
		return []LessonLegendEntry{
			{Label: "Array item identity fill", Color: "#c45d75"},
			{Label: "Midpoint comparison", Color: "#ffe29a"},
			{Label: "Active [low, high] range", Color: "#8bd5ff"},
			{Label: "Found / final ✓", Color: "#9bf0bb"},
			{Label: "Eliminated range", Color: "#ff9b9b"},
		}
	case "bst":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Compared key", Color: "#ffe29a"},
			{Label: "Current traversal pointer", Color: "#8bd5ff"},
			{Label: "Found / inserted ✓", Color: "#9bf0bb"},
		}
	case "bubble":
		return []LessonLegendEntry{
			{Label: "Original identity (#ID; color stays with value)", Color: "#c45d75"},
			{Label: "Adjacent comparison", Color: "#ffe29a"},
			{Label: "Final position ✓", Color: "#9bf0bb"},
		}
	case "bucketsort":
		return []LessonLegendEntry{
			{Label: "Item identity fill (survives copies)", Color: "#c45d75"},
			{Label: "Active bucket / input item", Color: "#ffe29a"},
			{Label: "Gather cursor", Color: "#8bd5ff"},
			{Label: "Confirmed output ✓", Color: "#9bf0bb"},
		}
	case "coinchange":
		return []LessonLegendEntry{
			{Label: "Coin identity fill", Color: "#c45d75"},
			{Label: "Coin considered / current amount", Color: "#ffe29a"},
			{Label: "DP dependency (reused result)", Color: "#8bd5ff"},
			{Label: "Reachable / final ✓", Color: "#9bf0bb"},
			{Label: "Unreachable amount", Color: "#ff9b9b"},
		}
	case "countingsort":
		return []LessonLegendEntry{
			{Label: "Item identity fill (stable copies)", Color: "#c45d75"},
			{Label: "Count increment / current item", Color: "#ffe29a"},
			{Label: "Count index / prefix total", Color: "#8bd5ff"},
			{Label: "Placed output ✓", Color: "#9bf0bb"},
		}
	case "dfs":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Neighbor inspection", Color: "#ffe29a"},
			{Label: "Current call cursor", Color: "#8bd5ff"},
			{Label: "Finished / check ✓", Color: "#9bf0bb"},
			{Label: "Call stack: newest unfinished on top", Color: "#bbcab9"},
		}
	case "dijkstra":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Frontier / current minimum", Color: "#ffe29a"},
			{Label: "Tentative distance relaxation", Color: "#8bd5ff"},
			{Label: "Finalized shortest distance ✓", Color: "#9bf0bb"},
			{Label: "Unreachable", Color: "#ff9b9b"},
		}
	case "fibonacci":
		return []LessonLegendEntry{
			{Label: "Indexed identity fill", Color: "#c45d75"},
			{Label: "Reused source (dashed witness)", Color: "#ffe29a"},
			{Label: "Final verified result ✓", Color: "#9bf0bb"},
			{Label: "? value not computed yet", Color: "#bbcab9"},
		}
	case "floydwarshall":
		return []LessonLegendEntry{
			{Label: "Node identity fill", Color: "#c45d75"},
			{Label: "Compared via intermediate k", Color: "#ffe29a"},
			{Label: "Improved DP dependency", Color: "#8bd5ff"},
			{Label: "Final shortest distance ✓", Color: "#9bf0bb"},
			{Label: "Unreachable pair", Color: "#ff9b9b"},
		}
	case "genetic":
		return []LessonLegendEntry{
			{Label: "Individual identity fill", Color: "#c45d75"},
			{Label: "Selected parent / fitness", Color: "#ffe29a"},
			{Label: "Crossover / mutation operation", Color: "#8bd5ff"},
			{Label: "Best individual / complete ✓", Color: "#9bf0bb"},
		}
	case "gradientdescent":
		return []LessonLegendEntry{
			{Label: "Current parameter / model", Color: "#ee8b9b"},
			{Label: "MSE slice", Color: "#aa92e7"},
			{Label: "Residual / tangent", Color: "#ffe29a"},
			{Label: "Discrete update target", Color: "#8bd5ff"},
			{Label: "Known fixed-b minimum", Color: "#9bf0bb"},
		}
	case "hanoitower":
		return []LessonLegendEntry{
			{Label: "Disk rank colors (coral → teal)", Color: "#c45d75"},
			{Label: "Disk moved this step", Color: "#ffe29a"},
			{Label: "Complete tower ✓", Color: "#9bf0bb"},
			{Label: "Move / from / to / aux labels", Color: "#bbcab9"},
		}
	case "heap":
		return []LessonLegendEntry{
			{Label: "Item identity fill (tree and array)", Color: "#c45d75"},
			{Label: "Candidate / comparison outline", Color: "#ffe29a"},
			{Label: "Root label", Color: "#8bd5ff"},
			{Label: "Final array slot ✓", Color: "#9bf0bb"},
			{Label: "Parent / child slot line", Color: "#bbcab9"},
		}
	case "insertion":
		return []LessonLegendEntry{
			{Label: "Original identity fill", Color: "#c45d75"},
			{Label: "Comparing / held key", Color: "#ffe29a"},
			{Label: "Conceptual hole / copy", Color: "#8bd5ff"},
			{Label: "Locally sorted prefix ✓", Color: "#9bf0bb"},
		}
	case "kmeans":
		return []LessonLegendEntry{
			{Label: "Sample identity fill", Color: "#c45d75"},
			{Label: "Current assignment / centroid", Color: "#ffe29a"},
			{Label: "Centroid movement", Color: "#8bd5ff"},
			{Label: "Converged cluster ✓", Color: "#9bf0bb"},
		}
	case "kmp":
		return []LessonLegendEntry{
			{Label: "Byte identity fill", Color: "#c45d75"},
			{Label: "Current comparison", Color: "#ffe29a"},
			{Label: "Failure-link fallback", Color: "#8bd5ff"},
			{Label: "Matched prefix / found ✓", Color: "#9bf0bb"},
			{Label: "Mismatch", Color: "#ff9b9b"},
		}
	case "knapsack":
		return []LessonLegendEntry{
			{Label: "Item identity fill", Color: "#c45d75"},
			{Label: "Include / exclude decision", Color: "#ffe29a"},
			{Label: "DP dependency", Color: "#8bd5ff"},
			{Label: "Chosen item / final value ✓", Color: "#9bf0bb"},
			{Label: "Infeasible cell", Color: "#ff9b9b"},
		}
	case "kosaraju":
		return []LessonLegendEntry{
			{Label: "Node identity fill (input index)", Color: "#c45d75"},
			{Label: "Current node / finish cursor", Color: "#ffe29a"},
			{Label: "SCC membership badge", Color: "#8bd5ff"},
			{Label: "Committed SCC ✓", Color: "#9bf0bb"},
			{Label: "Previous edge direction (L7/L16)", Color: "#bbcab9"},
		}
	case "kruskal":
		return []LessonLegendEntry{
			{Label: "Vertex identity fill", Color: "#c45d75"},
			{Label: "Candidate / compared roots", Color: "#ffe29a"},
			{Label: "Accepted edge / merged ✓", Color: "#9bf0bb"},
			{Label: "Rejected cycle edge ×", Color: "#ff9b9b"},
			{Label: "Component row", Color: "#8bd5ff"},
		}
	case "lcs":
		return []LessonLegendEntry{
			{Label: "Byte identity fill", Color: "#c45d75"},
			{Label: "Compared bytes / cursor", Color: "#ffe29a"},
			{Label: "DP dependency", Color: "#8bd5ff"},
			{Label: "Final verified sequence ✓", Color: "#9bf0bb"},
		}
	case "linearsearch":
		return []LessonLegendEntry{
			{Label: "Original array identity fill", Color: "#c45d75"},
			{Label: "Comparison outline", Color: "#ffe29a"},
			{Label: "Current index pointer", Color: "#8bd5ff"},
			{Label: "First match ✓", Color: "#9bf0bb"},
			{Label: "Rejected prefix (≠ target)", Color: "#ff9b9b"},
		}
	case "merge":
		return []LessonLegendEntry{
			{Label: "Identity fill (five fixed colors)", Color: "#c45d75"},
			{Label: "Comparison / half brackets", Color: "#ffe29a"},
			{Label: "Copy / aux / pointers", Color: "#8bd5ff"},
			{Label: "Global final / local merged ✓", Color: "#9bf0bb"},
		}
	case "minimax":
		return []LessonLegendEntry{
			{Label: "MAX triangle / MIN inverted triangle", Color: "#bbcab9"},
			{Label: "Active node / bounds", Color: "#ffe29a"},
			{Label: "Evaluated utility", Color: "#9bf0bb"},
			{Label: "Returned utility", Color: "#8bd5ff"},
			{Label: "Skipped search ×", Color: "#ff9b9b"},
		}
	case "nqueens":
		return []LessonLegendEntry{
			{Label: "Committed queen (row palette)", Color: "#c45d75"},
			{Label: "Candidate / conflict ray", Color: "#ffe29a"},
			{Label: "Solution recorded ✓", Color: "#9bf0bb"},
		}
	case "perceptron":
		return []LessonLegendEntry{
			{Label: "Sample identity fill", Color: "#c45d75"},
			{Label: "Misclassified / weight update", Color: "#ffe29a"},
			{Label: "Decision boundary", Color: "#8bd5ff"},
			{Label: "Correct / converged ✓", Color: "#9bf0bb"},
		}
	case "prim":
		return []LessonLegendEntry{
			{Label: "Input-order node identity fill", Color: "#c45d75"},
			{Label: "Mathematical cut (dashed)", Color: "#ffe29a"},
			{Label: "Accepted edge / tree check ✓", Color: "#9bf0bb"},
			{Label: "Obsolete heap candidate", Color: "#8bd5ff"},
			{Label: "Discarded cycle candidate", Color: "#ff9b9b"},
		}
	case "qlearning":
		return []LessonLegendEntry{
			{Label: "Agent", Color: "#ee8b9b"},
			{Label: "Selected action", Color: "#ffe29a"},
			{Label: "Transition preview", Color: "#8bd5ff"},
			{Label: "Tied candidate actions", Color: "#aa92e7"},
			{Label: "Start / goal ✓", Color: "#9bf0bb"},
			{Label: "Collision", Color: "#ff9b9b"},
		}
	case "quick":
		return []LessonLegendEntry{
			{Label: "Item identity fill", Color: "#c45d75"},
			{Label: "Comparison / pivot", Color: "#ffe29a"},
			{Label: "Active range", Color: "#8bd5ff"},
			{Label: "Final pivot position ✓", Color: "#9bf0bb"},
		}
	case "radixsort":
		return []LessonLegendEntry{
			{Label: "Item identity fill (stable copies)", Color: "#c45d75"},
			{Label: "Current digit / active item", Color: "#ffe29a"},
			{Label: "Scatter / gather bucket", Color: "#8bd5ff"},
			{Label: "Confirmed output ✓", Color: "#9bf0bb"},
		}
	case "segmenttree":
		return []LessonLegendEntry{
			{Label: "Array / node identity fill", Color: "#c45d75"},
			{Label: "Query range", Color: "#ffe29a"},
			{Label: "Node visited / combined", Color: "#8bd5ff"},
			{Label: "Included result ✓", Color: "#9bf0bb"},
		}
	case "selection":
		return []LessonLegendEntry{
			{Label: "Identity colors (coral → teal)", Color: "#c45d75"},
			{Label: "Comparison", Color: "#ffe29a"},
			{Label: "i / j / minimum flags", Color: "#8bd5ff"},
			{Label: "Sorted prefix ✓", Color: "#9bf0bb"},
		}
	case "sudoku":
		return []LessonLegendEntry{
			{Label: "Given (bold + dot)", Color: "#eff5ed"},
			{Label: "Guess / candidate", Color: "#c45d75"},
			{Label: "Row, column, box check", Color: "#ffe29a"},
			{Label: "Conflict ×", Color: "#ff9b9b"},
			{Label: "Verified solution ✓", Color: "#9bf0bb"},
		}
	case "toposort":
		return []LessonLegendEntry{
			{Label: "Task identity fill (input index)", Color: "#c45d75"},
			{Label: "Current dependency / task", Color: "#ffe29a"},
			{Label: "Resolved dependency", Color: "#8bd5ff"},
			{Label: "Ready FIFO / ordered check ✓", Color: "#9bf0bb"},
			{Label: "Blocked text (not cycle proof)", Color: "#ff9b9b"},
		}
	case "trie":
		return []LessonLegendEntry{
			{Label: "Word / character identity fill", Color: "#c45d75"},
			{Label: "Prefix traversal", Color: "#ffe29a"},
			{Label: "Current node / new edge", Color: "#8bd5ff"},
			{Label: "Inserted word / found ✓", Color: "#9bf0bb"},
			{Label: "Missing prefix ×", Color: "#ff9b9b"},
		}
	case "unionfind":
		return []LessonLegendEntry{
			{Label: "Element identity fill", Color: "#c45d75"},
			{Label: "Find traversal / compared root", Color: "#ffe29a"},
			{Label: "Parent pointer / union link", Color: "#8bd5ff"},
			{Label: "Same component / complete ✓", Color: "#9bf0bb"},
		}
	default:
		return nil
	}
}
