package avl

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
		ID:       "avl",
		Name:     "AVL Tree (Self-Balancing)",
		Category: catalog.CategoryTrees,
		Renderer: "tree",
		Documentation: catalog.Documentation{
			Summary:         "A self-balancing binary search tree maintaining a maximum height difference of 1 between left and right subtrees via single and double rotations.",
			TimeComplexity:  "O(log n)",
			SpaceComplexity: "O(n)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Balance Factor [-1, 0, 1]",
				"Single & Double Rotations",
				"Height-Balanced Invariant",
			},
			HowItWorks: "Inserts keys recursively like a standard BST while maintaining node heights. When unwinding recursion, each ancestor checks its balance factor (bf = left.Height - right.Height). If balance factor exceeds [-1, 1], one of four rotation cases restores balance while preserving BST order: Left-Left (Right rotation where pivot rises and root descends right), Right-Right (Left rotation where pivot rises and root descends left), Left-Right (Left rotation on child, then Right rotation on root), or Right-Left (Right rotation on child, then Left rotation on root). Rotations maintain in-order key ordering because the inner transferred subtree sits strictly between the pivot and root values.",
		},
		DefaultInput: "[10, 20, 30, 40, 50, 25]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func insert(node *Node, val int) *Node {"},
		{ID: "L2", Number: 2, Code: "    if node == nil { return &Node{Val: val, Height: 1} }"},
		{ID: "L3", Number: 3, Code: "    if val < node.Val { node.Left = insert(node.Left, val) }"},
		{ID: "L4", Number: 4, Code: "    else if val > node.Val { node.Right = insert(node.Right, val) }"},
		{ID: "L5", Number: 5, Code: "    else { return node }"},
		{ID: "L6", Number: 6, Code: "    node.Height = 1 + max(height(node.Left), height(node.Right))"},
		{ID: "L7", Number: 7, Code: "    bf := balanceFactor(node)"},
		{ID: "L8", Number: 8, Code: "    if bf > 1 && val < node.Left.Val { return rightRotate(node) } // LL"},
		{ID: "L9", Number: 9, Code: "    if bf < -1 && val > node.Right.Val { return leftRotate(node) } // RR"},
		{ID: "L10", Number: 10, Code: "    if bf > 1 && val > node.Left.Val { // LR"},
		{ID: "L11", Number: 11, Code: "        node.Left = leftRotate(node.Left); return rightRotate(node)"},
		{ID: "L12", Number: 12, Code: "    }"},
		{ID: "L13", Number: 13, Code: "    if bf < -1 && val < node.Right.Val { // RL"},
		{ID: "L14", Number: 14, Code: "        node.Right = rightRotate(node.Right); return leftRotate(node)"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    return node"},
		{ID: "L17", Number: 17, Code: "}"},
		{ID: "L18", Number: 18, Code: "func rightRotate(y *Node) *Node {"},
		{ID: "L19", Number: 19, Code: "    x := y.Left; y.Left = x.Right; x.Right = y"},
		{ID: "L20", Number: 20, Code: "    updateHeight(y); updateHeight(x); return x"},
		{ID: "L21", Number: 21, Code: "}"},
		{ID: "L22", Number: 22, Code: "func leftRotate(x *Node) *Node {"},
		{ID: "L23", Number: 23, Code: "    y := x.Right; x.Right = y.Left; y.Left = x"},
		{ID: "L24", Number: 24, Code: "    updateHeight(x); updateHeight(y); return y"},
		{ID: "L25", Number: 25, Code: "}"},
	}
}

type buildInput struct {
	Values *[]int `json:"values"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	if input == nil {
		return Trace([]int{10, 20, 30, 40, 50, 25}), nil
	}

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("avl: input cannot be empty")
	}
	if string(trimmed) == "null" {
		return nil, fmt.Errorf("avl: input cannot be null")
	}

	if trimmed[0] == '[' {
		var arr []int
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return nil, fmt.Errorf("avl: invalid integer array: %w", err)
		}
		if len(arr) == 0 {
			return nil, fmt.Errorf("avl: values array cannot be empty")
		}
		return Trace(arr), nil
	}

	if trimmed[0] == '{' {
		var custom buildInput
		dec := json.NewDecoder(bytes.NewReader(trimmed))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&custom); err != nil {
			return nil, fmt.Errorf("avl: invalid input object: %w", err)
		}
		if custom.Values == nil {
			return nil, fmt.Errorf("avl: values field cannot be null or missing")
		}
		if len(*custom.Values) == 0 {
			return nil, fmt.Errorf("avl: values array cannot be empty")
		}
		return Trace(*custom.Values), nil
	}

	return nil, fmt.Errorf("avl: input must be a JSON array or object with values field")
}

func init() {
	catalog.Register(Lesson{})
}
