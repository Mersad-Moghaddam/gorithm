package trie

import (
	"fmt"
	"sort"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type childPair struct {
	ch   rune
	node *Node
}

func getChildren(n *Node) []childPair {
	var children []childPair
	for ch, node := range n.Children {
		children = append(children, childPair{ch, node})
	}
	sort.Slice(children, func(i, j int) bool { return children[i].ch < children[j].ch })
	return children
}

type tracer struct {
	b       *trace.Builder
	root    *Node
	nodeIDs map[*Node]string
	seq     int
	state   renderers.TrieLessonState
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

func (tr *tracer) add(line string, action trace.ActionType, explanation string, vars map[string]any) {
	s := tr.state
	s.Nodes = nil
	s.TerminalIDs = nil
	var walk func(*Node, string)
	walk = func(n *Node, parent string) {
		id := tr.idFor(n)
		label := string(n.Char)
		if parent == "" {
			label = "root"
		}
		s.Nodes = append(s.Nodes, renderers.TreeNode{ID: id, ParentID: parent, Val: int(n.Char), Label: label})
		if n.IsEnd {
			s.TerminalIDs = append(s.TerminalIDs, id)
		}
		for _, child := range getChildren(n) {
			walk(child.node, id)
		}
	}
	walk(tr.root, "")
	s.Runes = append([]rune(nil), s.Runes...)
	s.PathIDs = append([]string(nil), s.PathIDs...)
	tr.b.AddStep(line, action, explanation, s, vars)
}

func (tr *tracer) operation(word, op string) {
	tr.state = renderers.TrieLessonState{Operation: op, Word: word, Runes: []rune(word), RuneIndex: -1, ByteOffset: -1, Phase: "begin"}
	start, rootLine, runeLine, lookup, advance, missing, done := "L1", "L2", "L3", "L4", "L7", "", "L10"
	if op == "search" {
		start, rootLine, runeLine, lookup, advance, missing, done = "L11", "L12", "L13", "L14", "L16", "L15", "L18"
	}
	if op == "startsWith" {
		start, rootLine, runeLine, lookup, advance, missing, done = "L20", "L21", "L22", "L23", "L25", "L24", "L27"
	}
	vars := func() map[string]any {
		return map[string]any{"op": op, "word": word, "index": tr.state.RuneIndex, "byteOffset": tr.state.ByteOffset}
	}
	tr.add(start, trace.ActionInit, fmt.Sprintf("Begin %s(%q).", op, word), vars())
	curr := tr.root
	tr.state.ActiveID = "root"
	tr.state.PathIDs = []string{"root"}
	tr.state.Phase = "root"
	tr.add(rootLine, trace.ActionSet, "Start at root.", vars())
	index := 0
	for offset, ch := range word {
		tr.state.RuneIndex = index
		tr.state.ByteOffset = offset
		tr.state.NextID = ""
		tr.state.Created = false
		tr.state.Phase = "rune"
		tr.add(runeLine, trace.ActionVisit, fmt.Sprintf("Read rune %q at rune index %d, byte offset %d.", ch, index, offset), vars())
		child, exists := curr.Children[ch]
		tr.state.NextID = tr.idFor(child)
		tr.state.Phase = "compare"
		tr.b.IncComparisons()
		tr.add(lookup, trace.ActionCompare, fmt.Sprintf("Look for edge %q from %s.", ch, tr.state.ActiveID), vars())
		if !exists {
			if op != "insert" {
				tr.state.Phase = "missing"
				tr.state.ResultKnown = true
				tr.state.Found = false
				tr.add(missing, trace.ActionDone, fmt.Sprintf("Missing edge %q; %s(%q) = false.", ch, op, word), vars())
				return
			}
			child = &Node{Char: ch, Children: make(map[rune]*Node)}
			curr.Children[ch] = child
			tr.state.NextID = tr.idFor(child)
			tr.state.Created = true
			tr.state.Phase = "create"
			tr.b.IncOperations()
			tr.add("L5", trace.ActionPush, fmt.Sprintf("Create edge %q and node %s.", ch, tr.state.NextID), vars())
		} else if op == "insert" {
			tr.state.Phase = "reuse"
			tr.add("L6", trace.ActionVisit, fmt.Sprintf("Reuse existing edge %q; identity stays %s.", ch, tr.state.NextID), vars())
		}
		curr = child
		tr.state.ActiveID = tr.idFor(curr)
		tr.state.NextID = ""
		tr.state.PathIDs = append(tr.state.PathIDs, tr.state.ActiveID)
		tr.state.Phase = "advance"
		tr.b.IncOperations()
		tr.add(advance, trace.ActionVisit, fmt.Sprintf("Follow edge %q to %s.", ch, tr.state.ActiveID), vars())
		index++
	}
	tr.state.RuneIndex = -1
	tr.state.ByteOffset = -1
	tr.state.Created = false
	if op == "insert" {
		curr.IsEnd = true
		tr.b.IncOperations()
		tr.state.Phase = "terminal"
		tr.add("L9", trace.ActionDone, fmt.Sprintf("Mark %s as word end for %q.", tr.state.ActiveID, word), vars())
	} else {
		if op == "search" {
			tr.b.IncComparisons()
		}
		tr.state.ResultKnown = true
		tr.state.Found = op == "startsWith" || curr.IsEnd
	}
	tr.state.Phase = "result"
	explanation := fmt.Sprintf("Insert(%q) complete; word ends here.", word)
	if op == "search" {
		explanation = fmt.Sprintf("Path exists; word ends here = %t. Search(%q) = %t.", curr.IsEnd, word, curr.IsEnd)
	}
	if op == "startsWith" {
		explanation = fmt.Sprintf("Path exists. StartsWith(%q) = true; word-end flag is not required.", word)
	}
	tr.add(done, trace.ActionDone, explanation, vars())
}

// Trace preserves direct API defaults, including the first nonempty variadic prefix.
func Trace(words []string, searchWord string, prefix ...string) []trace.Step {
	if len(words) == 0 {
		words = []string{"go", "golang", "good", "god"}
	}
	if searchWord == "" {
		searchWord = "gold"
	}
	pref := "go"
	if len(prefix) > 0 && prefix[0] != "" {
		pref = prefix[0]
	}
	root := &Node{Children: make(map[rune]*Node)}
	tr := tracer{b: trace.NewBuilder(), root: root, nodeIDs: map[*Node]string{root: "root"}, state: renderers.TrieLessonState{Phase: "init", RuneIndex: -1, ByteOffset: -1}}
	tr.add("L1", trace.ActionInit, "Initialize trie root.", map[string]any{"words": append([]string(nil), words...), "search": searchWord, "prefix": pref})
	for _, word := range words {
		tr.operation(word, "insert")
	}
	tr.operation(searchWord, "search")
	tr.operation(pref, "startsWith")
	steps := tr.b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].Action == trace.ActionCompare {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
