package sudoku

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"strings"
	"testing"
)

func TestSudokuExhaustedAfterRollbackCaption(t *testing.T) {
	steps := Trace(sampleSolvableBoard())
	for i, step := range steps {
		s := step.State.(renderers.SudokuLessonState)
		if s.Phase != "exhausted" {
			continue
		}
		removed, placed := false, false
		for j := i - 1; j >= 0; j-- {
			prev := steps[j].State.(renderers.SudokuLessonState)
			if prev.Depth < s.Depth {
				break
			}
			if prev.Row != s.Row || prev.Col != s.Col || prev.Depth != s.Depth {
				continue
			}
			if prev.Phase == "select" {
				break
			}
			removed = removed || prev.Phase == "remove"
			placed = placed || prev.Phase == "place"
		}
		if !removed || !placed {
			continue
		}
		svg := renderers.RenderSudokuLessonSVG(step, 320, 450)
		if !strings.Contains(svg, "All candidates exhausted; backtrack") || strings.Contains(svg, "All 9 rejected") {
			t.Fatal("exhausted caption misstates recursive branch failure")
		}
		return
	}
	t.Fatal("fixture lacks exhausted cell with earlier placement and rollback")
}

func TestSudokuBuildRejectsInvalidJSONAndShape(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `{"board":null}`, `[]`, `[[1]]`, `{"board":"bad"}`, `{"board":[[1.5]]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, value := range []int{-1, 10} {
		board := CanonicalBoard
		board[0][2] = value
		raw, _ := json.Marshal(board)
		if _, err := (Lesson{}).Build(raw); err == nil {
			t.Errorf("accepted digit %d", value)
		}
	}
	board := CanonicalBoard
	board[0][2] = 5
	raw, _ := json.Marshal(board)
	if _, err := (Lesson{}).Build(raw); err == nil {
		t.Fatal("accepted duplicate givens")
	}
}

func TestSudokuBuildDimensionsAndAlternateForms(t *testing.T) {
	for _, size := range []int{8, 10} {
		rows := make([][]int, 9)
		for r := range rows {
			rows[r] = make([]int, 9)
		}
		rows[0] = make([]int, size)
		raw, _ := json.Marshal(rows)
		if _, err := (Lesson{}).Build(raw); err == nil {
			t.Fatalf("accepted row length %d", size)
		}
		rows = make([][]int, size)
		for r := range rows {
			rows[r] = make([]int, 9)
		}
		raw, _ = json.Marshal(rows)
		if _, err := (Lesson{}).Build(raw); err == nil {
			t.Fatalf("accepted row count %d", size)
		}
	}
	solved, _ := Solve(CanonicalBoard)
	for _, value := range []any{solved, map[string]any{"board": solved}, map[string]any{}} {
		raw, _ := json.Marshal(value)
		if _, err := (Lesson{}).Build(raw); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{strings.Replace((Lesson{}).Meta().DefaultInput, "5,3,0", "5,3,null", 1), `{"board":[null,null,null,null,null,null,null,null,null]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Fatal("accepted null entry")
		}
	}
	unsat := [9][9]int{}
	unsat[0] = [9]int{1, 2, 3, 4, 5, 6, 7, 8, 0}
	unsat[1][8] = 9
	raw, _ := json.Marshal(unsat)
	steps, err := (Lesson{}).Build(raw)
	if err != nil || steps[len(steps)-1].State.(renderers.SudokuLessonState).Status != "unsatisfiable" {
		t.Fatal("valid unsatisfiable rejected")
	}
}

