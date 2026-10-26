//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type Player struct {
	lesson      catalog.Lesson
	steps       []trace.Step
	stepIdx     int
	playing     bool
	speed       float64
	doc         js.Value
	mu          sync.Mutex
	stopChannel chan struct{}
}

func NewPlayer(doc js.Value) *Player {
	return &Player{
		speed: 1.0,
		doc:   doc,
	}
}

func (p *Player) LoadLesson(l catalog.Lesson) {
	p.LoadLessonWithInput(l, nil)
}

func (p *Player) LoadLessonWithInput(l catalog.Lesson, input []byte) {
	p.Pause()

	p.mu.Lock()
	p.lesson = l
	steps, err := l.Build(input)
	if err != nil {
		fmt.Printf("Error building lesson: %v\n", err)
		steps, _ = l.Build(nil)
	}
	p.steps = steps
	p.stepIdx = 0
	p.mu.Unlock()

	p.renderLessonShell()
	p.RenderStep(0)
}

func (p *Player) Shuffle() {
	p.mu.Lock()
	l := p.lesson
	p.mu.Unlock()
	if l == nil {
		return
	}
	shuffled := generateShuffledInput(l.Meta())
	p.LoadLessonWithInput(l, shuffled)
}

func generateShuffledInput(m catalog.Meta) []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	switch m.ID {
	case "bubble", "selection", "insertion", "merge", "quick", "heap", "countingsort", "radixsort", "bucketsort":
		length := 6 + r.Intn(3)
		arr := make([]int, length)
		for i := range arr {
			arr[i] = 10 + r.Intn(88)
		}
		data, _ := json.Marshal(arr)
		return data

	case "linearsearch":
		arr := make([]int, 7)
		for i := range arr {
			arr[i] = 10 + r.Intn(85)
		}
		target := arr[r.Intn(len(arr))]
		if r.Float64() < 0.25 {
			target = 99
		}
		data, _ := json.Marshal(map[string]any{"array": arr, "target": target})
		return data

	case "binarysearch":
		arr := make([]int, 8)
		cur := 5 + r.Intn(10)
		for i := range arr {
			arr[i] = cur
			cur += 4 + r.Intn(12)
		}
		target := arr[r.Intn(len(arr))]
		data, _ := json.Marshal(map[string]any{"array": arr, "target": target})
		return data

	case "bst", "avl":
		length := 6 + r.Intn(2)
		used := make(map[int]bool)
		arr := make([]int, 0, length)
		for len(arr) < length {
			v := 10 + r.Intn(85)
			if !used[v] {
				used[v] = true
				arr = append(arr, v)
			}
		}
		if m.ID == "bst" {
			data, _ := json.Marshal(map[string]any{
				"initial": arr,
				"ops": []map[string]any{
					{"type": "insert", "value": 10 + r.Intn(85)},
					{"type": "search", "value": arr[r.Intn(len(arr))]},
				},
			})
			return data
		}
		data, _ := json.Marshal(arr)
		return data

	case "hanoitower":
		disks := 3 + r.Intn(2)
		data, _ := json.Marshal(disks)
		return data

	case "fibonacci":
		n := 5 + r.Intn(5)
		data, _ := json.Marshal(map[string]any{"n": n})
		return data

	case "nqueens":
		n := 4 + r.Intn(2)
		data, _ := json.Marshal(map[string]any{"n": n})
		return data

	case "coinchange":
		amount := 8 + r.Intn(8)
		data, _ := json.Marshal(map[string]any{"coins": []int{1, 2, 5}, "amount": amount})
		return data

	case "kmeans":
		pts := make([]map[string]int, 12)
		for i := range pts {
			pts[i] = map[string]int{
				"x": 120 + r.Intn(560),
				"y": 100 + r.Intn(280),
			}
		}
		data, _ := json.Marshal(map[string]any{"k": 3, "points": pts})
		return data

	default:
		return []byte(m.DefaultInput)
	}
}

