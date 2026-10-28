package hanoitower

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestHanoiLessonInputBounds(t *testing.T) {
	for _, raw := range []string{"1", "8", `{"disks":1,"n":8}`, `{"n":2}`, `{"count":8}`, `{}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
	}
	steps, err := (Lesson{}).Build(json.RawMessage(`{"disks":1,"n":8,"count":2}`))
	if err != nil || steps[0].State.(renderers.HanoiTowerLessonState).TotalDisks != 1 {
		t.Fatal("alias precedence changed")
	}
	for _, raw := range []string{"0", "9", "-1", "null", "3.5", `"3"`, `[]`, `{`, `{"disks":null}`, `{"disks":"3"}`, `{"disks":0,"n":3}`, `{"n":3,"count":false}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}

func TestHanoiAnimationTrace(t *testing.T) {
	for n := 0; n <= 5; n++ {
		steps := Trace(n)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		var moves []Move
		var previous [3][]int
		for d := n; d > 0; d-- {
			previous[0] = append(previous[0], d)
		}
		for i, step := range steps {
			s, ok := step.State.(renderers.HanoiTowerLessonState)
			if !ok {
				t.Fatalf("missing typed state: %T", step.State)
			}
			count := 0
			seen := make(map[int]bool)
			for _, peg := range s.Pegs {
				for i, d := range peg {
					count++
					if d < 1 || d > n || seen[d] {
						t.Fatal("disk multiset changed")
					}
					seen[d] = true
					if i > 0 && peg[i-1] <= d {
						t.Fatal("illegal disk order")
					}
				}
			}
			if count != n {
				t.Fatal("disk count changed")
			}
			if step.LineID == "L6" {
				if !reflect.DeepEqual(s.BeforePegs, clonePegs(previous)) {
					t.Fatal("wrong motion origin")
				}
				m := Move{s.Disk, s.From, s.To}
				moves = append(moves, m)
				p := previous[m.From]
				if len(p) == 0 || p[len(p)-1] != m.Disk {
					t.Fatal("not top disk")
				}
				previous[m.From] = p[:len(p)-1]
				previous[m.To] = append(previous[m.To], m.Disk)
			} else if s.Disk != -1 || s.From != -1 || s.To != -1 {
				t.Fatal("inactive move fields")
			}
			if !reflect.DeepEqual(clonePegs(previous), s.Pegs) {
				t.Fatal("board changed outside move")
			}
			if s.MoveCount != len(moves) || step.Metrics.Operations != len(moves) {
				t.Fatal("wrong move metrics")
			}
			if step.LineID == "L2" && s.Calls[len(s.Calls)-1].N != step.Variables["n"] {
				t.Fatal("wrong active call")
			}
			if step.LineID == "L5" || step.LineID == "L7" {
				c := s.Calls[len(s.Calls)-1]
				if c.N-1 != step.Variables["n"] {
					t.Fatal("pending call already entered")
				}
				next := steps[i+1].State.(renderers.HanoiTowerLessonState)
				want := renderers.HanoiTowerCall{N: c.N - 1, From: c.From, To: c.Aux, Aux: c.To}
				if step.LineID == "L7" {
					want = renderers.HanoiTowerCall{N: c.N - 1, From: c.Aux, To: c.To, Aux: c.From}
				}
				if len(next.Calls) != len(s.Calls)+1 || next.Calls[len(next.Calls)-1] != want {
					t.Fatal("wrong recursive role permutation")
				}
			}
			if out := renderers.RenderHanoiTowerLessonSVG(step, 320, 450); !strings.Contains(out, `data-lesson="hanoitower"`) {
				t.Fatal("real trace renderer missing")
			}
		}
		if !reflect.DeepEqual(moves, Solve(n)) {
			t.Fatalf("n=%d moves differ", n)
		}
		final := steps[len(steps)-1].State.(renderers.HanoiTowerLessonState)
		if len(final.Calls) != 0 || len(final.Pegs[2]) != n {
			t.Fatal("incomplete result")
		}
	}
}

func TestHanoiSnapshotsIndependent(t *testing.T) {
	steps := Trace(3)
	before, _ := json.Marshal(steps)
	for i, step := range steps {
		s := step.State.(renderers.HanoiTowerLessonState)
		if step.LineID != "L6" {
			continue
		}
		s.Pegs[s.To][len(s.Pegs[s.To])-1] = 99
		s.BeforePegs[s.From][len(s.BeforePegs[s.From])-1] = 98
		s.Calls[0].N = 97
		next := steps[i+1].State.(renderers.HanoiTowerLessonState)
		if next.Calls[0].N != 3 {
			t.Fatal("call aliases")
		}
		for _, p := range next.Pegs {
			for _, d := range p {
				if d > 3 {
					t.Fatal("peg aliases")
				}
			}
		}
		for j, other := range steps {
			if j == i {
				continue
			}
			var original []json.RawMessage
			if err := json.Unmarshal(before, &original); err != nil {
				t.Fatal(err)
			}
			current, _ := json.Marshal(other)
			if string(current) != string(original[j]) {
				t.Fatalf("snapshot %d aliases move %d", j, i)
			}
		}
		return
	}
	t.Fatal("missing move")
}
