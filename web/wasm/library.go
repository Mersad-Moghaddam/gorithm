//go:build js && wasm

package main

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"syscall/js"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func on(el js.Value, event string, fn func(js.Value)) {
	el.Call("addEventListener", event, js.FuncOf(func(_ js.Value, args []js.Value) any { fn(args[0]); return nil }))
}

func initLibrary(doc js.Value, player *Player) {
	get := func(id string) js.Value { return doc.Call("getElementById", id) }
	metas := catalog.List()
	var cards, filters strings.Builder
	categories := []catalog.Category{}
	counts := map[catalog.Category]int{}
	for _, m := range metas {
		if counts[m.Category] == 0 {
			categories = append(categories, m.Category)
		}
		counts[m.Category]++
	}
	fmt.Fprintf(&filters, `<button class="filter-button" data-category="" aria-pressed="true">All algorithms <span>%d</span></button>`, len(metas))
	for _, cat := range categories {
		fmt.Fprintf(&filters, `<button class="filter-button" data-category="%s" aria-pressed="false">%s <span>%d</span></button>`, html.EscapeString(string(cat)), html.EscapeString(string(cat)), counts[cat])
	}
	for _, m := range metas {
		fmt.Fprintf(&cards, `<a class="algorithm-card" id="card-%s" href="#algorithm=%s"><div class="card-illustration" aria-hidden="true">%s</div><p class="card-category">%s</p><h3>%s</h3><p class="card-summary">%s</p><div class="card-bottom"><span class="card-complexity">%s</span><span class="card-open">Explore algorithm ↗</span></div></a>`, html.EscapeString(m.ID), url.QueryEscape(m.ID), categoryArt(m.Renderer), html.EscapeString(string(m.Category)), html.EscapeString(m.Name), html.EscapeString(m.Documentation.Summary), html.EscapeString(m.Documentation.TimeComplexity))
	}
	get("algorithm-grid").Set("innerHTML", cards.String())
	get("category-filters").Set("innerHTML", filters.String())
	get("catalog-total").Set("textContent", fmt.Sprintf("%d algorithms. Endless lightbulb moments.", len(metas)))
	search := get("algo-search")
	search.Set("disabled", false)
	var category catalog.Category
	update := func() {
		count := 0
		for _, m := range metas {
			match := catalog.Matches(m, search.Get("value").String(), category)
			get("card-"+m.ID).Set("hidden", !match)
			if match {
				count++
			}
		}
		get("result-count").Set("textContent", fmt.Sprintf("%d algorithms to explore", count))
		get("empty-results").Set("hidden", count != 0)
		get("clear-filters").Set("hidden", category == "" && search.Get("value").String() == "")
		buttons := get("category-filters").Call("querySelectorAll", "button")
		for i := 0; i < buttons.Length(); i++ {
			b := buttons.Index(i)
			b.Call("setAttribute", "aria-pressed", fmt.Sprint(b.Get("dataset").Get("category").String() == string(category)))
		}
	}
	on(search, "input", func(js.Value) { update() })
	on(get("category-filters"), "click", func(e js.Value) {
		b := e.Get("target").Call("closest", "button")
		if !b.IsNull() {
			category = catalog.Category(b.Get("dataset").Get("category").String())
			update()
		}
	})
	for _, id := range []string{"clear-filters", "reset-search"} {
		on(get(id), "click", func(js.Value) { category = ""; search.Set("value", ""); update(); search.Call("focus") })
	}
	lastID := ""
	route := func() {
		player.Pause()
		if get("custom-modal").Get("open").Bool() {
			get("custom-modal").Call("close")
		}
		hash := strings.TrimPrefix(js.Global().Get("location").Get("hash").String(), "#")
		id, err := url.QueryUnescape(strings.TrimPrefix(hash, "algorithm="))
		lesson, ok := catalog.Get(id)
		valid := strings.HasPrefix(hash, "algorithm=") && err == nil && ok
		get("workspace").Set("hidden", !valid)
		get("library").Set("hidden", valid)
		get("navigation-notice").Set("hidden", true)
		if valid {
			lastID = id
			player.LoadLesson(lesson)
			get("viz-title").Call("focus")
			js.Global().Call("scrollTo", 0, 0)
		} else {
			if hash != "" && hash != "discovery" && hash != "main" {
				get("navigation-notice").Set("textContent", "That algorithm link was not found. Explore the library below.")
				get("navigation-notice").Set("hidden", false)
				js.Global().Get("history").Call("replaceState", js.Null(), "", js.Global().Get("location").Get("pathname").String())
			}
			if lastID != "" {
				card := get("card-" + lastID)
				if !card.Get("hidden").Bool() {
					card.Call("focus")
				} else {
					search.Call("focus")
				}
			}
			if hash == "discovery" {
				search.Call("focus")
				get("discovery").Call("scrollIntoView")
			}
		}
	}
	on(doc.Call("querySelector", ".skip-link"), "click", func(e js.Value) {
		e.Call("preventDefault")
		get("main").Call("focus")
		get("main").Call("scrollIntoView")
	})
	on(js.Global(), "hashchange", func(js.Value) { route() })
	on(doc, "keydown", func(e js.Value) {
		target := e.Get("target")
		tag := target.Get("tagName").String()
		if e.Get("key").String() == "/" && !e.Get("ctrlKey").Bool() && !e.Get("metaKey").Bool() && tag != "INPUT" && tag != "TEXTAREA" && tag != "SELECT" && !target.Get("isContentEditable").Bool() && !get("library").Get("hidden").Bool() {
			e.Call("preventDefault")
			search.Call("focus")
		}
	})
	update()
	route()
}

func categoryArt(renderer string) string {
	content := `<rect x="5" y="40" width="16" height="30" rx="3"/><rect x="30" y="12" width="16" height="58" rx="3" opacity=".45"/><rect x="55" y="28" width="16" height="42" rx="3"/><rect x="80" y="2" width="16" height="68" rx="3" opacity=".65"/><rect x="105" y="22" width="16" height="48" rx="3"/>`
	switch renderer {
	case "graph", "tree":
		content = `<g fill="none" stroke="currentColor" stroke-width="1.5"><path d="M16 56L58 14L109 49L58 65L16 56M58 14V65"/><circle cx="16" cy="56" r="9"/><circle cx="58" cy="14" r="11"/><circle cx="109" cy="49" r="9"/><circle cx="58" cy="65" r="7"/></g>`
	case "table":
		content = `<g fill="none" stroke="currentColor" opacity=".7"><rect x="10" y="5" width="105" height="63" rx="5"/><path d="M10 26H115M10 47H115M45 5V68M80 5V68"/></g><rect x="45" y="26" width="35" height="21" opacity=".8"/>`
	case "hanoi":
		content = `<path d="M22 6V65M65 6V65M108 6V65" stroke="currentColor" opacity=".4"/><rect x="1" y="53" width="42" height="12" rx="5"/><rect x="6" y="37" width="32" height="12" rx="5" opacity=".7"/><rect x="54" y="53" width="22" height="12" rx="5" opacity=".5"/>`
	}
	return `<svg viewBox="0 0 130 78" fill="currentColor">` + content + `</svg>`
}
