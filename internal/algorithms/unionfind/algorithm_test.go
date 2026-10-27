package unionfind

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestDSUBasicOperations(t *testing.T) {
	dsu := New(5)
	if dsu.Count() != 5 {
		t.Fatalf("expected initial count 5, got %d", dsu.Count())
	}

	for i := 0; i < 5; i++ {
		if dsu.Find(i) != i {
			t.Fatalf("expected initial parent of %d to be %d, got %d", i, i, dsu.Find(i))
		}
	}

	// 0 - 1
	if !dsu.Union(0, 1) {
		t.Fatalf("expected Union(0, 1) to be true")
	}
	if dsu.Count() != 4 {
		t.Fatalf("expected count 4 after first union, got %d", dsu.Count())
	}
	if !dsu.Connected(0, 1) {
		t.Fatalf("expected 0 and 1 to be connected")
	}

	// Redundant union
	if dsu.Union(0, 1) {
		t.Fatalf("expected redundant Union(0, 1) to be false")
	}
	if dsu.Count() != 4 {
		t.Fatalf("expected count to remain 4, got %d", dsu.Count())
	}

	// 1 - 2 (transitive connection 0 - 2)
	if !dsu.Union(1, 2) {
		t.Fatalf("expected Union(1, 2) to be true")
	}
	if dsu.Count() != 3 {
		t.Fatalf("expected count 3, got %d", dsu.Count())
	}
	if !dsu.Connected(0, 2) {
		t.Fatalf("expected 0 and 2 to be connected transitively")
	}

	// 3 and 4 are not yet connected to {0, 1, 2}
	if dsu.Connected(0, 3) || dsu.Connected(1, 4) {
		t.Fatalf("expected 3 and 4 to be disconnected from {0, 1, 2}")
	}
}

func TestDSUUnionByRank(t *testing.T) {
	dsu := New(4)

	// Union(0, 1): equal ranks (0, 0) -> rank becomes 1
	dsu.Union(0, 1)
	root01 := dsu.Find(0)
	if dsu.rank[root01] != 1 {
		t.Fatalf("expected rank of merged root to be 1, got %d", dsu.rank[root01])
	}

	// Union(2, 3): equal ranks (0, 0) -> rank becomes 1
	dsu.Union(2, 3)
	root23 := dsu.Find(2)
	if dsu.rank[root23] != 1 {
		t.Fatalf("expected rank of merged root to be 1, got %d", dsu.rank[root23])
	}

	// Union(0, 2): both have rank 1 -> new root rank becomes 2
	dsu.Union(0, 2)
	finalRoot := dsu.Find(0)
	if dsu.rank[finalRoot] != 2 {
		t.Fatalf("expected final root rank to be 2, got %d", dsu.rank[finalRoot])
	}
}

func TestDSUPathCompression(t *testing.T) {
	dsu := New(4)
	// Manually construct chain: 3 -> 2 -> 1 -> 0
	dsu.parent[3] = 2
	dsu.parent[2] = 1
	dsu.parent[1] = 0
	dsu.parent[0] = 0

	root := dsu.Find(3)
	if root != 0 {
		t.Fatalf("expected Find(3) to be 0, got %d", root)
	}

	// After Find(3) with path compression, parent[3] and parent[2] must point directly to root 0
	if dsu.parent[3] != 0 {
		t.Fatalf("expected parent[3] to be compressed to 0, got %d", dsu.parent[3])
	}
	if dsu.parent[2] != 0 {
		t.Fatalf("expected parent[2] to be compressed to 0, got %d", dsu.parent[2])
	}
}

