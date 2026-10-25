//go:build js && wasm

package main

import (
	"fmt"
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
	p.Pause()

	p.mu.Lock()
	p.lesson = l
	steps, err := l.Build(nil)
	if err != nil {
		fmt.Printf("Error building lesson: %v\n", err)
		p.steps = nil
	} else {
		p.steps = steps
	}
	p.stepIdx = 0
	p.mu.Unlock()

	p.renderLessonShell()
	p.RenderStep(0)
}

func (p *Player) renderLessonShell() {
	meta := p.lesson.Meta()

	p.doc.Call("getElementById", "viz-title").Set("innerText", meta.Name)
	p.doc.Call("getElementById", "viz-category").Set("innerText", string(meta.Category))
	p.doc.Call("getElementById", "viz-complexity").Set("innerText", fmt.Sprintf(
		"Time: %s | Space: %s", meta.Documentation.TimeComplexity, meta.Documentation.SpaceComplexity,
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

	// Code container
	var codeHTML strings.Builder
	for _, line := range p.lesson.Source() {
		codeHTML.WriteString(fmt.Sprintf(
			`<div class="code-line" id="line-%s"><span class="line-num">%d</span><span class="line-text">%s</span></div>`,
			line.ID, line.Number, escapeHTML(line.Code),
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
	allLines := p.doc.Call("querySelectorAll", ".code-line")
	for i := 0; i < allLines.Length(); i++ {
		allLines.Index(i).Get("classList").Call("remove", "active")
	}
	if step.LineID != "" {
		activeEl := p.doc.Call("getElementById", "line-"+step.LineID)
		if !activeEl.IsNull() && !activeEl.IsUndefined() {
			activeEl.Get("classList").Call("add", "active")
			activeEl.Call("scrollIntoView", map[string]any{"block": "nearest", "behavior": "smooth"})
		}
	}

	// 3. Update scrubber & step counter
	p.doc.Call("getElementById", "scrubber").Set("value", idx)
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
			varsHTML.WriteString(fmt.Sprintf(`<span class="var-pill"><span class="var-name">%s</span><span class="var-val">%v</span></span>`, k, v))
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
	p.playing = true
	p.doc.Call("getElementById", "btn-play").Set("innerText", "Pause")
	p.stopChannel = make(chan struct{})

	go func() {
		for {
			select {
			case <-p.stopChannel:
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
				time.Sleep(interval)

				js.Global().Call("requestAnimationFrame", js.FuncOf(func(this js.Value, args []js.Value) any {
					p.Next()
					return nil
				}))
			}
		}
	}()
}

func (p *Player) Pause() {
	if !p.playing {
		return
	}
	p.playing = false
	p.doc.Call("getElementById", "btn-play").Set("innerText", "Play")
	if p.stopChannel != nil {
		close(p.stopChannel)
		p.stopChannel = nil
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
