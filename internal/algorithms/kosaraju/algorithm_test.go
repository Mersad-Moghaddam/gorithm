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
			name:     "single node graph",
			nodes:    []string{"A"},
			edges:    [][2]string{},
			expected: [][]string{{"A"}},
		},
		{
			name:  "fully connected graph",
			nodes: []string{"A", "B", "C", "D"},
			edges: [][2]string{
				{"A", "B"}, {"B", "A"},
				{"A", "C"}, {"C", "A"},
				{"A", "D"}, {"D", "A"},
				{"B", "C"}, {"C", "B"},
				{"B", "D"}, {"D", "B"},
				{"C", "D"}, {"D", "C"},
			},
			expected: [][]string{
				{"A", "B", "C", "D"},
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

func TestKosarajuInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid inputs
	if _, err := lesson.Build(nil); err != nil {
		t.Fatalf("expected nil input to succeed, got: %v", err)
	}
	if _, err := lesson.Build([]byte("")); err != nil {
		t.Fatalf("expected empty input to succeed, got: %v", err)
	}
	if _, err := lesson.Build([]byte("   ")); err != nil {
		t.Fatalf("expected whitespace input to succeed, got: %v", err)
	}

	// Invalid inputs
	invalidInputs := []struct {
		name  string
		input string
	}{
		{"null literal", "null"},
		{"whitespace null", "  null  "},
		{"malformed JSON", "{not valid json"},
		{"empty JSON object", "{}"},
		{"empty nodes array", `{"nodes":[]}`},
		{"node with empty ID", `{"nodes":[{"id":""}]}`},
		{"duplicate node ID", `{"nodes":[{"id":"A"},{"id":"A"}]}`},
		{"edge references nonexistent node", `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"B"}]}`},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build([]byte(tc.input))
			if err == nil {
				t.Fatalf("expected error for input %q, got nil error and %d steps", tc.input, len(steps))
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

func TestKosarajuWorklist(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	foundFinishOrder := false
	maxWorklistLen := 0
	finalWorklistLen := -1

	for _, s := range steps {
		state, ok := s.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("expected step state to be GraphState, got %T", s.State)
		}

		if state.WorklistLabel != "Finish Stack (LIFO)" {
			t.Errorf("expected WorklistLabel 'Finish Stack (LIFO)', got %q", state.WorklistLabel)
		}

		if len(state.Worklist) > maxWorklistLen {
			maxWorklistLen = len(state.Worklist)
		}

		// When Pass 1 completes (L6), finish stack should have all 7 nodes
		if s.LineID == "L6" {
			foundFinishOrder = true
			if len(state.Worklist) != 7 {
				t.Fatalf("expected 7 items in finish stack at L6, got %d", len(state.Worklist))
			}
		}

		if s.LineID == "L17" {
			finalWorklistLen = len(state.Worklist)
		}
	}

	if !foundFinishOrder {
		t.Error("expected to hit step L6 marking finish order completion")
	}
	if maxWorklistLen != 7 {
		t.Errorf("expected max worklist length 7, got %d", maxWorklistLen)
	}
	if finalWorklistLen != 0 {
		t.Errorf("expected final worklist length 0 after all pops, got %d", finalWorklistLen)
	}
}

func TestKosarajuSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	// Verify deep copy across steps by checking distinct slice headers
	for i := 0; i < len(steps)-1; i++ {
		s1 := steps[i].State.(renderers.GraphState)
		s2 := steps[i+1].State.(renderers.GraphState)

		if len(s1.Nodes) > 0 && len(s2.Nodes) > 0 {
			if &s1.Nodes[0] == &s2.Nodes[0] {
				t.Fatalf("steps %d and %d share identical Nodes slice backing array", i, i+1)
			}
		}
		if len(s1.Edges) > 0 && len(s2.Edges) > 0 {
			if &s1.Edges[0] == &s2.Edges[0] {
				t.Fatalf("steps %d and %d share identical Edges slice backing array", i, i+1)
			}
		}
		if len(s1.Worklist) > 0 && len(s2.Worklist) > 0 {
			if &s1.Worklist[0] == &s2.Worklist[0] {
				t.Fatalf("steps %d and %d share identical Worklist slice backing array", i, i+1)
			}
		}
	}

	// Verify mutating step 0 does not alter subsequent steps
	firstState := steps[0].State.(renderers.GraphState)
	firstNodeLabel := firstState.Nodes[0].Label
	firstState.Nodes[0].Label = "CORRUPTED"

	secondState := steps[1].State.(renderers.GraphState)
	if secondState.Nodes[0].Label == "CORRUPTED" {
		t.Fatalf("mutation of step 0 affected step 1; snapshot was not deep copied")
	}
	firstState.Nodes[0].Label = firstNodeLabel
}

func TestKosarajuSCCComponentGroupings(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.GraphState)
	if !ok {
		t.Fatalf("expected GraphState in last step, got %T", lastStep.State)
	}

	compsRaw, ok := lastStep.Variables["components"].([][]string)
	if !ok {
		t.Fatalf("expected [][]string for components in last step, got %T", lastStep.Variables["components"])
	}

	expected := [][]string{
		{"A", "B", "C"},
		{"D", "E"},
		{"F", "G"},
	}

	if !reflect.DeepEqual(normalizeSCCs(compsRaw), normalizeSCCs(expected)) {
		t.Fatalf("expected components %v, got %v", expected, compsRaw)
	}

	// Verify nodes within the same component share a state, and different components have distinct states
	nodeStateMap := make(map[string]string)
	for _, n := range state.Nodes {
		nodeStateMap[n.ID] = n.State
	}

	sccStatesFound := make(map[string]string)
	for _, comp := range compsRaw {
		firstNodeState := nodeStateMap[comp[0]]
		for _, u := range comp {
			if nodeStateMap[u] != firstNodeState {
				t.Errorf("node %s in SCC %v has state %q, expected %q", u, comp, nodeStateMap[u], firstNodeState)
			}
		}
		sccStatesFound[firstNodeState] = comp[0]
	}

	if len(sccStatesFound) != len(compsRaw) {
		t.Errorf("expected %d distinct states for %d SCCs, got %d", len(compsRaw), len(compsRaw), len(sccStatesFound))
	}
}

func TestKosarajuSingleNodeGraph(t *testing.T) {
	lesson := Lesson{}
	input := LessonInput{
		Nodes: []renderers.GraphNode{{ID: "A"}},
		Edges: []renderers.GraphEdge{},
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	steps, err := lesson.Build(bytes)
	if err != nil {
		t.Fatalf("failed to build single node graph: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	compsRaw, ok := lastStep.Variables["components"].([][]string)
	if !ok {
		t.Fatalf("expected [][]string for components, got %T", lastStep.Variables["components"])
	}
	if len(compsRaw) != 1 || len(compsRaw[0]) != 1 || compsRaw[0][0] != "A" {
		t.Fatalf("expected single component [['A']], got %v", compsRaw)
	}
}

func TestKosarajuFullyConnectedGraph(t *testing.T) {
	lesson := Lesson{}
	input := LessonInput{
		Nodes: []renderers.GraphNode{
			{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"},
		},
		Edges: []renderers.GraphEdge{
			{From: "A", To: "B"}, {From: "B", To: "A"},
			{From: "A", To: "C"}, {From: "C", To: "A"},
			{From: "A", To: "D"}, {From: "D", To: "A"},
			{From: "B", To: "C"}, {From: "C", To: "B"},
			{From: "B", To: "D"}, {From: "D", To: "B"},
			{From: "C", To: "D"}, {From: "D", To: "C"},
		},
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	steps, err := lesson.Build(bytes)
	if err != nil {
		t.Fatalf("failed to build fully connected graph: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	compsRaw, ok := lastStep.Variables["components"].([][]string)
	if !ok {
		t.Fatalf("expected [][]string for components, got %T", lastStep.Variables["components"])
	}
	if len(compsRaw) != 1 || len(compsRaw[0]) != 4 {
		t.Fatalf("expected 1 component with 4 vertices, got %v", compsRaw)
	}
}

func TestKosarajuPhase2Transposition(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build: %v", err)
	}

	var p2Step *renderers.GraphState
	foundL7 := false
	for _, s := range steps {
		if s.LineID == "L7" {
			foundL7 = true
			st := s.State.(renderers.GraphState)
			p2Step = &st
			break
		}
	}

	if !foundL7 || p2Step == nil {
		t.Fatal("step L7 (Phase 2 transposition) not found")
	}

	// Verify all edges in L7 are reversed relative to default graph
	origFromTo := map[[2]string]bool{
		{"A", "B"}: true, {"B", "C"}: true, {"C", "A"}: true,
		{"C", "D"}: true, {"D", "E"}: true, {"E", "D"}: true,
		{"E", "F"}: true, {"F", "G"}: true, {"G", "F"}: true,
	}

	for _, e := range p2Step.Edges {
		reversed := [2]string{e.To, e.From}
		if !origFromTo[reversed] {
			t.Errorf("edge %s -> %s in Phase 2 was not reversed from original", e.From, e.To)
		}
	}
}
