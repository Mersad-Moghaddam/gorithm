package bst

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "bst",
		Name:     "Binary Search Tree (BST)",
		Category: catalog.CategoryTrees,
		Renderer: "tree",
		Documentation: catalog.Documentation{
			Summary:         "A node-based binary tree data structure where each node stores a key greater than all keys in its left subtree and smaller than all keys in its right subtree.",
			TimeComplexity:  "O(log n) avg, O(n) worst",
			SpaceComplexity: "O(n)",
			KeyConcepts:     []string{"Left < Root < Right Invariant", "Recursive Descent", "In-Order Traversal"},
			HowItWorks:      "Maintains a sorted binary tree structure. Insertion and search recursively traverse left if the target is smaller or right if greater. Deletion handles three cases: leaf removal, single-child bypass, or replacing a two-child node with its in-order successor (minimum in the right subtree).",
		},
		DefaultInput: `{"initial": [50, 30, 70, 20, 40, 60, 80], "ops": [{"type": "insert", "value": 35}, {"type": "search", "value": 40}, {"type": "delete", "value": 30}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func (t *BST) Insert(val int) {"},
		{ID: "L2", Number: 2, Code: "    t.Root = insert(t.Root, val)"},
		{ID: "L3", Number: 3, Code: "}"},
		{ID: "L4", Number: 4, Code: "func insert(node *Node, val int) *Node {"},
		{ID: "L5", Number: 5, Code: "    if node == nil { return &Node{Val: val} }"},
		{ID: "L6", Number: 6, Code: "    if val < node.Val { node.Left = insert(node.Left, val) }"},
		{ID: "L7", Number: 7, Code: "    else if val > node.Val { node.Right = insert(node.Right, val) }"},
		{ID: "L8", Number: 8, Code: "    return node"},
		{ID: "L9", Number: 9, Code: "}"},
		{ID: "L10", Number: 10, Code: "func (t *BST) Search(val int) bool {"},
		{ID: "L11", Number: 11, Code: "    return search(t.Root, val)"},
		{ID: "L12", Number: 12, Code: "}"},
		{ID: "L13", Number: 13, Code: "func search(node *Node, val int) bool {"},
		{ID: "L14", Number: 14, Code: "    if node == nil { return false }"},
		{ID: "L15", Number: 15, Code: "    if val == node.Val { return true }"},
		{ID: "L16", Number: 16, Code: "    if val < node.Val { return search(node.Left, val) }"},
		{ID: "L17", Number: 17, Code: "    return search(node.Right, val)"},
		{ID: "L18", Number: 18, Code: "}"},
		{ID: "L19", Number: 19, Code: "func (t *BST) Delete(val int) {"},
		{ID: "L20", Number: 20, Code: "    t.Root = deleteNode(t.Root, val)"},
		{ID: "L21", Number: 21, Code: "}"},
		{ID: "L22", Number: 22, Code: "func deleteNode(node *Node, val int) *Node {"},
		{ID: "L23", Number: 23, Code: "    if node == nil { return nil }"},
		{ID: "L24", Number: 24, Code: "    if val < node.Val { node.Left = deleteNode(node.Left, val) }"},
		{ID: "L25", Number: 25, Code: "    else if val > node.Val { node.Right = deleteNode(node.Right, val) }"},
		{ID: "L26", Number: 26, Code: "    else {"},
		{ID: "L27", Number: 27, Code: "        if node.Left == nil { return node.Right }"},
		{ID: "L28", Number: 28, Code: "        if node.Right == nil { return node.Left }"},
		{ID: "L29", Number: 29, Code: "        min := minNode(node.Right)"},
		{ID: "L30", Number: 30, Code: "        node.Val = min.Val"},
		{ID: "L31", Number: 31, Code: "        node.Right = deleteNode(node.Right, min.Val)"},
		{ID: "L32", Number: 32, Code: "    }"},
		{ID: "L33", Number: 33, Code: "    return node"},
		{ID: "L34", Number: 34, Code: "}"},
	}
}

type buildInput struct {
	Initial []int       `json:"initial"`
	Ops     []Operation `json:"ops"`
}

func defaultInitial() []int {
	return []int{50, 30, 70, 20, 40, 60, 80}
}

func defaultOps() []Operation {
	return []Operation{
		{Type: "insert", Value: 35},
		{Type: "search", Value: 40},
		{Type: "delete", Value: 30},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(defaultInitial(), defaultOps()), nil
	}
	if bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("input cannot be null")
	}

	var custom buildInput
	if err := json.Unmarshal(trimmed, &custom); err != nil {
		return nil, fmt.Errorf("malformed JSON input: %w", err)
	}

	if len(custom.Ops) == 0 {
		return nil, fmt.Errorf("ops cannot be empty")
	}

	for i, op := range custom.Ops {
		switch op.Type {
		case "insert", "search", "delete":
		default:
			return nil, fmt.Errorf("op %d: invalid operation type %q", i, op.Type)
		}
	}

	initial := custom.Initial
	if initial == nil {
		initial = defaultInitial()
	}

	return Trace(initial, custom.Ops), nil
}

func init() {
	catalog.Register(Lesson{})
}
