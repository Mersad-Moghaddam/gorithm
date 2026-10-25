package fibonacci

// Compute calculates the n-th Fibonacci number and returns the result with the full DP table.
// ponytail: O(n) table space retained for visualization; O(1) two-variable rotation possible if only final value needed.
func Compute(n int) (int, []int) {
	if n <= 0 {
		return 0, []int{0}
	}
	dp := make([]int, n+1)
	dp[0] = 0
	dp[1] = 1
	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}
	return dp[n], dp
}
