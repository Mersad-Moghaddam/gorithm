package gradientdescent

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type epochRecord struct {
	epoch int
	w     float64
	b     float64
	loss  float64
	valid bool
}

func buildTableState(records []epochRecord, activeEpoch int, activeState string) renderers.TableState {
	total := len(records)
	maxRows := 10
	start := 0
	if activeEpoch >= maxRows {
		start = activeEpoch - maxRows + 1
	}
	end := start + maxRows
	if end > total {
		end = total
	}

	numRows := end - start
	if numRows <= 0 {
		numRows = 1
		start = 0
		end = 1
	}

	rowHeaders := make([]string, numRows)
	cells := make([]renderers.TableCell, 0, numRows*4)

	for r := 0; r < numRows; r++ {
		epIdx := start + r
		rowHeaders[r] = fmt.Sprintf("ep %d", epIdx)

		rec := epochRecord{epoch: epIdx}
		if epIdx < len(records) {
			rec = records[epIdx]
		}

		st := "normal"
		if epIdx == activeEpoch {
			st = activeState
		}

		wVal, bVal, lossVal := "-", "-", "-"
		if rec.valid {
			wVal = fmt.Sprintf("%.4f", rec.w)
			bVal = fmt.Sprintf("%.4f", rec.b)
			lossVal = fmt.Sprintf("%.4f", rec.loss)
		}

		cells = append(cells,
			renderers.TableCell{Row: r, Col: 0, Value: fmt.Sprintf("%d", epIdx), State: st},
			renderers.TableCell{Row: r, Col: 1, Value: wVal, State: st},
			renderers.TableCell{Row: r, Col: 2, Value: bVal, State: st},
			renderers.TableCell{Row: r, Col: 3, Value: lossVal, State: st},
		)
	}

	return renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: []string{"Epoch", "Weight", "Bias", "MSE Loss"},
		Cells:      cells,
	}
}

// Trace records the step-by-step execution of gradient descent linear regression.
func Trace(x, y []float64, lr float64, epochs int) []trace.Step {
	b := trace.NewBuilder()
	n := len(x)

	if n == 0 || n != len(y) || epochs <= 0 {
		emptyState := renderers.TableState{
			RowHeaders: []string{"ep 0"},
			ColHeaders: []string{"Epoch", "Weight", "Bias", "MSE Loss"},
			Cells: []renderers.TableCell{
				{Row: 0, Col: 0, Value: "0", State: "normal"},
				{Row: 0, Col: 1, Value: "-", State: "normal"},
				{Row: 0, Col: 2, Value: "-", State: "normal"},
				{Row: 0, Col: 3, Value: "-", State: "normal"},
			},
		}
		b.AddStep("L1", trace.ActionInit, "Invalid or empty dataset for gradient descent.", emptyState, map[string]any{"epoch": 0, "loss": 0.0, "w": 0.0, "b": 0.0})
		b.AddStep("L17", trace.ActionDone, "Training completed with empty inputs.", emptyState, map[string]any{"epoch": 0, "loss": 0.0, "w": 0.0, "b": 0.0})
		return b.Steps()
	}

	w, bVal := 0.0, 0.0
	records := make([]epochRecord, epochs)
	for i := 0; i < epochs; i++ {
		records[i] = epochRecord{epoch: i, valid: false}
	}

	// Initial step
	records[0] = epochRecord{epoch: 0, w: w, b: bVal, loss: 0.0, valid: true}
	initState := buildTableState(records, 0, "active")
	b.AddStep("L2", trace.ActionInit, "Initialize parameters: weight w = 0.0000, bias b = 0.0000.", initState, map[string]any{
		"epoch": 0,
		"w":     w,
		"b":     bVal,
		"loss":  0.0,
	})

	var lastLoss float64

	for epoch := 0; epoch < epochs; epoch++ {
		curState := buildTableState(records, epoch, "active")
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Epoch %d: evaluate forward predictions and cost.", epoch), curState, map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  lastLoss,
		})

		var cost, dw, db float64
		for i := 0; i < n; i++ {
			pred := w*x[i] + bVal
			err := pred - y[i]
			cost += err * err
			dw += err * x[i]
			db += err
		}

		mse := cost / float64(n)
		lastLoss = mse
		records[epoch] = epochRecord{epoch: epoch, w: w, b: bVal, loss: mse, valid: true}

		lossState := buildTableState(records, epoch, "active")
		b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Epoch %d: computed Mean Squared Error loss = %.4f.", epoch, mse), lossState, map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"dw":    dw,
			"db":    db,
		})

		w -= lr * (2.0 / float64(n)) * dw
		b.IncOperations()
		wState := buildTableState(records, epoch, "computed")
		b.AddStep("L14", trace.ActionSet, fmt.Sprintf("Epoch %d: updated weight w = %.4f.", epoch, w), wState, map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"dw":    dw,
		})

		bVal -= lr * (2.0 / float64(n)) * db
		b.IncOperations()
		records[epoch] = epochRecord{epoch: epoch, w: w, b: bVal, loss: mse, valid: true}
		bState := buildTableState(records, epoch, "computed")
		b.AddStep("L15", trace.ActionSet, fmt.Sprintf("Epoch %d: updated bias b = %.4f.", epoch, bVal), bState, map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"db":    db,
		})
	}

	finalState := buildTableState(records, epochs-1, "computed")
	b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Training complete after %d epochs. Final weight w = %.4f, bias b = %.4f, MSE loss = %.4f.", epochs, w, bVal, lastLoss), finalState, map[string]any{
		"epoch": epochs,
		"w":     w,
		"b":     bVal,
		"loss":  lastLoss,
	})

	return b.Steps()
}
