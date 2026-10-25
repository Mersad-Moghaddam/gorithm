package minimax

import (
	"fmt"
	"math"
)

const (
	Infinity    = math.MaxInt32
	NegInfinity = math.MinInt32
)

// Node represents a state node in a minimax game tree.
// ponytail: static tree evaluation with alpha-beta cutoffs; transposition tables and iterative deepening needed for unbounded state spaces.
type Node struct {
	ID       string  `json:"id"`
	Val      int     `json:"val"`
	IsMax    bool    `json:"isMax"`
	Depth    int     `json:"depth"`
	Children []*Node `json:"children"`
	Parent   *Node   `json:"-"`
}

// BuildTree constructs a game tree from terminal scores and branching factor.
func BuildTree(scores []int, branching int) *Node {
	if len(scores) == 0 {
		return &Node{ID: "root", Val: 0, IsMax: true}
	}
	if branching <= 1 {
		branching = 2
	}
	return buildSubtree(scores, branching, 0, true, "n0")
}

func buildSubtree(scores []int, branching int, depth int, isMax bool, id string) *Node {
	node := &Node{
		ID:    id,
		Depth: depth,
		IsMax: isMax,
	}
	if len(scores) <= 1 {
		node.Val = scores[0]
		return node
	}

	chunkSize := (len(scores) + branching - 1) / branching
	for i := 0; i < branching; i++ {
		start := i * chunkSize
		if start >= len(scores) {
			break
		}
		end := start + chunkSize
		if end > len(scores) {
			end = len(scores)
		}
		childScores := scores[start:end]
		childID := fmt.Sprintf("%s-%d", id, i+1)
		child := buildSubtree(childScores, branching, depth+1, !isMax, childID)
		child.Parent = node
		node.Children = append(node.Children, child)
	}

	return node
}

// Minimax evaluates the game tree node using alpha-beta pruning.
func Minimax(node *Node, alpha, beta int, isMax bool) int {
	if len(node.Children) == 0 {
		return node.Val
	}

	if isMax {
		maxEval := NegInfinity
		for _, child := range node.Children {
			eval := Minimax(child, alpha, beta, false)
			if eval > maxEval {
				maxEval = eval
			}
			if eval > alpha {
				alpha = eval
			}
			if beta <= alpha {
				break
			}
		}
		node.Val = maxEval
		return maxEval
	}

	minEval := Infinity
	for _, child := range node.Children {
		eval := Minimax(child, alpha, beta, true)
		if eval < minEval {
			minEval = eval
		}
		if eval < beta {
			beta = eval
		}
		if beta <= alpha {
			break
		}
	}
	node.Val = minEval
	return minEval
}

// Solve returns the optimal value and best child node move for the root.
func Solve(root *Node) (int, *Node) {
	if root == nil {
		return 0, nil
	}
	if len(root.Children) == 0 {
		return root.Val, root
	}

	alpha := NegInfinity
	beta := Infinity
	var bestChild *Node

	if root.IsMax {
		bestVal := NegInfinity
		for _, child := range root.Children {
			eval := Minimax(child, alpha, beta, false)
			if eval > bestVal {
				bestVal = eval
				bestChild = child
			}
			if eval > alpha {
				alpha = eval
			}
			if beta <= alpha {
				break
			}
		}
		root.Val = bestVal
		return bestVal, bestChild
	}

	bestVal := Infinity
	for _, child := range root.Children {
		eval := Minimax(child, alpha, beta, true)
		if eval < bestVal {
			bestVal = eval
			bestChild = child
		}
		if eval < beta {
			beta = eval
		}
		if beta <= alpha {
			break
		}
	}
	root.Val = bestVal
	return bestVal, bestChild
}
