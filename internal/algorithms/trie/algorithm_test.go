package trie

import (
	"reflect"
	"strings"
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
		state, ok := step.State.(renderers.TrieLessonState)
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

func TestTrieLessonTerminalVsPrefix(t *testing.T) {
	steps := Trace([]string{"go", "golang", "good", "god", "go"}, "gol", "gol")
	seen := map[string]string{}
	search, prefix := false, false
	for _, step := range steps {
		s := step.State.(renderers.TrieLessonState)
		for _, n := range s.Nodes {
			key := n.ParentID + ":" + n.Label
			if id, ok := seen[key]; ok && id != n.ID {
				t.Fatal("shared edge changed identity")
			}
			seen[key] = n.ID
		}
		if s.ResultKnown && s.Operation == "search" {
			search = true
			if s.Found {
				t.Fatal("gol is not a word")
			}
		}
		if s.ResultKnown && s.Operation == "startsWith" {
			prefix = true
			if !s.Found {
				t.Fatal("gol is a prefix")
			}
		}
		if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
			t.Fatal("invalid hold")
		}
	}
	if !search || !prefix {
		t.Fatal("missing explicit results")
	}
	if len(steps[len(steps)-1].State.(renderers.TrieLessonState).TerminalIDs) != 4 {
		t.Fatal("duplicate word allocated terminal")
	}
}

func TestTrieLessonRunesAndEscaping(t *testing.T) {
	words := []string{"é猫", "é", "<&\"", ""}
	steps := Trace(words, "é猫", "é")
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	unicode := false
	for _, step := range steps {
		s := step.State.(renderers.TrieLessonState)
		if s.Operation == "search" && s.RuneIndex == 1 && s.ByteOffset == 2 {
			unicode = true
		}
	}
	if !unicode {
		t.Fatal("rune index confused with byte offset")
	}
	last := steps[len(steps)-1].State.(renderers.TrieLessonState)
	rootEnd := false
	for _, id := range last.TerminalIDs {
		if id == "root" {
			rootEnd = true
		}
	}
	if !rootEnd {
		t.Fatal("empty word must mark root terminal")
	}
	words[0] = "changed"
	if steps[0].Variables["words"].([]string)[0] != "é猫" {
		t.Fatal("input aliases initial variables")
	}
	if len(steps[0].State.(renderers.TrieLessonState).Nodes) != 1 {
		t.Fatal("early tree mutated")
	}
}

func TestTrieLessonValidation(t *testing.T) {
	for _, input := range []string{`[null]`, `{"words":[null]}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted null word: %s", input)
		}
	}
	for _, input := range []string{`null`, `{`, `{"words":null}`, `{"search":null}`, `{"words":[1]}`, `{"prefix":1}`, `{"words":["` + strings.Repeat("猫", 33) + `"]}`, `{"search":"` + strings.Repeat("x", 33) + `"}`, `{"prefix":"` + strings.Repeat("x", 33) + `"}`, `[` + strings.Repeat(`"x",`, 32) + `"x"]`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	for _, input := range []string{`{}`, `[]`, `[""]`, `{"words":["` + strings.Repeat("猫", 32) + `"],"search":"` + strings.Repeat("猫", 32) + `","prefix":"` + strings.Repeat("猫", 32) + `"}`, `[` + strings.Repeat(`"x",`, 31) + `"x"]`} {
		if _, err := (Lesson{}).Build([]byte(input)); err != nil {
			t.Errorf("rejected valid input: %v", err)
		}
	}
}

func TestTrieLessonSnapshotsAndDefaults(t *testing.T) {
	steps := Trace(nil, "")
	built, err := (Lesson{}).Build(nil)
	if err != nil || !reflect.DeepEqual(steps, built) {
		t.Fatal("defaults changed")
	}
	results := map[string]bool{}
	for i, step := range steps {
		s := step.State.(renderers.TrieLessonState)
		if s.ResultKnown {
			results[s.Operation] = s.Found
		}
		if step.Action == "compare" && i+1 < len(steps) && step.HoldMS+steps[i+1].HoldMS != 1800 {
			t.Fatal("compare pair budget")
		}
	}
	if results["search"] || !results["startsWith"] {
		t.Fatal("default gold/go results")
	}
	custom := Trace([]string{"ab", "ab"}, "ab", "a")
	var early renderers.TrieLessonState
	for _, step := range custom {
		s := step.State.(renderers.TrieLessonState)
		if s.Phase == "terminal" {
			early = s
			break
		}
	}
	last := custom[len(custom)-1].State.(renderers.TrieLessonState)
	last.Nodes[0].Label = "mutated"
	last.TerminalIDs[0] = "mutated"
	last.PathIDs[0] = "mutated"
	last.Runes[0] = 'x'
	if early.Nodes[0].Label != "root" || early.TerminalIDs[0] == "mutated" || early.PathIDs[0] != "root" || early.Runes[0] != 'a' {
		t.Fatal("snapshots alias mutable data")
	}
}
