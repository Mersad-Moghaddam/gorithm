package selection

func Sort(arr []int) []int {
	n := len(arr)
	res := make([]int, n)
	copy(res, arr)

	for i := 0; i < n-1; i++ {
		minIdx := i
		for j := i + 1; j < n; j++ {
			if res[j] < res[minIdx] {
				minIdx = j
			}
		}
		res[i], res[minIdx] = res[minIdx], res[i]
	}
	return res
}