func TestSudokuTypedTrace(t *testing.T) {
	solved, _ := Solve(CanonicalBoard)
	unsat := [9][9]int{}
	unsat[0] = [9]int{1, 2, 3, 4, 5, 6, 7, 8, 0}
	unsat[1][8] = 9
	invalid := CanonicalBoard
	invalid[0][2] = 5
	outside := CanonicalBoard
	outside[0][2] = 10
	rollback := sampleSolvableBoard()
	for name, board := range map[string][9][9]int{"default": CanonicalBoard, "solved": solved, "unsatisfiable": unsat, "invalid": invalid, "outside": outside, "rollback": rollback} {
		t.Run(name, func(t *testing.T) {
			steps := Trace(board)
			if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
				t.Fatal(err)
			}
			oracle, ok := Solve(board)
			last := steps[len(steps)-1].State.(renderers.SudokuLessonState)
			want := "unsatisfiable"
			if ok {
				want = "solved"
			} else if !validateBoard(board) {
				want = "invalid"
			}
			if last.Status != want || last.Board != oracle {
				t.Fatalf("terminal mismatch %s", last.Status)
			}
			if name == "solved" && len(steps) != 2 {
				t.Fatal("solved board tried candidates")
			}
			removed, placed := false, false
			for i, step := range steps {
				s := step.State.(renderers.SudokuLessonState)
				if name == "default" {
					svg := renderers.RenderSudokuLessonSVG(step, 320, 450)
					if !strings.Contains(svg, `data-phase="`+s.Phase+`"`) {
						t.Fatal("real snapshot render mismatch")
					}
				}
				for r := 0; r < 9; r++ {
					for c := 0; c < 9; c++ {
						if s.Given[r][c] != (board[r][c] != 0) || (s.Given[r][c] && s.Board[r][c] != board[r][c]) {
							t.Fatal("given changed")
						}
					}
				}
				if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
					t.Fatal("invalid hold")
				}
				if i == 0 {
					continue
				}
				prev := steps[i-1].State.(renderers.SudokuLessonState)
				switch step.LineID {
				case "L6", "L7":
					if s.Board != prev.Board {
						t.Fatal("candidate changed board")
					}
				case "L8", "L12":
					changed := 0
					for r := 0; r < 9; r++ {
						for c := 0; c < 9; c++ {
							if s.Board[r][c] != prev.Board[r][c] {
								changed++
								if r != s.Row || c != s.Col || s.Given[r][c] {
									t.Fatal("wrong cell changed")
								}
							}
						}
					}
					if changed != 1 {
						t.Fatal("mutation not one cell")
					}
					if step.LineID == "L8" {
						if !IsValid(prev.Board, s.Row, s.Col, s.Candidate) || s.Board[s.Row][s.Col] != s.Candidate {
							t.Fatal("invalid placement")
						}
						if !placed && name == "default" && (s.Row != 0 || s.Col != 2 || s.Candidate != 4) {
							t.Fatal("first placement")
						}
						placed = true
						if steps[i-1].HoldMS+step.HoldMS != 1800 {
							t.Fatal("pair timing")
						}
					} else {
						removed = true
						if s.Board[s.Row][s.Col] != 0 {
							t.Fatal("rollback not zero")
						}
					}
				}
				if s.Phase == "reject" {
					if len(s.Conflicts) != 1 || len(s.ConflictKinds) != 1 {
						t.Fatal("missing witness")
					}
					p := s.Conflicts[0]
					if name == "default" && s.Row == 0 && s.Col == 2 && s.Candidate == 1 && (p != [2]int{0, 7} || s.ConflictKinds[0] != "row") {
						t.Fatal("first row witness order changed")
					}
					if s.Board[p[0]][p[1]] != s.Candidate {
						t.Fatal("false witness")
					}
					kind := s.ConflictKinds[0]
					if kind == "row" && p[0] != s.Row || kind == "column" && p[1] != s.Col || kind == "box" && (p[0]/3 != s.Row/3 || p[1]/3 != s.Col/3) {
						t.Fatal("wrong witness kind")
					}
				}
			}
			if name == "rollback" && !removed {
				t.Fatal("fixture lacks L12")
			}
			first := steps[0].State.(renderers.SudokuLessonState)
			next := steps[1].State.(renderers.SudokuLessonState)
			first.Cells[0].Value = "changed"
			first.RowHeaders[0] = "changed"
			if next.Cells[0].Value == "changed" || next.RowHeaders[0] == "changed" {
				t.Fatal("shared table slices")
			}
			for i, step := range steps {
				s := step.State.(renderers.SudokuLessonState)
				if len(s.Conflicts) > 0 {
					original := s.Conflicts[0]
					s.Conflicts[0] = [2]int{-1, -1}
					s.ConflictKinds[0] = "changed"
					for j := i + 1; j < len(steps); j++ {
						other := steps[j].State.(renderers.SudokuLessonState)
						if len(other.Conflicts) > 0 && (other.Conflicts[0] == [2]int{-1, -1} || other.ConflictKinds[0] == "changed") {
							t.Fatal("shared conflicts")
						}
					}
					s.Conflicts[0] = original
					break
				}
			}
		})
	}
}
