package renderers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestRenderTreeSVGDynamicSizing(t *testing.T) {
	// Case 1: Default (small node count, shallow depth)
	stepDefault := trace.Step{
		State: TreeState{
			Nodes: []TreeNode{
				{ID: "0", Val: 10, X: 400, Y: 60, ParentID: ""},
				{ID: "1", Val: 5, X: 300, Y: 130, ParentID: "0"},
				{ID: "2", Val: 15, X: 500, Y: 130, ParentID: "0"},
			},
		},
	}
	svgDefault := RenderTreeSVG(stepDefault, 800, 450)
	if !strings.Contains(svgDefault, `r="22.0"`) {
		t.Errorf("expected r=22.0 in default tree, got %s", svgDefault)
	}
	if !strings.Contains(svgDefault, `font-size="13.0"`) {
		t.Errorf("expected font-size=13.0 in default tree, got %s", svgDefault)
	}

	// Case 2: Tree depth >= 4 (nodes at depth 0, 1, 2, 3, 4)
	var nodesDepth4 []TreeNode
	for d := 0; d <= 4; d++ {
		parentID := ""
		if d > 0 {
			parentID = fmt.Sprintf("%d", d-1)
		}
		nodesDepth4 = append(nodesDepth4, TreeNode{
			ID:       fmt.Sprintf("%d", d),
			Val:      d,
			X:        400,
			Y:        float64(60 + d*70),
			ParentID: parentID,
		})
	}
	svgDepth4 := RenderTreeSVG(trace.Step{State: TreeState{Nodes: nodesDepth4}}, 800, 450)
	if !strings.Contains(svgDepth4, `r="16.0"`) {
		t.Errorf("expected r=16.0 for depth >= 4, got %s", svgDepth4)
	}
	if !strings.Contains(svgDepth4, `font-size="10.0"`) {
		t.Errorf("expected font-size=10.0 for depth >= 4, got %s", svgDepth4)
	}

	// Case 3: Node count > 15 with shallow depth
	var nodesCount16 []TreeNode
	nodesCount16 = append(nodesCount16, TreeNode{ID: "0", Val: 0, X: 400, Y: 60, ParentID: ""})
	for i := 1; i <= 16; i++ {
		nodesCount16 = append(nodesCount16, TreeNode{
			ID:       fmt.Sprintf("%d", i),
			Val:      i,
			X:        float64(50 + i*40),
			Y:        130,
			ParentID: "0",
		})
	}
	svgCount16 := RenderTreeSVG(trace.Step{State: TreeState{Nodes: nodesCount16}}, 800, 450)
	if !strings.Contains(svgCount16, `r="16.0"`) {
		t.Errorf("expected r=16.0 for count > 15, got %s", svgCount16)
	}
	if !strings.Contains(svgCount16, `font-size="10.0"`) {
		t.Errorf("expected font-size=10.0 for count > 15, got %s", svgCount16)
	}

	// Case 4: Tree depth >= 5
	var nodesDepth5 []TreeNode
	for d := 0; d <= 5; d++ {
		parentID := ""
		if d > 0 {
			parentID = fmt.Sprintf("%d", d-1)
		}
		nodesDepth5 = append(nodesDepth5, TreeNode{
			ID:       fmt.Sprintf("%d", d),
			Val:      d,
			X:        400,
			Y:        float64(60 + d*70),
			ParentID: parentID,
		})
	}
	svgDepth5 := RenderTreeSVG(trace.Step{State: TreeState{Nodes: nodesDepth5}}, 800, 450)
	if !strings.Contains(svgDepth5, `r="13.0"`) {
		t.Errorf("expected r=13.0 for depth >= 5, got %s", svgDepth5)
	}
	if !strings.Contains(svgDepth5, `font-size="8.5"`) {
		t.Errorf("expected font-size=8.5 for depth >= 5, got %s", svgDepth5)
	}
}

func TestRenderTableSVGDynamicFontSize(t *testing.T) {
	// Case 1: Default large cells (cellW >= 36, cellH >= 28)
	stepDefault := trace.Step{
		State: TableState{
			RowHeaders: []string{"R0", "R1"},
			ColHeaders: []string{"C0", "C1"},
			Cells: []TableCell{
				{Row: 0, Col: 0, Value: "42", State: "normal"},
			},
		},
	}
	svgDefault := RenderTableSVG(stepDefault, 800, 450)
	if !strings.Contains(svgDefault, `font-size="12.0"`) {
		t.Errorf("expected font-size=12.0 for default table, got %s", svgDefault)
	}

	// Case 2: Medium cells (cellW < 36.0 || cellH < 28.0)
	// (width - startX - 30) / numCols = (800 - 100 - 30) / 22 = 670 / 22 ≈ 30.45 (< 36.0, >= 26.0)
	cols22 := make([]string, 22)
	for i := range cols22 {
		cols22[i] = fmt.Sprintf("%d", i)
	}
	stepMedium := trace.Step{
		State: TableState{
			RowHeaders: []string{"R0"},
			ColHeaders: cols22,
			Cells: []TableCell{
				{Row: 0, Col: 0, Value: "99", State: "normal"},
			},
		},
	}
	svgMedium := RenderTableSVG(stepMedium, 800, 450)
	if !strings.Contains(svgMedium, `font-size="10.0"`) {
		t.Errorf("expected font-size=10.0 for medium cells, got %s", svgMedium)
	}

	// Case 3: Small cells (cellW < 26.0 || cellH < 22.0)
	// (width - startX - 30) / numCols = 670 / 30 ≈ 22.3 (< 26.0)
	cols30 := make([]string, 30)
	for i := range cols30 {
		cols30[i] = fmt.Sprintf("%d", i)
	}
	stepSmall := trace.Step{
		State: TableState{
			RowHeaders: []string{"R0"},
			ColHeaders: cols30,
			Cells: []TableCell{
				{Row: 0, Col: 0, Value: "7", State: "normal"},
			},
		},
	}
	svgSmall := RenderTableSVG(stepSmall, 800, 450)
	if !strings.Contains(svgSmall, `font-size="8.5"`) {
		t.Errorf("expected font-size=8.5 for small cells, got %s", svgSmall)
	}
}
