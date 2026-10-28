package coinchange

import (
	"encoding/json"
	"reflect"
	"testing"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestLessonInputBounds(t *testing.T) {
	for _, input := range []string{`null`, `{`, `[]`, `{"coins":null}`, `{"amount":null}`, `{"coins":[]}`, `{"coins":[0]}`, `{"coins":[-1]}`, `{"coins":[201]}`, `{"amount":-1}`, `{"amount":201}`, `{"amount":1.5}`, `{"coins":"1"}`, `{"coins":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	for _, input := range []string{``, `{}`, `{"amount":0,"coins":[1]}`, `{"amount":200,"coins":[200]}`, `{"amount":0,"coins":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err != nil {
			t.Errorf("rejected valid input %s: %v", input, err)
		}
	}
}

func TestCoinChangeTypedTrace(t *testing.T) {
	for _, fixture := range []struct{coins []int; amount int}{ {[]int{1,3,4},6}, {[]int{1,2,5},11}, {[]int{2},3}, {[]int{1},0}, {[]int{1},-1}, {nil,3}, {[]int{1,1,3},6}, {[]int{0,1},2} } {
		steps := Trace(fixture.coins, fixture.amount)
		want, dp := MinCoins(fixture.coins, fixture.amount)
		last := steps[len(steps)-1].State.(renderers.CoinChangeLessonState)
		if !last.HasResult || last.Result != want || !reflect.DeepEqual(last.DP, dp) { t.Fatalf("oracle mismatch %+v: %+v / %v", fixture, last, dp) }
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil { t.Fatal(err) }
		for i, step := range steps {
			s := step.State.(renderers.CoinChangeLessonState)
			if step.LineID == "L8" {
				prev := steps[i-1].State.(renderers.CoinChangeLessonState)
				if s.SourceIndex < 0 || s.SourceIndex >= len(s.DP) || s.Candidate >= prev.DP[s.Index] || s.DP[s.Index] != s.Candidate || s.Previous != prev.DP[s.Index] { t.Fatal("non-improving write") }
				if step.HoldMS + steps[i-1].HoldMS != 1800 { t.Fatal("pair budget") }
			}
			if step.LineID == "L7" && i > 0 && !reflect.DeepEqual(s.DP, steps[i-1].State.(renderers.CoinChangeLessonState).DP) { t.Fatal("comparison changed table") }
			if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 { t.Fatal("invalid hold") }
			if renderers.RenderCoinChangeLessonSVG(step, 320, 450) == "" { t.Fatal("empty real-trace scene") }
		}
	}
}

func TestCoinChangeSnapshotIsolation(t *testing.T) {
	coins := []int{1,3,4}
	steps := Trace(coins, 6)
	coins[0] = 99
	a := steps[0].State.(renderers.CoinChangeLessonState)
	b := steps[1].State.(renderers.CoinChangeLessonState)
	if a.Coins[0] != 1 { t.Fatal("caller aliases coin snapshot") }
	a.Coins[0], a.DP[0], a.Cells[0].Value, a.ColHeaders[0] = 55, 55, "bad", "bad"
	if b.Coins[0] != 1 || b.DP[0] != 0 || b.Cells[0].Value == "bad" || b.ColHeaders[0] == "bad" { t.Fatal("snapshots share backing arrays") }
}
