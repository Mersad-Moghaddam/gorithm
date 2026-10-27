package trie

import (
	"fmt"
	"sort"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type tracer struct {
	b       *trace.Builder
	nodeIDs map[*Node]string
	root    *Node
	seq     int
}

func newTracer() *tracer {
	root := &Node{
		Char:     '*',
		Children: make(map[rune]*Node),
	}
	tr := &tracer{
		b:       trace.NewBuilder(),
		nodeIDs: make(map[*Node]string),
		root:    root,
	}
	tr.nodeIDs[root] = "root"
	return tr
}

func (tr *tracer) idFor(n *Node) string {
	if n == nil {
		return ""
	}
	if id, ok := tr.nodeIDs[n]; ok {
		return id
	}
	tr.seq++
	id := fmt.Sprintf("node-%d", tr.seq)
	tr.nodeIDs[n] = id
	return id
}

type childPair struct {
	ch   rune
	node *Node
}

func getChildren(n *Node) []childPair {
	var list []childPair
	for ch, child := range n.Children {
		list = append(list, childPair{ch: ch, node: child})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ch < list[j].ch
	})
	return list
}

func (tr *tracer) treeState(states map[*Node]string) renderers.TreeState {
	if tr.root == nil {
		return renderers.TreeState{Nodes: []renderers.TreeNode{}}
	}

	var leaves []*Node
	var maxDepth int
	var findLeaves func(n *Node, depth int)
	findLeaves = func(n *Node, depth int) {
		if depth > maxDepth {
			maxDepth = depth
		}
		children := getChildren(n)
		if len(children) == 0 {
			leaves = append(leaves, n)
			return
		}
		for _, cp := range children {
			findLeaves(cp.node, depth+1)
		}
	}
	findLeaves(tr.root, 0)

	leafX := make(map[*Node]float64)
	numLeaves := len(leaves)
	minX, maxX := 60.0, 740.0
	if numLeaves <= 1 {
		if numLeaves == 1 {
			leafX[leaves[0]] = 400.0
		}
	} else {
		stepX := (maxX - minX) / float64(numLeaves)
		for i, leaf := range leaves {
			leafX[leaf] = minX + (float64(i)+0.5)*stepX
		}
	}

	yStep := 60.0
	if maxDepth > 5 {
		yStep = 340.0 / float64(maxDepth)
		if yStep < 35.0 {
			yStep = 35.0
		}
	}

	nodeX := make(map[*Node]float64)
	var computeX func(n *Node) float64
	computeX = func(n *Node) float64 {
		children := getChildren(n)
		if len(children) == 0 {
			nodeX[n] = leafX[n]
			return leafX[n]
		}
		firstX := computeX(children[0].node)
		lastX := firstX
		for i := 1; i < len(children); i++ {
			lastX = computeX(children[i].node)
		}
		x := (firstX + lastX) / 2.0
		nodeX[n] = x
		return x
	}
	computeX(tr.root)

	var nodes []renderers.TreeNode
	var buildNodes func(n *Node, parentID string, depth int)
	buildNodes = func(n *Node, parentID string, depth int) {
		id := tr.idFor(n)
		st := "normal"
		if states != nil {
			if s, ok := states[n]; ok {
				st = s
			}
		}
		nodes = append(nodes, renderers.TreeNode{
			ID:       id,
			Val:      int(n.Char),
			Label:    string(n.Char),
			X:        nodeX[n],
			Y:        50.0 + float64(depth)*yStep,
			ParentID: parentID,
			State:    st,
		})

		for _, cp := range getChildren(n) {
			buildNodes(cp.node, id, depth+1)
		}
	}
	buildNodes(tr.root, "", 0)

	return renderers.TreeState{Nodes: nodes}
}

func (tr *tracer) traceInsert(word string) {
	tr.b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Begin Insert(%q).", word),
		tr.treeState(nil),
		map[string]any{"op": "insert", "word": word})

	curr := tr.root
	tr.b.AddStep("L2", trace.ActionSet, "Set curr to root node.",
		tr.treeState(map[*Node]string{curr: "active"}),
		map[string]any{"curr": "root", "word": word})

	for i, ch := range word {
		tr.b.AddStep("L3", trace.ActionVisit, fmt.Sprintf("Examine character '%c' (index %d/%d of %q).", ch, i+1, len(word), word),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch), "index": i})

		tr.b.IncComparisons()
		tr.b.AddStep("L4", trace.ActionCompare, fmt.Sprintf("Check if edge '%c' exists from current node.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch)})

		child, exists := curr.Children[ch]
		if !exists {
			child = &Node{
				Char:     ch,
				Children: make(map[rune]*Node),
			}
			curr.Children[ch] = child
			tr.b.IncOperations()
			tr.b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Edge '%c' missing. Created new node '%c'.", ch, ch),
				tr.treeState(map[*Node]string{child: "inserted", curr: "active"}),
				map[string]any{"word": word, "char": string(ch), "created": true})
		} else {
			tr.b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Edge '%c' exists. Reusing existing node.", ch),
				tr.treeState(map[*Node]string{child: "matched", curr: "active"}),
				map[string]any{"word": word, "char": string(ch), "created": false})
		}

		curr = child
		tr.b.IncOperations()
		tr.b.AddStep("L7", trace.ActionVisit, fmt.Sprintf("Advance curr to node '%c'.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch), "curr": string(ch)})
	}

	curr.IsEnd = true
	tr.b.IncOperations()
	tr.b.AddStep("L9", trace.ActionDone, fmt.Sprintf("Marked node '%c' as end of word %q.", curr.Char, word),
		tr.treeState(map[*Node]string{curr: "matched"}),
		map[string]any{"word": word, "isEnd": true})

	tr.b.AddStep("L10", trace.ActionDone, fmt.Sprintf("Insert(%q) completed.", word),
		tr.treeState(nil),
		map[string]any{"word": word})
}

