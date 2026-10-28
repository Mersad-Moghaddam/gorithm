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
			HowItWorks: "Inserts keys recursively like a standard BST while maintaining node heights. When unwinding recursion, each ancestor checks its balance factor (left height minus right height). If balance factor exceeds [-1, 1], one of four rotation cases restores balance: Left-Left (Right rotation), Right-Right (Left rotation), Left-Right (Left rotation on child, Right rotation on root), or Right-Left (Right rotation on child, Left rotation on root).",
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
	Values []int `json:"values"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	values := []int{10, 20, 30, 40, 50, 25}
	if len(input) > 0 {
		raw := bytes.TrimSpace(input)
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			return nil, fmt.Errorf("AVL input must be an array or object")
		}
		var custom []int
		if raw[0] == '[' {
			if err := json.Unmarshal(raw, &custom); err != nil {
				return nil, fmt.Errorf("invalid AVL values: %w", err)
			}
		} else {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
				return nil, fmt.Errorf("AVL input must be an array or object")
			}
			if v, exists := fields["values"]; exists {
				if bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
					return nil, fmt.Errorf("AVL values must be an integer array")
				}
				if err := json.Unmarshal(v, &custom); err != nil {
					return nil, fmt.Errorf("invalid AVL values: %w", err)
				}
			}
		}
		if len(custom) > 32 {
			return nil, fmt.Errorf("AVL supports at most 32 values")
		}
		for _, v := range custom {
			if v < -1000000 || v > 1000000 {
				return nil, fmt.Errorf("AVL values must be within ±1000000")
			}
		}
		if len(custom) > 0 {
			values = custom
		}
	}
	return Trace(values), nil
}

func init() {
	catalog.Register(Lesson{})
}
