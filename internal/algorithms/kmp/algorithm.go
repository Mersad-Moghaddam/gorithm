package kmp

// ComputeLPS computes the Longest Proper Prefix that is also Suffix array for the pattern.
func ComputeLPS(pattern string) []int {
	m := len(pattern)
	lps := make([]int, m)
	if m == 0 {
		return lps
	}

	length := 0
	i := 1

	for i < m {
		if pattern[i] == pattern[length] {
			length++
			lps[i] = length
			i++
		} else {
			if length != 0 {
				length = lps[length-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	return lps
}

// Search finds the first occurrence of pattern in text using the Knuth-Morris-Pratt algorithm.
// Returns the 0-based starting index, or -1 if not found.
// ponytail: basic KMP with prefix-suffix table; upgrade to Boyer-Moore or Two-Way if sublinear average performance needed.
func Search(text, pattern string) int {
	n := len(text)
	m := len(pattern)

	if m == 0 {
		return 0
	}
	if n < m {
		return -1
	}

	lps := ComputeLPS(pattern)
	i := 0
	j := 0

	for i < n {
		if text[i] == pattern[j] {
			i++
			j++
		}

		if j == m {
			return i - j
		} else if i < n && text[i] != pattern[j] {
			if j != 0 {
				j = lps[j-1]
			} else {
				i++
			}
		}
	}

	return -1
}
