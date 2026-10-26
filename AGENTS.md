# Agent guide: Gorithm Living Lab

## Source of truth and working scope

Gorithm is an educational algorithm library and interactive execution viewer. Preserve discoverability, accessible browser behavior, and alignment between algorithm state and displayed Go source when changing its presentation.

Read current source before editing. `docs/superpowers/specs/2026-10-26-living-lab-design.md` and `docs/superpowers/plans/2026-10-26-living-lab.md` are historical design/implementation records, not proof of current behavior or instructions to recreate superseded code. README architecture labels can also lag source (for example, the current Docker runtime is Alpine, not distroless).

Check `git status --short` before work. Other agents may own changes in the same worktree; read unfamiliar diffs and coordinate overlapping edits rather than reverting them. During the current redesign, controls/library behavior and `web/static/themes.css` have separate editors. Prefer the smallest change in the existing owner file; avoid new managers, generic component systems, configuration layers, or dependencies without a concrete need.

## Architecture and responsibility boundaries

| Path | Responsibility |
| --- | --- |
| `go.mod` | Module `github.com/Mersad-Moghaddam/gorithm`, Go 1.24 baseline; currently no external Go module requirements. |
| `cmd/gorithm/main.go` | Native HTTP process, `PORT` (default 8080), timeouts, signal-driven graceful shutdown. |
| `internal/server/server.go` | Embedded shell/static serving, `/healthz`, security headers, WASM MIME type. No algorithm execution API is needed for playback. |
| `web/embed.go` | Embeds `static/*` into the native server at compile time. Rebuild server after rebuilding assets. |
| `internal/catalog/catalog.go` | `Lesson` (`Meta`, `Source`, `Build`), metadata/documentation, registration order, lookup, and source-line ID validation. |
| `internal/catalog/search.go` | Pure metadata matching: optional exact category, case-insensitive whitespace-separated query words, every word required. Search covers ID, name, category, summary, and key concepts. Keep this logic testable outside WASM. |
| `internal/algorithms/*` | Lesson implementations, JSON input handling, source listings, trace generation, and algorithm tests. |
| `internal/trace/types.go`, `internal/trace/trace.go` | `Step`, `SourceLine`, actions, metrics, and trace builder. Algorithms supply state snapshots. |
| `web/wasm/main.go` | `js && wasm` entrypoint, blank imports registering lessons, player/control/library wiring, preset inputs. |
| `web/wasm/library.go` | Catalog cards and category filters, result/empty state, search state, hash navigation, route focus, skip link, search shortcut, decorative category SVG. |
| `web/wasm/controls.go` | Playback bindings, speed/range controls, accessible tabs, native dataset dialog, validation/presets, global playback shortcuts. |
| `web/wasm/player.go` | Current lesson/trace/index, SVG renderer dispatch, source highlight, metrics/variables/explanation, shuffle, playback timing and cancellation. |
| `web/renderers/{array,graph,tree,table,hanoi}.go` | Pure Go SVG fragment generation from trace state; browser-independent tests. |
| `web/renderers/syntax.go` | Escaped Go-source HTML with syntax classes. |
| `web/static/index.html` | Semantic shell and stable element IDs, startup/error UI, minimal WASM bootstrap, native theme popover/radios and persistence. |
| `web/static/styles.css` | Base layout, components, source/state styling, responsive and accessibility rules. |
| `web/static/themes.css` | Loaded after base CSS: four theme palettes plus mobile-first, iOS-inspired surface/control overrides and glass fallback. Read both stylesheets before changing specificity. |
| `scripts/build_wasm.sh`, `Makefile` | Compile client, copy matching Go runtime shim, then build embedded native server. |
| `scripts/check_ui.mjs`, `scripts/audit_ui.mjs` | External-tool browser regression and accessibility checks; not application dependencies. |

Application uses Go/WASM, `syscall/js`, HTML, CSS, and SVG. No frontend framework, npm build pipeline, or third-party JavaScript runtime library is needed. Small inline JavaScript handles bootstrap and theme selection; `wasm_exec.js` is Go's required runtime shim. Google Fonts links remain in the shell, with CSS fallbacks; “no runtime dependencies” does not mean zero external font requests or no JavaScript.

## Trace and playback contracts

