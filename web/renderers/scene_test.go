package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"strings"
	"testing"
)

type testScene struct{}

func (testScene) RenderSVG(w, h int) string { return fmt.Sprintf("<text>%d × %d</text>", w, h) }

func TestSceneDispatch(t *testing.T) {
	if got := RenderStepSVG(trace.Step{State: testScene{}}, "array", 320, 180); got != "<text>320 × 180</text>" {
		t.Fatal(got)
	}
	for _, renderer := range []string{"array", "unknown", "graph", "tree", "table", "hanoi"} {
		if got := RenderStepSVG(trace.Step{}, renderer, 800, 450); !strings.Contains(got, "available") {
			t.Fatalf("%s: %s", renderer, got)
		}
	}
}

func TestEqualValuesKeepExplicitIdentities(t *testing.T) {
	state := ArrayState{Values: []int{4, 4}, IDs: []string{"first", "second"}, Current: -1}
	a := RenderArraySVG(trace.Step{State: state}, 800, 450)
	state.IDs = []string{"second", "first"}
	b := RenderArraySVG(trace.Step{State: state}, 800, 450)
	for _, key := range []string{`data-key="array-first"`, `data-key="array-second"`} {
		if strings.Count(a, key) != 1 || strings.Count(b, key) != 1 {
			t.Fatal("missing unique identity", key)
		}
	}
	if strings.Index(a, "array-first") >= strings.Index(a, "array-second") || strings.Index(b, "array-second") >= strings.Index(b, "array-first") {
		t.Fatal("identities must move with values")
	}
}

func TestTeachingLabelsEscapeMarkup(t *testing.T) {
	label := `<script> & "label"`
	states := []struct {
		renderer string
		state    any
	}{
		{"graph", GraphState{Nodes: []GraphNode{{ID: label, Label: label}}, Caption: label, WorklistLabel: label, Worklist: []WorklistItem{{ID: "a", Label: label}}}},
		{"tree", TreeState{Nodes: []TreeNode{{ID: label, Label: label}}}},
		{"table", TableState{RowHeaders: []string{label}, ColHeaders: []string{label}, Cells: []TableCell{{Value: label}}, Caption: label}},
	}
	for _, tt := range states {
		got := RenderStepSVG(trace.Step{State: tt.state}, tt.renderer, 800, 450)
		if strings.Contains(got, "<script>") || !strings.Contains(got, "&lt;script&gt;") {
			t.Fatalf("unsafe %s: %s", tt.renderer, got)
		}
	}
}

func TestTableDependencies(t *testing.T) {
	state := TableState{RowHeaders: []string{"r"}, ColHeaders: []string{"0", "1"}, Dependencies: []TableDependency{{FromRow: 0, FromCol: 0, ToRow: 0, ToCol: 1}}}
	got := RenderTableSVG(trace.Step{State: state}, 800, 450)
	if !strings.Contains(got, `marker-end="url(#dependency-arrow)"`) {
		t.Fatal("missing directed dependency", got)
	}
}