func (p *Player) renderLessonShell() {
	meta := p.lesson.Meta()

	p.doc.Call("getElementById", "viz-title").Set("innerText", meta.Name)
	p.doc.Call("getElementById", "viz-summary").Set("textContent", meta.Documentation.Summary)
	p.doc.Call("getElementById", "svg-stage").Call("setAttribute", "aria-label", meta.Name+" execution visualization")
	controls := p.doc.Call("querySelectorAll", "#workspace button, #workspace input, #workspace select")
	for i := 0; i < controls.Length(); i++ {
		controls.Index(i).Set("disabled", len(p.steps) == 0)
	}
	p.doc.Call("getElementById", "viz-category").Set("innerText", string(meta.Category))
	p.doc.Call("getElementById", "viz-complexity").Set("innerText", fmt.Sprintf(
		"Time: %s\nSpace: %s", meta.Documentation.TimeComplexity, meta.Documentation.SpaceComplexity,
	))

	// Documentation tab
	p.doc.Call("getElementById", "doc-summary").Set("innerText", meta.Documentation.Summary)
	p.doc.Call("getElementById", "doc-how").Set("innerText", meta.Documentation.HowItWorks)
	p.doc.Call("getElementById", "doc-complexity-detail").Set("innerText", fmt.Sprintf(
		"Time Complexity: %s\nSpace Complexity: %s", meta.Documentation.TimeComplexity, meta.Documentation.SpaceComplexity,
	))

	var conceptsHTML strings.Builder
	for _, c := range meta.Documentation.KeyConcepts {
		conceptsHTML.WriteString(fmt.Sprintf(`<span class="concept-badge">%s</span>`, c))
	}
	p.doc.Call("getElementById", "doc-concepts").Set("innerHTML", conceptsHTML.String())

	// Code container with syntax highlighting
	var codeHTML strings.Builder
	for _, line := range p.lesson.Source() {
		codeHTML.WriteString(fmt.Sprintf(
			`<div class="code-line" id="line-%s"><span class="line-num">%d</span><span class="line-text">%s</span></div>`,
			line.ID, line.Number, renderers.HighlightGoCode(line.Code),
		))
	}
	p.doc.Call("getElementById", "code-container").Set("innerHTML", codeHTML.String())

	// Scrubber max
	scrubber := p.doc.Call("getElementById", "scrubber")
	maxSteps := len(p.steps) - 1
	if maxSteps < 0 {
		maxSteps = 0
	}
	scrubber.Set("max", maxSteps)
	scrubber.Set("value", 0)

	// Stage legend
	p.renderLegend(meta.Category)
}

type legendEntry struct {
	label string
	color string
}

func (p *Player) renderLegend(cat catalog.Category) {
	legendEl := p.doc.Call("getElementById", "stage-legend")
	if legendEl.IsNull() || legendEl.IsUndefined() {
		return
	}

	var items []legendEntry
	switch cat {
	case catalog.CategorySorting:
		items = []legendEntry{
			{"Comparing", "#fde047"},
			{"Swapped", "#f87171"},
			{"Sorted", "#4ade80"},
			{"Pivot", "#c084fc"},
		}
	case catalog.CategoryGraphs:
		items = []legendEntry{
			{"Visiting", "#38bdf8"},
			{"Visited", "#4ade80"},
			{"Queued", "#fde047"},
			{"Shortest Path", "#c084fc"},
		}
	case catalog.CategoryTrees:
		items = []legendEntry{
			{"Active", "#38bdf8"},
			{"Matched", "#4ade80"},
			{"Inserted", "#c084fc"},
			{"Deleted", "#f87171"},
		}
	case catalog.CategoryDP:
		items = []legendEntry{
			{"Active", "#38bdf8"},
			{"Computed", "#4ade80"},
			{"Source Reference", "#c084fc"},
		}
	case catalog.CategoryAI:
		items = []legendEntry{
			{"Active", "#38bdf8"},
			{"Best / Optimal", "#4ade80"},
			{"Candidate", "#fde047"},
			{"Loss / Penalty", "#f87171"},
		}
	case catalog.CategoryBacktracking:
		items = []legendEntry{
			{"Trying", "#38bdf8"},
			{"Valid Solution", "#4ade80"},
			{"Conflict", "#f87171"},
			{"Backtracked", "#c084fc"},
		}
	case catalog.CategoryRecursion:
		items = []legendEntry{
			{"Active Frame", "#38bdf8"},
			{"Base Case", "#4ade80"},
			{"Call Push", "#fde047"},
			{"Return Unwind", "#c084fc"},
		}
	case catalog.CategorySearching:
		items = []legendEntry{
			{"Scanning", "#38bdf8"},
			{"Matched", "#4ade80"},
			{"Mismatch", "#f87171"},
			{"Eliminated", "#94a3b8"},
		}
	default:
		items = []legendEntry{
			{"Active", "#38bdf8"},
			{"Complete", "#4ade80"},
		}
	}

	var sb strings.Builder
	for _, it := range items {
		sb.WriteString(fmt.Sprintf(
			`<div class="legend-chip"><span class="legend-dot" style="background-color: %s;"></span><span class="legend-text">%s</span></div>`,
			it.color, it.label,
		))
	}
	legendEl.Set("innerHTML", sb.String())
}

