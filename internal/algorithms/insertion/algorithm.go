package insertion

// ponytail: in-place sort on copy; upgrade to cmp.Ordered generic when multi-type needed.
func Sort(arr []int) []int {
	n := len(arr)
	res := make([]int, n)
	copy(res, arr)

	for i := 1; i < n; i++ {
		key := res[i]
		j := i - 1
		for j >= 0 && res[j] > key {
			res[j+1] = res[j]
			j--
		}
		res[j+1] = key
	}
	return res
}
