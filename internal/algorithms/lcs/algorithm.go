package lcs

// Solve computes the Longest Common Subsequence of s1 and s2 and returns the LCS string and DP matrix.
// ponytail: 2D table O(m*n) space retained for visualization and backtracking; 2-row buffer O(min(m, n)) space possible if only length needed.
func Solve(s1, s2 string) (string, [][]int) {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] >= dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	res := make([]byte, dp[m][n])
	idx := dp[m][n] - 1
	i, j := m, n
	for i > 0 && j > 0 {
		if s1[i-1] == s2[j-1] {
			res[idx] = s1[i-1]
			idx--
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return string(res), dp
}
