package coinchange

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestMinCoins(t *testing.T) {
	tests := []struct {
		name       string
		coins      []int
		amount     int
		wantCoins  int
		wantDPSize int
	}{
		{
			name:       "standard amount 11",
			coins:      []int{1, 2, 5},
			amount:     11,
			wantCoins:  3,
			wantDPSize: 12,
		},
		{
			name:       "impossible amount",
			coins:      []int{2},
			amount:     3,
			wantCoins:  -1,
			wantDPSize: 4,
		},
		{
			name:       "zero amount",
			coins:      []int{1},
			amount:     0,
			wantCoins:  0,
			wantDPSize: 1,
		},
		{
			name:       "larger amount",
			coins:      []int{1, 2, 5, 10},
			amount:     27,
			wantCoins:  4,
			wantDPSize: 28,
		},
		{
			name:       "greedy failure case 1, 5, 6, 9 amount 11",
			coins:      []int{1, 5, 6, 9},
			amount:     11,
			wantCoins:  2,
			wantDPSize: 12,
		},
		{
			name:       "negative amount",
			coins:      []int{1, 2, 5},
			amount:     -1,
			wantCoins:  -1,
			wantDPSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotDP := MinCoins(tt.coins, tt.amount)
			if gotMin != tt.wantCoins {
				t.Errorf("MinCoins(%v, %d) = %d, want %d", tt.coins, tt.amount, gotMin, tt.wantCoins)
			}
			if len(gotDP) != tt.wantDPSize {
				t.Errorf("len(gotDP) = %d, want %d", len(gotDP), tt.wantDPSize)
			}
		})
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "coinchange" {
		t.Errorf("expected ID 'coinchange', got %q", meta.ID)
	}
	if meta.Name != "Coin Change (DP)" {
		t.Errorf("expected Name 'Coin Change (DP)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(amount · |coins|)" {
		t.Errorf("expected TimeComplexity 'O(amount · |coins|)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(amount)" {
		t.Errorf("expected SpaceComplexity 'O(amount)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Unbounded Knapsack",
		"Overlapping Subproblems",
		"Bottom-up Tabulation",
		"Min Optimization",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	registered, ok := catalog.Get("coinchange")
	if !ok {
		t.Fatal("lesson 'coinchange' not registered in catalog")
	}
	if registered.Meta().ID != "coinchange" {
		t.Errorf("expected registered ID 'coinchange', got %q", registered.Meta().ID)
	}
}

func TestTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
	if minCoins, ok := lastStep.Variables["minCoins"].(int); !ok || minCoins != 3 {
		t.Fatalf("expected minCoins 3, got %v", lastStep.Variables["minCoins"])
	}

	// Verify headers and cells on every step
	wantColHeaders := make([]string, 12)
	for i := 0; i <= 11; i++ {
		wantColHeaders[i] = strconv.Itoa(i)
	}

	for i, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
		}
		if !reflect.DeepEqual(tableState.RowHeaders, []string{"dp[amount]"}) {
			t.Fatalf("step %d RowHeaders = %v, want ['dp[amount]']", i, tableState.RowHeaders)
		}
		if !reflect.DeepEqual(tableState.ColHeaders, wantColHeaders) {
			t.Fatalf("step %d ColHeaders = %v, want %v", i, tableState.ColHeaders, wantColHeaders)
		}
		if len(tableState.Cells) != 12 {
			t.Fatalf("step %d has %d cells, want 12", i, len(tableState.Cells))
		}
		for c, cell := range tableState.Cells {
			if cell.Row != 0 || cell.Col != c {
				t.Fatalf("step %d cell %d has pos (%d, %d), want (0, %d)", i, c, cell.Row, cell.Col, c)
			}
		}
	}

	// Verify base case step (L3)
	var foundBase bool
	for _, step := range steps {
		if step.LineID == "L3" {
			foundBase = true
			ts := step.State.(renderers.TableState)
			if ts.Cells[0].Value != "0" || ts.Cells[0].State != "computed" {
				t.Fatalf("base case step cell 0 value=%q state=%q, want '0' computed", ts.Cells[0].Value, ts.Cells[0].State)
			}
		}
	}
	if !foundBase {
		t.Fatal("base case step L3 not found in trace")
	}
}

func TestTableStateDependenciesAndTeaching(t *testing.T) {
	steps := Trace([]int{1, 2, 5}, 11)
	if len(steps) == 0 {
		t.Fatal("trace is empty")
	}

	var foundEvaluation, foundUpdate bool

	for i, step := range steps {
		ts := step.State.(renderers.TableState)

		if step.LineID == "L7" {
			canTake, ok := step.Variables["canTake"].(bool)
			if !ok {
				t.Fatalf("step %d L7 missing canTake variable", i)
			}

			if canTake {
				foundEvaluation = true
				coin := step.Variables["coin"].(int)
				amount := step.Variables["i"].(int)
				subproblem := step.Variables["subproblem"].(int)

				if len(ts.Dependencies) != 1 {
					t.Fatalf("step %d L7: expected 1 dependency, got %d", i, len(ts.Dependencies))
				}
				dep := ts.Dependencies[0]
				if dep.FromRow != 0 || dep.FromCol != subproblem || dep.ToRow != 0 || dep.ToCol != amount {
					t.Fatalf("step %d L7: dependency mismatch %+v, want (0,%d)->(0,%d)", i, dep, subproblem, amount)
				}

				if ts.Cells[subproblem].State != "source" {
					t.Fatalf("step %d L7: source cell state=%q, want 'source'", i, ts.Cells[subproblem].State)
				}
				if ts.Cells[amount].State != "active" {
					t.Fatalf("step %d L7: active cell state=%q, want 'active'", i, ts.Cells[amount].State)
				}

				wantPrefix := fmt.Sprintf("Testing coin %d for amount %d: dp[%d - %d] + 1 =", coin, amount, amount, coin)
				if !strings.HasPrefix(step.Explanation, wantPrefix) {
					t.Fatalf("step %d L7: explanation %q does not start with %q", i, step.Explanation, wantPrefix)
				}
				if !strings.HasPrefix(ts.Caption, wantPrefix) {
					t.Fatalf("step %d L7: caption %q does not start with %q", i, ts.Caption, wantPrefix)
				}
			} else {
				if len(ts.Dependencies) != 0 {
					t.Fatalf("step %d L7 (c > i): expected 0 dependencies, got %d", i, len(ts.Dependencies))
				}
			}
		}

		if step.LineID == "L8" {
			foundUpdate = true
			amount := step.Variables["i"].(int)
			subproblem := step.Variables["subproblem"].(int)

			if len(ts.Dependencies) != 1 {
				t.Fatalf("step %d L8: expected 1 dependency, got %d", i, len(ts.Dependencies))
			}
			dep := ts.Dependencies[0]
			if dep.FromRow != 0 || dep.FromCol != subproblem || dep.ToRow != 0 || dep.ToCol != amount {
				t.Fatalf("step %d L8: dependency mismatch %+v", i, dep)
			}
			if ts.Cells[amount].State != "computed" {
				t.Fatalf("step %d L8: updated cell state=%q, want 'computed'", i, ts.Cells[amount].State)
			}
			if ts.Cells[subproblem].State != "source" {
				t.Fatalf("step %d L8: source cell state=%q, want 'source'", i, ts.Cells[subproblem].State)
			}
		}
	}

	if !foundEvaluation {
		t.Fatal("no coin evaluation step L7 (canTake=true) observed")
	}
	if !foundUpdate {
		t.Fatal("no DP update step L8 observed")
	}
}

func TestSnapshotImmutability(t *testing.T) {
	steps := Trace([]int{1, 2, 5}, 11)
	if len(steps) < 2 {
		t.Fatalf("too few steps: %d", len(steps))
	}

	// 1. Verify slice backing pointers are completely distinct across steps
	for i := 0; i < len(steps)-1; i++ {
		ts1 := steps[i].State.(renderers.TableState)
		ts2 := steps[i+1].State.(renderers.TableState)

		if len(ts1.Cells) > 0 && len(ts2.Cells) > 0 {
			p1 := unsafe.SliceData(ts1.Cells)
			p2 := unsafe.SliceData(ts2.Cells)
			if p1 == p2 {
				t.Fatalf("steps %d and %d share Cells backing array (%p == %p)", i, i+1, p1, p2)
			}
		}

		if len(ts1.Dependencies) > 0 && len(ts2.Dependencies) > 0 {
			p1 := unsafe.SliceData(ts1.Dependencies)
			p2 := unsafe.SliceData(ts2.Dependencies)
			if p1 == p2 {
				t.Fatalf("steps %d and %d share Dependencies backing array (%p == %p)", i, i+1, p1, p2)
			}
		}
	}

	// 2. Mutate step 0 cells and verify step 1 is unmodified
	ts0 := steps[0].State.(renderers.TableState)
	origVal1 := steps[1].State.(renderers.TableState).Cells[0].Value
	ts0.Cells[0].Value = "MUTATED_VAL"
	ts0.Cells[0].State = "MUTATED_STATE"

	if steps[1].State.(renderers.TableState).Cells[0].Value != origVal1 {
		t.Fatal("mutating step 0 cells leaked into step 1 cells")
	}
}

func TestImpossibleAmount(t *testing.T) {
	// 1. Direct trace with negative amount
	negSteps := Trace([]int{1, 2, 5}, -1)
	if len(negSteps) == 0 || negSteps[0].LineID != "L2" {
		t.Fatalf("expected step L2 on negative amount trace, got %+v", negSteps)
	}

	// 2. Unreachable positive amount
	steps := Trace([]int{2}, 3)
	if len(steps) == 0 {
		t.Fatal("trace is empty")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.LineID != "L12" {
		t.Fatalf("impossible amount should finish at L12, got %q", lastStep.LineID)
	}
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected ActionDone, got %s", lastStep.Action)
	}
	if minCoins, ok := lastStep.Variables["minCoins"].(int); !ok || minCoins != -1 {
		t.Fatalf("expected minCoins -1, got %v", lastStep.Variables["minCoins"])
	}
	if result, ok := lastStep.Variables["result"].(int); !ok || result != -1 {
		t.Fatalf("expected result -1, got %v", lastStep.Variables["result"])
	}

	ts := lastStep.State.(renderers.TableState)
	if ts.Cells[3].Value != "∞" {
		t.Fatalf("expected impossible amount cell 3 to have value '∞', got %q", ts.Cells[3].Value)
	}
}

func TestInputValidation(t *testing.T) {
	lesson := Lesson{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "nil / empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "whitespace only",
			input:   "   \n  \t  ",
			wantErr: false,
		},
		{
			name:    "valid custom input",
			input:   `{"coins": [2, 5], "amount": 7}`,
			wantErr: false,
		},
		{
			name:    "null JSON literal",
			input:   "null",
			wantErr: true,
		},
		{
			name:    "malformed JSON syntax",
			input:   `{"coins": [1, 2], "amount":`,
			wantErr: true,
		},
		{
			name:    "non-object JSON array",
			input:   `[1, 2, 3]`,
			wantErr: true,
		},
		{
			name:    "non-object JSON primitive",
			input:   `"invalid"`,
			wantErr: true,
		},
		{
			name:    "coins is null",
			input:   `{"coins": null, "amount": 5}`,
			wantErr: true,
		},
		{
			name:    "amount is null",
			input:   `{"coins": [1, 2], "amount": null}`,
			wantErr: true,
		},
		{
			name:    "coins is empty",
			input:   `{"coins": [], "amount": 5}`,
			wantErr: true,
		},
		{
			name:    "negative coin denomination",
			input:   `{"coins": [1, -2, 5], "amount": 5}`,
			wantErr: true,
		},
		{
			name:    "zero coin denomination",
			input:   `{"coins": [0, 5], "amount": 5}`,
			wantErr: true,
		},
		{
			name:    "negative amount",
			input:   `{"coins": [1, 2], "amount": -1}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw json.RawMessage
			if tt.input != "" {
				raw = json.RawMessage(tt.input)
			}
			steps, err := lesson.Build(raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for input %q: %v", tt.input, err)
				}
				if len(steps) == 0 {
					t.Fatalf("expected non-empty steps for input %q", tt.input)
				}
			}
		})
	}
}

func TestRenderTableSVG(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	for i, step := range steps {
		svg := renderers.RenderTableSVG(step, 800, 450)
		if svg == "" {
			t.Errorf("step %d produced empty SVG", i)
		}
	}
}
