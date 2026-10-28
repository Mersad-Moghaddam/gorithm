package nqueens

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestNQueensInputValidation(t *testing.T) {
	for _, raw := range []string{`0`, `9`, `-1`, `null`, `{"n":null}`, `{"n":"4"}`, `{"n":0}`, `{"n":9}`, `4.5`, `{`, `[]`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("Build(%s) accepted invalid input", raw)
		}
	}
	for _, raw := range []string{``, `{}`, `1`, `{"n":4}`, `8`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Errorf("Build(%s): %v", raw, err)
		}
	}
}

func TestNQueensSourceMatchesInitializationAndCompletion(t *testing.T) {
	lines := (Lesson{}).Source()
	steps := Trace(4)
	if !strings.Contains(lines[1].Code, "board[i] = -1") {
		t.Fatal("source omits empty-board initialization")
	}
	if got := steps[len(steps)-1].LineID; got != "L4" {
		t.Fatalf("completion highlights %s, want closing line L4", got)
	}
}

func TestNQueensSnapshots(t *testing.T) {
	for n := 1; n <= 4; n++ {
		steps := Trace(n)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		var solutions [][]int
		for i, step := range steps {
			s, ok := step.State.(renderers.NQueensLessonState)
			if !ok {
				t.Fatalf("step %d: %T", i, step.State)
			}
			if svg := renderers.RenderNQueensLessonSVG(step, 320, 450); !strings.Contains(svg, `data-phase="`+s.Phase+`"`) {
				t.Fatalf("step %d missing dedicated scene", i)
			}
			switch step.HoldMS {
			case 0, 350, 600, 1200, 1800:
			default:
				t.Fatalf("invalid hold %d", step.HoldMS)
			}
			if step.LineID == "L7" {
				solutions = append(solutions, append([]int(nil), s.Board...))
				if !reflect.DeepEqual(s.LastSolution, s.Board) {
					t.Fatal("solution not retained")
				}
				if i == len(steps)-1 {
					t.Fatal("solution stops search")
				}
			}
			if s.Phase == "place" && !isSafe(s.Row, s.Col, s.Board) {
				t.Fatal("unsafe placement")
			}
			if s.Phase == "conflict" {
				r, c := s.Conflict[0], s.Conflict[1]
				if r < 0 || r >= s.Row || s.Board[r] != c {
					t.Fatal("missing conflict witness")
				}
				first := -1
				for r := 0; r < s.Row; r++ {
					if s.Board[r] == s.Col || absNQueens(s.Board[r]-s.Col) == s.Row-r {
						first = r
						break
					}
				}
				if first != r {
					t.Fatal("not first conflict")
				}
			}
			if i > 0 {
				prev := steps[i-1].State.(renderers.NQueensLessonState)
				if s.Phase == "try" && !reflect.DeepEqual(s.Board, prev.Board) {
					t.Fatal("candidate committed")
				}
				if s.Phase == "remove" {
					for r, c := range s.Board {
						if r == s.Row {
							if c != -1 || prev.Board[r] != s.Col {
								t.Fatal("wrong removal")
							}
						} else if c != prev.Board[r] {
							t.Fatal("other row changed")
						}
					}
				}
				if prev.Phase == "safe" && (step.LineID != "L12" || steps[i-1].HoldMS+step.HoldMS != 1800) {
					t.Fatal("safe/place budget")
				}
			}
		}
		if !reflect.DeepEqual(solutions, Solve(n)) {
			t.Fatalf("n=%d solutions %v != %v", n, solutions, Solve(n))
		}
		if n == 4 && !reflect.DeepEqual(solutions, [][]int{{1, 3, 0, 2}, {2, 0, 3, 1}}) {
			t.Fatal("n=4 solution order changed")
		}
		last := steps[len(steps)-1].State.(renderers.NQueensLessonState)
		if last.Phase != "done" || last.Solutions != len(solutions) {
			t.Fatal("wrong final result")
		}
		for _, c := range last.Board {
			if c != -1 {
				t.Fatal("final board not empty")
			}
		}
	}
	steps := Trace(4)
	for i, s := range steps {
		if s.LineID == "L7" {
			original := steps[i].State.(renderers.NQueensLessonState)
			later := steps[i+1].State.(renderers.NQueensLessonState)
			later.Board[0] = 99
			later.LastSolution[0] = 99
			later.Cells[0].Value = "changed"
			later.RowHeaders[0] = "changed"
			if original.Board[0] == 99 || original.LastSolution[0] == 99 || original.Cells[0].Value == "changed" || original.RowHeaders[0] == "changed" {
				t.Fatal("snapshot aliased")
			}
			break
		}
	}
	if !reflect.DeepEqual(Trace(0), Trace(4)) {
		t.Fatal("direct Trace default changed")
	}
}

func absNQueens(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
