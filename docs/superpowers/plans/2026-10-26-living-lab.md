# Living Lab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the dense initial studio with an immersive, searchable algorithm library and a focused learning workspace.

**Architecture:** Keep the registry and deterministic Go/WASM player. Render library metadata once, filter locally, and use URL fragments for lesson navigation. Replace the static shell and stylesheet; preserve player element IDs where their responsibilities remain unchanged.

**Tech Stack:** Go, WebAssembly, semantic HTML, CSS, existing SVG renderers; no new runtime dependencies.

**Spec:** `docs/superpowers/specs/2026-10-26-living-lab-design.md`

## Global Constraints

- Dark green surfaces, lime accents, bold typography, immersive presentation.
- Search is local and case-insensitive across names, IDs, categories, summaries, and key concepts. All query words must match the searchable text.
- Every registered algorithm must be reachable; no invented difficulty or complexity metadata.
- Preserve all existing playback, input, source, documentation, metrics, and variable features.
- Honor `prefers-reduced-motion`. No automatic playback on discovery or lesson entry.
- No framework migration or external search service is needed.
- Commit and push verified implementation; exclude local browser-companion files.

## Review Focus

- Empty, mixed-case, whitespace-heavy, and multiword queries must combine correctly with category selection (Task 1).
- Unknown fragments, direct entry, back/forward, and filtered return paths must recover without losing discovery context (Task 2).
- Switching lesson or leaving while playback is scheduled must not advance the wrong lesson (Task 3).
- Editable controls, dialog dismissal, and tab navigation must not accidentally trigger playback or lose keyboard focus (Tasks 2–3).
- Narrow screens, reduced motion, and failed WASM loading must retain understandable, usable controls (Tasks 2–4).

## File Responsibilities

- `internal/catalog/search.go` and `search_test.go`: pure metadata matching, runnable in native Go tests.
- `web/wasm/library.go`: registry-driven cards, filters, result state, fragment navigation, focus restoration.
- `web/wasm/main.go`: startup and existing controls; remove obsolete sidebar code and bind library behavior.
- `web/wasm/player.go`: workspace metadata, safe playback lifecycle, step presentation.
- `web/static/index.html`: replacement discovery/workspace shell, dialog and loading/error states.
- `web/static/styles.css`: replacement Living Lab visual system and responsive motion rules.
- `scripts/check_ui.mjs`: browser regression check, only if available browser tooling supports it without adding runtime dependencies.
- `.gitignore`, `README.md`, and design/plan documents: local artifact exclusions, revised usage, delivery record.

### Task 1: Search contract

**Interfaces:** Add `catalog.Matches(meta Meta, query string, category Category) bool`; empty category means All. Consumes existing `Meta` and `Documentation`.

- [ ] Write table-driven `TestMatches` in `internal/catalog/search_test.go`. Fixture: ID `dijkstra`, name `Dijkstra's Algorithm`, category `CategoryGraphs`, summary `Find shortest paths`, concepts `non-negative weights`. Assert empty query matches; `  SHORTEST weights ` matches; `dijkstra` matches; `graphs` matches; `shortest unknown` does not; matching text with `CategorySorting` does not; `CategoryGraphs` does.
- [ ] Run `go test ./internal/catalog`; confirm failure from missing `Matches`.
- [ ] Implement `Matches` using `strings.Fields`, lowercase text, and all-word substring matching. Filter category first.
- [ ] Run `go test ./internal/catalog`; require PASS.

### Task 2: Discovery and navigation redesign

**Interfaces:** Add `initLibrary(doc js.Value, player *Player)` in `web/wasm/library.go`. It owns query/category state, registry cards, hashchange handling, and library/workspace visibility. Use `#algorithm=<URL-encoded registry ID>` for lessons and empty fragment for library. Existing `bindControls` remains responsible for playback and input.

- [ ] Load frontend-design skill before styling. Inspect full existing control code and renderer styling contracts before replacing shell.
- [ ] Add `.superpowers/` and worktree directory, if used, to `.gitignore`; set up isolated execution workspace using required worktree workflow.
- [ ] Replace shell with header, discovery hero (“Make complex click.”), labeled search, registry-populated category filters, result status, grid, resettable empty state, and hidden workspace. Retain IDs consumed by player and controls.
- [ ] Implement `initLibrary`: render escaped metadata and genuine links once; filter using Task 1; derive categories and counts from registry; use delegated events to avoid repeated callback allocation. Preserve DOM and filter state across navigation.
- [ ] Replace automatic first-lesson load and sidebar startup with `initLibrary`. Delete obsolete accordion, drawer, and old name-only search handlers.
- [ ] Implement hash navigation: validate IDs with `catalog.Get`, pause on navigation, load valid lessons without autoplay, focus lesson heading, and restore originating visible card or search on return. Unknown IDs show a library notice and normalize fragment with `replaceState`.
- [ ] Replace stylesheet with ink-green/lime tokens, bold hero, illustrated cards, readable hierarchy, responsive grid and focus states. Category graphics remain decorative; reduced motion disables entrances and hover movement.
- [ ] Verify in browser: empty initial fragment shows library; every registered card exists; search and category intersect; no results offers reset; `/` focuses search only outside editable controls; direct link loads correct lesson; back/forward and back-to-library preserve filters. Check at 390px and 1440px widths.
- [ ] Run `make build` and `go test ./...`; require successful compilation and passing tests.