func (p *Player) RenderStep(idx int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.steps) == 0 {
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(p.steps) {
		idx = len(p.steps) - 1
	}
	p.stepIdx = idx
	step := p.steps[idx]

	// 1. Render Visualizer SVG
	rendererType := p.lesson.Meta().Renderer
	var svgContent string
	switch rendererType {
	case "graph":
		svgContent = renderers.RenderGraphSVG(step, 800, 450)
	case "tree":
		svgContent = renderers.RenderTreeSVG(step, 800, 450)
	case "table":
		svgContent = renderers.RenderTableSVG(step, 800, 450)
	case "hanoi":
		svgContent = renderers.RenderHanoiSVG(step, 800, 450)
	default:
		svgContent = renderers.RenderArraySVG(step, 800, 450)
	}
	p.doc.Call("getElementById", "svg-stage").Set("innerHTML", svgContent)

	// 2. Update line highlight & auto-scroll
	allLines := p.doc.Call("querySelectorAll", ".code-line.active")
	for i := 0; i < allLines.Length(); i++ {
		allLines.Index(i).Get("classList").Call("remove", "active")
	}
	if step.LineID != "" {
		activeEl := p.doc.Call("getElementById", "line-"+step.LineID)
		if !activeEl.IsNull() && !activeEl.IsUndefined() {
			activeEl.Get("classList").Call("add", "active")
			codeBox := p.doc.Call("getElementById", "code-container")
			if !codeBox.IsNull() && !codeBox.IsUndefined() && !activeEl.IsNull() && !activeEl.IsUndefined() {
				offsetTop := activeEl.Get("offsetTop").Float()
				boxHeight := codeBox.Get("clientHeight").Float()
				codeBox.Set("scrollTop", offsetTop-boxHeight/2.0)
			}
		}
	}

	// 3. Update scrubber & step counter
	scrubber := p.doc.Call("getElementById", "scrubber")
	scrubber.Set("value", idx)
	scrubber.Call("setAttribute", "aria-valuetext", fmt.Sprintf("Step %d of %d", idx+1, len(p.steps)))
	p.doc.Call("getElementById", "step-counter").Set("innerText", fmt.Sprintf("Step %d of %d", idx+1, len(p.steps)))

	// 4. Update Action Pill
	actionEl := p.doc.Call("getElementById", "viz-action")
	if !actionEl.IsNull() && !actionEl.IsUndefined() {
		actionEl.Set("innerText", strings.ToUpper(string(step.Action)))
		actionEl.Set("className", fmt.Sprintf("action-pill action-%s", step.Action))
	}

	// 5. Update Live Variables Pill Deck
	varsEl := p.doc.Call("getElementById", "step-variables")
	if !varsEl.IsNull() && !varsEl.IsUndefined() {
		var varsHTML strings.Builder
		for k, v := range step.Variables {
			varsHTML.WriteString(fmt.Sprintf(`<span class="var-pill"><span class="var-name">%s</span><span class="var-val">%s</span></span>`, html.EscapeString(k), html.EscapeString(fmt.Sprint(v))))
		}
		varsEl.Set("innerHTML", varsHTML.String())
	}

	// 6. Update Explanation
	p.doc.Call("getElementById", "step-explanation").Set("innerText", step.Explanation)

	// 7. Update Metrics
	p.doc.Call("getElementById", "metric-comps").Set("innerText", fmt.Sprintf("%d", step.Metrics.Comparisons))
	p.doc.Call("getElementById", "metric-swaps").Set("innerText", fmt.Sprintf("%d", step.Metrics.Swaps))
	p.doc.Call("getElementById", "metric-ops").Set("innerText", fmt.Sprintf("%d", step.Metrics.Operations))
}

