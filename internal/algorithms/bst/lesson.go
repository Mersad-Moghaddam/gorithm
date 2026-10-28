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

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	initial := []int{50, 30, 70, 20, 40, 60, 80}
	ops := []Operation{
		{Type: "insert", Value: 35},
		{Type: "search", Value: 40},
		{Type: "delete", Value: 30},
	}

	if len(input) > 0 {
		raw := bytes.TrimSpace(input)
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			return nil, fmt.Errorf("BST input must be object or integer array")
		}
		if raw[0] == '[' {
			var values []int
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, err
			}
			if len(values) > 0 {
				initial = values
			}
		} else {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				return nil, err
			}
			if fields == nil {
				return nil, fmt.Errorf("BST input must be object")
			}
			if value, ok := fields["initial"]; ok {
				var values []int
				if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
					return nil, fmt.Errorf("initial cannot be null")
				}
				if err := json.Unmarshal(value, &values); err != nil {
					return nil, err
				}
				if len(values) > 0 {
					initial = values
				}
			}
			if value, ok := fields["ops"]; ok {
				var entries []struct {
					Type  string `json:"type"`
					Value *int   `json:"value"`
				}
				if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
					return nil, fmt.Errorf("ops cannot be null")
				}
				if err := json.Unmarshal(value, &entries); err != nil {
					return nil, err
				}
				if len(entries) > 0 {
					ops = make([]Operation, len(entries))
					for i, e := range entries {
						if e.Value == nil {
							return nil, fmt.Errorf("operation %d requires value", i)
						}
						ops[i] = Operation{Type: e.Type, Value: *e.Value}
					}
				}
			}
		}
	}
	if len(initial) > 32 || len(ops) > 64 {
		return nil, fmt.Errorf("BST permits at most 32 initial values and 64 operations")
	}
	for _, v := range initial {
		if v < -1000000 || v > 1000000 {
			return nil, fmt.Errorf("initial value outside [-1000000,1000000]")
		}
	}
	for _, op := range ops {
		if op.Type != "insert" && op.Type != "search" && op.Type != "delete" {
			return nil, fmt.Errorf("unknown BST operation %q", op.Type)
		}
		if op.Value < -1000000 || op.Value > 1000000 {
			return nil, fmt.Errorf("operation value outside [-1000000,1000000]")
		}
	}

	return Trace(initial, ops), nil
}

func init() {
	catalog.Register(Lesson{})
}
