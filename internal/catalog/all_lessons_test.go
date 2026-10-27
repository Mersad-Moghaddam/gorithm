package catalog_test

import (
	"encoding/json"
	"strings"
	"testing"

	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/astar"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/avl"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bellmanford"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/binarysearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bst"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bubble"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bucketsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/coinchange"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/countingsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dijkstra"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/fibonacci"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/floydwarshall"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/genetic"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/gradientdescent"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/hanoitower"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/heap"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/insertion"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmeans"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmp"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/knapsack"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kosaraju"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kruskal"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/lcs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/linearsearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/merge"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/minimax"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/nqueens"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/perceptron"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/prim"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/qlearning"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/quick"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/radixsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/segmenttree"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/selection"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/sudoku"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/toposort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/trie"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/unionfind"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestAllFortyAlgorithmsCatalog(t *testing.T) {
	metas := catalog.List()
	// Filter out test dummy if present
	var realMetas []catalog.Meta
	for _, m := range metas {
		if m.ID != "dummy" {
			realMetas = append(realMetas, m)
		}
	}
	if len(realMetas) != 40 {
		t.Fatalf("expected 40 algorithms registered in catalog, got %d", len(realMetas))
	}

	for _, meta := range realMetas {
		meta := meta
		t.Run(meta.ID, func(t *testing.T) {
			lesson, ok := catalog.Get(meta.ID)
			if !ok {
				t.Fatalf("algorithm %s not found in catalog", meta.ID)
			}

			// Validate DefaultInput builds valid trace
			steps, err := lesson.Build(json.RawMessage(meta.DefaultInput))
			if err != nil {
				t.Fatalf("lesson.Build(defaultInput) error: %v", err)
			}
			if len(steps) == 0 {
				t.Fatalf("lesson %s produced 0 steps from default input", meta.ID)
			}

			// Validate all line IDs in trace match source
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("source line validation error: %v", err)
			}

			// Render every step to SVG and check for invalid numeric values or anomalies
			for i, step := range steps {
				svg := renderers.RenderStepSVG(step, meta.Renderer, 800, 500)
				if strings.TrimSpace(svg) == "" {
					t.Errorf("step %d produced empty SVG output", i)
				}
				if strings.Contains(svg, "NaN") {
					t.Errorf("step %d SVG contains NaN", i)
				}
				if strings.Contains(svg, "Infinity") {
					t.Errorf("step %d SVG contains Infinity", i)
				}
			}
		})
	}
}
