package radixsort

// Sort performs Least-Significant-Digit (LSD) Radix Sort on non-negative integers.
func Sort(arr []int) []int {
	if len(arr) <= 1 {
		res := make([]int, len(arr))
		copy(res, arr)
		return res
	}

	res := make([]int, len(arr))
	copy(res, arr)

	maxVal := res[0]
	for _, v := range res {
		if v > maxVal {
			maxVal = v
		}
	}

	for exp := 1; maxVal/exp > 0; exp *= 10 {
		countingSortByDigit(res, exp)
	}

	return res
}

func countingSortByDigit(arr []int, exp int) {
	n := len(arr)
	output := make([]int, n)
	var count [10]int

	for i := 0; i < n; i++ {
		digit := (arr[i] / exp) % 10
		count[digit]++
	}

	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}

	for i := n - 1; i >= 0; i-- {
		digit := (arr[i] / exp) % 10
		output[count[digit]-1] = arr[i]
		count[digit]--
	}

	for i := 0; i < n; i++ {
		arr[i] = output[i]
	}
}
