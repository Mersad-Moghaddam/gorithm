package coinchange

// MinCoins finds minimum coins needed to make amount and returns the DP table.
// ponytail: 1D tabulation O(amount) space; early-exit or BFS possible if only shortest path needed.
func MinCoins(coins []int, amount int) (int, []int) {
	if amount < 0 {
		return -1, nil
	}
	dp := make([]int, amount+1)
	for i := 1; i <= amount; i++ {
		dp[i] = amount + 1
	}

	for i := 1; i <= amount; i++ {
		for _, c := range coins {
			if c <= i && dp[i-c]+1 < dp[i] {
				dp[i] = dp[i-c] + 1
			}
		}
	}

	if dp[amount] > amount {
		return -1, dp
	}
	return dp[amount], dp
}
