package bucketsort

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: range-based 5-bucket trace; upgrade to dynamic bucket count visualizer if array size warrants.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort with empty array.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"n": 0})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is empty.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort with single element.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Single element is sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	minVal, maxVal := arr[0], arr[0]
	for i := 1; i < n; i++ {
		b.IncComparisons()
		if arr[i] < minVal {
			minVal = arr[i]
		}
		if arr[i] > maxVal {
			maxVal = arr[i]
		}
	}
	b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Found range: min = %d, max = %d.", minVal, maxVal), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"min": minVal, "max": maxVal})

	if minVal == maxVal {
		b.AddStep("L20", trace.ActionSet, "All elements are equal; array is already sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, n-1),
			Current: -1,
		}, map[string]any{"status": "complete"})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is fully sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, n-1),
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	k := 5
	b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Set number of buckets k = %d.", k), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"k": k, "min": minVal, "max": maxVal})

	buckets := make([][]int, k)
	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Initialized %d empty buckets.", k), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, bucketVars(buckets, map[string]any{"k": k, "min": minVal, "max": maxVal}))

	// Scatter phase
	for i := 0; i < n; i++ {
		x := arr[i]
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Examine element arr[%d] = %d to place into bucket.", i, x), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, bucketVars(buckets, map[string]any{"i": i, "x": x}))

		idx := (x - minVal) * (k - 1) / (maxVal - minVal)
		b.IncOperations()
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Range hash: (%d - %d) * %d / %d = bucket index %d.", x, minVal, k-1, maxVal-minVal, idx), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{i},
			Current:   i,
		}, bucketVars(buckets, map[string]any{"i": i, "x": x, "bucketIdx": idx}))

		buckets[idx] = append(buckets[idx], x)
		b.IncOperations()
		b.AddStep("L9", trace.ActionPush, fmt.Sprintf("Appended %d to bucket %d.", x, idx), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, bucketVars(buckets, map[string]any{"i": i, "x": x, "bucketIdx": idx}))
	}

	// Local sort phase
	for i := 0; i < k; i++ {
		b.AddStep("L11", trace.ActionHighlight, fmt.Sprintf("Starting local insertion sort for bucket %d (%d elements).", i, len(buckets[i])), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, bucketVars(buckets, map[string]any{"bucketIdx": i}))

		bkt := buckets[i]
		for p := 1; p < len(bkt); p++ {
			key := bkt[p]
			q := p - 1
			for q >= 0 {
				b.IncComparisons()
				if bkt[q] > key {
					bkt[q+1] = bkt[q]
					b.IncSwaps()
					q--
				} else {
					break
				}
			}
			bkt[q+1] = key
			b.IncOperations()
		}
		buckets[i] = bkt

		b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Bucket %d sorted: %v.", i, buckets[i]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, bucketVars(buckets, map[string]any{"bucketIdx": i, "sortedBucket": fmt.Sprintf("%v", buckets[i])}))
	}

	// Gather phase
	b.AddStep("L14", trace.ActionSet, "Reset array write index to 0 for gathering sorted buckets.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: 0,
	}, bucketVars(buckets, map[string]any{"idx": 0}))

	writeIdx := 0
	for bIdx, bkt := range buckets {
		b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Gathering elements from bucket %d.", bIdx), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, writeIdx-1),
			Current: writeIdx,
		}, bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "writeIdx": writeIdx}))

		for _, x := range bkt {
			b.AddStep("L16", trace.ActionHighlight, fmt.Sprintf("Next element from bucket %d is %d.", bIdx, x), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  makeRange(0, writeIdx-1),
				Current: writeIdx,
			}, bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "val": x, "writeIdx": writeIdx}))

			arr[writeIdx] = x
			b.IncOperations()
			b.AddStep("L17", trace.ActionSet, fmt.Sprintf("Placed %d at arr[%d].", x, writeIdx), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  makeRange(0, writeIdx),
				Current: writeIdx,
			}, bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "val": x, "writeIdx": writeIdx}))

			writeIdx++
		}
	}

	b.AddStep("L20", trace.ActionHighlight, "All buckets gathered into array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  makeRange(0, n-1),
		Current: -1,
	}, bucketVars(buckets, map[string]any{"status": "gathered"}))

	b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is fully sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  makeRange(0, n-1),
		Current: -1,
	}, bucketVars(buckets, map[string]any{"status": "complete"}))

	return b.Steps()
}

func bucketVars(buckets [][]int, extra map[string]any) map[string]any {
	vars := make(map[string]any, len(extra)+len(buckets)+1)
	for k, v := range extra {
		vars[k] = v
	}
	allBuckets := make([]string, len(buckets))
	for i, b := range buckets {
		str := fmt.Sprintf("%v", b)
		vars[fmt.Sprintf("bucket_%d", i)] = str
		allBuckets[i] = str
	}
	vars["buckets"] = "[" + strings.Join(allBuckets, ", ") + "]"
	return vars
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}

func makeRange(start, end int) []int {
	if start > end {
		return []int{}
	}
	res := make([]int, end-start+1)
	for i := range res {
		res[i] = start + i
	}
	return res
}
