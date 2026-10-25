package trie

// Node represents a trie node storing child edges and end-of-word marker.
type Node struct {
	Char     rune
	Children map[rune]*Node
	IsEnd    bool
}

// Trie represents a prefix tree data structure.
// ponytail: map-based child storage; fixed [26]*Node array reduces memory indirection for lowercase ASCII.
type Trie struct {
	Root *Node
}

// New creates an empty prefix tree with initialized root.
func New() *Trie {
	return &Trie{
		Root: &Node{
			Children: make(map[rune]*Node),
		},
	}
}

// Insert adds word into the trie.
func (t *Trie) Insert(word string) {
	curr := t.Root
	for _, ch := range word {
		if _, ok := curr.Children[ch]; !ok {
			curr.Children[ch] = &Node{
				Char:     ch,
				Children: make(map[rune]*Node),
			}
		}
		curr = curr.Children[ch]
	}
	curr.IsEnd = true
}

// Search returns true if word is present in the trie.
func (t *Trie) Search(word string) bool {
	curr := t.Root
	for _, ch := range word {
		child, ok := curr.Children[ch]
		if !ok {
			return false
		}
		curr = child
	}
	return curr.IsEnd
}

// StartsWith returns true if there is any word in the trie starting with prefix.
func (t *Trie) StartsWith(prefix string) bool {
	curr := t.Root
	for _, ch := range prefix {
		child, ok := curr.Children[ch]
		if !ok {
			return false
		}
		curr = child
	}
	return true
}
