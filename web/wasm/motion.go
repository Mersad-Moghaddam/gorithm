//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"syscall/js"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func motionDuration(step trace.Step, speed float64, base int) int {
	if speed <= 0 || math.IsNaN(speed) || math.IsInf(speed, 0) {
		speed = 1
	}
	return int(math.Min(math.Round(float64(base)/speed), math.Floor(.6*float64(renderers.StepHoldMS(step, speed)))))
}

func (p *Player) initMotion() {
	media := js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)")
	p.motionListener = js.FuncOf(func(js.Value, []js.Value) any {
		if media.Get("matches").Bool() {
			p.cancelMotion()
		}
		return nil
	})
	media.Call("addEventListener", "change", p.motionListener)
	p.visibilityListener = js.FuncOf(func(js.Value, []js.Value) any {
		if p.doc.Get("hidden").Bool() {
			p.Pause()
		}
		return nil
	})
	p.doc.Call("addEventListener", "visibilitychange", p.visibilityListener)
	p.resizeListener = js.FuncOf(func(js.Value, []js.Value) any { p.updateStageOverflow(); return nil })
	js.Global().Call("addEventListener", "resize", p.resizeListener)
	if observer := js.Global().Get("ResizeObserver"); observer.Type() == js.TypeFunction {
		p.resizeObserver = observer.New(p.resizeListener)
		p.resizeObserver.Call("observe", p.elSvgStage.Get("parentElement"))
	}
}

func (p *Player) cancelMotion() {
	for _, animation := range p.animations {
		animation.Call("cancel")
	}
	p.animations = nil
	// Decorative emphasis clones have no effect after interruption.
	nodes := p.elSvgStage.Call("querySelectorAll", "[data-motion-outline]")
	for i := 0; i < nodes.Length(); i++ {
		nodes.Index(i).Call("remove")
	}
}

func finiteAttribute(node js.Value, name string) (float64, bool) {
	v, err := strconv.ParseFloat(node.Call("getAttribute", name).String(), 64)
	return v, err == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}

func entityPosition(node js.Value) ([2]float64, bool) {
	x, okX := finiteAttribute(node, "data-x")
	y, okY := finiteAttribute(node, "data-y")
	return [2]float64{x, y}, okX && okY
}

func (p *Player) stagePositions() map[string][2]float64 {
	positions := make(map[string][2]float64)
	duplicates := make(map[string]bool)
	nodes := p.elSvgStage.Call("querySelectorAll", "[data-entity]")
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		id := node.Call("getAttribute", "data-entity").String()
		if _, exists := positions[id]; exists {
			duplicates[id] = true
		}
		if pos, ok := entityPosition(node); ok && id != "" {
			positions[id] = pos
		}
	}
	for id := range duplicates {
		delete(positions, id)
	}
	return positions
}

func (p *Player) sizeStage() {
	w, h, minWidth := 800.0, 450.0, 0.0
	root := p.elSvgStage.Call("querySelector", "[data-lesson]")
	dedicated := !root.IsNull()
	if dedicated {
		if v, ok := finiteAttribute(root, "data-scene-width"); ok && v >= 1 && v <= 32768 {
			w = v
		}
		if v, ok := finiteAttribute(root, "data-scene-height"); ok && v >= 1 && v <= 32768 {
			h = v
		}
		if v, ok := finiteAttribute(root, "data-min-width"); ok && v >= 0 && v <= w {
			minWidth = v
		}
	}
	p.elSvgStage.Call("setAttribute", "viewBox", fmt.Sprintf("0 0 %g %g", w, h))
	p.elSvgStage.Get("style").Set("minWidth", fmt.Sprintf("%gpx", minWidth))
	p.elSvgStage.Get("classList").Call("toggle", "lesson-scene", dedicated)
	p.updateStageOverflow()
}

func (p *Player) updateStageOverflow() {
	viewport := p.elSvgStage.Get("parentElement")
	overflow := viewport.Get("scrollWidth").Int() > viewport.Get("clientWidth").Int()
	if overflow {
		viewport.Call("setAttribute", "role", "region")
		viewport.Call("setAttribute", "aria-label", "Algorithm diagram; scroll to explore")
		viewport.Call("setAttribute", "tabindex", "0")
	} else {
		for _, name := range []string{"role", "aria-label", "tabindex"} {
			viewport.Call("removeAttribute", name)
		}
	}
}

