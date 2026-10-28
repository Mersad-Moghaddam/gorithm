package knapsack

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type buildInput struct {
	Weights  []int `json:"weights"`
	Values   []int `json:"values"`
	Capacity *int  `json:"capacity"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "knapsack",
		Name:     "0/1 Knapsack (DP)",
		Category: catalog.CategoryDP,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "A fundamental dynamic programming algorithm that finds the maximum value subset of items that fits within a knapsack of limited weight capacity.",
			TimeComplexity:  "O(n · W)",
			SpaceComplexity: "O(n · W)",
			KeyConcepts:     []string{"2D State Matrix", "Include/Exclude Choice", "Pseudo-Polynomial Time"},
			HowItWorks:      "Constructs an (n+1) x (W+1) dynamic programming table where dp[i][c] represents the maximum value attainable using a subset of the first i items with weight capacity c. For each item and capacity, the algorithm chooses the better of two options: excluding the current item (taking dp[i-1][c]) or, if it fits, including the current item (taking dp[i-1][c-weight] + value). The final cell dp[n][W] holds the optimal solution.",
		},
		DefaultInput: `{"weights": [2, 3, 4, 5], "values": [3, 4, 5, 6], "capacity": 5}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Solve(weights, values []int, capacity int) (int, [][]int) {"},
		{ID: "L2", Number: 2, Code: "    n := len(weights)"},
		{ID: "L3", Number: 3, Code: "    dp := make([][]int, n+1)"},
		{ID: "L4", Number: 4, Code: "    for i := range dp { dp[i] = make([]int, capacity+1) }"},
		{ID: "L5", Number: 5, Code: "    for i := 1; i <= n; i++ {"},
		{ID: "L6", Number: 6, Code: "        w, v := weights[i-1], values[i-1]"},
		{ID: "L7", Number: 7, Code: "        for c := 0; c <= capacity; c++ {"},
		{ID: "L8", Number: 8, Code: "            if w <= c {"},
		{ID: "L9", Number: 9, Code: "                dp[i][c] = max(dp[i-1][c], dp[i-1][c-w]+v)"},
		{ID: "L10", Number: 10, Code: "            } else {"},
		{ID: "L11", Number: 11, Code: "                dp[i][c] = dp[i-1][c]"},
		{ID: "L12", Number: 12, Code: "            }"},
		{ID: "L13", Number: 13, Code: "        }"},
		{ID: "L14", Number: 14, Code: "    }"},
		{ID: "L15", Number: 15, Code: "    return dp[n][capacity], dp"},
		{ID: "L16", Number: 16, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	weights := []int{2, 3, 4, 5}
	values := []int{3, 4, 5, 6}
	capacity := 5

	if len(bytes.TrimSpace(input)) > 0 {
		var custom buildInput
		if bytes.Equal(bytes.TrimSpace(input), []byte("null")) {
			return nil, fmt.Errorf("knapsack input must be an object")
		}
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid knapsack input: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		for _, key := range []string{"weights", "values", "capacity"} {
			if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s cannot be null", key)
			}
		}
		if _, ok := fields["weights"]; ok {
			weights = custom.Weights
		}
		if _, ok := fields["values"]; ok {
			values = custom.Values
		}
		if custom.Capacity != nil {
			capacity = *custom.Capacity
		}
	}
	if len(weights) < 1 || len(weights) > 20 || len(weights) != len(values) {
		return nil, fmt.Errorf("require 1–20 items with equal weight/value lengths")
	}
	if capacity < 0 || capacity > 200 {
		return nil, fmt.Errorf("capacity must be in [0,200]")
	}
	for i, weight := range weights {
		if weight < 0 || weight > 200 || values[i] < 0 || values[i] > 1000000 {
			return nil, fmt.Errorf("item %d requires weight in [0,200] and value in [0,1000000]", i+1)
		}
	}

	return Trace(weights, values, capacity), nil
}

func init() {
	catalog.Register(Lesson{})
}
