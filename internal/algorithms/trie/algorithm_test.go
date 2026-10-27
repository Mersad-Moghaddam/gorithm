package trie

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestTrieCorrectness(t *testing.T) {
	tr := New()

	if tr.Search("") {
		t.Fatalf("Search(\"\") on empty trie want false")
	}
	if tr.Search("apple") {
		t.Fatalf("Search(\"apple\") on empty trie want false")
	}
	if tr.StartsWith("app") {
		t.Fatalf("StartsWith(\"app\") on empty trie want false")
	}

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

	tr.Insert("app")
	if !tr.Search("app") {
		t.Fatalf("Search(\"app\") after Insert(\"app\") want true")
	}

	words := []string{"go", "golang", "good", "god"}
	for _, w := range words {
		tr.Insert(w)
	}

	for _, w := range words {
		if !tr.Search(w) {
			t.Fatalf("Search(%q) want true", w)
		}
	}

	if tr.Search("gold") {
		t.Fatalf("Search(\"gold\") want false")
	}
	if tr.Search("g") {
		t.Fatalf("Search(\"g\") want false")
	}

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

		rootNode, hasRoot := nodeMap["root"]
		if !hasRoot {
			t.Fatalf("step %d missing root node", i)
		}
		if rootNode.Label != "*" {
			t.Errorf("step %d root label = %q, want '*'", i, rootNode.Label)
		}
		if rootNode.ParentID != "" {
			t.Errorf("step %d root parentID = %q, want empty", i, rootNode.ParentID)
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

func TestTrieSharedPrefixes(t *testing.T) {
	words := []string{"car", "card", "cat"}
	steps := Trace(words, "cat", "ca")

	foundSharedPrefixStep := false
	for _, step := range steps {
		if step.LineID == "L6" && strings.Contains(step.Explanation, "shared prefix") {
			foundSharedPrefixStep = true
			if created, ok := step.Variables["created"].(bool); !ok || created {
				t.Fatalf("expected created=false on shared prefix step, got %v", step.Variables["created"])
			}
		}
	}
	if !foundSharedPrefixStep {
		t.Fatal("expected at least one step explaining shared prefix reuse (L6)")
	}

	lastStep := steps[len(steps)-1]
	state := lastStep.State.(renderers.TreeState)

	// "car", "card", "cat" share 'c' and 'a'.
	// Nodes: root('*'), 'c', 'a', 'r', 'd', 't' -> exactly 6 nodes.
	if len(state.Nodes) != 6 {
		t.Fatalf("expected 6 nodes in final tree with shared prefixes, got %d", len(state.Nodes))
	}

	charCounts := make(map[string]int)
	for _, n := range state.Nodes {
		charCounts[n.Label]++
	}
	if charCounts["c"] != 1 {
		t.Fatalf("expected 1 shared node for 'c', got %d", charCounts["c"])
	}
	if charCounts["a"] != 1 {
		t.Fatalf("expected 1 shared node for 'a', got %d", charCounts["a"])
	}
}

func TestTrieWordTermination(t *testing.T) {
	steps := Trace([]string{"app", "apple"}, "apple", "app")

	var afterInserts renderers.TreeState
	for _, step := range steps {
		if step.LineID == "L10" && step.Explanation == "Insert(\"apple\") completed." {
			afterInserts = step.State.(renderers.TreeState)
			break
		}
	}
	if len(afterInserts.Nodes) == 0 {
		t.Fatal("failed to find state after inserting 'app' and 'apple'")
	}

	nodeMap := make(map[string]renderers.TreeNode)
	for _, n := range afterInserts.Nodes {
		nodeMap[n.ID] = n
	}

	// Trace "app": root -> 'a' -> 'p' -> 'p' (matched)
	// Trace "apple": ... -> 'l' (normal) -> 'e' (matched)
	var rootID, aID, p1ID, p2ID, lID, eID string
	for _, n := range afterInserts.Nodes {
		if n.ID == "root" {
			rootID = n.ID
		}
	}
	for _, n := range afterInserts.Nodes {
		if n.ParentID == rootID && n.Label == "a" {
			aID = n.ID
		}
	}
	for _, n := range afterInserts.Nodes {
		if n.ParentID == aID && n.Label == "p" {
			p1ID = n.ID
		}
	}
	for _, n := range afterInserts.Nodes {
		if n.ParentID == p1ID && n.Label == "p" {
			p2ID = n.ID
		}
	}
	for _, n := range afterInserts.Nodes {
		if n.ParentID == p2ID && n.Label == "l" {
			lID = n.ID
		}
	}
	for _, n := range afterInserts.Nodes {
		if n.ParentID == lID && n.Label == "e" {
			eID = n.ID
		}
	}

	if aID == "" || p1ID == "" || p2ID == "" || lID == "" || eID == "" {
		t.Fatalf("failed to locate path nodes: a=%s p1=%s p2=%s l=%s e=%s", aID, p1ID, p2ID, lID, eID)
	}

	if nodeMap[aID].State != "normal" {
		t.Errorf("prefix node 'a' state = %q, want 'normal'", nodeMap[aID].State)
	}
	if nodeMap[p1ID].State != "normal" {
		t.Errorf("prefix node 'p' (1st) state = %q, want 'normal'", nodeMap[p1ID].State)
	}
	if nodeMap[p2ID].State != "matched" {
		t.Errorf("word termination node 'p' (end of 'app') state = %q, want 'matched'", nodeMap[p2ID].State)
	}
	if nodeMap[lID].State != "normal" {
		t.Errorf("prefix node 'l' state = %q, want 'normal'", nodeMap[lID].State)
	}
	if nodeMap[eID].State != "matched" {
		t.Errorf("word termination node 'e' (end of 'apple') state = %q, want 'matched'", nodeMap[eID].State)
	}
}

func TestTrieSearchOperations(t *testing.T) {
	// 1. Search success
	stepsSuccess := Trace([]string{"golang"}, "golang", "go")
	var searchSuccessStep *trace.Step
	for i := range stepsSuccess {
		step := &stepsSuccess[i]
		if step.LineID == "L18" && step.Variables["found"] == true {
			searchSuccessStep = step
			break
		}
	}
	if searchSuccessStep == nil {
		t.Fatal("expected search success step on L18 with found=true")
	}

	// 2. Search failure: missing edge
	stepsMissingEdge := Trace([]string{"golang"}, "gold", "go")
	var searchMissingStep *trace.Step
	for i := range stepsMissingEdge {
		step := &stepsMissingEdge[i]
		if step.LineID == "L15" && step.Variables["found"] == false {
			searchMissingStep = step
			break
		}
	}
	if searchMissingStep == nil {
		t.Fatal("expected search failure step on L15 with found=false for missing edge")
	}

	// 3. Search failure: prefix exists but isEnd is false
	stepsPrefixOnly := Trace([]string{"golang"}, "gola", "go")
	var searchPrefixOnlyStep *trace.Step
	for i := range stepsPrefixOnly {
		step := &stepsPrefixOnly[i]
		if step.LineID == "L18" && step.Variables["found"] == false {
			searchPrefixOnlyStep = step
			break
		}
	}
	if searchPrefixOnlyStep == nil {
		t.Fatal("expected search failure step on L18 with found=false when isEnd=false")
	}
}

func TestTriePrefixOperations(t *testing.T) {
	// 1. StartsWith success
	stepsSuccess := Trace([]string{"golang"}, "golang", "go")
	var prefixSuccessStep *trace.Step
	for i := range stepsSuccess {
		step := &stepsSuccess[i]
		if step.LineID == "L27" && step.Variables["found"] == true {
			prefixSuccessStep = step
			break
		}
	}
	if prefixSuccessStep == nil {
		t.Fatal("expected StartsWith success step on L27 with found=true")
	}

	// 2. StartsWith failure: missing edge
	stepsFail := Trace([]string{"golang"}, "golang", "zoo")
	var prefixFailStep *trace.Step
	for i := range stepsFail {
		step := &stepsFail[i]
		if step.LineID == "L24" && step.Variables["found"] == false {
			prefixFailStep = step
			break
		}
	}
	if prefixFailStep == nil {
		t.Fatal("expected StartsWith failure step on L24 with found=false")
	}
}

func TestTrieSnapshotImmutability(t *testing.T) {
	steps := Trace([]string{"go", "golang"}, "go", "go")
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	type nodeSnapshot struct {
		id    string
		state string
		label string
	}
	initialSnapshots := make([][]nodeSnapshot, len(steps))
	for i, step := range steps {
		ts := step.State.(renderers.TreeState)
		for _, n := range ts.Nodes {
			initialSnapshots[i] = append(initialSnapshots[i], nodeSnapshot{
				id:    n.ID,
				state: n.State,
				label: n.Label,
			})
		}
	}

	// Mutate first step
	ts0 := steps[0].State.(renderers.TreeState)
	ts0.Nodes[0].ID = "MUTATED_ID"
	ts0.Nodes[0].State = "corrupted"
	ts0.Nodes[0].Label = "X"

	// Verify all other steps remain unaltered
	for i := 1; i < len(steps); i++ {
		ts := steps[i].State.(renderers.TreeState)
		for j, n := range ts.Nodes {
			orig := initialSnapshots[i][j]
			if n.ID != orig.id || n.State != orig.state || n.Label != orig.label {
				t.Fatalf("step %d was mutated by modifying step 0: got %+v, want %+v", i, n, orig)
			}
		}
	}

	// Mutate last step
	lastIdx := len(steps) - 1
	tsLast := steps[lastIdx].State.(renderers.TreeState)
	tsLast.Nodes[0].ID = "LAST_MUTATED"
	tsLast.Nodes[0].State = "deleted"

	// Verify step 1 still matches initial snapshot
	ts1 := steps[1].State.(renderers.TreeState)
	for j, n := range ts1.Nodes {
		orig := initialSnapshots[1][j]
		if n.ID != orig.id || n.State != orig.state || n.Label != orig.label {
			t.Fatalf("step 1 was mutated by modifying last step: got %+v, want %+v", n, orig)
		}
	}

	// Verify slice backing arrays are distinct between consecutive steps
	for i := 0; i < len(steps)-1; i++ {
		tsA := steps[i].State.(renderers.TreeState)
		tsB := steps[i+1].State.(renderers.TreeState)
		if len(tsA.Nodes) > 0 && len(tsB.Nodes) > 0 {
			if &tsA.Nodes[0] == &tsB.Nodes[0] {
				t.Fatalf("steps %d and %d share backing slice array", i, i+1)
			}
		}
	}
}

func TestTrieBuildValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []json.RawMessage{
		[]byte(`null`),
		[]byte(`not json`),
		[]byte(`{"bad": 1}`),
		[]byte(`{"words": []}`),
		[]byte(`{"words": [""]}`),
		[]byte(`{"words": ["go", ""]}`),
		[]byte(`[]`),
		[]byte(`[""]`),
		[]byte(`["go", ""]`),
		[]byte(`[1, 2]`),
		[]byte(`123`),
		[]byte(`"just a string"`),
		[]byte(`{"words": ["good"], "extra": "invalid"}`),
	}

	for _, input := range invalidInputs {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("Build(%s) want error, got nil", string(input))
		}
	}

	validInputs := []json.RawMessage{
		nil,
		[]byte(``),
		[]byte(`   `),
		[]byte(`{"words": ["cat", "car", "cart"], "search": "car", "prefix": "ca"}`),
		[]byte(`["cat", "car"]`),
		[]byte(`{"words": ["cat"]}`),
	}

	for _, input := range validInputs {
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("Build(%s) want success, got error: %v", string(input), err)
		}
		if len(steps) == 0 {
			t.Fatalf("Build(%s) produced 0 steps", string(input))
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("Build(%s) ValidateSourceLineIDs failed: %v", string(input), err)
		}
	}
}
