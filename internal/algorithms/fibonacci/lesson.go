package fibonacci

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "fibonacci",
		Name:     "Fibonacci (DP)",
		Category: catalog.CategoryDP,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Computes Fibonacci numbers iteratively using dynamic programming tabulation to store and reuse previous subproblem results.",
			TimeComplexity:  "O(n)",
			SpaceComplexity: "O(n)",
			KeyConcepts:     []string{"Optimal Substructure", "Overlapping Subproblems", "Tabulation", "Bottom-Up DP"},
			HowItWorks:      "Allocates a 1D DP table of size n+1. Sets base cases dp[0] = 0 and dp[1] = 1. Iterates from i = 2 up to n, computing dp[i] = dp[i-1] + dp[i-2] in O(1) time per entry by looking up already solved subproblems.",
		},
		DefaultInput: `{"n": 8}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Compute(n int) (int, []int) {"},
		{ID: "L2", Number: 2, Code: "    if n <= 0 {"},
		{ID: "L3", Number: 3, Code: "        return 0, []int{0}"},
		{ID: "L4", Number: 4, Code: "    }"},
		{ID: "L5", Number: 5, Code: "    dp := make([]int, n+1)"},
		{ID: "L6", Number: 6, Code: "    dp[0] = 0"},
		{ID: "L7", Number: 7, Code: "    dp[1] = 1"},
		{ID: "L8", Number: 8, Code: "    for i := 2; i <= n; i++ {"},
		{ID: "L9", Number: 9, Code: "        dp[i] = dp[i-1] + dp[i-2]"},
		{ID: "L10", Number: 10, Code: "    }"},
		{ID: "L11", Number: 11, Code: "    return dp[n], dp"},
		{ID: "L12", Number: 12, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	n := 8
	if len(input) > 0 {
		var payload struct {
			N *int `json:"n"`
		}
		if err := json.Unmarshal(input, &payload); err == nil && payload.N != nil {
			n = *payload.N
		} else {
			var directInt int
			if err := json.Unmarshal(input, &directInt); err == nil {
				n = directInt
			}
		}
	}
	return Trace(n), nil
}

func init() {
	catalog.Register(Lesson{})
}
