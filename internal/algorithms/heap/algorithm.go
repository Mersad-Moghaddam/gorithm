package heap

// Sort returns a sorted copy of arr using Heap Sort.
func Sort(arr []int) []int {
	n := len(arr)
	res := make([]int, n)
	copy(res, arr)

	if n <= 1 {
		return res
	}

	for i := n/2 - 1; i >= 0; i-- {
		heapify(res, n, i)
	}

	for i := n - 1; i > 0; i-- {
		res[0], res[i] = res[i], res[0]
		heapify(res, i, 0)
	}

	return res
}

func heapify(arr []int, n, i int) {
	largest := i
	l := 2*i + 1
	r := 2*i + 2

	if l < n && arr[l] > arr[largest] {
		largest = l
	}
	if r < n && arr[r] > arr[largest] {
		largest = r
	}

	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, n, largest)
	}
}
