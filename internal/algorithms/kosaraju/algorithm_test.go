package kosaraju

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func normalizeSCCs(sccs [][]string) [][]string {
	res := make([][]string, len(sccs))
	for i, c := range sccs {
		norm := make([]string, len(c))
		copy(norm, c)
		sort.Strings(norm)
		res[i] = norm
	}
	sort.Slice(res, func(i, j int) bool {
		if len(res[i]) == 0 || len(res[j]) == 0 {
			return len(res[i]) < len(res[j])
		}
		return res[i][0] < res[j][0]
	})
	return res
}

func TestFindSCCsCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []string
		edges    [][2]string
		expected [][]string
	}{
		{
			name:  "three components chain",
			nodes: []string{"A", "B", "C", "D", "E", "F", "G"},
			edges: [][2]string{
				{"A", "B"}, {"B", "C"}, {"C", "A"},
				{"C", "D"},
				{"D", "E"}, {"E", "D"},
				{"E", "F"},
				{"F", "G"}, {"G", "F"},
			},
			expected: [][]string{
				{"A", "B", "C"},
				{"D", "E"},
				{"F", "G"},
			},
		},
		{
			name:  "single component cycle",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{
				{"A", "B"}, {"B", "C"}, {"C", "A"},
			},
			expected: [][]string{
				{"A", "B", "C"},
			},
		},
		{
			name:  "dag all singletons",
			nodes: []string{"A", "B", "C"},
			edges: [][2]string{
				{"A", "B"}, {"B", "C"},
			},
			expected: [][]string{
				{"A"},
				{"B"},
				{"C"},
			},
		},
		{
			name:  "disconnected components",
			nodes: []string{"A", "B", "C", "D"},
			edges: [][2]string{
				{"A", "B"}, {"B", "A"},
				{"C", "D"}, {"D", "C"},
			},
			expected: [][]string{
				{"A", "B"},
				{"C", "D"},
			},
		},
		{
			name:     "empty graph",
			nodes:    []string{},
			edges:    [][2]string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSCCs(tt.nodes, tt.edges)
			normGot := normalizeSCCs(got)
			normExp := normalizeSCCs(tt.expected)
			if len(normGot) == 0 && len(normExp) == 0 {
				return
			}
			if !reflect.DeepEqual(normGot, normExp) {
				t.Fatalf("expected %v, got %v", normExp, normGot)
			}
		})
	}
}

func TestKosarajuTraceValidation(t *testing.T) {
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

func TestKosarajuTraceVariablesAndState(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	foundComponents := false
	for _, s := range steps {
		if _, ok := s.Variables["components"]; ok {
			foundComponents = true
			break
		}
	}
	if !foundComponents {
		t.Error("expected trace steps to contain 'components' in variables")
	}

	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.KosarajuLessonState)
	if !ok {
		t.Fatalf("expected KosarajuLessonState in last step, got %T", lastStep.State)
	}
	if len(state.Nodes) == 0 {
		t.Error("expected non-empty nodes in graph state")
	}
}

func TestKosarajuCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := LessonInput{
		Nodes: []renderers.GraphNode{
			{ID: "1", Label: "1", X: 100, Y: 100},
			{ID: "2", Label: "2", X: 200, Y: 100},
		},
		Edges: []renderers.GraphEdge{
			{From: "1", To: "2"},
			{From: "2", To: "1"},
		},
	}
	bytes, err := json.Marshal(customInput)
	if err != nil {
		t.Fatalf("failed to marshal custom input: %v", err)
	}
	steps, err := lesson.Build(bytes)
	if err != nil {
		t.Fatalf("build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}
}