### Task 3: Focused workspace and reliable interactions

**Interfaces:** Keep `Player.LoadLesson`, `RenderStep`, and existing control IDs. Add `viz-summary` populated from lesson metadata. Use native `<dialog id="custom-modal">` with `showModal()`/`close()` and restore focus to invoking button. Code/Learn tabs expose `role`, `aria-selected`, `aria-controls`, and keyboard selection.

- [ ] Rebuild workspace as canvas plus inspector on desktop, stacked on mobile. Place narration and playback beneath canvas; maintain metrics, variables, legend, source and documentation. Use local code scrolling.
- [ ] Update `renderLessonShell` to populate summary and SVG accessible name. Ensure loading/unavailable states disable controls and meaningful step availability restores them.
- [ ] Update tabs for keyboard ArrowLeft/ArrowRight/Home/End selection and proper focus/state. Replace modal display toggles with native dialog methods, retaining dataset validation and presets; escape error text safely.
- [ ] Inspect keyboard handlers and gate shortcuts when library, dialog, editable element, button, or select owns input; preserve visible controls for every action.
- [ ] Inspect playback cancellation before changing it. Current implementation sleeps before scheduling callbacks and reads mutable stop-channel state. Add cancellation checks tied to the captured playback session before scheduled `Next`, so pausing or navigating cannot execute stale callbacks. Use systematic-debugging and test-driven-development workflows for confirmed failures.
- [ ] Preserve deterministic SVG/source synchronization. Existing renderer replaces SVG innerHTML each step; do not claim persistent-element interpolation. Use restrained entrance/state highlighting and control transitions, with reduced-motion overrides, instead of a generic animation engine.
- [ ] Browser checks: step and timeline update source/narration together; pause/navigation prevents late steps; speed and restart work; modal traps focus, Escape restores it, invalid JSON stays visible without data loss; tabs work by keyboard; reduced motion removes decorative animation.
- [ ] Run `make build` and `go test ./...`; require PASS.

### Task 4: Verify, review, commit and push

**Interfaces:** Existing `make build` compiles WASM and server; `go test ./...` runs native suite. Browser check uses project server and external development tooling only.

- [ ] Improve WASM bootstrap failure handling: catch both instantiation and `go.run` rejection, show readable error and reload action; never report runtime ready before initialization succeeds.
- [ ] If browser tooling is available, leave `scripts/check_ui.mjs` with documented invocation and assertions for Task 2 navigation/filter cases, Task 3 interactions, WASM failure, mobile overflow, and reduced motion. Otherwise retain native search regression test and document browser verification limitation precisely.
- [ ] Inspect rendered library and workspace at desktop/mobile sizes and keyboard-only navigation. Confirm no page-level horizontal overflow and no controls hidden beneath overlays.
- [ ] Update README with search, category filtering, shareable links, keyboard navigation, and verification commands. Mark completed plan steps with actual evidence.
- [ ] Run final `go test ./...`, `make build`, relevant browser checks, and `git diff --check`. Review staged diff for generated files or unrelated user changes.
- [ ] Request independent review using requesting-code-review workflow; resolve material findings and repeat affected checks.
- [ ] Commit approved design, plan, code, and checks with `feat: redesign Gorithm as a discovery-first living lab` (or cohesive task commits). Push current branch to configured upstream; verify push output and clean tracked working tree. Do not force-push.

## Plan Self-Review

All spec sections map to Tasks 1–4. Search matcher has native coverage; browser-dependent navigation, accessibility, motion, and failure cases have explicit checks. Player lifecycle is investigated before changing behavior. Renderer replacement limits are stated rather than hidden behind animation promises. No new runtime package is required.

## Execution result

Tasks 1–4 implemented and verified. Native catalog regression, full Go suite,
WASM/server build, and Chromium regression pass. Independent review found and
resolved discovery-history, skip-link routing, and open-dialog navigation bugs.
Browser checks cover those fixes and pause cancellation. Additional requested
theme picker offers four persisted palettes with native popover and radio controls.
Build ran inside golang:1.27 because host Go wrapper drops target environment
variables. UI transitions preserve deterministic SVG frame rendering rather than
adding persistent-object interpolation. Delivery uses isolated `living-lab` branch.
