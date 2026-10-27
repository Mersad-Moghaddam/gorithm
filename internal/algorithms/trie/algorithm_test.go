package trie

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestTrieCorrectness(t *testing.T) {
	tr := New()

	// Empty trie
	if tr.Search("") {
		t.Fatalf("Search(\"\") on empty trie want false")
	}
	if tr.Search("apple") {
		t.Fatalf("Search(\"apple\") on empty trie want false")
	}
	if tr.StartsWith("app") {
		t.Fatalf("StartsWith(\"app\") on empty trie want false")
	}

	// Insert "apple"
	tr.Insert("apple")
	if !tr.Search("apple") {
		t.Fatalf("Search(\"apple\") after Insert(\"apple\") want true")
	}
	if tr.Search("app") {
		t.Fatalf("Search(\"app\") want false (prefix only, not full word)")
	}
	if !tr.StartsWith("app") {
		t.Fatalf("StartsWith(\"app\") want true")
	}
	if !tr.StartsWith("apple") {
		t.Fatalf("StartsWith(\"apple\") want true")
	}
	if tr.StartsWith("applause") {
		t.Fatalf("StartsWith(\"applause\") want false")
	}

	// Insert "app"
	tr.Insert("app")
	if !tr.Search("app") {
		t.Fatalf("Search(\"app\") after Insert(\"app\") want true")
	}

	// Insert multiple words
	words := []string{"go", "golang", "good", "god"}
	for _, w := range words {
		tr.Insert(w)
	}

	for _, w := range words {
		if !tr.Search(w) {
			t.Fatalf("Search(%q) want true", w)
		}
	}

	// Search non-existing
	if tr.Search("gold") {
		t.Fatalf("Search(\"gold\") want false")
	}
	if tr.Search("g") {
		t.Fatalf("Search(\"g\") want false")
	}

	// StartsWith
	if !tr.StartsWith("go") {
		t.Fatalf("StartsWith(\"go\") want true")
	}
	if !tr.StartsWith("gola") {
		t.Fatalf("StartsWith(\"gola\") want true")
	}
	if tr.StartsWith("gold") {
		t.Fatalf("StartsWith(\"gold\") want false")
	}
}

func TestTrieLessonMeta(t *testing.T) {
	lesson, ok := catalog.Get("trie")
	if !ok {
		t.Fatalf("trie lesson not registered in catalog")
	}

	meta := lesson.Meta()
	if meta.ID != "trie" {
		t.Errorf("meta.ID = %q, want 'trie'", meta.ID)
	}
	if meta.Name != "Trie (Prefix Tree)" {
		t.Errorf("meta.Name = %q, want 'Trie (Prefix Tree)'", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("meta.Category = %q, want %q", meta.Category, catalog.CategoryTrees)
	}
	if meta.Renderer != "tree" {
		t.Errorf("meta.Renderer = %q, want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(L) where L is key length" {
		t.Errorf("TimeComplexity = %q, want 'O(L) where L is key length'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(ALPHABET_SIZE · L · N)" {
		t.Errorf("SpaceComplexity = %q, want 'O(ALPHABET_SIZE · L · N)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("Difficulty = %q, want 'Medium'", meta.Documentation.Difficulty)
	}

	wantConcepts := []string{"Shared Prefixes", "Word End Flags", "Fast Dictionary Search", "Autocomplete"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, wantConcepts) {
		t.Errorf("KeyConcepts = %v, want %v", meta.Documentation.KeyConcepts, wantConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("Summary should not be empty")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("HowItWorks should not be empty")
	}
}

func TestTrieTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("Build returned empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}

	for i, step := range steps {
		state, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}
		if len(state.Nodes) == 0 {
			t.Fatalf("step %d has 0 tree nodes", i)
		}
		nodeMap := make(map[string]renderers.TreeNode)
		for _, n := range state.Nodes {
			nodeMap[n.ID] = n
		}
		for _, n := range state.Nodes {
			if n.ParentID != "" {
				if _, exists := nodeMap[n.ParentID]; !exists {
					t.Errorf("step %d: node %s has unknown parent %s", i, n.ID, n.ParentID)
				}
			}
		}
	}
}

func TestTrieCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"words": ["cat", "car", "cart"],
		"search": "car",
		"prefix": "ca"
	}`)

	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build failed with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("custom input trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed on custom input: %v", err)
	}
}

