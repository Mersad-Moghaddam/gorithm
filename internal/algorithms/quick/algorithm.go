package quick

// Sort returns a sorted copy of arr using Quick Sort.
// ponytail: Lomuto partition chosen for simplicity; upgrade to dual-pivot/introsort if worst-case recursion depth is a concern.
func Sort(arr []int) []int {
	if len(arr) <= 1 {
		res := make([]int, len(arr))
		copy(res, arr)
		return res
	}

	res := make([]int, len(arr))
	copy(res, arr)
	quickSort(res, 0, len(res)-1)
	return res
}

func quickSort(arr []int, low, high int) {
	if low < high {
		p := partition(arr, low, high)
		quickSort(arr, low, p-1)
		quickSort(arr, p+1, high)
	}
}

func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