- For a fixed input, preserve deterministic trace generation and meaningful `Step.LineID` references to `Lesson.Source()`. A source line can execute repeatedly; do not interpret frame/source alignment as one unique frame per source line.
- Keep SVG state, highlighted source line, scrubber/index, explanation, action, variables, and metrics tied to the same selected step in `RenderStep`. Seeking and reverse stepping use stored snapshots, not a rerun with new random input.
- `trace.Builder.AddStep` copies the variables map shallowly and stores state directly. It does not deep-copy nested slices/maps. Algorithms must clone mutable state before recording it; later mutation must not rewrite earlier frames.
- Shuffle may choose random input; normal playback must not introduce randomness. Keep IDs and output ordering stable where presentation depends on them.
- Loading, resetting, scrubbing, leaving a lesson, and opening custom input must stop old playback where appropriate. Playback begins only by explicit user action.
- `Play` captures its own stop channel. Both the timer wait and queued `requestAnimationFrame` callback check cancellation. Preserve this lifecycle: a paused or replaced run must not advance a later lesson. Release one-shot `js.Func` callbacks after use. Keep persistent event handlers alive for their intended lifetime.
- Verify rapid Play/Pause/Play, pause before the timer fires, navigation while playing, custom data while playing, final-step behavior, and seek/reset after cancellation when changing scheduling. A native `go test ./...` alone cannot exercise WASM event behavior.

## Native interaction, keyboard, focus, and history

- Keep cards as real anchors with `#algorithm=<escaped ID>` links and controls as buttons, labeled inputs, range, and select. Preserve native activation, modifier-key navigation, copy/open-link behavior, and range keyboard behavior; avoid broad event interception.
- Library search and category selection remain intact when entering a lesson and returning. Back/Forward and direct lesson links work. Unknown or malformed lesson hashes return to the library with a visible notice rather than a panic or broken workspace.
- Lesson navigation focuses `#viz-title`; returning focuses the prior visible card or search fallback. `#discovery` returns to library discovery. Skip-to-content focuses main without replacing the current lesson route. Navigation closes an open dataset dialog and cancels playback.
- `/` focuses search in library context. Space toggles playback; Left/Right step; R restarts in workspace context. Do not steal keys from editable content, inputs, selects, buttons, links, dialogs, tab navigation, or modified browser shortcuts. Check existing guards before adding shortcuts.
- Tabs retain `role=tablist/tab/tabpanel`, correct `aria-selected`/`aria-controls`, roving `tabIndex`, Left/Right and Home/End behavior, and hidden inactive panels. Tab arrows must not also advance playback through bubbling.
- Custom data uses native `<dialog>`/`showModal()`, Escape cancellation, focus on input, and focus restoration to its opener. Invalid JSON or lesson input stays in the dialog with a readable error; valid apply resets to the new trace. Preserve the current dataset on validation failure.
- Theme picker uses native popover and radio semantics. Keep theme storage best-effort when `localStorage` is unavailable, allowlist stored values, and restore focus after selection.
- Preserve `[hidden]` behavior, visible focus, labels, meaningful SVG names, polite search status, and runtime error/reload feedback. Disabled startup controls must not imply an engine that is ready before WASM loads.

## Visual direction and acceptance checks

User-requested direction: mobile-first, iOS-inspired glass across **eight themes: Living Lab (`lab`), Midnight Blue (`midnight`), Warm Paper (`paper`), Plum (`plum`), Nordic Frost (`nord`), Retro Amber (`amber`), Nordic Rose (`rose`), and Obsidian Mono (`monochrome`)**. Preserve the algorithm-first information hierarchy, restrained translucent surfaces, rounded controls, generous touch targets, and readable typography. This is a cross-theme requirement, not a dark-theme-only polish pass.

- Use existing theme tokens and CSS for presentation. Keep opaque `var(--panel)` surfaces as the baseline; enhance glass through feature queries for blur/color mixing. Unsupported blur and reduced-transparency preference must retain readable surfaces and borders.
- Test at 320px, common phone width, tablet, and desktop. Inspect library, theme popover, workspace transport, source/variables, and dataset dialog. No document-level horizontal overflow; source regions may scroll locally. Do not hide important controls to make a screenshot fit.
- Keep touch targets around 44 CSS pixels or larger, form text legible on mobile, safe-area padding, and wrapping for long lesson names and values. Verify actual keyboard focus rings and dialog/popover positioning, not only pointer interaction.
- Check contrast in all four themes: body/muted text, placeholders, selected filters/tabs, buttons, errors, code tokens, active source line, SVG labels, and state legend. Aim for WCAG AA (4.5:1 normal text, 3:1 large text and essential UI boundaries). Glass contrast must hold against the composited background.
- SVG renderers contain dark-canvas colors. Warm Paper deliberately keeps a dark visualization canvas; do not lighten it without checking every renderer's state colors and labels. Color changes must retain explanations/legend cues, not rely on color alone.
- Respect `prefers-reduced-motion`: remove decorative animation/transitions and smooth scrolling. Reduced motion does not remove explicit step/play controls or require autoplay. Verify preference behavior in browser.
- Screenshots and automated axe checks complement manual keyboard, touch-layout, contrast, and fallback review; neither proves full accessibility alone.

## Input and HTML/SVG trust boundaries

Metadata, source IDs, fixed icons, and legend strings originate in repository code. Dataset JSON, URL hashes, storage values, and text derived from datasets are user-controlled. Validate JSON and lesson-specific limits/types before replacing a trace; successful JSON parsing alone is not sufficient.

