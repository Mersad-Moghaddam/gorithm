package minimax

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
		ID:       "minimax",
		Name:     "Minimax & Alpha-Beta Pruning",
		Category: catalog.CategoryAI,
		Renderer: "tree",
		Documentation: catalog.Documentation{
			Summary:         "An adversarial search algorithm determining optimal moves in zero-sum games. Alpha-Beta pruning discards branches that cannot influence the final outcome, reducing complexity from O(b^d) down to O(b^(d/2)).",
			TimeComplexity:  "O(b^(d/2)) best, O(b^d) worst",
			SpaceComplexity: "O(b · d)",
			Difficulty:      "Hard",
			KeyConcepts:     []string{"Adversarial Search", "Alpha-Beta Cutoffs", "Game Theory", "Zero-Sum"},
			HowItWorks:      "Minimax alternates between the Maximizer (seeking highest score) and Minimizer (seeking lowest score). Alpha maintains the best guaranteed score for Maximizer, while Beta maintains the best guaranteed score for Minimizer. If Beta <= Alpha at any point, the remaining sibling branches are pruned because rational play makes them unreachable.",
		},
		DefaultInput: `{"scores": [-1, 4, 2, 6, -3, -5, 0, 7], "branching": 2}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func minimax(node, depth, alpha, beta, isMax) int {"},
		{ID: "L2", Number: 2, Code: "    if isTerminal(node) {"},
		{ID: "L3", Number: 3, Code: "        return node.Value"},
		{ID: "L4", Number: 4, Code: "    }"},
		{ID: "L5", Number: 5, Code: "    if isMax {"},
		{ID: "L6", Number: 6, Code: "        maxEval := -Infinity"},
		{ID: "L7", Number: 7, Code: "        for _, child := range node.Children {"},
		{ID: "L8", Number: 8, Code: "            eval := minimax(child, depth+1, alpha, beta, false)"},
		{ID: "L9", Number: 9, Code: "            maxEval = max(maxEval, eval)"},
		{ID: "L10", Number: 10, Code: "            alpha = max(alpha, eval)"},
		{ID: "L11", Number: 11, Code: "            if beta <= alpha { break } // Beta cutoff"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "        return maxEval"},
		{ID: "L14", Number: 14, Code: "    } else {"},
		{ID: "L15", Number: 15, Code: "        minEval := +Infinity"},
		{ID: "L16", Number: 16, Code: "        for _, child := range node.Children {"},
		{ID: "L17", Number: 17, Code: "            eval := minimax(child, depth+1, alpha, beta, true)"},
		{ID: "L18", Number: 18, Code: "            minEval = min(minEval, eval)"},
		{ID: "L19", Number: 19, Code: "            beta = min(beta, eval)"},
		{ID: "L20", Number: 20, Code: "            if beta <= alpha { break } // Alpha cutoff"},
		{ID: "L21", Number: 21, Code: "        }"},
		{ID: "L22", Number: 22, Code: "        return minEval"},
		{ID: "L23", Number: 23, Code: "    }"},
		{ID: "L24", Number: 24, Code: "}"},
	}
}

type buildInput struct {
	Scores    []int `json:"scores"`
	Branching int   `json:"branching"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	scores := []int{-1, 4, 2, 6, -3, -5, 0, 7}
	branching := 2

	if len(input) > 0 {
		raw := bytes.TrimSpace(input)
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			return nil, fmt.Errorf("minimax: expected scores array or object")
		}
		var scoreRaw json.RawMessage
		if raw[0] == '[' {
			scoreRaw = raw
		} else {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				return nil, fmt.Errorf("minimax: %w", err)
			}
			scoreRaw = fields["scores"]
			if v, ok := fields["branching"]; ok {
				if bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
					return nil, fmt.Errorf("minimax: branching must be integer")
				}
				if err := json.Unmarshal(v, &branching); err != nil {
					return nil, err
				}
			}
		}
		if scoreRaw != nil {
			var values []*int
			if err := json.Unmarshal(scoreRaw, &values); err != nil {
				return nil, err
			}
			if len(values) == 0 {
				return nil, fmt.Errorf("minimax: scores require 1–64 integers")
			}
			scores = make([]int, len(values))
			for i, v := range values {
				if v == nil {
					return nil, fmt.Errorf("minimax: score %d must be integer", i)
				}
				scores[i] = *v
			}
		}
	}
	if err := validateMinimaxInput(scores, branching); err != nil {
		return nil, err
	}

	return Trace(scores, branching), nil
}

func validateMinimaxInput(scores []int, branching int) error {
	if len(scores) < 1 || len(scores) > 64 {
		return fmt.Errorf("minimax: scores require 1–64 integers")
	}
	if branching < 2 || branching > 8 {
		return fmt.Errorf("minimax: branching must be 2–8")
	}
	for _, v := range scores {
		if v < -999999 || v > 999999 {
			return fmt.Errorf("minimax: scores must be within [-999999,999999]")
		}
	}
	return nil
}

func init() {
	catalog.Register(Lesson{})
}
