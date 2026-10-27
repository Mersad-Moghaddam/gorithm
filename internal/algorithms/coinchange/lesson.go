package coinchange

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type buildInput struct {
	Coins  *[]int `json:"coins"`
	Amount *int   `json:"amount"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "coinchange",
		Name:     "Coin Change (DP)",
		Category: catalog.CategoryDP,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Finds the minimum number of coins needed to make up a given amount using bottom-up dynamic programming tabulation.",
			TimeComplexity:  "O(amount · |coins|)",
			SpaceComplexity: "O(amount)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Unbounded Knapsack",
				"Overlapping Subproblems",
				"Bottom-up Tabulation",
				"Min Optimization",
			},
			HowItWorks: "Constructs a 1D DP table of size amount+1 where dp[i] stores the fewest coins required to form amount i. With dp[0] = 0 and all other cells initialized to infinity, each subproblem i is solved by evaluating each available coin c <= i: if using coin c yields fewer coins (dp[i-c] + 1 < dp[i]), dp[i] is updated. Once the table is populated, dp[amount] provides the minimum coins, or -1 if the amount cannot be made.",
		},
		DefaultInput: `{"coins": [1, 2, 5], "amount": 11}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func MinCoins(coins []int, amount int) (int, []int) {"},
		{ID: "L2", Number: 2, Code: "    if amount < 0 { return -1, nil }"},
		{ID: "L3", Number: 3, Code: "    dp := make([]int, amount+1)"},
		{ID: "L4", Number: 4, Code: "    for i := 1; i <= amount; i++ { dp[i] = amount + 1 }"},
		{ID: "L5", Number: 5, Code: "    for i := 1; i <= amount; i++ {"},
		{ID: "L6", Number: 6, Code: "        for _, c := range coins {"},
		{ID: "L7", Number: 7, Code: "            if c <= i && dp[i-c]+1 < dp[i] {"},
		{ID: "L8", Number: 8, Code: "                dp[i] = dp[i-c] + 1"},
		{ID: "L9", Number: 9, Code: "            }"},
		{ID: "L10", Number: 10, Code: "        }"},
		{ID: "L11", Number: 11, Code: "    }"},
		{ID: "L12", Number: 12, Code: "    if dp[amount] > amount { return -1, dp }"},
		{ID: "L13", Number: 13, Code: "    return dp[amount], dp"},
		{ID: "L14", Number: 14, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	coins := []int{1, 2, 5}
	amount := 11

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(coins, amount), nil
	}

	if bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("input cannot be null")
	}

	if trimmed[0] != '{' {
		return nil, fmt.Errorf("malformed JSON input: expected JSON object")
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &rawMap); err != nil {
		return nil, fmt.Errorf("malformed JSON input: %w", err)
	}

	if rawCoins, ok := rawMap["coins"]; ok && bytes.Equal(bytes.TrimSpace(rawCoins), []byte("null")) {
		return nil, fmt.Errorf("coins cannot be null")
	}
	if rawAmount, ok := rawMap["amount"]; ok && bytes.Equal(bytes.TrimSpace(rawAmount), []byte("null")) {
		return nil, fmt.Errorf("amount cannot be null")
	}

	var custom buildInput
	if err := json.Unmarshal(trimmed, &custom); err != nil {
		return nil, fmt.Errorf("malformed JSON input: %w", err)
	}

	if custom.Coins != nil {
		if len(*custom.Coins) == 0 {
			return nil, fmt.Errorf("coins cannot be empty")
		}
		for _, c := range *custom.Coins {
			if c <= 0 {
				return nil, fmt.Errorf("coin denominations must be positive, got %d", c)
			}
		}
		coins = *custom.Coins
	}

	if custom.Amount != nil {
		if *custom.Amount < 0 {
			return nil, fmt.Errorf("amount cannot be negative: %d", *custom.Amount)
		}
		amount = *custom.Amount
	}

	return Trace(coins, amount), nil
}

func init() {
	catalog.Register(Lesson{})
}