Prefer `textContent`/`innerText` for plain strings. When building HTML/SVG strings, escape user-derived text with `html.EscapeString` (including attribute quotes), and use URL escaping for route parameters. Generated trusted markup can use `innerHTML`; escaping the whole SVG would destroy rendering.

Do not assume all existing interpolation is already safe: some SVG table/header paths interpolate strings directly. Player variable pills escape both names and values. Audit the actual data path when changing them or accepting new string input. `escapeHTML` in player only handles `&`, `<`, and `>`; it is not an attribute sanitizer. Do not expand the trusted-content boundary to include uploaded lessons or arbitrary user markup without explicit validation/escaping work.

## Build and verification

Run from repository root with Go 1.24 or later. For product changes, required baseline:

```sh
go test ./...
make build
```

CI additionally runs `go test -v -race ./...`, `./scripts/build_wasm.sh`, and native `go build -v ./cmd/gorithm`. Run race checks for shared-state/trace changes or CI parity. Format changed Go files with `gofmt`. For documentation-only edits, check source claims and `git diff --check`; do not claim runtime verification from a prose review.

`make build` rebuilds WASM before native embedding; `make run` then starts `./bin/gorithm`. The server embeds files at build time, so restarting an old binary does not pick up CSS/WASM edits. `web/static/app.wasm` is generated. Copy `wasm_exec.js` from the **same Go toolchain** used for WASM (`lib/wasm`, or older `misc/wasm` path); do not hand-edit that runtime. Both the build script and Dockerfile refresh the shim before embedding.

### Local Docker-backed Go wrapper

On this workstation, `/home/mersad/.local/bin/go` invokes Docker and does not forward `GOOS`/`GOARCH`. Therefore host-side `GOOS=js GOARCH=wasm go build ...` may still target native Linux; `go env GOROOT` can also report a container-only path. Inspect `command -v go` and the wrapper when environment behavior is unexpected.

Portable explicit-container build, with target environment set inside the container and matching shim copied there:

```sh
docker run --rm --user "$(id -u):$(id -g)" \
  --mount "type=bind,src=$PWD,dst=/src" --workdir /src \
  -e HOME=/tmp -e GOCACHE=/tmp/go-build \
  golang:1.24 sh -eu -c '
    go test ./...
    GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/static/app.wasm ./web/wasm
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/wasm_exec.js
    mkdir -p bin
    go build -o bin/gorithm ./cmd/gorithm
  '
```

This uses Docker daemon access and writes build outputs to the mounted worktree. A changed toolchain may change the tracked shim; inspect that diff rather than silently restoring a mismatched version. Verify a real browser loads the resulting module.

### Browser checks

Start the freshly built server (default port 8080). Playwright, Chromium, and axe are external developer tools, not application dependencies. Use an existing installation or install outside this repository; keep npm manifests and `node_modules` out of the application solely for these checks.

```sh
NODE_PATH=/path/to/external/node_modules APP_URL=http://localhost:8080 \
  node scripts/check_ui.mjs
NODE_PATH=/path/to/external/node_modules APP_URL=http://localhost:8080 \
  node scripts/audit_ui.mjs
```

`check_ui.mjs` needs `playwright`, supports `CHROMIUM` (default `/usr/bin/chromium`), and checks discovery, themes/storage, history, cancellation, dialog validation, tabs, overflow, and failed WASM loading. Its lesson-count assertion currently expects 40; update deliberately when changing the catalog. It saves screenshots under `/tmp/opencode`.

`audit_ui.mjs` also needs `@axe-core/playwright`; currently its Chromium path is hardcoded `/usr/bin/chromium` and its default app URL differs, so set `APP_URL` explicitly. It checks four-theme accessibility, library widths 320/768/1440, workspace at 320px, and opens catalog lessons. Read scripts for exact coverage; a lesson opening successfully does not verify its full trace or all data inputs. Supplement with the interaction and visual checks above when touching those areas.

## Graph evidence and honest completion

Use codebase-memory graph tools first for structural discovery when available: `list_projects`/`index_status`, `search_graph`, relevant `trace_path`, then exact source. Confirm the project root matches this worktree; a graph rooted at `/home/mersad/Projects/gorithm` may describe different revisions even when its generation is recent.

Check `check_index_coverage` for cited paths and bounded scopes behind negative/exhaustive claims; follow pagination. Coverage is best-effort, not proof of completeness. Static HTML/CSS and untracked work may be excluded or absent. Read current worktree files and use literal/config searches when coverage, freshness, or repository identity is insufficient. Never infer “unused,” “safe,” or “all checked” solely from an empty graph result.

Record commands actually run, observed outcomes, and remaining limits. Missing browser tools, stale assets, partial graph coverage, concurrent edits, or checks not run must remain explicit. A successful build is not proof of browser behavior. A local commit is not a push: claim a push/deployment only after actual command output confirms it, and identify the branch/remote or deployment result. Do not commit, push, or edit unrelated files without task authorization.
