package knapsack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSolveCorrectness(t *testing.T) {
	tests := []struct {
		name      string
		weights   []int
		values    []int
		capacity  int
		wantMax   int
		wantTable [][]int
	}{
		{
			name:     "default example",
			weights:  []int{2, 3, 4, 5},
			values:   []int{3, 4, 5, 6},
			capacity: 5,
			wantMax:  7, // items 0 and 1: weights 2+3=5, values 3+4=7
		},
		{
			name:     "zero capacity",
			weights:  []int{1, 2, 3},
			values:   []int{10, 20, 30},
			capacity: 0,
			wantMax:  0,
		},
		{
			name:     "empty items",
			weights:  []int{},
			values:   []int{},
			capacity: 10,
			wantMax:  0,
		},
		{
			name:     "all items fit",
			weights:  []int{1, 2, 3},
			values:   []int{6, 10, 12},
			capacity: 10,
			wantMax:  28,
		},
		{
			name:     "no item fits",
			weights:  []int{5, 6, 7},
			values:   []int{10, 20, 30},
			capacity: 4,
			wantMax:  0,
		},
		{
			name:     "single item fits",
			weights:  []int{3},
			values:   []int{25},
			capacity: 3,
			wantMax:  25,
		},
		{
			name:     "single item does not fit",
			weights:  []int{5},
			values:   []int{25},
			capacity: 3,
			wantMax:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMax, gotTable := Solve(tt.weights, tt.values, tt.capacity)
			if gotMax != tt.wantMax {
				t.Errorf("Solve() gotMax = %d, want %d", gotMax, tt.wantMax)
			}
			if len(gotTable) != len(tt.weights)+1 {
				t.Fatalf("Solve() table rows = %d, want %d", len(gotTable), len(tt.weights)+1)
			}
			for r, row := range gotTable {
				if len(row) != tt.capacity+1 {
					t.Fatalf("Solve() table row %d len = %d, want %d", r, len(row), tt.capacity+1)
				}
			}
			if gotTable[len(tt.weights)][tt.capacity] != tt.wantMax {
				t.Errorf("Solve() table end cell = %d, want %d", gotTable[len(tt.weights)][tt.capacity], tt.wantMax)
			}
		})
	}
}

func TestKnapsackMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "knapsack" {
		t.Errorf("expected ID 'knapsack', got %q", meta.ID)
	}
	if meta.Name != "0/1 Knapsack (DP)" {
		t.Errorf("expected Name '0/1 Knapsack (DP)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n · W)" {
		t.Errorf("expected TimeComplexity 'O(n · W)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n · W)" {
		t.Errorf("expected SpaceComplexity 'O(n · W)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"2D State Matrix", "Include/Exclude Choice", "Pseudo-Polynomial Time"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("knapsack")
	if !ok {
		t.Fatal("lesson 'knapsack' not registered in catalog")
	}
	if l.Meta().ID != "knapsack" {
		t.Errorf("expected registered ID 'knapsack', got %q", l.Meta().ID)
	}
}

func TestKnapsackTraceValidation(t *testing.T) {
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
	if maxVal, ok := lastStep.Variables["maxValue"].(int); !ok || maxVal != 7 {
		t.Fatalf("expected maxValue 7, got %v", lastStep.Variables["maxValue"])
	}
}

func TestKnapsackTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"weights": [1, 2, 3], "values": [6, 10, 12], "capacity": 5}`)
	steps, err := lesson.Build(json.RawMessage(input))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
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
	// Items 1 and 2 (w=2,v=10 and w=3,v=12) give 22. Items 0 and 2 (w=1,v=6 and w=3,v=12) give 18.
	// Weights 1+2=3 < 5 (v=16). Weights 1+2+3=6 > 5. Weights 2+3=5 (v=22).
	if maxVal, ok := lastStep.Variables["maxValue"].(int); !ok || maxVal != 22 {
		t.Fatalf("expected maxValue 22, got %v", lastStep.Variables["maxValue"])
	}
}

func TestKnapsackRenderTableSVG(t *testing.T) {
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

func TestKnapsackInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid empty input returns default steps
	if steps, err := lesson.Build(nil); err != nil || len(steps) == 0 {
		t.Fatalf("expected nil input to succeed, got err=%v, steps=%d", err, len(steps))
	}
	if steps, err := lesson.Build(json.RawMessage("")); err != nil || len(steps) == 0 {
		t.Fatalf("expected empty input to succeed, got err=%v, steps=%d", err, len(steps))
	}

	// Invalid inputs
	invalidCases := []struct {
		name  string
		input string
	}{
		{"null literal", "null"},
		{"malformed JSON", `{"weights": [1, 2,}`},
		{"mismatched weights/values lengths", `{"weights": [1, 2], "values": [1], "capacity": 5}`},
		{"negative capacity", `{"weights": [1], "values": [1], "capacity": -1}`},
		{"missing capacity", `{"weights": [1], "values": [1]}`},
		{"missing weights", `{"values": [1], "capacity": 5}`},
		{"missing values", `{"weights": [1], "capacity": 5}`},
		{"negative weight", `{"weights": [-1], "values": [1], "capacity": 5}`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Errorf("expected error for input %s, got nil", tc.input)
			}
		})
	}
}

func TestKnapsackTableState(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil) // weights=[2, 3, 4, 5], values=[3, 4, 5, 6], capacity=5
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	weights := []int{2, 3, 4, 5}
	values := []int{3, 4, 5, 6}
	capacity := 5
	n := len(weights)

	// Check table dimensions on every step
	for i, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState, got %T", i, step.State)
		}
		if len(tableState.RowHeaders) != n+1 {
			t.Errorf("step %d RowHeaders len = %d, want %d", i, len(tableState.RowHeaders), n+1)
		}
		if len(tableState.ColHeaders) != capacity+1 {
			t.Errorf("step %d ColHeaders len = %d, want %d", i, len(tableState.ColHeaders), capacity+1)
		}
		if len(tableState.Cells) != (n+1)*(capacity+1) {
			t.Errorf("step %d Cells len = %d, want %d", i, len(tableState.Cells), (n+1)*(capacity+1))
		}
	}

	// Verify RowHeaders format: Item 0 (w:0, v:0), Item 1 (w:2, v:3), etc.
	firstState := steps[0].State.(renderers.TableState)
	if firstState.RowHeaders[0] != "Item 0 (w:0, v:0)" {
		t.Errorf("RowHeaders[0] = %q, want %q", firstState.RowHeaders[0], "Item 0 (w:0, v:0)")
	}
	for i := 1; i <= n; i++ {
		wantHeader := fmt.Sprintf("Item %d (w:%d, v:%d)", i, weights[i-1], values[i-1])
		if firstState.RowHeaders[i] != wantHeader {
			t.Errorf("RowHeaders[%d] = %q, want %q", i, firstState.RowHeaders[i], wantHeader)
		}
	}

	// Verify ColHeaders format: "0", "1", ..., "5"
	for c := 0; c <= capacity; c++ {
		if firstState.ColHeaders[c] != strconv.Itoa(c) {
			t.Errorf("ColHeaders[%d] = %q, want %q", c, firstState.ColHeaders[c], strconv.Itoa(c))
		}
	}

	// Check compare step dependencies and active/source cell states
	for stepIdx, step := range steps {
		if step.Action == trace.ActionCompare {
			tableState := step.State.(renderers.TableState)
			iVar, ok1 := step.Variables["i"].(int)
			cVar, ok2 := step.Variables["c"].(int)
			wVar, ok3 := step.Variables["w"].(int)
			if !ok1 || !ok2 || !ok3 {
				continue
			}

			// Active cell (i, c)
			var activeCell *renderers.TableCell
			for idx := range tableState.Cells {
				cell := &tableState.Cells[idx]
				if cell.Row == iVar && cell.Col == cVar {
					activeCell = cell
					break
				}
			}
			if activeCell == nil || activeCell.State != "active" {
				t.Errorf("step %d (compare i=%d, c=%d): cell (%d, %d) state = %v, want active", stepIdx, iVar, cVar, iVar, cVar, activeCell)
			}

			// Check dependencies
			if wVar <= cVar {
				// Fits: exclude source (i-1, c) and include source (i-1, c-w)
				expectedDeps := 2
				if wVar == 0 {
					expectedDeps = 1
				}
				if len(tableState.Dependencies) != expectedDeps {
					t.Errorf("step %d (compare i=%d, c=%d): expected %d dependencies, got %d", stepIdx, iVar, cVar, expectedDeps, len(tableState.Dependencies))
				}
				for _, dep := range tableState.Dependencies {
					if dep.ToRow != iVar || dep.ToCol != cVar {
						t.Errorf("step %d: dependency target is (%d, %d), want (%d, %d)", stepIdx, dep.ToRow, dep.ToCol, iVar, cVar)
					}
				}
			} else {
				// Cannot fit: exclude source (i-1, c) only
				if len(tableState.Dependencies) != 1 {
					t.Errorf("step %d (compare i=%d, c=%d): expected 1 dependency, got %d", stepIdx, iVar, cVar, len(tableState.Dependencies))
				}
				if len(tableState.Dependencies) == 1 {
					dep := tableState.Dependencies[0]
					if dep.FromRow != iVar-1 || dep.FromCol != cVar || dep.ToRow != iVar || dep.ToCol != cVar {
						t.Errorf("step %d: unexpected dependency %+v", stepIdx, dep)
					}
				}
			}
		}
	}

	// Final step check
	lastStep := steps[len(steps)-1]
	lastTable := lastStep.State.(renderers.TableState)
	var finalCell *renderers.TableCell
	for idx := range lastTable.Cells {
		cell := &lastTable.Cells[idx]
		if cell.Row == n && cell.Col == capacity {
			finalCell = cell
			break
		}
	}
	if finalCell == nil || finalCell.Value != "7" {
		t.Fatalf("final cell value = %v, want 7", finalCell)
	}
	if finalCell.State != "computed" {
		t.Errorf("final cell state = %q, want 'computed'", finalCell.State)
	}

	// Caption check
	if !strings.Contains(lastTable.Caption, "Optimal value: 7") {
		t.Errorf("expected caption to contain 'Optimal value: 7', got %q", lastTable.Caption)
	}
	if !strings.Contains(lastTable.Caption, "Item 1, Item 2") {
		t.Errorf("expected caption to list selected items 'Item 1, Item 2', got %q", lastTable.Caption)
	}
}

func TestKnapsackSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	// Record original cell values and states across all steps
	type cellSnapshot struct {
		val   string
		state string
	}
	type stepRecord struct {
		cells    []cellSnapshot
		depCount int
	}

	records := make([]stepRecord, len(steps))
	for i, step := range steps {
		ts := step.State.(renderers.TableState)
		rec := stepRecord{
			cells:    make([]cellSnapshot, len(ts.Cells)),
			depCount: len(ts.Dependencies),
		}
		for j, c := range ts.Cells {
			rec.cells[j] = cellSnapshot{val: c.Value, state: c.State}
		}
		records[i] = rec
	}

	// Mutate each step's Cells and Dependencies in-place
	for _, step := range steps {
		ts := step.State.(renderers.TableState)
		for j := range ts.Cells {
			ts.Cells[j].Value = "MUTATED"
			ts.Cells[j].State = "MUTATED"
		}
		ts.Dependencies = append(ts.Dependencies, renderers.TableDependency{FromRow: 999, FromCol: 999, ToRow: 999, ToCol: 999})
	}

	// Generate fresh steps and verify that original records matched
	freshSteps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build fresh trace: %v", err)
	}

	for i, step := range freshSteps {
		ts := step.State.(renderers.TableState)
		rec := records[i]
		if len(ts.Cells) != len(rec.cells) {
			t.Fatalf("step %d cell count changed", i)
		}
		for j, c := range ts.Cells {
			if c.Value != rec.cells[j].val || c.State != rec.cells[j].state {
				t.Fatalf("step %d cell %d mutated: got (%q, %q), want (%q, %q)", i, j, c.Value, c.State, rec.cells[j].val, rec.cells[j].state)
			}
		}
		if len(ts.Dependencies) != rec.depCount {
			t.Fatalf("step %d dependency count mutated: got %d, want %d", i, len(ts.Dependencies), rec.depCount)
		}
	}
}

func TestKnapsackSelectedItems(t *testing.T) {
	tests := []struct {
		name         string
		weights      []int
		values       []int
		capacity     int
		wantMax      int
		wantSelected []int
	}{
		{
			name:         "default",
			weights:      []int{2, 3, 4, 5},
			values:       []int{3, 4, 5, 6},
			capacity:     5,
			wantMax:      7,
			wantSelected: []int{1, 2},
		},
		{
			name:         "all items fit",
			weights:      []int{1, 2, 3},
			values:       []int{6, 10, 12},
			capacity:     10,
			wantMax:      28,
			wantSelected: []int{1, 2, 3},
		},
		{
			name:         "no item fits",
			weights:      []int{5, 6, 7},
			values:       []int{10, 20, 30},
			capacity:     4,
			wantMax:      0,
			wantSelected: []int{},
		},
		{
			name:         "zero capacity",
			weights:      []int{1, 2},
			values:       []int{10, 20},
			capacity:     0,
			wantMax:      0,
			wantSelected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := Trace(tt.weights, tt.values, tt.capacity)
			lastStep := steps[len(steps)-1]
			if maxVal, ok := lastStep.Variables["maxValue"].(int); !ok || maxVal != tt.wantMax {
				t.Errorf("maxValue = %v, want %d", lastStep.Variables["maxValue"], tt.wantMax)
			}
			selected, ok := lastStep.Variables["selectedItems"].([]int)
			if !ok {
				t.Fatalf("selectedItems missing or wrong type: %T", lastStep.Variables["selectedItems"])
			}
			if len(selected) == 0 && len(tt.wantSelected) == 0 {
				// match
			} else if !reflect.DeepEqual(selected, tt.wantSelected) {
				t.Errorf("selectedItems = %v, want %v", selected, tt.wantSelected)
			}
		})
	}
}
