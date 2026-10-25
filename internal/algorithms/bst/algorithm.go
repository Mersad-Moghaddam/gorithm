package bst

// Node represents a binary search tree node.
type Node struct {
	Val   int
	Left  *Node
	Right *Node
}

// BST represents a binary search tree data structure.
// ponytail: unbalanced recursive BST; AVL or Red-Black self-balancing needed for guaranteed O(log n) height.
type BST struct {
	Root *Node
}

// New creates an empty binary search tree.
func New() *BST {
	return &BST{}
}

// Insert adds val into the BST preserving the left < root < right invariant.
func (t *BST) Insert(val int) {
	t.Root = insert(t.Root, val)
}

func insert(node *Node, val int) *Node {
	if node == nil {
		return &Node{Val: val}
	}
	if val < node.Val {
		node.Left = insert(node.Left, val)
	} else if val > node.Val {
		node.Right = insert(node.Right, val)
	}
	return node
}

// Search checks if val exists in the BST.
func (t *BST) Search(val int) bool {
	return search(t.Root, val)
}

func search(node *Node, val int) bool {
	if node == nil {
		return false
	}
	if val == node.Val {
		return true
	} else if val < node.Val {
		return search(node.Left, val)
	}
	return search(node.Right, val)
}

// Delete removes val from the BST if present.
func (t *BST) Delete(val int) {
	t.Root = deleteNode(t.Root, val)
}

func deleteNode(node *Node, val int) *Node {
	if node == nil {
		return nil
	}
	if val < node.Val {
		node.Left = deleteNode(node.Left, val)
	} else if val > node.Val {
		node.Right = deleteNode(node.Right, val)
	} else {
		if node.Left == nil {
			return node.Right
		}
		if node.Right == nil {
			return node.Left
		}
		min := minNode(node.Right)
		node.Val = min.Val
		node.Right = deleteNode(node.Right, min.Val)
	}
	return node
}

func minNode(node *Node) *Node {
	curr := node
	for curr.Left != nil {
		curr = curr.Left
	}
	return curr
}

// InOrder returns the tree keys via in-order traversal (sorted).
func (t *BST) InOrder() []int {
	var res []int
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		walk(n.Left)
		res = append(res, n.Val)
		walk(n.Right)
	}
	walk(t.Root)
	return res
}
