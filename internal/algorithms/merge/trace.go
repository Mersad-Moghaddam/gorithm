package merge

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	ids := make([]int, len(arr))
	for i := range ids {
		ids[i] = i
	}
	low, high := 0, 0
	for _, v := range arr {
		low = min(low, v)
		high = max(high, v)
	}
	s := renderers.MergeLessonState{Left: -1, Mid: -1, Right: -1, I: -1, J: -1, K: -1, CopyIndex: -1, MinValue: low, MaxValue: high}
	emit := func(line string, action trace.ActionType, explanation string, current int, comparing, swapped, sorted []int, vars map[string]any) {
		out := s
		out.ArrayState = renderers.ArrayState{Values: cloneSlice(arr), Current: current, Comparing: cloneSlice(comparing), Swapped: cloneSlice(swapped), Sorted: cloneSlice(sorted)}
		out.IDs = cloneSlice(ids)
		out.Aux = cloneSlice(s.Aux)
		out.AuxIDs = cloneSlice(s.AuxIDs)
		out.AuxFilled = append([]bool(nil), s.AuxFilled...)
		out.Stack = append([]renderers.MergeLessonRange(nil), s.Stack...)
		b.AddStep(line, action, explanation, out, vars)
	}
	emit("L1", trace.ActionInit, "Initialize Merge Sort. Split ranges, then copy their stable merge through auxiliary storage.", -1, nil, nil, nil, map[string]any{"n": len(arr)})
	var visit func(int, int)
	visit = func(left, right int) {
		parent := s
		s = renderers.MergeLessonState{Left: left, Mid: -1, Right: right, I: -1, J: -1, K: -1, CopyIndex: -1, MinValue: low, MaxValue: high, Stack: append(append([]renderers.MergeLessonRange(nil), parent.Stack...), renderers.MergeLessonRange{Left: left, Mid: -1, Right: right})}
		defer func() { s = parent }()
		b.IncOperations()
		emit("L2", trace.ActionHighlight, fmt.Sprintf("Check range [%d..%d] for a base case.", left, right), -1, nil, nil, nil, map[string]any{"left": left, "right": right})
		if left >= right {
			emit("L3", trace.ActionVisit, fmt.Sprintf("Range [%d..%d] has one element; return.", left, right), left, nil, nil, nil, map[string]any{"left": left, "right": right})
			return
		}
		mid := left + (right-left)/2
		s.Mid = mid
		s.Stack[len(s.Stack)-1].Mid = mid
		b.IncOperations()
		emit("L5", trace.ActionSet, fmt.Sprintf("Split [%d..%d] at mid = %d.", left, right, mid), mid, nil, nil, nil, map[string]any{"left": left, "mid": mid, "right": right})
		emit("L6", trace.ActionVisit, fmt.Sprintf("Sort left range [%d..%d] first.", left, mid), mid, nil, nil, nil, map[string]any{"left": left, "mid": mid})
		visit(left, mid)
		emit("L7", trace.ActionVisit, fmt.Sprintf("Left range returned; sort right range [%d..%d].", mid+1, right), mid+1, nil, nil, nil, map[string]any{"mid+1": mid + 1, "right": right})
		visit(mid+1, right)
		emit("L8", trace.ActionHighlight, fmt.Sprintf("Both halves [%d..%d] and [%d..%d] are ready to merge.", left, mid, mid+1, right), -1, nil, nil, nil, map[string]any{"left": left, "mid": mid, "right": right})
		emit("L10", trace.ActionVisit, fmt.Sprintf("Begin merging [%d..%d] using separate auxiliary storage.", left, right), -1, nil, nil, nil, map[string]any{"left": left, "mid": mid, "right": right})
		s.Aux = make([]int, right-left+1)
		s.AuxIDs = make([]int, len(s.Aux))
		s.AuxFilled = make([]bool, len(s.Aux))
		for i := range s.AuxIDs {
			s.AuxIDs[i] = -1
		}
		b.IncOperations()
		emit("L11", trace.ActionSet, fmt.Sprintf("Allocate %d empty auxiliary cells. Unwritten cells are not zero values.", len(s.Aux)), -1, nil, nil, nil, map[string]any{"aux_size": len(s.Aux)})
		s.I, s.J, s.K = left, mid+1, 0
		b.IncOperations()
		emit("L12", trace.ActionSet, fmt.Sprintf("Set i = %d, j = %d, k = 0.", s.I, s.J), -1, nil, nil, nil, map[string]any{"i": s.I, "j": s.J, "k": s.K})
		copyAux := func(line string, from int) {
			s.Aux[s.K] = arr[from]
			s.AuxIDs[s.K] = ids[from]
			s.AuxFilled[s.K] = true
			b.IncOperations()
			vars := map[string]any{"val": arr[from], "k": s.K}
			if line == "L15" || line == "L21" {
				vars["i"] = from
			} else {
				vars["j"] = from
			}
			emit(line, trace.ActionSet, fmt.Sprintf("Copy arr[%d] = %d (#%d) into aux[%d]; main memory remains unchanged.", from, arr[from], ids[from], s.K), from, nil, nil, nil, vars)
		}
		for s.I <= mid && s.J <= right {
			b.IncComparisons()
			emit("L14", trace.ActionCompare, fmt.Sprintf("Compare %d and %d. Equal values take the left identity first.", arr[s.I], arr[s.J]), -1, []int{s.I, s.J}, nil, nil, map[string]any{"i": s.I, "j": s.J, "val_i": arr[s.I], "val_j": arr[s.J]})
			if arr[s.I] <= arr[s.J] {
				copyAux("L15", s.I)
				s.I++
			} else {
				copyAux("L17", s.J)
				s.J++
			}
			s.K++
		}
		for s.I <= mid {
			copyAux("L21", s.I)
			s.I++
			s.K++
		}
		for s.J <= right {
			copyAux("L22", s.J)
			s.J++
			s.K++
		}
		for p := range s.Aux {
			s.CopyIndex = p
			arr[left+p] = s.Aux[p]
			ids[left+p] = s.AuxIDs[p]
			b.IncOperations()
			emit("L23", trace.ActionSet, fmt.Sprintf("Copy aux[%d] = %d (#%d) back to arr[%d]. Other main cells still hold their old values.", p, s.Aux[p], s.AuxIDs[p], left+p), left+p, nil, []int{left + p}, nil, map[string]any{"index": left + p, "val": s.Aux[p]})
		}
		s.CopyIndex = -1
		emit("L24", trace.ActionHighlight, fmt.Sprintf("Range [%d..%d] is merged: %v. Return to its caller.", left, right, arr[left:right+1]), -1, nil, nil, nil, map[string]any{"left": left, "right": right})
	}
	if len(arr) > 1 {
		visit(0, len(arr)-1)
	}
	sorted := make([]int, len(arr))
	for i := range sorted {
		sorted[i] = i
	}
	emit("L9", trace.ActionDone, "Merge Sort complete. Every main-array position is final.", -1, nil, nil, sorted, map[string]any{"status": "complete"})
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L14" && (steps[i+1].LineID == "L15" || steps[i+1].LineID == "L17") {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}

func cloneSlice(s []int) []int { out := make([]int, len(s)); copy(out, s); return out }
