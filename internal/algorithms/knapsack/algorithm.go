package knapsack

// Solve computes 0/1 Knapsack maximum value and returns the completed DP table.
// ponytail: 2D table O(n*W) space retained for visualization; 1D row buffer possible if only max value needed.
func Solve(weights, values []int, capacity int) (int, [][]int) {
	n := min(len(weights), len(values))
	if capacity < 0 {
		capacity = 0
	}
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, capacity+1)
	}

	for i := 1; i <= n; i++ {
		w := weights[i-1]
		v := values[i-1]
		for c := 0; c <= capacity; c++ {
			if w <= c {
				dp[i][c] = max(dp[i-1][c], dp[i-1][c-w]+v)
			} else {
				dp[i][c] = dp[i-1][c]
			}
		}
	}

	return dp[n][capacity], dp
}
