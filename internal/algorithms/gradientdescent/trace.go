package gradientdescent

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type pointRecord struct {
	pred      float64
	err       float64
	predValid bool
	errValid  bool
}

func cloneTableState(s renderers.TableState) renderers.TableState {
	rowHeaders := make([]string, len(s.RowHeaders))
	copy(rowHeaders, s.RowHeaders)

	colHeaders := make([]string, len(s.ColHeaders))
	copy(colHeaders, s.ColHeaders)

	cells := make([]renderers.TableCell, len(s.Cells))
	copy(cells, s.Cells)

	var deps []renderers.TableDependency
	if len(s.Dependencies) > 0 {
		deps = make([]renderers.TableDependency, len(s.Dependencies))
		copy(deps, s.Dependencies)
	}

	return renderers.TableState{
		Caption:      s.Caption,
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

func buildTableState(
	x, y []float64,
	records []pointRecord,
	activeRow int,
	activeStage string, // "init", "pred", "err", "computed", "all_computed"
	caption string,
) renderers.TableState {
	n := len(x)
	rowHeaders := make([]string, n)
	cells := make([]renderers.TableCell, 0, n*4)

	for i := 0; i < n; i++ {
		rowHeaders[i] = fmt.Sprintf("pt %d", i)

		predStr, errStr := "-", "-"
		if records[i].predValid {
			predStr = fmt.Sprintf("%.4f", records[i].pred)
		}
		if records[i].errValid {
			errStr = fmt.Sprintf("%.4f", records[i].err)
		}

		var xState, yState, predState, errState string
		if activeStage == "all_computed" {
			xState, yState, predState, errState = "computed", "computed", "computed", "computed"
		} else if activeStage == "init" {
			xState, yState, predState, errState = "normal", "normal", "normal", "normal"
		} else if i == activeRow {
			switch activeStage {
			case "pred":
				xState, yState, predState = "active", "active", "active"
				errState = "normal"
				if records[i].errValid {
					errState = "computed"
				}
			case "err":
				xState, yState, predState, errState = "active", "active", "active", "active"
			case "computed":
				xState, yState, predState, errState = "computed", "computed", "computed", "computed"
			default:
				xState, yState, predState, errState = "active", "active", "active", "active"
			}
		} else if records[i].errValid {
			xState, yState, predState, errState = "computed", "computed", "computed", "computed"
		} else {
			xState, yState, predState, errState = "normal", "normal", "normal", "normal"
		}

		cells = append(cells,
			renderers.TableCell{Row: i, Col: 0, Value: fmt.Sprintf("%.2f", x[i]), State: xState},
			renderers.TableCell{Row: i, Col: 1, Value: fmt.Sprintf("%.2f", y[i]), State: yState},
			renderers.TableCell{Row: i, Col: 2, Value: predStr, State: predState},
			renderers.TableCell{Row: i, Col: 3, Value: errStr, State: errState},
		)
	}

	return renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: []string{"x", "y", "y_pred", "residual"},
		Cells:      cells,
	}
}

// Trace records the step-by-step execution of gradient descent linear regression.
func Trace(x, y []float64, lr float64, epochs int) []trace.Step {
	b := trace.NewBuilder()
	n := len(x)

	if n == 0 || n != len(y) || epochs <= 0 {
		emptyState := renderers.TableState{
			RowHeaders: []string{"pt 0"},
			ColHeaders: []string{"x", "y", "y_pred", "residual"},
			Cells: []renderers.TableCell{
				{Row: 0, Col: 0, Value: "-", State: "normal"},
				{Row: 0, Col: 1, Value: "-", State: "normal"},
				{Row: 0, Col: 2, Value: "-", State: "normal"},
				{Row: 0, Col: 3, Value: "-", State: "normal"},
			},
		}
		b.AddStep("L1", trace.ActionInit, "Invalid or empty dataset for gradient descent.", cloneTableState(emptyState), map[string]any{"epoch": 0, "loss": 0.0, "w": 0.0, "b": 0.0, "dw": 0.0, "db": 0.0})
		b.AddStep("L17", trace.ActionDone, "Training completed with empty inputs.", cloneTableState(emptyState), map[string]any{"epoch": 0, "loss": 0.0, "w": 0.0, "b": 0.0, "dw": 0.0, "db": 0.0})
		return b.Steps()
	}

	w, bVal := 0.0, 0.0
	records := make([]pointRecord, n)

	// Initial step
	initState := buildTableState(x, y, records, -1, "init", "Initialize parameters: w = 0.0000, b = 0.0000")
	b.AddStep("L2", trace.ActionInit, "Initialize parameters: weight w = 0.0000, bias b = 0.0000.", cloneTableState(initState), map[string]any{
		"epoch": 0,
		"w":     w,
		"b":     bVal,
		"loss":  0.0,
		"dw":    0.0,
		"db":    0.0,
	})

	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Initialize cost history tracking for %d epochs.", epochs), cloneTableState(initState), map[string]any{
		"epoch": 0,
		"w":     w,
		"b":     bVal,
		"loss":  0.0,
		"dw":    0.0,
		"db":    0.0,
	})

	var lastLoss float64

	for epoch := 0; epoch < epochs; epoch++ {
		for k := 0; k < n; k++ {
			records[k] = pointRecord{}
		}

		epCaption := fmt.Sprintf("Epoch %d: evaluate forward predictions and cost across %d points.", epoch, n)
		epState := buildTableState(x, y, records, -1, "init", epCaption)
		b.AddStep("L4", trace.ActionHighlight, epCaption, cloneTableState(epState), map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  lastLoss,
			"dw":    0.0,
			"db":    0.0,
		})

		var cost, dw, db float64
		resetState := buildTableState(x, y, records, -1, "init", fmt.Sprintf("Epoch %d: cost=0, dw=0, db=0", epoch))
		b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Epoch %d: initialize loss and gradient accumulators to zero.", epoch), cloneTableState(resetState), map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  lastLoss,
			"cost":  cost,
			"dw":    dw,
			"db":    db,
		})

		for i := 0; i < n; i++ {
			pred := w*x[i] + bVal
			records[i].pred = pred
			records[i].predValid = true
			predCaption := fmt.Sprintf("Epoch %d, pt %d: x=%.2f, y=%.2f -> y_pred=%.4f", epoch, i, x[i], y[i], pred)
			predState := buildTableState(x, y, records, i, "pred", predCaption)
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Point %d (x=%.2f, y=%.2f): compute prediction y_pred = w*x + b = (%.4f)*(%.2f) + %.4f = %.4f.", i, x[i], y[i], w, x[i], bVal, pred), cloneTableState(predState), map[string]any{
				"epoch": epoch,
				"i":     i,
				"x":     x[i],
				"y":     y[i],
				"pred":  pred,
				"w":     w,
				"b":     bVal,
				"loss":  lastLoss,
				"dw":    dw,
				"db":    db,
			})

			err := pred - y[i]
			records[i].err = err
			records[i].errValid = true
			errCaption := fmt.Sprintf("Epoch %d, pt %d: residual = %.4f - %.2f = %.4f", epoch, i, pred, y[i], err)
			errState := buildTableState(x, y, records, i, "err", errCaption)
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Point %d: compute residual error = y_pred - y = %.4f - %.2f = %.4f.", i, pred, y[i], err), cloneTableState(errState), map[string]any{
				"epoch": epoch,
				"i":     i,
				"x":     x[i],
				"y":     y[i],
				"pred":  pred,
				"err":   err,
				"w":     w,
				"b":     bVal,
				"loss":  lastLoss,
				"dw":    dw,
				"db":    db,
			})

			cost += err * err
			dw += err * x[i]
			db += err
			b.IncOperations()

			compCaption := fmt.Sprintf("Epoch %d: accumulated dw=%.4f, db=%.4f after pt %d", epoch, dw, db, i)
			compState := buildTableState(x, y, records, i, "computed", compCaption)
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Point %d: accumulate gradients (dw += %.4f, db += %.4f) and cost (cost += %.4f).", i, err*x[i], err, err*err), cloneTableState(compState), map[string]any{
				"epoch": epoch,
				"i":     i,
				"x":     x[i],
				"y":     y[i],
				"pred":  pred,
				"err":   err,
				"cost":  cost,
				"dw":    dw,
				"db":    db,
				"w":     w,
				"b":     bVal,
				"loss":  cost / float64(i+1),
			})
		}

		mse := cost / float64(n)
		lastLoss = mse

		updateCaption := fmt.Sprintf("Epoch %d: Loss=%.4f, updating w (dw=%.4f) and b (db=%.4f)", epoch, mse, dw, db)

		lossState := buildTableState(x, y, records, -1, "all_computed", updateCaption)
		b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Epoch %d: computed Mean Squared Error loss = %.4f.", epoch, mse), cloneTableState(lossState), map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"cost":  cost,
			"dw":    dw,
			"db":    db,
		})

		w -= lr * (2.0 / float64(n)) * dw
		b.IncOperations()
		wState := buildTableState(x, y, records, -1, "all_computed", updateCaption)
		b.AddStep("L14", trace.ActionSet, fmt.Sprintf("Epoch %d: updated weight w = %.4f.", epoch, w), cloneTableState(wState), map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"cost":  cost,
			"dw":    dw,
			"db":    db,
		})

		bVal -= lr * (2.0 / float64(n)) * db
		b.IncOperations()
		bState := buildTableState(x, y, records, -1, "all_computed", updateCaption)
		b.AddStep("L15", trace.ActionSet, fmt.Sprintf("Epoch %d: updated bias b = %.4f.", epoch, bVal), cloneTableState(bState), map[string]any{
			"epoch": epoch,
			"w":     w,
			"b":     bVal,
			"loss":  mse,
			"cost":  cost,
			"dw":    dw,
			"db":    db,
		})
	}

	finalCaption := fmt.Sprintf("Final: Loss=%.4f, w=%.4f, b=%.4f", lastLoss, w, bVal)
	finalState := buildTableState(x, y, records, -1, "all_computed", finalCaption)
	b.AddStep("L17", trace.ActionDone, fmt.Sprintf("Training complete after %d epochs. Final parameters: weight w = %.4f, bias b = %.4f, MSE loss = %.4f.", epochs, w, bVal, lastLoss), cloneTableState(finalState), map[string]any{
		"epoch": epochs,
		"w":     w,
		"b":     bVal,
		"loss":  lastLoss,
		"dw":    0.0,
		"db":    0.0,
	})

	return b.Steps()
}
