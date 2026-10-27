package trie

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "trie",
		Name:     "Trie (Prefix Tree)",
		Category: catalog.CategoryTrees,
		Renderer: "tree",
		Documentation: catalog.Documentation{
			Summary:         "An ordered tree data structure used to store an associative array where keys are usually strings. All descendants of a node share a common prefix.",
			TimeComplexity:  "O(L) where L is key length",
			SpaceComplexity: "O(ALPHABET_SIZE · L · N)",
			Difficulty:      "Medium",
			KeyConcepts:     []string{"Shared Prefixes", "Word End Flags", "Fast Dictionary Search", "Autocomplete"},
			HowItWorks:      "Words are inserted character-by-character from the root, sharing common prefix pathways. Each node stores child pointers keyed by character, plus a flag indicating whether the node completes a valid word. Search and prefix queries follow the path of characters in O(L) time independently of the number of words stored in the tree.",
		},
		DefaultInput: `{"words": ["go", "golang", "good", "god"], "search": "gold", "prefix": "go"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func (t *Trie) Insert(word string) {"},
		{ID: "L2", Number: 2, Code: "    curr := t.Root"},
		{ID: "L3", Number: 3, Code: "    for _, ch := range word {"},
		{ID: "L4", Number: 4, Code: "        if _, ok := curr.Children[ch]; !ok {"},
		{ID: "L5", Number: 5, Code: "            curr.Children[ch] = &Node{Char: ch, Children: make(map[rune]*Node)}"},
		{ID: "L6", Number: 6, Code: "        }"},
		{ID: "L7", Number: 7, Code: "        curr = curr.Children[ch]"},
		{ID: "L8", Number: 8, Code: "    }"},
		{ID: "L9", Number: 9, Code: "    curr.IsEnd = true"},
		{ID: "L10", Number: 10, Code: "}"},
		{ID: "L11", Number: 11, Code: "func (t *Trie) Search(word string) bool {"},
		{ID: "L12", Number: 12, Code: "    curr := t.Root"},
		{ID: "L13", Number: 13, Code: "    for _, ch := range word {"},
		{ID: "L14", Number: 14, Code: "        child, ok := curr.Children[ch]"},
		{ID: "L15", Number: 15, Code: "        if !ok { return false }"},
		{ID: "L16", Number: 16, Code: "        curr = child"},
		{ID: "L17", Number: 17, Code: "    }"},
		{ID: "L18", Number: 18, Code: "    return curr.IsEnd"},
		{ID: "L19", Number: 19, Code: "}"},
		{ID: "L20", Number: 20, Code: "func (t *Trie) StartsWith(prefix string) bool {"},
		{ID: "L21", Number: 21, Code: "    curr := t.Root"},
		{ID: "L22", Number: 22, Code: "    for _, ch := range prefix {"},
		{ID: "L23", Number: 23, Code: "        child, ok := curr.Children[ch]"},
		{ID: "L24", Number: 24, Code: "        if !ok { return false }"},
		{ID: "L25", Number: 25, Code: "        curr = child"},
		{ID: "L26", Number: 26, Code: "    }"},
		{ID: "L27", Number: 27, Code: "    return true"},
		{ID: "L28", Number: 28, Code: "}"},
	}
}

type buildInput struct {
	Words  []string `json:"words"`
	Search string   `json:"search"`
	Prefix string   `json:"prefix"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	words := []string{"go", "golang", "good", "god"}
	search := "gold"
	prefix := "go"

	if len(input) > 0 {
		var custom buildInput
		if err := json.Unmarshal(input, &custom); err == nil && (len(custom.Words) > 0 || custom.Search != "" || custom.Prefix != "") {
			if len(custom.Words) > 0 {
				words = custom.Words
			}
			if custom.Search != "" {
				search = custom.Search
			}
			if custom.Prefix != "" {
				prefix = custom.Prefix
			}
		} else {
			var customList []string
			if err := json.Unmarshal(input, &customList); err == nil && len(customList) > 0 {
				words = customList
			}
		}
	}

	return Trace(words, search, prefix), nil
}

func init() {
	catalog.Register(Lesson{})
}
