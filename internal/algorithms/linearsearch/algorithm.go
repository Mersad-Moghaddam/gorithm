package linearsearch

// Search performs a linear search on arr for target and returns index, or -1 if not found.
func Search(arr []int, target int) int {
	for i := 0; i < len(arr); i++ {
		if arr[i] == target {
			return i
		}
	}
	return -1
}
