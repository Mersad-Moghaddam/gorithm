package dfs

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestTraverseCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		adj      map[string][]string
		start    string
		expected []string
	}{
		{
			name: "canonical tree-like graph",
			adj: map[string][]string{
				"A": {"B", "C"},
				"B": {"D", "E"},
				"C": {"E", "F"},
				"D": {},
				"E": {},
				"F": {},
			},
			start:    "A",
			expected: []string{"A", "B", "D", "E", "C", "F"},
		},
		{
			name: "graph with cycle",
			adj: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			start:    "A",
			expected: []string{"A", "B", "C"},
		},
		{
			name: "single node",
			adj: map[string][]string{
				"A": {},
			},
			start:    "A",
			expected: []string{"A"},
		},
		{
			name:     "empty start",
			adj:      map[string][]string{"A": {"B"}},
			start:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Traverse(tt.adj, tt.start)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestDFSTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}

func TestDFSTraceExecution(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	// Verify all nodes are marked visited in final step
	finalStep := steps[len(steps)-1]
	state, ok := finalStep.State.(renderers.DFSLessonState)
	if !ok {
		t.Fatalf("expected DFSLessonState, got %T", finalStep.State)
	}

	for _, node := range state.Nodes {
		if node.State != "visited" {
			t.Errorf("expected node %s to be visited, got %q", node.ID, node.State)
		}
	}

	for _, edge := range state.Edges {
		if edge.Active {
			t.Errorf("expected edge (%s -> %s) to be inactive at end", edge.From, edge.To)
		}
	}
}

func TestDFSCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := json.RawMessage(`"B"`)
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if steps[0].Variables["start"] != "B" {
		t.Fatalf("expected start node B, got %v", steps[0].Variables["start"])
	}
}
