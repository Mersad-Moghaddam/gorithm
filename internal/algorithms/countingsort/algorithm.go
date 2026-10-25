package countingsort

// ponytail: non-negative integers only; upgrade to offset shift if negatives needed.
func Sort(arr []int) []int {
	n := len(arr)
	if n <= 1 {
		res := make([]int, n)
		copy(res, arr)
		return res
	}

	max := arr[0]
	for i := 1; i < n; i++ {
		if arr[i] > max {
			max = arr[i]
		}
	}

	count := make([]int, max+1)
	for i := 0; i < n; i++ {
		count[arr[i]]++
	}

	for i := 1; i <= max; i++ {
		count[i] += count[i-1]
	}

	output := make([]int, n)
	for i := n - 1; i >= 0; i-- {
		output[count[arr[i]]-1] = arr[i]
		count[arr[i]]--
	}

	return output
}
