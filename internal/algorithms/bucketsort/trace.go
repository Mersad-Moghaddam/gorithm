package bucketsort

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type bucketItem struct {
	val int
	id  string
}

// ponytail: range-based 5-bucket trace; upgrade to dynamic bucket count visualizer if array size warrants.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	n := len(initial)
	arr := cloneSlice(initial)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("bucketsort-%d", i)
	}

	snapshot := func(caption string, current int, comparing, swapped, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:    cloneSlice(arr),
			IDs:       cloneSlice(ids),
			Sorted:    cloneSlice(sorted),
			Current:   current,
			Comparing: cloneSlice(comparing),
			Swapped:   cloneSlice(swapped),
			Caption:   caption,
		}
	}

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort with empty array.", snapshot("Initialize Bucket Sort: empty array", -1, nil, nil, nil), map[string]any{"n": 0})
		b.AddStep("L3", trace.ActionHighlight, "Array is empty (n = 0); nothing to sort.", snapshot("Empty array: n <= 1", -1, nil, nil, nil), map[string]any{"n": 0})
		b.AddStep("L20", trace.ActionHighlight, "Return empty array.", snapshot("Return empty array", -1, nil, nil, nil), map[string]any{"status": "complete"})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is empty.", snapshot("Bucket Sort completed! Array is empty", -1, nil, nil, nil), map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort with single element.", snapshot("Initialize Bucket Sort: single element", -1, nil, nil, nil), map[string]any{"n": 1})
		b.AddStep("L2", trace.ActionSet, "Array length is 1.", snapshot("Array length is 1", -1, nil, nil, nil), map[string]any{"n": 1})
		b.AddStep("L3", trace.ActionHighlight, "Single element array (n <= 1) is already sorted.", snapshot("Single element is sorted", 0, nil, nil, []int{0}), map[string]any{"n": 1})
		b.AddStep("L20", trace.ActionHighlight, "Return sorted single-element array.", snapshot("Return sorted array", -1, nil, nil, []int{0}), map[string]any{"status": "complete"})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Single element is sorted.", snapshot("Bucket Sort completed! Single element is sorted", -1, nil, nil, []int{0}), map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Bucket Sort: distribute elements across k=5 buckets, sort locally, then gather.", snapshot("Initialize Bucket Sort: distribute, sort buckets, concatenate", -1, nil, nil, nil), map[string]any{"n": n})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(fmt.Sprintf("%d values to sort", n), -1, nil, nil, nil), map[string]any{"n": n})

	b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Array length %d > 1: proceed to determine value range.", n), snapshot(fmt.Sprintf("Array length %d > 1: proceed with sorting", n), -1, nil, nil, nil), map[string]any{"n": n})

	// L4: minVal, maxVal
	minVal, maxVal := arr[0], arr[0]
	b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Scan range: start with min = %d, max = %d from arr[0].", minVal, maxVal), snapshot(fmt.Sprintf("Initial range candidate: [%d, %d] from arr[0]", minVal, maxVal), 0, []int{0}, nil, nil), map[string]any{"min": minVal, "max": maxVal})

	for i := 1; i < n; i++ {
		b.IncComparisons()
		changed := false
		if arr[i] < minVal {
			minVal = arr[i]
			changed = true
		}
		if arr[i] > maxVal {
			maxVal = arr[i]
			changed = true
		}
		if changed {
			b.IncOperations()
			b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Examined arr[%d] = %d: updated range to [min=%d, max=%d].", i, arr[i], minVal, maxVal), snapshot(fmt.Sprintf("Inspecting arr[%d]=%d: updated range [%d, %d]", i, arr[i], minVal, maxVal), i, []int{i}, nil, nil), map[string]any{"min": minVal, "max": maxVal, "i": i})
		}
	}

	b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Determined range: min = %d, max = %d (spread = %d).", minVal, maxVal, maxVal-minVal), snapshot(fmt.Sprintf("Range established: [%d, %d] (spread = %d)", minVal, maxVal, maxVal-minVal), -1, nil, nil, nil), map[string]any{"min": minVal, "max": maxVal, "spread": maxVal - minVal})

	if minVal == maxVal {
		allSorted := makeRange(0, n-1)
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("All elements are identical (%d); array is already sorted.", minVal), snapshot(fmt.Sprintf("All elements equal (%d): already sorted", minVal), -1, nil, nil, allSorted), map[string]any{"min": minVal, "max": maxVal})
		b.AddStep("L20", trace.ActionHighlight, "Return array without bucket distribution since all values are equal.", snapshot("Return array: uniform range already sorted", -1, nil, nil, allSorted), map[string]any{"status": "complete"})
		b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is fully sorted.", snapshot("Bucket Sort completed! Array is fully sorted", -1, nil, nil, allSorted), map[string]any{"status": "complete"})
		return b.Steps()
	}

	k := 5
	b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Set number of buckets k = %d to partition range [%d, %d].", k, minVal, maxVal), snapshot(fmt.Sprintf("Partition range [%d, %d] into k = %d buckets", minVal, maxVal, k), -1, nil, nil, nil), map[string]any{"k": k, "min": minVal, "max": maxVal})

	buckets := make([][]bucketItem, k)
	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Initialized %d empty buckets: %s.", k, formatBuckets(buckets)), snapshot(fmt.Sprintf("Initialized %d empty buckets", k), -1, nil, nil, nil), bucketVars(buckets, map[string]any{"k": k, "min": minVal, "max": maxVal}))

	// Scatter phase
	for i := 0; i < n; i++ {
		x := arr[i]
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Scatter element %d of %d: arr[%d] = %d.", i+1, n, i, x), snapshot(fmt.Sprintf("Scatter arr[%d] = %d: map to bucket", i, x), i, []int{i}, nil, nil), bucketVars(buckets, map[string]any{"i": i, "x": x}))

		idx := (x - minVal) * (k - 1) / (maxVal - minVal)
		b.IncOperations()
		formulaCaption := fmt.Sprintf("Formula: (%d - %d) * (%d - 1) / (%d - %d) = %d / %d = bucket %d", x, minVal, k, maxVal, minVal, (x-minVal)*(k-1), maxVal-minVal, idx)
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Calculated bucket index for %d: %s.", x, formulaCaption), snapshot(formulaCaption, i, []int{i}, nil, nil), bucketVars(buckets, map[string]any{"i": i, "x": x, "bucketIdx": idx, "formula": formulaCaption}))

		buckets[idx] = append(buckets[idx], bucketItem{val: x, id: ids[i]})
		b.IncOperations()
		pushCaption := fmt.Sprintf("Bucket %d receives %d; contents now: %s", idx, x, formatBucket(buckets[idx]))
		b.AddStep("L9", trace.ActionPush, fmt.Sprintf("Appended %d to bucket %d; contents now: %s.", x, idx, formatBucket(buckets[idx])), snapshot(pushCaption, i, nil, nil, nil), bucketVars(buckets, map[string]any{"i": i, "x": x, "bucketIdx": idx}))
	}

	// Local sort phase
	for i := 0; i < k; i++ {
		bkt := buckets[i]
		b.AddStep("L11", trace.ActionHighlight, fmt.Sprintf("Start local insertion sort for bucket %d (%d elements): %s.", i, len(bkt), formatBucket(bkt)), snapshot(fmt.Sprintf("Sorting bucket %d (%d items): %s", i, len(bkt), formatBucket(bkt)), -1, nil, nil, nil), bucketVars(buckets, map[string]any{"bucketIdx": i}))

		if len(bkt) <= 1 {
			b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Bucket %d has %d element(s); already sorted: %s.", i, len(bkt), formatBucket(bkt)), snapshot(fmt.Sprintf("Bucket %d sorted: %s", i, formatBucket(bkt)), -1, nil, nil, nil), bucketVars(buckets, map[string]any{"bucketIdx": i, "sortedBucket": formatBucket(bkt)}))
		} else {
			for p := 1; p < len(bkt); p++ {
				key := bkt[p]
				q := p - 1
				for q >= 0 {
					b.IncComparisons()
					if bkt[q].val > key.val {
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
			b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Bucket %d sorted with insertion sort: %s.", i, formatBucket(buckets[i])), snapshot(fmt.Sprintf("Bucket %d sorted: %s", i, formatBucket(buckets[i])), -1, nil, nil, nil), bucketVars(buckets, map[string]any{"bucketIdx": i, "sortedBucket": formatBucket(buckets[i])}))
		}
	}

	// Gather phase
	sorted := make([]int, 0, n)
	b.AddStep("L14", trace.ActionSet, "Reset array write index to 0 for concatenating sorted buckets.", snapshot("Gather phase: reset write index to 0", 0, nil, nil, sorted), bucketVars(buckets, map[string]any{"writeIdx": 0}))

	writeIdx := 0
	for bIdx, bkt := range buckets {
		b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Gathering sorted bucket %d (%d elements): %s.", bIdx, len(bkt), formatBucket(bkt)), snapshot(fmt.Sprintf("Concatenating bucket %d (%d items): %s", bIdx, len(bkt), formatBucket(bkt)), writeIdx, nil, nil, sorted), bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "writeIdx": writeIdx}))

		for _, item := range bkt {
			b.AddStep("L16", trace.ActionHighlight, fmt.Sprintf("Read next element %d from bucket %d to place at arr[%d].", item.val, bIdx, writeIdx), snapshot(fmt.Sprintf("Next from bucket %d: %d -> arr[%d]", bIdx, item.val, writeIdx), writeIdx, nil, nil, sorted), bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "val": item.val, "writeIdx": writeIdx}))

			pos := -1
			for p := writeIdx; p < n; p++ {
				if ids[p] == item.id {
					pos = p
					break
				}
			}
			if pos != -1 && pos != writeIdx {
				arr[writeIdx], arr[pos] = arr[pos], arr[writeIdx]
				ids[writeIdx], ids[pos] = ids[pos], ids[writeIdx]
			} else {
				arr[writeIdx] = item.val
				ids[writeIdx] = item.id
			}
			b.IncOperations()
			sorted = append(sorted, writeIdx)

			placedCaption := fmt.Sprintf("Placed %d at arr[%d]; green prefix [0..%d] is settled", item.val, writeIdx, writeIdx)
			b.AddStep("L17", trace.ActionSet, fmt.Sprintf("Placed %d at arr[%d] from bucket %d.", item.val, writeIdx, bIdx), snapshot(placedCaption, writeIdx, nil, nil, sorted), bucketVars(buckets, map[string]any{"bucketIdx": bIdx, "val": item.val, "writeIdx": writeIdx}))

			writeIdx++
		}
	}

	b.AddStep("L20", trace.ActionHighlight, fmt.Sprintf("All %d buckets concatenated into array: %v.", k, arr), snapshot("Concatenation complete: all buckets merged", -1, nil, nil, sorted), bucketVars(buckets, map[string]any{"status": "gathered"}))

	b.AddStep("L21", trace.ActionDone, "Bucket Sort completed! Array is fully sorted.", snapshot("Bucket Sort completed! Array is fully sorted", -1, nil, nil, sorted), bucketVars(buckets, map[string]any{"status": "complete"}))

	return b.Steps()
}

func bucketVars(buckets [][]bucketItem, extra map[string]any) map[string]any {
	vars := make(map[string]any, len(extra)+len(buckets)+1)
	for k, v := range extra {
		vars[k] = v
	}
	allBuckets := make([]string, len(buckets))
	for i, b := range buckets {
		str := formatBucket(b)
		vars[fmt.Sprintf("bucket_%d", i)] = str
		allBuckets[i] = str
	}
	vars["buckets"] = "[" + strings.Join(allBuckets, ", ") + "]"
	return vars
}

func formatBucket(bkt []bucketItem) string {
	if len(bkt) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for i, it := range bkt {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(strconv.Itoa(it.val))
	}
	sb.WriteByte(']')
	return sb.String()
}

func formatBuckets(buckets [][]bucketItem) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, bkt := range buckets {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(formatBucket(bkt))
	}
	sb.WriteByte(']')
	return sb.String()
}

func cloneSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	c := make([]T, len(s))
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
