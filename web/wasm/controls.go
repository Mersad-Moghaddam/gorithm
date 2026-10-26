//go:build js && wasm

package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"syscall/js"
)

func bindLabControls(doc js.Value, p *Player) {
	get := func(id string) js.Value { return doc.Call("getElementById", id) }
	for id, action := range map[string]func(){"btn-play": p.TogglePlay, "btn-prev": p.Prev, "btn-next": p.Next, "btn-reset": p.Reset, "btn-shuffle": p.Shuffle} {
		fn := action
		on(get(id), "click", func(js.Value) { fn() })
	}
	on(get("speed-select"), "change", func(e js.Value) {
		if speed, err := strconv.ParseFloat(e.Get("target").Get("value").String(), 64); err == nil {
			p.SetSpeed(speed)
		}
	})
	on(get("scrubber"), "input", func(e js.Value) {
		p.Pause()
		if idx, err := strconv.Atoi(e.Get("target").Get("value").String()); err == nil {
			p.Seek(idx)
		}
	})
	tabs := []string{"tab-code-btn", "tab-doc-btn"}
	selectTab := func(index int, focus bool) {
		for i, id := range tabs {
			tab := get(id)
			selected := i == index
			tab.Call("setAttribute", "aria-selected", strconv.FormatBool(selected))
			if selected {
				tab.Set("tabIndex", 0)
			} else {
				tab.Set("tabIndex", -1)
			}
			panel := get(strings.TrimSuffix(id, "-btn"))
			panel.Set("hidden", !selected)
			if selected && focus {
				tab.Call("focus")
			}
		}
	}
	for i, id := range tabs {
		idx := i
		on(get(id), "click", func(js.Value) { selectTab(idx, false) })
		on(get(id), "keydown", func(e js.Value) {
			switch e.Get("key").String() {
			case "ArrowLeft", "ArrowRight":
				e.Call("preventDefault")
				selectTab(1-idx, true)
			case "Home":
				e.Call("preventDefault")
				selectTab(0, true)
			case "End":
				e.Call("preventDefault")
				selectTab(1, true)
			}
		})
	}
	modal, input, errBox := get("custom-modal"), get("custom-input-val"), get("modal-error")
	closeModal := func() { modal.Call("close"); get("btn-custom-input").Call("focus") }
	on(get("btn-custom-input"), "click", func(js.Value) {
		if p.lesson == nil {
			return
		}
		p.Pause()
		input.Set("value", p.lesson.Meta().DefaultInput)
		errBox.Set("hidden", true)
		id := p.lesson.Meta().ID
		get("preset-row").Set("hidden", !strings.Contains("|bubble|selection|insertion|merge|quick|heap|countingsort|radixsort|bucketsort|linearsearch|binarysearch|bst|hanoitower|fibonacci|nqueens|coinchange|", "|"+id+"|"))
		modal.Call("showModal")
		input.Call("focus")
	})
	for _, id := range []string{"btn-modal-close", "btn-modal-cancel"} {
		on(get(id), "click", func(js.Value) { closeModal() })
	}
	on(modal, "cancel", func(e js.Value) { e.Call("preventDefault"); closeModal() })
	on(get("btn-modal-apply"), "click", func(js.Value) {
		if p.lesson == nil {
			return
		}
		val := strings.TrimSpace(input.Get("value").String())
		if !json.Valid([]byte(val)) {
			errBox.Set("textContent", "Enter valid JSON, for example the dataset shown when opening this dialog.")
			errBox.Set("hidden", false)
			return
		}
		if val == "" {
			errBox.Set("textContent", "Enter a JSON dataset before applying.")
			errBox.Set("hidden", false)
			return
		}
		if _, err := p.lesson.Build([]byte(val)); err != nil {
			errBox.Set("textContent", err.Error())
			errBox.Set("hidden", false)
			return
		}
		p.LoadLessonWithInput(p.lesson, []byte(val))
		closeModal()
	})
	chips := doc.Call("querySelectorAll", ".preset-chip")
	for i := 0; i < chips.Length(); i++ {
		on(chips.Index(i), "click", func(e js.Value) {
			if p.lesson != nil {
				input.Set("value", getPresetInput(e.Get("currentTarget").Get("dataset").Get("preset").String(), p.lesson.Meta().ID))
			}
		})
	}
	on(doc, "keydown", func(e js.Value) {
		target := e.Get("target")
		tag := target.Get("tagName").String()
		if get("workspace").Get("hidden").Bool() || modal.Get("open").Bool() || target.Get("isContentEditable").Bool() || tag == "INPUT" || tag == "TEXTAREA" || tag == "SELECT" || tag == "BUTTON" || tag == "A" || e.Get("ctrlKey").Bool() || e.Get("metaKey").Bool() || e.Get("altKey").Bool() {
			return
		}
		switch e.Get("code").String() {
		case "Space":
			e.Call("preventDefault")
			p.TogglePlay()
		case "ArrowRight":
			e.Call("preventDefault")
			p.Next()
		case "ArrowLeft":
			e.Call("preventDefault")
			p.Prev()
		case "KeyR":
			e.Call("preventDefault")
			p.Reset()
		}
	})
}
