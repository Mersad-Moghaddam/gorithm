package bubble

func Sort(arr []int) []int {
	n := len(arr)
	res := make([]int, n)
	copy(res, arr)

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if res[j] > res[j+1] {
				res[j], res[j+1] = res[j+1], res[j]
			}
		}
	}
	return res
}
