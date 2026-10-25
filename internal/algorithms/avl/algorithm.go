package avl

// Node represents a node in the AVL tree.
type Node struct {
	Val    int
	Height int
	Left   *Node
	Right  *Node
}

// AVL represents a self-balancing binary search tree.
// ponytail: recursive pointer AVL tree; parent pointers and deletion rebalancing needed for full mutable dictionary.
type AVL struct {
	Root *Node
}

// New creates an empty AVL tree.
func New() *AVL {
	return &AVL{}
}

func height(n *Node) int {
	if n == nil {
		return 0
	}
	return n.Height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func updateHeight(n *Node) {
	if n != nil {
		n.Height = 1 + max(height(n.Left), height(n.Right))
	}
}

// BalanceFactor returns height(Left) - height(Right).
func BalanceFactor(n *Node) int {
	if n == nil {
		return 0
	}
	return height(n.Left) - height(n.Right)
}

// rightRotate performs a Right (LL) rotation around y.
func rightRotate(y *Node) *Node {
	x := y.Left
	t2 := x.Right

	x.Right = y
	y.Left = t2

	updateHeight(y)
	updateHeight(x)
	return x
}

// leftRotate performs a Left (RR) rotation around x.
func leftRotate(x *Node) *Node {
	y := x.Right
	t2 := y.Left

	y.Left = x
	x.Right = t2

	updateHeight(x)
	updateHeight(y)
	return y
}

// Insert adds val into the AVL tree, balancing via rotations.
func (t *AVL) Insert(val int) {
	t.Root = insert(t.Root, val)
}

func insert(node *Node, val int) *Node {
	if node == nil {
		return &Node{Val: val, Height: 1}
	}

	if val < node.Val {
		node.Left = insert(node.Left, val)
	} else if val > node.Val {
		node.Right = insert(node.Right, val)
	} else {
		return node
	}

	updateHeight(node)
	bf := BalanceFactor(node)

	// Left-Left (LL) Case -> Right Rotation
	if bf > 1 && val < node.Left.Val {
		return rightRotate(node)
	}

	// Right-Right (RR) Case -> Left Rotation
	if bf < -1 && val > node.Right.Val {
		return leftRotate(node)
	}

	// Left-Right (LR) Case -> Left Rotate Left Child, then Right Rotate Root
	if bf > 1 && val > node.Left.Val {
		node.Left = leftRotate(node.Left)
		return rightRotate(node)
	}

	// Right-Left (RL) Case -> Right Rotate Right Child, then Left Rotate Root
	if bf < -1 && val < node.Right.Val {
		node.Right = rightRotate(node.Right)
		return leftRotate(node)
	}

	return node
}

// Search checks if val exists in the AVL tree.
func (t *AVL) Search(val int) bool {
	curr := t.Root
	for curr != nil {
		if val == curr.Val {
			return true
		} else if val < curr.Val {
			curr = curr.Left
		} else {
			curr = curr.Right
		}
	}
	return false
}

// InOrder returns keys in ascending order.
func (t *AVL) InOrder() []int {
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
