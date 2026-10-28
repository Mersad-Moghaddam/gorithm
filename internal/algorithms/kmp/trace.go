package kmp

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace follows Search's byte comparisons, including its post-advance mismatch check.
func Trace(text, pattern string) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.KMPLessonState{Text: text, Pattern: pattern, LPS: make([]int, len(pattern)), Phase: "init", LPSIndex: -1, PrefixLength: -1, FallbackFrom: -1, FallbackTo: -1, LPSKnownThrough: -1, MatchIndex: -1}
	for k := range s.LPS {
		s.LPS[k] = -1
	}
	if len(pattern) > 0 {
		s.LPS[0] = 0
		s.LPSKnownThrough = 0
	}
	add := func(line string, action trace.ActionType, phase, why string) {
		s.Phase = phase
		s.Shift = s.I - s.J
		copyState := s
		copyState.LPS = append([]int(nil), s.LPS...)
		vars := map[string]any{"i": s.I, "j": s.J, "shift": s.Shift}
		if s.ResultKnown {
			vars["matchIndex"] = s.MatchIndex
		}
		b.AddStep(line, action, why, copyState, vars)
	}
	finish := func(line string, result int) {
		s.ResultKnown = true
		s.MatchIndex = result
		why := fmt.Sprintf("Pattern found at byte offset %d.", result)
		if result < 0 {
			why = "Pattern not found."
		}
		if len(pattern) == 0 {
			why = "Empty pattern returns byte offset 0; no comparisons needed."
		}
		add(line, trace.ActionDone, "done", why)
	}
	add("L1", trace.ActionInit, "init", "Search uses byte offsets, not Unicode character positions.")
	if len(pattern) == 0 {
		finish("empty", 0)
		return b.Steps()
	}
	add("L2", trace.ActionSet, "init", fmt.Sprintf("Text has %d bytes; pattern has %d bytes.", len(text), len(pattern)))
	if len(text) < len(pattern) {
		finish("short", -1)
		return b.Steps()
	}
	s.LPSIndex = 0
	s.PrefixLength = 0
	add("L3", trace.ActionSet, "lps", "Inside computeLPS: lps[0] = 0; other entries remain unknown.")
	length := 0
	for scan := 1; scan < len(pattern); {
		s.LPSIndex = scan
		s.PrefixLength = length
		s.FallbackFrom = -1
		s.FallbackTo = -1
		b.IncComparisons()
		add("L3", trace.ActionCompare, "lps", fmt.Sprintf("Inside computeLPS: compare pattern bytes %d and %d.", scan, length))
		if pattern[scan] == pattern[length] {
			length++
			s.LPS[scan] = length
			s.LPSKnownThrough = scan
			s.PrefixLength = length
			b.IncOperations()
			add("L3", trace.ActionSet, "lps", fmt.Sprintf("Inside computeLPS: matching suffix extends prefix; lps[%d] = %d.", scan, length))
			scan++
		} else if length > 0 {
			old := length
			length = s.LPS[length-1]
			s.FallbackFrom = old
			s.FallbackTo = length
			s.PrefixLength = length
			b.IncOperations()
			add("L3", trace.ActionSet, "lps", fmt.Sprintf("Inside computeLPS: unequal bytes; prefix falls from %d to %d, scan stays %d.", old, length, scan))
		} else {
			s.LPS[scan] = 0
			s.LPSKnownThrough = scan
			b.IncOperations()
			add("L3", trace.ActionSet, "lps", fmt.Sprintf("Inside computeLPS: no matching proper prefix; lps[%d] = 0.", scan))
			scan++
		}
	}
	s.LPSIndex = -1
	s.PrefixLength = -1
	s.FallbackFrom = -1
	s.FallbackTo = -1
	add("L4", trace.ActionSet, "init", "Start text pointer i = 0 and pattern pointer j = 0.")
	for s.I < len(text) {
		s.LPSIndex = -1
		s.FallbackFrom = -1
		s.FallbackTo = -1
		add("L5", trace.ActionHighlight, "advance", "Text remains; compare current bytes.")
		b.IncComparisons()
		add("L6", trace.ActionCompare, "compare", fmt.Sprintf("Compare text[%d] and pattern[%d].", s.I, s.J))
		if text[s.I] == pattern[s.J] {
			s.I++
			s.J++
			add("L7", trace.ActionSet, "advance", "Equal bytes: advance both pointers; matched prefix grows.")
		}
		if s.J == len(pattern) {
			add("L9", trace.ActionHighlight, "advance", "Pattern pointer reached its end: full match.")
			finish("L10", s.I-s.J)
			return kmpLessonTiming(b.Steps())
		}
		if s.I < len(text) {
			b.IncComparisons()
			if text[s.I] != pattern[s.J] {
				add("L11", trace.ActionHighlight, "mismatch", fmt.Sprintf("Unequal bytes at text[%d] and pattern[%d]: ≠.", s.I, s.J))
				if s.J > 0 {
					s.LPSIndex = s.J - 1
					s.FallbackFrom = s.J
					s.FallbackTo = s.LPS[s.J-1]
					add("L12", trace.ActionHighlight, "mismatch", fmt.Sprintf("Read lps[%d] = %d; text pointer stays fixed.", s.LPSIndex, s.FallbackTo))
					s.J = s.FallbackTo
					b.IncOperations()
					add("L13", trace.ActionSet, "fallback", fmt.Sprintf("j = lps[%d] = %d; i stays %d. Shift alignment to %d.", s.LPSIndex, s.J, s.I, s.I-s.J))
				} else {
					add("L14", trace.ActionHighlight, "mismatch", "No matched prefix to reuse; advance text pointer.")
					s.I++
					b.IncOperations()
					add("L15", trace.ActionSet, "advance", "Advance i; j remains 0.")
				}
			}
		}
	}
	add("L18", trace.ActionHighlight, "advance", "Text pointer reached end without full match.")
	finish("L19", -1)
	return kmpLessonTiming(b.Steps())
}

func kmpLessonTiming(steps []trace.Step) []trace.Step {
	for k := 1; k < len(steps); k++ {
		if steps[k-1].Action == trace.ActionCompare && steps[k].Action == trace.ActionSet {
			steps[k-1].HoldMS = 600
			steps[k].HoldMS = 1200
		}
	}
	return steps
}