func (tr *tracer) traceSearch(word string) bool {
	tr.b.AddStep("L11", trace.ActionInit, fmt.Sprintf("Begin Search(%q).", word),
		tr.treeState(nil),
		map[string]any{"op": "search", "word": word})

	curr := tr.root
	tr.b.AddStep("L12", trace.ActionSet, "Set curr to root node.",
		tr.treeState(map[*Node]string{curr: "active"}),
		map[string]any{"curr": "root", "word": word})

	for i, ch := range word {
		tr.b.AddStep("L13", trace.ActionVisit, fmt.Sprintf("Search character '%c' (index %d/%d of %q).", ch, i+1, len(word), word),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch), "index": i})

		tr.b.IncComparisons()
		tr.b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Lookup child edge '%c' from current node.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch)})

		child, exists := curr.Children[ch]
		if !exists {
			tr.b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Edge '%c' not found. Search(%q) = false.", ch, word),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"word": word, "char": string(ch), "found": false})
			return false
		}

		curr = child
		tr.b.IncOperations()
		tr.b.AddStep("L16", trace.ActionVisit, fmt.Sprintf("Found edge '%c'. Advance curr to child node.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"word": word, "char": string(ch), "curr": string(ch)})
	}

	tr.b.IncComparisons()
	if curr.IsEnd {
		tr.b.AddStep("L18", trace.ActionDone, fmt.Sprintf("Node '%c' has isEnd=true. Search(%q) = true.", curr.Char, word),
			tr.treeState(map[*Node]string{curr: "matched"}),
			map[string]any{"word": word, "found": true})
		return true
	}

	tr.b.AddStep("L18", trace.ActionDone, fmt.Sprintf("Prefix matched but node '%c' has isEnd=false. Search(%q) = false.", curr.Char, word),
		tr.treeState(map[*Node]string{curr: "active"}),
		map[string]any{"word": word, "found": false})
	return false
}

func (tr *tracer) traceStartsWith(prefix string) bool {
	tr.b.AddStep("L20", trace.ActionInit, fmt.Sprintf("Begin StartsWith(%q).", prefix),
		tr.treeState(nil),
		map[string]any{"op": "startsWith", "prefix": prefix})

	curr := tr.root
	tr.b.AddStep("L21", trace.ActionSet, "Set curr to root node.",
		tr.treeState(map[*Node]string{curr: "active"}),
		map[string]any{"curr": "root", "prefix": prefix})

	for i, ch := range prefix {
		tr.b.AddStep("L22", trace.ActionVisit, fmt.Sprintf("Check prefix character '%c' (index %d/%d of %q).", ch, i+1, len(prefix), prefix),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"prefix": prefix, "char": string(ch), "index": i})

		tr.b.IncComparisons()
		tr.b.AddStep("L23", trace.ActionCompare, fmt.Sprintf("Lookup child edge '%c' from current node.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"prefix": prefix, "char": string(ch)})

		child, exists := curr.Children[ch]
		if !exists {
			tr.b.AddStep("L24", trace.ActionDone, fmt.Sprintf("Edge '%c' not found. StartsWith(%q) = false.", ch, prefix),
				tr.treeState(map[*Node]string{curr: "active"}),
				map[string]any{"prefix": prefix, "char": string(ch), "found": false})
			return false
		}

		curr = child
		tr.b.IncOperations()
		tr.b.AddStep("L25", trace.ActionVisit, fmt.Sprintf("Found edge '%c'. Advance curr.", ch),
			tr.treeState(map[*Node]string{curr: "active"}),
			map[string]any{"prefix": prefix, "char": string(ch), "curr": string(ch)})
	}

	tr.b.AddStep("L27", trace.ActionDone, fmt.Sprintf("All characters matched. StartsWith(%q) = true.", prefix),
		tr.treeState(map[*Node]string{curr: "matched"}),
		map[string]any{"prefix": prefix, "found": true})
	return true
}

// Trace records execution steps for inserting words and searching.
// ponytail: deterministic slice sorting on children; can cache layout tree if re-traced on large datasets.
func Trace(words []string, searchWord string, prefix ...string) []trace.Step {
	tr := newTracer()
	pref := "go"
	if len(prefix) > 0 && prefix[0] != "" {
		pref = prefix[0]
	}

	if len(words) == 0 {
		words = []string{"go", "golang", "good", "god"}
	}
	if searchWord == "" {
		searchWord = "gold"
	}

	tr.b.AddStep("L1", trace.ActionInit, "Trie (Prefix Tree) initialized with root node.",
		tr.treeState(nil),
		map[string]any{"words": words, "search": searchWord, "prefix": pref})

	for _, w := range words {
		tr.traceInsert(w)
	}

	tr.traceSearch(searchWord)

	if pref != "" {
		tr.traceStartsWith(pref)
	}

	return tr.b.Steps()
}
