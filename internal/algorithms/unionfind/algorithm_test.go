package unionfind

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
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
	}
}

func TestLessonBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	custom := json.RawMessage(`{
		"n": 4,
		"operations": [
			{"op": "union", "i": 0, "j": 1},
			{"op": "union", "i": 2, "j": 3},
			{"op": "find", "i": 1},
			{"op": "connected", "i": 0, "j": 3}
		]
	}`)

	steps, err := lesson.Build(custom)
	if err != nil {
		t.Fatalf("Build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatalf("expected steps from custom input")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed for custom input: %v", err)
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
