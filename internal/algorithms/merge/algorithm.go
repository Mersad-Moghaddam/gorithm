package merge

// Sort sorts arr using merge sort and returns a new sorted slice.
func Sort(arr []int) []int {
	if len(arr) <= 1 {
		res := make([]int, len(arr))
		copy(res, arr)
		return res
	}

	res := make([]int, len(arr))
	copy(res, arr)
	mergeSort(res, 0, len(res)-1)
	return res
}

func mergeSort(arr []int, left, right int) {
	if left >= right {
		return
	}
	mid := left + (right-left)/2
	mergeSort(arr, left, mid)
	mergeSort(arr, mid+1, right)
	merge(arr, left, mid, right)
}

func merge(arr []int, left, mid, right int) {
	aux := make([]int, right-left+1)
	i, j, k := left, mid+1, 0

	for i <= mid && j <= right {
		if arr[i] <= arr[j] {
			aux[k] = arr[i]
			i++
		} else {
			aux[k] = arr[j]
			j++
		}
		k++
	}

	for i <= mid {
		aux[k] = arr[i]
		i++
		k++
	}

	for j <= right {
		aux[k] = arr[j]
		j++
		k++
	}

	for p := 0; p < len(aux); p++ {
		arr[left+p] = aux[p]
	}
}
