package bucketsort

// ponytail: 5 buckets with insertion sort; upgrade to dynamic k=len(arr) if variance/distribution requires it.
func Sort(arr []int) []int {
	n := len(arr)
	if n <= 1 {
		res := make([]int, n)
		copy(res, arr)
		return res
	}

	minVal, maxVal := arr[0], arr[0]
	for _, v := range arr[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if minVal == maxVal {
		res := make([]int, n)
		copy(res, arr)
		return res
	}

	k := 5
	buckets := make([][]int, k)
	for _, x := range arr {
		idx := (x - minVal) * (k - 1) / (maxVal - minVal)
		buckets[idx] = append(buckets[idx], x)
	}

	res := make([]int, 0, n)
	for i := 0; i < k; i++ {
		insertionSort(buckets[i])
		res = append(res, buckets[i]...)
	}
	return res
}

func insertionSort(bucket []int) {
	for i := 1; i < len(bucket); i++ {
		key := bucket[i]
		j := i - 1
		for j >= 0 && bucket[j] > key {
			bucket[j+1] = bucket[j]
			j--
		}
		bucket[j+1] = key
	}
}