func TestDSUOutOfBounds(t *testing.T) {
	dsu := New(3)
	if dsu.Find(-1) != -1 || dsu.Find(5) != -1 {
		t.Fatalf("expected Find with out-of-bound index to return -1")
	}
	if dsu.Union(-1, 0) || dsu.Union(0, 5) {
		t.Fatalf("expected Union with out-of-bound index to return false")
	}
	if dsu.Connected(-1, 0) || dsu.Connected(0, 5) {
		t.Fatalf("expected Connected with out-of-bound index to return false")
	}
	if dsu.findPath(-1) != nil || dsu.findPath(5) != nil {
		t.Fatalf("expected findPath with out-of-bound index to return nil")
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "unionfind" {
		t.Errorf("expected ID 'unionfind', got %q", meta.ID)
	}
	if meta.Name != "Disjoint Set Union (Union-Find)" {
		t.Errorf("expected Name 'Disjoint Set Union (Union-Find)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("expected Category %q, got %q", catalog.CategoryTrees, meta.Category)
	}
	if meta.Renderer != "graph" {
		t.Errorf("expected Renderer 'graph', got %q", meta.Renderer)
	}
	if meta.Documentation.Summary == "" {
		t.Errorf("expected non-empty Summary")
	}
	if meta.Documentation.TimeComplexity != "O(α(n)) amortized inverse Ackermann" {
		t.Errorf("expected TimeComplexity 'O(α(n)) amortized inverse Ackermann', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("expected SpaceComplexity 'O(n)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}
	expectedConcepts := []string{"Equivalence Relations", "Representative Roots", "Path Compression", "Union by Rank"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.HowItWorks == "" {
		t.Errorf("expected non-empty HowItWorks")
	}
}

func TestLessonBuildAndTrace(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatalf("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	// Verify trace structure
	for i, step := range steps {
		if step.LineID == "" {
			t.Errorf("step %d missing LineID", i)
		}
		if step.Explanation == "" {
			t.Errorf("step %d missing Explanation", i)
		}
		if step.Variables == nil {
			t.Errorf("step %d missing Variables", i)
		}
		if _, ok := step.Variables["parent"]; !ok {
			t.Errorf("step %d missing 'parent' in Variables", i)
		}

		state, ok := step.State.(renderers.GraphState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.GraphState", i, step.State)
		}
		if len(state.Nodes) != 5 {
			t.Fatalf("step %d: expected 5 nodes in GraphState, got %d", i, len(state.Nodes))
		}
		if state.WorklistLabel != "Operations Queue" {
			t.Fatalf("step %d: expected WorklistLabel 'Operations Queue', got %q", i, state.WorklistLabel)
		}
	}
}

func TestWorklistOperationsQueue(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// First step should have all 7 operations
	initState := steps[0].State.(renderers.GraphState)
	if len(initState.Worklist) != 7 {
		t.Fatalf("expected initial worklist to have 7 operations, got %d", len(initState.Worklist))
	}
	if initState.Worklist[0].Label != "Union(0, 1)" {
		t.Errorf("expected first worklist item 'Union(0, 1)', got %q", initState.Worklist[0].Label)
	}

	// Final step should have empty worklist
	lastState := steps[len(steps)-1].State.(renderers.GraphState)
	if len(lastState.Worklist) != 0 {
		t.Fatalf("expected final worklist to be empty, got %d items", len(lastState.Worklist))
	}
}

func TestGraphStateEdgesAndForest(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Step 0: Initial state - each node points to itself as root, so 0 inter-node parent edges
	initState := steps[0].State.(renderers.GraphState)
	if len(initState.Edges) != 0 {
		t.Fatalf("expected 0 edges in initial parent forest, got %d", len(initState.Edges))
	}

	// All edges across all steps must be directed parent pointers (From != To)
	for i, step := range steps {
		state := step.State.(renderers.GraphState)
		for _, edge := range state.Edges {
			if !edge.Directed {
				t.Fatalf("step %d: expected directed edge, got undirected %v", i, edge)
			}
			if edge.From == edge.To {
				t.Fatalf("step %d: unexpected self-loop edge %v", i, edge)
			}
		}
	}
}

func TestPathCompressionInTrace(t *testing.T) {
	// Construct chain: 0-1, 2-3, then 0-2 (giving 3 -> 2 -> 0), then Find(3)
	input := json.RawMessage(`{
		"n": 4,
		"operations": [
			{"op": "union", "i": 0, "j": 1},
			{"op": "union", "i": 2, "j": 3},
			{"op": "union", "i": 0, "j": 2},
			{"op": "find", "i": 3}
		]
	}`)

	lesson := Lesson{}
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundL7 := false
	for _, step := range steps {
		if step.LineID == "L7" {
			foundL7 = true
			if step.Action != trace.ActionSet {
				t.Errorf("expected ActionSet for L7, got %v", step.Action)
			}
			if !strings.Contains(step.Explanation, "Path compression") {
				t.Errorf("expected 'Path compression' in explanation, got %q", step.Explanation)
			}
			state := step.State.(renderers.GraphState)
			// Node 3 should now point directly to root 0
			foundEdge3To0 := false
			for _, e := range state.Edges {
				if e.From == "3" && e.To == "0" {
					foundEdge3To0 = true
					if !e.Active {
						t.Errorf("expected active edge for compressed pointer 3 -> 0")
					}
				}
			}
			if !foundEdge3To0 {
				t.Errorf("expected edge 3 -> 0 in GraphState during path compression step")
			}
		}
	}

	if !foundL7 {
		t.Fatalf("expected path compression step (L7) in trace")
	}
}

func TestUnionByRankInTrace(t *testing.T) {
	// Union equal ranks (0, 0) -> hits L16
	// Then Union unequal ranks (rank 1 vs rank 0) -> hits L14 or L15
	input := json.RawMessage(`{
		"n": 3,
		"operations": [
			{"op": "union", "i": 0, "j": 1},
			{"op": "union", "i": 0, "j": 2}
		]
	}`)

	lesson := Lesson{}
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundL16 := false
	foundL14orL15 := false

	for _, step := range steps {
		if step.LineID == "L16" {
			foundL16 = true
			if !strings.Contains(step.Explanation, "Ranks equal") {
				t.Errorf("expected explanation to mention equal ranks, got %q", step.Explanation)
			}
		}
		if step.LineID == "L14" || step.LineID == "L15" {
			foundL14orL15 = true
			if !strings.Contains(step.Explanation, "rank[") {
				t.Errorf("expected explanation to mention rank comparison, got %q", step.Explanation)
			}
		}
	}

	if !foundL16 {
		t.Errorf("expected equal rank step L16 in trace")
	}
	if !foundL14orL15 {
		t.Errorf("expected unequal rank step L14/L15 in trace")
	}
}

func TestCycleSameSetCheckInTrace(t *testing.T) {
	// Redundant union of elements already in the same set
	input := json.RawMessage(`{
		"n": 3,
		"operations": [
			{"op": "union", "i": 0, "j": 1},
			{"op": "union", "i": 0, "j": 1}
		]
	}`)

	lesson := Lesson{}
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundL13 := false
	for _, step := range steps {
		if step.LineID == "L13" {
			foundL13 = true
			if !strings.Contains(step.Explanation, "Already in same set") {
				t.Errorf("expected explanation to state 'Already in same set', got %q", step.Explanation)
			}
			state := step.State.(renderers.GraphState)
			if !strings.Contains(state.Caption, "Already in same set") {
				t.Errorf("expected caption to state 'Already in same set', got %q", state.Caption)
			}
			if step.Variables["redundant"] != true {
				t.Errorf("expected redundant=true in variables")
			}
		}
	}

	if !foundL13 {
		t.Fatalf("expected same-set detection step L13 in trace")
	}
}

func TestConnectedQueryInTrace(t *testing.T) {
	input := json.RawMessage(`{
		"n": 3,
		"operations": [
			{"op": "union", "i": 0, "j": 1},
			{"op": "connected", "i": 0, "j": 1},
			{"op": "connected", "i": 0, "j": 2}
		]
	}`)

	lesson := Lesson{}
	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundTrue := false
	foundFalse := false

	for _, step := range steps {
		if step.LineID == "L20" {
			state := step.State.(renderers.GraphState)
			if strings.Contains(state.Caption, "Connected(0, 1) = true") {
				foundTrue = true
				if step.Variables["connected"] != true {
					t.Errorf("expected connected=true in variables")
				}
			}
			if strings.Contains(state.Caption, "Connected(0, 2) = false") {
				foundFalse = true
				if step.Variables["connected"] != false {
					t.Errorf("expected connected=false in variables")
				}
			}
		}
	}

	if !foundTrue {
		t.Errorf("expected Connected query showing outcome true in caption")
	}
	if !foundFalse {
		t.Errorf("expected Connected query showing outcome false in caption")
	}
}

func TestGraphStateSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify deep copy across all consecutive step pairs
	for i := 0; i < len(steps)-1; i++ {
		s1 := steps[i].State.(renderers.GraphState)
		s2 := steps[i+1].State.(renderers.GraphState)

		if len(s1.Nodes) > 0 && len(s2.Nodes) > 0 {
			if &s1.Nodes[0] == &s2.Nodes[0] {
				t.Fatalf("steps %d and %d share Nodes backing array", i, i+1)
			}
		}
		if len(s1.Edges) > 0 && len(s2.Edges) > 0 {
			if &s1.Edges[0] == &s2.Edges[0] {
				t.Fatalf("steps %d and %d share Edges backing array", i, i+1)
			}
		}
		if len(s1.Worklist) > 0 && len(s2.Worklist) > 0 {
			if &s1.Worklist[0] == &s2.Worklist[0] {
				t.Fatalf("steps %d and %d share Worklist backing array", i, i+1)
			}
		}
	}

	// Mutate step 0 state and ensure other steps remain untouched
	s0 := steps[0].State.(renderers.GraphState)
	s0.Nodes[0].State = "MUTATED"
	if len(s0.Worklist) > 0 {
		s0.Worklist[0].Label = "MUTATED"
	}

	s1 := steps[1].State.(renderers.GraphState)
	if s1.Nodes[0].State == "MUTATED" {
		t.Fatalf("step 1 was mutated when step 0 was changed")
	}
	if len(s1.Worklist) > 0 && s1.Worklist[0].Label == "MUTATED" {
		t.Fatalf("step 1 worklist was mutated when step 0 worklist was changed")
	}
}

func TestInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{"malformed json", json.RawMessage(`{not a valid json}`)},
		{"null literal", json.RawMessage(`null`)},
		{"zero n", json.RawMessage(`{"n": 0}`)},
		{"negative n", json.RawMessage(`{"n": -5}`)},
		{"negative index i in union", json.RawMessage(`{"n": 3, "operations": [{"op": "union", "i": -1, "j": 1}]}`)},
		{"out of bounds index j in union", json.RawMessage(`{"n": 3, "operations": [{"op": "union", "i": 0, "j": 3}]}`)},
		{"negative index in find", json.RawMessage(`{"n": 3, "operations": [{"op": "find", "i": -1}]}`)},
		{"out of bounds index in find", json.RawMessage(`{"n": 3, "operations": [{"op": "find", "i": 4}]}`)},
		{"out of bounds index in connected", json.RawMessage(`{"n": 3, "operations": [{"op": "connected", "i": 0, "j": 5}]}`)},
		{"unknown operation", json.RawMessage(`{"n": 3, "operations": [{"op": "disconnect", "i": 0, "j": 1}]}`)},
		{"unknown field", json.RawMessage(`{"n": 3, "unexpected": 42}`)},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}

	// Valid inputs
	validInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{"nil input", nil},
		{"empty input", json.RawMessage(``)},
		{"custom valid", json.RawMessage(`{"n": 3, "operations": [{"op": "union", "i": 0, "j": 1}]}`)},
		{"custom n only", json.RawMessage(`{"n": 3}`)},
		{"empty operations", json.RawMessage(`{"n": 3, "operations": []}`)},
	}

	for _, tc := range validInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
			if len(steps) == 0 {
				t.Fatalf("expected non-empty steps for %s", tc.name)
			}
		})
	}
}

func TestCatalogRegistration(t *testing.T) {
	l, ok := catalog.Get("unionfind")
	if !ok {
		t.Fatalf("expected 'unionfind' to be registered in catalog")
	}
	if l.Meta().ID != "unionfind" {
		t.Fatalf("expected registered lesson ID 'unionfind', got %q", l.Meta().ID)
	}
}

func TestGraphSVGRendering(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	for i, step := range steps {
		svg := renderers.RenderGraphSVG(step, 800, 450)
		if svg == "" {
			t.Fatalf("step %d: empty SVG rendered", i)
		}
		if !strings.Contains(svg, "Operations Queue") {
			t.Fatalf("step %d: missing 'Operations Queue' in rendered SVG", i)
		}
	}
}
