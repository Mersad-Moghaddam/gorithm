//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
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
	input.Call("setAttribute", "aria-describedby", "modal-description modal-error")
	clearError := func() {
		errBox.Set("hidden", true)
		input.Call("removeAttribute", "aria-invalid")
	}
	showError := func(message string) {
		errBox.Set("textContent", message)
		errBox.Set("hidden", false)
		input.Call("setAttribute", "aria-invalid", "true")
		input.Call("focus")
	}
	on(input, "input", func(js.Value) { clearError() })
	closeModal := func() { modal.Call("close"); get("btn-custom-input").Call("focus") }
	on(get("btn-custom-input"), "click", func(js.Value) {
		if p.lesson == nil {
			return
		}
		p.Pause()
		defaultInput := p.lesson.Meta().DefaultInput
		if !json.Valid([]byte(defaultInput)) {
			encoded, _ := json.Marshal(defaultInput)
			defaultInput = string(encoded)
		}
		input.Set("value", defaultInput)
		clearError()
		id := p.lesson.Meta().ID
		get("preset-row").Set("hidden", !strings.Contains("|bubble|selection|insertion|merge|quick|heap|countingsort|radixsort|bucketsort|linearsearch|binarysearch|bst|hanoitower|fibonacci|nqueens|coinchange|", "|"+id+"|"))
		modal.Call("showModal")
		input.Call("focus")
	})
	for _, id := range []string{"btn-modal-close", "btn-modal-cancel"} {
		on(get(id), "click", func(js.Value) { closeModal() })
	}
	on(modal, "cancel", func(e js.Value) { e.Call("preventDefault"); closeModal() })
	on(modal, "click", func(e js.Value) {
		if e.Get("target").Equal(modal) {
			rect := modal.Call("getBoundingClientRect")
			clientX := e.Get("clientX").Float()
			clientY := e.Get("clientY").Float()
			if clientX < rect.Get("left").Float() || clientX > rect.Get("right").Float() ||
				clientY < rect.Get("top").Float() || clientY > rect.Get("bottom").Float() {
				closeModal()
			}
		}
	})
	on(input, "keydown", func(e js.Value) {
		if (e.Get("ctrlKey").Bool() || e.Get("metaKey").Bool()) && e.Get("key").String() == "Enter" {
			e.Call("preventDefault")
			get("btn-modal-apply").Call("click")
		}
	})
	on(get("btn-modal-apply"), "click", func(js.Value) {
		if p.lesson == nil {
			return
		}
		val := strings.TrimSpace(input.Get("value").String())
		if val == "" {
			showError("Enter a JSON dataset before applying.")
			return
		}
		if err := validateInputShape(val, p.lesson.Meta().ID, p.lesson.Meta().DefaultInput); err != nil {
			showError(err.Error())
			return
		}
		if _, err := p.lesson.Build([]byte(val)); err != nil {
			showError(err.Error())
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
				clearError()
			}
		})
	}
	on(doc, "keydown", func(e js.Value) {
		target := e.Get("target")
		tag := target.Get("tagName").String()
		if !target.Call("closest", ".canvas-viewport").IsNull() {
			return
		}
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

// validateInputShape catches JSON that permissive lesson decoders silently ignore.
// ponytail: shape validation only; algorithm-specific limits remain in Build.
func validateInputShape(input, id, example string) error {
	var value any
	if err := json.Unmarshal([]byte(input), &value); err != nil {
		return fmt.Errorf("Enter valid JSON, for example the dataset shown when opening this dialog.")
	}
	if array, ok := value.([]any); ok && len(array) == 0 {
		return fmt.Errorf("Enter a non-empty dataset; an empty array would use the default dataset.")
	}
	if !json.Valid([]byte(example)) {
		encoded, _ := json.Marshal(example)
		example = string(encoded)
	}
	examples := []string{example}
	// Preserve alternate formats already supported by lesson decoders.
	switch id {
	case "linearsearch", "binarysearch":
		examples = append(examples, `[1]`, `{"arr":[1],"array":[1],"target":1}`)
	case "bst":
		examples = append(examples, `[1]`)
	case "avl":
		examples = append(examples, `{"values":[1]}`)
	case "hanoitower":
		examples = append(examples, `{"n":1,"disks":1,"count":1}`)
	case "fibonacci", "nqueens":
		examples = append(examples, `1`)
	case "trie":
		examples = append(examples, `["word"]`)
	case "segmenttree", "minimax":
		examples = append(examples, `[1]`)
	case "lcs":
		examples = append(examples, `["a","b"]`, `{"text":"a","pattern":"b"}`)
	case "kmp":
		examples = append(examples, `["a","b"]`)
	case "sudoku":
		examples = append(examples, `[[1]]`)
	case "dfs", "bfs", "bellmanford", "toposort":
		examples = append(examples, `{"start":"A","nodes":[{"id":"A","label":"A","x":1,"y":1}],"edges":[{"from":"A","to":"B","weight":1,"directed":true}]}`)
	}
	for _, raw := range examples {
		var sample any
		if json.Unmarshal([]byte(raw), &sample) == nil && matchesInputShape(value, sample) {
			return nil
		}
	}
	return fmt.Errorf("Dataset has the wrong shape or field types. Use the example shown when opening this dialog.")
}

func matchesInputShape(value, sample any) bool {
	switch want := sample.(type) {
	case []any:
		got, ok := value.([]any)
		if !ok {
			return false
		}
		for _, item := range got {
			if len(want) > 0 && !matchesInputShape(item, want[0]) {
				return false
			}
		}
		return true
	case map[string]any:
		got, ok := value.(map[string]any)
		if !ok {
			return false
		}
		if len(got) == 0 {
			return len(want) == 0
		}
		matched := false
		for key, item := range got {
			if expected, known := want[key]; known {
				if !matchesInputShape(item, expected) {
					return false
				}
				matched = true
			}
		}
		return matched
	case float64:
		_, ok := value.(float64)
		return ok
	case string:
		_, ok := value.(string)
		return ok
	case bool:
		_, ok := value.(bool)
		return ok
	}
	return false
}