func (p *Player) Next() {
	if p.stepIdx < len(p.steps)-1 {
		p.RenderStep(p.stepIdx + 1)
	} else {
		p.Pause()
	}
}

func (p *Player) Prev() {
	if p.stepIdx > 0 {
		p.RenderStep(p.stepIdx - 1)
	}
}

func (p *Player) Seek(idx int) {
	p.RenderStep(idx)
}

func (p *Player) Reset() {
	p.Pause()
	p.RenderStep(0)
}

func (p *Player) Play() {
	if p.playing {
		return
	}
	p.mu.Lock()
	if len(p.steps) > 0 && p.stepIdx >= len(p.steps)-1 {
		p.mu.Unlock()
		p.RenderStep(0)
	} else {
		p.mu.Unlock()
	}
	p.playing = true
	p.updatePlayButtonUI(true)
	p.stopChannel = make(chan struct{})
	stop := p.stopChannel

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				p.mu.Lock()
				currentIdx := p.stepIdx
				maxIdx := len(p.steps) - 1
				sp := p.speed
				p.mu.Unlock()

				if currentIdx >= maxIdx {
					p.Pause()
					return
				}

				interval := time.Duration(float64(500*time.Millisecond) / sp)
				select {
				case <-stop:
					return
				case <-time.After(interval):
				}

				var cb js.Func
				cb = js.FuncOf(func(this js.Value, args []js.Value) any {
					select {
					case <-stop:
					default:
						p.Next()
					}
					cb.Release()
					return nil
				})
				js.Global().Call("requestAnimationFrame", cb)
			}
		}
	}()
}

func (p *Player) Pause() {
	if !p.playing {
		return
	}
	p.playing = false
	p.updatePlayButtonUI(false)
	if p.stopChannel != nil {
		close(p.stopChannel)
		p.stopChannel = nil
	}
}

func (p *Player) updatePlayButtonUI(playing bool) {
	labelEl := p.doc.Call("getElementById", "btn-play-label")
	iconEl := p.doc.Call("getElementById", "btn-play-icon")
	btn := p.doc.Call("getElementById", "btn-play")
	if !btn.IsNull() && !btn.IsUndefined() {
		if playing {
			btn.Call("setAttribute", "aria-label", "Pause playback (Space)")
			btn.Set("title", "Pause playback (Space)")
		} else {
			btn.Call("setAttribute", "aria-label", "Play playback (Space)")
			btn.Set("title", "Play playback (Space)")
		}
	}
	if !labelEl.IsNull() && !labelEl.IsUndefined() {
		if playing {
			labelEl.Set("innerText", "Pause")
		} else {
			labelEl.Set("innerText", "Play")
		}
	} else if !btn.IsNull() && !btn.IsUndefined() {
		if playing {
			btn.Set("innerText", "Pause")
		} else {
			btn.Set("innerText", "Play")
		}
	}
	if !iconEl.IsNull() && !iconEl.IsUndefined() {
		if playing {
			iconEl.Set("innerHTML", `<svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="4" width="4" height="16" rx="1"></rect><rect x="14" y="4" width="4" height="16" rx="1"></rect></svg>`)
		} else {
			iconEl.Set("innerHTML", `<svg class="play-icon" width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>`)
		}
	}
}

func (p *Player) TogglePlay() {
	if p.playing {
		p.Pause()
	} else {
		p.Play()
	}
}

func (p *Player) SetSpeed(speed float64) {
	p.mu.Lock()
	p.speed = speed
	p.mu.Unlock()
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
