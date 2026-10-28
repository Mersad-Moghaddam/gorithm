package qlearning

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestQLearningLessonBellmanNumbers(t *testing.T) {
	env := DefaultGridWorld()
	a := NewQAgent(env, .5, .9, .3)
	for i := range a.Q[8] {
		a.Q[8][i] = 100
	}
	next, r, done := env.Step(Coord{2, 1}, ActionRight)
	q, target := a.BellmanUpdate(Coord{2, 1}, ActionRight, r, next, done)
	if next != (Coord{2, 2}) || r != 10 || !done || q != 5 || target != 10 {
		t.Fatalf("terminal: %v %g %v %g %g", next, r, done, q, target)
	}
	a.Q[env.StateIndex(Coord{2, 1})][ActionRight] = 5
	q, target = a.BellmanUpdate(Coord{2, 0}, ActionRight, 0, Coord{2, 1}, false)
	if q != 2.25 || target != 4.5 {
		t.Fatalf("nonterminal %g %g", q, target)
	}
}

func TestQLearningLessonCollisionsAndTies(t *testing.T) {
	env := DefaultGridWorld()
	a := NewQAgent(env, .5, .9, .3)
	for _, x := range []struct {
		s Coord
		a Action
	}{{Coord{0, 0}, ActionUp}, {Coord{1, 0}, ActionRight}} {
		n, r, d := env.Step(x.s, x.a)
		if n != x.s || r != 0 || d {
			t.Fatal("collision", n, r, d)
		}
	}
	best, _ := a.BestActions(env.Start)
	act, _ := a.MaxQ(env.Start)
	if !reflect.DeepEqual(best, []Action{0, 1, 2, 3}) || act != ActionUp {
		t.Fatal(best, act)
	}
}

func TestQLearningLessonValidation(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `[]`, `{"alpha":0}`, `{"alpha":null}`, `{"gamma":0}`, `{"epsilon":0}`, `{"episodes":0}`, `{"episodes":21}`, `{"alpha":1.01}`, `{"gamma":-1}`, `{"epsilon":2}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"episodes":1,"alpha":1,"gamma":1,"epsilon":1}`, `{"episodes":20}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Errorf("rejected %s: %v", raw, err)
		}
	}
	for _, env := range []*GridWorld{NewGridWorld(0, 3, Coord{}, Coord{}, nil), NewGridWorld(9, 8, Coord{}, Coord{}, nil), NewGridWorld(2, 2, Coord{-1, 0}, Coord{}, nil), NewGridWorld(2, 2, Coord{}, Coord{2, 0}, nil), NewGridWorld(2, 2, Coord{}, Coord{1, 1}, []Coord{{3, 0}})} {
		steps := Trace(env, 1, .5, .9, .3)
		if len(steps) != 1 || steps[0].State.(renderers.QLearningLessonState).Phase != "invalid-input" {
			t.Fatal("invalid env not bounded")
		}
	}
	if steps := Trace(nil, 1, math.NaN(), .9, .3); len(steps) != 1 {
		t.Fatal("nonfinite input")
	}
	if steps := Trace(nil, 1, math.Inf(-1), .9, .3); len(steps) != 1 {
		t.Fatal("negative infinity silently defaulted")
	}
}

func TestQLearningLessonSeedAndBudget(t *testing.T) {
	steps := Trace(nil, 5, .5, .9, .3)
	if !reflect.DeepEqual(steps, TraceWithSeed(nil, 5, .5, .9, .3, 7)) {
		t.Fatal("seed drift")
	}
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	counts := map[int]int{}
	goals := 0
	for i, step := range steps {
		s := step.State.(renderers.QLearningLessonState)
		if svg := renderers.RenderQLearningLessonSVG(step, 320, 650); !strings.Contains(svg, `data-entity="agent:`) || strings.Contains(svg, "NaN") {
			t.Fatal("mid-trace SVG", i)
		}
		if step.LineID == "L7" {
			counts[s.Episode]++
			if s.State == s.NextState && s.Reward != 0 {
				t.Fatal("collision reward")
			}
		}
		if s.EpisodeStopReason == "goal" {
			goals++
		}
		if step.LineID == "L11" {
			prev := steps[i-1].State.(renderers.QLearningLessonState)
			if s.State != prev.NextState {
				t.Fatal("state assignment")
			}
		}
		if step.LineID == "L4" && s.Episode > 1 {
			prev := steps[i-1].State.(renderers.QLearningLessonState)
			if !reflect.DeepEqual(prev.Q, s.Q) {
				t.Fatal("reset Q")
			}
		}
		if step.LineID == "L10" {
			if steps[i-1].HoldMS != 600 || step.HoldMS != 1200 {
				t.Fatal("target/update timing")
			}
			if s.NewQ != s.OldQ+s.Alpha*(s.TDTarget-s.OldQ) || !s.UpdateValid {
				t.Fatal("update")
			}
			if s.Terminal && (s.BestNextQ != 0 || s.TDTarget != 10) {
				t.Fatal("terminal future")
			}
		}
	}
	for i, step := range steps {
		if step.LineID == "L7" {
			s := step.State.(renderers.QLearningLessonState)
			if s.State != s.NextState {
				for j := i + 1; j <= i+3; j++ {
					if steps[j].State.(renderers.QLearningLessonState).State != s.State {
						t.Fatal("agent moved before L11")
					}
				}
			}
		}
	}
	for _, n := range counts {
		if n > 12 {
			t.Fatal(n)
		}
	}
	if goals == 0 {
		t.Fatal("missing goal fixture")
	}
	unreachable := NewGridWorld(2, 2, Coord{}, Coord{1, 1}, []Coord{{0, 1}, {1, 0}})
	capped := Trace(unreachable, 1, .5, .9, .3)
	s := capped[len(capped)-2].State.(renderers.QLearningLessonState)
	if s.Phase != "episode-end" || s.EpisodeStopReason != "step-limit" || s.EpisodeStep != 12 {
		t.Fatal(s)
	}
	first := steps[0].State.(renderers.QLearningLessonState)
	last := steps[len(steps)-1].State.(renderers.QLearningLessonState)
	last.Q[0][0] = 999
	last.Obstacles[0] = [2]int{0, 0}
	if first.Q[0][0] != 0 || first.Obstacles[0] != [2]int{1, 1} {
		t.Fatal("aliased snapshot")
	}
}