func (p *Player) animateStage(previous map[string][2]float64, durationMS int) {
	if js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)").Get("matches").Bool() {
		return
	}
	nodes := p.elSvgStage.Call("querySelectorAll", "[data-entity]")
	destinations := p.stagePositions()
	box := p.elSvgStage.Get("viewBox").Get("baseVal")
	height := box.Get("height").Float()
	for i := 0; i < nodes.Length(); i++ {
		node := nodes.Index(i)
		if node.Get("animate").Type() != js.TypeFunction {
			continue
		}
		id := node.Call("getAttribute", "data-entity").String()
		end, valid := destinations[id]
		if !valid {
			continue
		}
		start, exists := previous[id]
		if !exists {
			start, exists = previous[node.Call("getAttribute", "data-from-entity").String()]
		}
		if !exists {
			// Outline new entities without inventing an incoming position.
			node.Get("style").Set("outline", "2px solid #8bd5ff")
			continue
		}
		node.Get("style").Set("transformBox", "view-box")
		node.Get("style").Set("transformOrigin", "0 0")
		if start == end {
			if node.Call("getAttribute", "data-emphasis").String() == "compare" {
				// Only renderer-marked decorative outlines are duplicated, never labels.
				outline := node.Call("querySelector", "[data-compare-outline]")
				if !outline.IsNull() {
					clone := outline.Call("cloneNode", true)
					clone.Call("removeAttribute", "id")
					clone.Call("setAttribute", "data-motion-outline", "")
					clone.Call("setAttribute", "aria-hidden", "true")
					node.Call("appendChild", clone)
					ms := motionDuration(p.steps[p.stepIdx], p.speed, 200)
					p.animations = append(p.animations, clone.Call("animate", js.ValueOf([]any{map[string]any{"opacity": 0}, map[string]any{"opacity": 1}}), js.ValueOf(map[string]any{"duration": ms, "fill": "none"})))
				}
			}
			continue
		}
		points := [][2]float64{start, end}
		switch node.Call("getAttribute", "data-motion").String() {
		case "swap-up", "swap-down":
			y := math.Min(start[1], end[1]) - 64
			if node.Call("getAttribute", "data-motion").String() == "swap-down" {
				y = math.Max(start[1], end[1]) + 24
			}
			points = [][2]float64{start, {(start[0] + end[0]) / 2, math.Max(0, math.Min(height, y))}, end}
		case "lift":
			y, ok := finiteAttribute(node, "data-clear-y")
			if !ok || y < 0 || y > height {
				continue
			}
			points = [][2]float64{start, {start[0], y}, {end[0], y}, end}
		case "path":
			var raw [][]*float64
			if json.Unmarshal([]byte(node.Call("getAttribute", "data-path").String()), &raw) != nil || len(raw) < 2 || len(raw) > 65 {
				continue
			}
			points = nil
			for _, point := range raw {
				if len(point) != 2 || point[0] == nil || point[1] == nil || math.IsNaN(*point[0]) || math.IsNaN(*point[1]) || math.IsInf(*point[0], 0) || math.IsInf(*point[1], 0) {
					points = nil
					break
				}
				points = append(points, [2]float64{*point[0], *point[1]})
			}
			if len(points) < 2 || math.Hypot(points[0][0]-start[0], points[0][1]-start[1]) > .01 || math.Hypot(points[len(points)-1][0]-end[0], points[len(points)-1][1]-end[1]) > .01 {
				continue
			}
		case "", "linear":
		default:
			continue
		}
		frames := make([]any, len(points))
		for j, point := range points {
			frames[j] = map[string]any{"transform": fmt.Sprintf("translate(%gpx, %gpx)", point[0], point[1])}
		}
		p.animations = append(p.animations, node.Call("animate", js.ValueOf(frames), js.ValueOf(map[string]any{"duration": durationMS, "fill": "none", "easing": "ease-in-out"})))
	}
}
