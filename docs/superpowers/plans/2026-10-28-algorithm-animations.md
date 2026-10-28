# Algorithm Animations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. User explicitly requested one implementation subagent per algorithm.

**Goal:** Teach all 40 algorithms through colorful, smooth, correctly paced, algorithm-specific visualizations matching the approved sorting and gradient-descent studies.

**Architecture:** Keep Go trace snapshots authoritative; enrich individual lessons with typed teaching state and pure Go SVG renderers. A single shared playback owner handles native browser animation of keyed SVG entities, cancellation, timing, and renderer dispatch. Existing application shell and theme design remain intact.

**Tech Stack:** Go 1.24+, Go/WASM, `syscall/js`, HTML, CSS, SVG, native Web Animations API; existing external Playwright/axe tools for browser checks.

**Spec:** `docs/superpowers/specs/2026-10-28-algorithm-animations-design.md` and both HTML references linked there.

## Global Constraints

- All 40 registered lessons receive an individual implementation worker and review.
- No framework, animation dependency, generic scene engine, or separate JavaScript algorithm implementation.
- Default meaningful operation: approximately 1800 ms at 1×. Normal movement 750 ms; compare-only changes 200 ms.
- Compare/commit pairs share a budget: compare 600 ms, commit/result 1200 ms. Single decision without mutation: 1800 ms.
- Motion completes before another step begins. Clamp movement to at most 60% of the current step's hold time.
- No autoplay on load. Preserve existing controls, themes, layout, focus/history, validation, and source/trace correspondence.
- Deep-copy mutable snapshot fields. Never infer identity from value or current index.
- Source and tests win over stale README claims. Do not commit generated binaries, unrelated work, or temporary preview state.

## Review Focus

1. Equal-valued entities must retain distinct identity during swaps, merges, and rotations: foundation browser test and sorting duplicate fixtures.
2. Pause/navigation/input changes during active motion must prevent late callbacks from changing a later lesson: foundation cancellation suite.
3. Signed, zero, constant, and diverging numeric ranges must remain finite and truthful: renderer geometry tests and AI numerical fixtures.
4. Custom labels containing HTML syntax and Unicode must remain escaped and readable: renderer injection suite and trie/KMP/LCS fixtures.
5. Large diagrams at 320px and reduced motion must retain information, readable labels, and keyboard control: final browser matrix.

---

## Current-source evidence and limitations

Planning baseline: commit `49815a34b5cd2f77321f254c54de80d7bbe8e53e`, worktree initially clean. `web/wasm/main.go` imports 40 lessons. `Player.RenderStep` in `web/wasm/player.go:327-414` replaces SVG with `innerHTML`, then updates source, scrubber, variables, explanation, metrics. `Play` uses `500ms / speed`; old CSS transitions on replaced rectangles cannot provide reliable object continuity. `trace.Builder.AddStep` shallow-copies variables but stores state directly.

Graph project `home-mersad-Projects-gorithm`, root `/home/mersad/Projects/gorithm`, initial generation `2026-10-27T07:56:32Z`. Structural query `(Build|Render.*|.*State|Meta|Play)` returned 226 results over two pages, fully read. `RenderStep` direct bidirectional trace returned six callers and seven callees, no pagination. Exact snippet and worktree source were compared. Coverage on all 120 algorithm lesson/trace/algorithm paths and bounded renderer/WASM/trace scopes reported `metadata_match`/`no_recorded_issue`; this is best-effort, not completeness proof. Shared-path recheck after automatic refresh used generation `2026-10-28T05:39:17Z` and again reported no recorded issues for the 11 cited WASM/renderer/trace/browser/static paths. Individual planning task books require current-source reads and record actual constraints. Refresh evidence when execution begins.

## Task books: complete catalog and assignment

| Task book | Individual implementation workers |
|---|---|
| [Sorting](2026-10-28-animation-sorting.md) | bubble, selection, insertion, merge, quick, heap, countingsort, radixsort, bucketsort |
| [Graphs](2026-10-28-animation-graphs.md) | bfs, dfs, dijkstra, bellmanford, floydwarshall, astar, prim, kruskal, toposort, kosaraju |
| [Trees and searching](2026-10-28-animation-trees-search.md) | bst, avl, trie, segmenttree, unionfind, linearsearch, binarysearch, kmp |
| [DP and puzzles](2026-10-28-animation-dp-puzzles.md) | fibonacci, coinchange, knapsack, lcs, nqueens, sudoku, hanoitower |
| [AI](2026-10-28-animation-ai.md) | gradientdescent, kmeans, perceptron, genetic, minimax, qlearning |

Every task-book section is mandatory, not an alternative design. Shared contract below takes precedence for timing, public dispatch, ownership, and SVG animation attributes. Per-lesson types/functions use the exact spelling in their task-book interface blocks. Do not derive Go identifiers mechanically from route IDs.

## File responsibilities and ownership

| Owner | Files |
|---|---|
| Foundation/integrator only | `internal/trace/types.go`, `web/wasm/player.go`, new `web/wasm/motion.go`, `web/wasm/controls.go`, `web/static/styles.css`, `web/static/index.html`, new `web/renderers/lesson.go`, new `web/renderers/lesson_test.go`, new `scripts/check_animations.mjs` |
| Individual algorithm worker | Its existing `internal/algorithms/<id>/{lesson,trace,algorithm_test}.go`; new `web/renderers/<id>_lesson.go` and `<id>_lesson_test.go`; additional test file only when needed for clarity |
| Pure algorithm edits | Existing `algorithm.go` only for a verified trace/algorithm discrepancy or necessary boundary validation; report and test before changing behavior |
| Existing generic renderers | Retain fallback. Shared owner alone may extract small reused helpers into `web/renderers/lesson.go`; no concurrent edits by algorithm workers |

Do not let 40 agents edit `player.go`, theme files, registry switches, or one common state struct. The independent renderer files exist to isolate genuinely distinct diagrams, not to duplicate generic rendering wholesale.

## Shared integration contract (freeze before algorithm workers start)

### Trace timing

Add `HoldMS int \`json:"holdMs,omitempty"\`` to `trace.Step`. Zero means 1800ms; accepted explicit values are 350, 600, 1200, 1800. Existing `Builder.AddStep` signature stays unchanged. Workers may set `steps[i].HoldMS` on the completed local slice, without altering indices or dropping steps. Pair budgets apply only to adjacent semantically linked compare/mutate frames; do not globally give every compare a short hold. Trace tests assert every nonzero hold is allowed and a selected compare/commit pair totals 1800ms.

Shared helper in `web/renderers/lesson.go`: `func StepHoldMS(step trace.Step, speed float64) int`. Invalid/nonfinite/nonpositive speed falls back to 1; divide validated base hold by speed, round to nearest integer, minimum 100ms. Unknown nonzero HoldMS falls back to 1800ms. A speed change settles active motion and restarts the current step's wait at the new duration without advancing the index.

### Dispatch and legend

`func RenderLessonSVG(id string, step trace.Step, width, height int) (string, bool)` in `web/renderers/lesson.go`: explicit switch on registered lesson ID, returning the dedicated renderer output and `true`; unsupported/unconverted lesson returns `"", false` so the existing renderer-type switch remains usable during incremental work. Integrator adds one case only after the worker's tests pass.

`type LessonLegendEntry struct { Label, Color string }` and `func LessonLegend(id string) []LessonLegendEntry` live in the same file. Integrator writes an explicit per-lesson legend based on task-book semantics. `renderLessonShell` prefers this legend and falls back to current category legend. Never show graph shortest-path legend for clustering or binary-search mismatch legend for a sort.

### SVG element contract

Each dedicated renderer returns SVG **fragment**, not outer document or `<svg>`. Root group carries `data-lesson="<id>"`, `data-phase="<phase>"`, and a meaningful escaped `<title>` or existing SVG accessible-name update. Final snapshot is fully visible with animation disabled.

Root group also carries `data-scene-width`, `data-scene-height`, and optional `data-min-width` in SVG/CSS pixels respectively. Foundation validates finite dimensions in [1,32768], defaults missing/invalid dimensions to 800×450 and minimum width to 0, sets outer viewBox and scene minimum CSS width, and resets all three on every render including fallback. Minimum width may not exceed scene width. Foundation owns scroll-region name/focus. Numerical workers compute padded finite axis domains once after building the trace and copy the same explicit domain fields into every typed snapshot; renderers never scan hidden history or rerun training to find bounds.

Movable entities are `<g data-entity="<stable role:id>" data-x="<finite float>" data-y="<finite float>" transform="translate(x y)">`. Children use local coordinates. IDs are unique within one scene; main/output/buffer copies use role-prefixed IDs. Fill is stable by identity. Do not use raw user strings as selector syntax: collect nodes, read attribute values, and match through a Go map.

Optional incoming motion metadata on the destination group:

- `data-motion="linear"` (default), `"swap-up"`, `"swap-down"`, `"lift"`, or `"path"`.
- Swap-up/down: midpoint at average x, respectively 64 units above or 24 units below baseline; cap against scene bounds. Movement is one 750ms-or-shorter animation, not three trace frames.
- Lift (Hanoi): up to clear tallest stack, horizontal travel, then down. Worker supplies finite `data-clear-y` in SVG units.
- Path: `data-path` is JSON array of `[x,y]` samples from previous position to final position. At least two, at most 65, all finite. Curve samples are renderer-computed presentation geometry from stored before/after values, not a second algorithm execution. Use this for gradient descent. Escape serialized attribute text. Invalid metadata means snap to final position, not crash.
- All animated nodes must have corresponding previous identities, except newly created nodes, which appear immediately with an outline. A destination copy/transfer may specify `data-from-entity="<previous role:id>"`; use that source position when no same-key previous entity exists. The source remains visible if still present in destination snapshot. This is geometry aliasing, not identity merging. Do not animate copied buffers as if originals vanished unless algorithm state really moved them.

Replacement strategy: collect previous final positions before replacing SVG; install final snapshot; run native `Element.animate` on matching new groups with keyframe CSS `transform: translate(...)`. Final transform remains in SVG attribute. Cancel animations to reveal final attributes. This deliberately avoids a DOM diff framework. Fixed graph/tree layouts use straight coordinate transitions; edges can redraw immediately unless task requires explicit edge rewiring markers. Label transition steps as transitions to the selected state; do not imply animation intermediates are additional algorithm snapshots.

Set animated groups' `transform-box: view-box; transform-origin: 0 0`; coordinates are SVG user units and must be tested under responsive viewBox scaling. Use animation options `fill: "none"` so completed effects cannot override later snapshots. Before a new render, cancel and clear retained animations; root query is scoped to `#svg-stage`, never global document SVGs. Feature-detect `Element.animate`; unsupported browsers render final state immediately.

All animated entities and aliases use untransformed parents in one scene coordinate system; no nested parent translation/scale. CSS keyframes use explicit lengths, `translate(100px, 50px)`, not SVG attribute syntax. Composition is replacement; no competing CSS transform. Path first/last samples must equal captured source/final destination within 0.01 SVG units; otherwise snap. Test motion under 0.5×, 1×, and 1.5× rendered SVG scale.

For responsive sizing, use existing `.canvas-viewport` as named focusable scroll region only when overflow exists. Shared owner may remove its SVG `max-height` constraint for dedicated scenes and set viewBox from lesson output extents. Do not set outer document min-width. Essential labels at fixed 800-unit width cannot become readable merely by setting font-size to 14 units and scaling down to 280 CSS pixels: either reflow the scene or retain a scrollable 1:1 minimum width. No automatic viewport panning during manual keyboard interaction.

### Player API

Keep public `RenderStep(idx int)` as instant authoritative rendering used by Seek/Prev/reset/load. Add private `renderStep(idx int, animate bool)`; Next invokes it with animation only for adjacent forward progression. Explicit Next/Prev while playing must pause first; autoplay uses private advance path to avoid pausing its own loop.

Add `func (p *Player) cancelMotion()` and `func (p *Player) animateStage(previous map[string][2]float64, durationMS int)` in `web/wasm/motion.go`. Player stores `animations []js.Value`; cancel every retained animation on interruption and clear references. Use no animation completion callbacks: native animation objects need no retained `js.Func`. Respect `matchMedia('(prefers-reduced-motion: reduce)')` per transition; switching preference mid-motion cancels active animations through one lifetime listener. Release any one-shot requestAnimationFrame callbacks exactly once.

Movement duration is `min(round(750/speed), floor(0.6*StepHoldMS(step,speed)))`, with the same speed sanitization as StepHoldMS. Thus a 1200ms commit permits 720ms movement at 1×. Unchanged-position groups with `data-emphasis="compare"` receive one 200ms/speed outline-opacity reveal, likewise capped; keep label and final outline visible at all times (animate a decorative outline duplicate only). No looping pulse. Hold clock begins when selected snapshot is installed and includes motion. Final autoplay snapshot receives its full hold before automatic stop; otherwise Pause would cancel its arrival immediately. Test final moving frame at midpoint and after stop.

Pause must cancel motion even if `playing == false` (manual Next can animate). Keep stop-channel/run cancellation checks before timer and inside queued RAF. Wait for RAF advancement to complete before scheduling the next hold; avoid multiple queued advances on a throttled tab. Returning from hidden tab must not replay accumulated steps. All old run callbacks must fail cancellation check after reset/load/navigation.

## Task F: foundation and one end-to-end pilot

**Files:** shared-owner files above; pilot uses Bubble worker's own files.

**Consumes:** existing `trace.Step`, `Builder.AddStep`, `Player` control bindings.
**Produces:** exact timing/dispatch/motion APIs above and browser harness for all algorithm workers.

- [ ] Add `TestStepHoldMS` with zero/default=1800; 600 at 2×=300; 1200 at 0.5×=2400; invalid speed (0, negative, NaN, infinity)=1×; invalid HoldMS=1800; minimum=100.
- [ ] Add `TestLessonMotionMarkup` against the Bubble pilot fixture, parsing rendered fragment wrapped in `<svg>` with `encoding/xml`: stable unique IDs and finite geometry. Add malicious-label/no-event-attribute assertions to the first string-bearing renderer fixture (trie), not numeric Bubble input.
- [ ] Run `go test ./web/renderers -run 'TestStepHoldMS|TestLessonMotionMarkup'`; record missing-interface failure before adding APIs.
- [ ] Implement timing and dispatch fallback. Add animation engine with contracts above; remove misleading transition declarations only in replaced new renderers, keep unaffected fallback behavior.
- [ ] Extend browser checks in `scripts/check_animations.mjs`, using existing external-tool import style. Test animated Next, pause after 150ms, pause before timer, rapid Play/Pause/Play, seek during swap, reset during swap, lesson change, dialog open/apply, speed change, final step replay, and reduced-motion toggling. After cancellation wait at least 2100ms at 1× and assert index, SVG root phase, active source ID, and metric text remain stable.
- [ ] Assert two equal-valued bars retain different `data-entity` IDs and colors across a swap using a duplicate fixture. Capture positions at start, midpoint, end; require midpoint differs from both endpoints when motion allowed, and no running animation under reduced motion.
- [ ] Wire Bubble pilot via its individual task. Check screenshot side-by-side with approved sorting reference before spreading implementation.
- [ ] Run `go test ./...`, `make build`, start fresh embedded server, run foundation browser tests. Commit only reviewed foundation and pilot files with `feat: add teaching animation playback`.

## Task A: 40 algorithm implementations

Execute one fresh implementation subagent for each route ID in the task-book table, not one worker implementing a whole category. Task-book grouping is documentation organization only. Use up to four workers concurrently when paths do not overlap. Each receives the global prompt below plus its complete task section. Review each before integration. Start gradientdescent after Bubble pilot to validate curve-path support; then remaining workers can proceed.

For each algorithm:

- [ ] Read current source and the assigned task's exact Source IDs; compare against baseline.
- [ ] Add the task's semantic trace/renderer tests, run and observe expected failures.
- [ ] Implement typed teaching state, actual trace snapshots, dedicated renderer and explanations; preserve validated algorithm results.
- [ ] Run algorithm and renderer package tests, format Go, report changes and exact dispatch/legend additions needed.
- [ ] Reviewer checks source alignment, snapshot copying, data-dependent labels, negative/failure cases, and precise task acceptance assertions.
- [ ] Integrator adds dispatch/legend case and the task's scenario to `scripts/check_animations.mjs`; builds and reviews real browser animation.
- [ ] Commit reviewed task with `feat: teach <id> with algorithm-specific animation`.

### Copy-ready global worker prompt

```text
Implement ONLY algorithm task <ID> from <TASK_BOOK_PATH>, including its exact
per-algorithm prompt. Read AGENTS.md, the root animation plan and design spec,
and both approved HTML references before editing. The references specify
visual grammar; Go source specifies actual algorithm semantics.

Use required execution/TDD/design skills. Inspect git status and actual source.
You own internal/algorithms/<ID>/lesson.go, trace.go, algorithm_test.go,
web/renderers/<ID>_lesson.go and <ID>_lesson_test.go only. Do not edit player,
registry/dispatch, generic renderer files, shell, themes, other algorithms,
generated assets, or dependencies. Ask integrator for shared-file changes.

If algorithm.go has a verified discrepancy, report it first. Only coordinator's
explicit narrow ownership grant permits that edit; default worker owns no
algorithm.go changes. Include any task-book resolved validation decisions and
lesson_animation_test.go ownership in your assigned scope.

Follow root plan's timing, entity identity, motion attributes, cancellation,
escaping, immutable snapshots, accessibility, and no-extra-microframes contract.
Use each exact symbol/type declared by your task. Emit renderable final state
without browser animation. Trace state, source highlight and metrics must agree.
Do not recreate algorithms in JS or invent results from narration strings.

First add specified meaningful failing assertions. Implement smallest change
that passes them. Run gofmt, algorithm tests, and renderer tests. Return exact
files, test commands/results, remaining risks, dispatch function, state type,
legend entries, and browser fixture/step expected to prove the distinctive
operation. No commit/push by workers; integrator commits reviewed changes.

Graph evidence is supplied by parent at dispatch time. If MCP unavailable,
use exact source reads and state that limitation. Never claim browser checks
or test passes you did not run. Do not overwrite other workers' changes.
```

The coordinator must replace every placeholder, append the complete individual task section, and supply refreshed graph generation/coverage. Do not send only a short task title.

## Task V: integration, accessibility, and push

**Files:** `scripts/check_animations.mjs`, `scripts/audit_ui.mjs`, `scripts/check_ui.mjs`, only fixes justified by observed failures.

- [ ] Enumerate catalog routes from live library and compare exact set with the 40 plan IDs. Every route must dispatch dedicated scene, not silently fall back.
- [ ] For every lesson, run default and specified small fixture; capture initial, distinctive operation, and final/failed outcome. Validate narration operands, root `data-phase`, source line, and terminal result. A lesson opening is not acceptance.
- [ ] Test 320/390/768/1440px with keyboard-only transport and custom dialog; check document has no horizontal overflow, scene may scroll locally, essential labels meet rendered-size requirement.
- [ ] Test all eight themes (`lab`, `midnight`, `paper`, `plum`, `nord`, `amber`, `rose`, `monochrome`), reduced motion, reduced transparency. Existing audit already covers all eight themes; extend its workspace coverage beyond the current merge-only check. Record contrast review for SVG labels/outlines on actual composited surfaces.
- [ ] Run `go test ./...`, `go test -race ./...`, `make build`. On workstation Docker-backed Go wrapper, use explicit Go container with target environment set inside it as documented in AGENTS.md; match `wasm_exec.js` toolchain. Never call native-only compile a WASM build.
- [ ] Start newly built server and run `NODE_PATH=<external modules> APP_URL=http://localhost:8080 node scripts/check_ui.mjs`, `scripts/check_animations.mjs`, and `scripts/audit_ui.mjs`. Use real installed paths. Record commands and results; missing browser tools block browser completion claims.
- [ ] Run `git diff --check`, inspect all staged paths, full review, and ensure no preview server key or temporary state is included. Durable reference HTML files intentionally belong in design docs.
- [ ] Inspect current branch, upstream, and remotes with `git branch --show-current`, `git rev-parse --abbrev-ref --symbolic-full-name '@{u}'`, `git remote -v`. Push reviewed commits to existing intended upstream using normal `git push`, never force-push. If no unambiguous upstream exists, ask user for destination. Confirm command output and report branch/remote/commit.

Push authorization came with original implementation request. Current request is planning: save and review this plan first; do not present product work as implemented or pushed.

## Resolved cross-book integration decisions

These decisions supersede task-book requests for parent to choose shared behavior. Append them to affected worker prompts.

- Motion/viewport protocol is frozen above: `data-entity`, not `data-entity-id`; root dimension attributes define custom canvas extents. Do not introduce alternative spelling or persistent renderer state.
- Playback completion depends on selected index reaching final snapshot and its hold completing, never on `ActionDone`. N-Queens solution events can occur mid-trace; Floyd–Warshall negative-cycle terminal `ActionHighlight` is still a terminal outcome at last index.
- Tree removal: destination excludes deleted entity; show removal/copy explanation and snap disappearance. No generic exit-animation system required. Rewired edges render final topology with old-parent dashed witness only when explicitly stored in lesson state.
- Numeric plot domains are computed after trace construction from finite snapshot data, stored identically in every relevant typed state, padded by 10% with ±1 fallback for constant range. Add explicit domain fields to K-means/perceptron/genetic/Q-learning plot states if plotted extents require them; use task-specific names `XMin, XMax, YMin, YMax float64`. Gradient fields are enumerated in AI task book. Never use unseen history to imply algorithm knowledge; domains are presentation framing only.
- Shared file access is coordinator-only. Algorithm workers may modify their own lesson decoder for the validation decisions here and in their book. Pure algorithm fixes require coordinator's explicit narrow grant after a failing discrepancy test.
- Real missing algorithm events (e.g. union-find compression or BST relink) may gain snapshots at truthful existing Source IDs, as explicitly required by their task. This does not authorize cosmetic lift/landing frames, changed algorithm output, or fabricated comparisons. Update tests that pin previous trace counts deliberately.

### Remaining input boundaries, fixed for implementation

Apply in the relevant `Lesson.Build`, not browser-only checks. Preserve supported JSON alternate formats and omitted-field defaults. Explicit malformed JSON, null, wrong types, duplicate graph node IDs, missing referenced endpoints, nonfinite coordinates/weights, and invalid operation names return errors; direct Trace APIs retain documented guard behavior. Use signed 32-bit safe numeric ranges independent of native architecture. Test one accepted boundary and one rejected boundary for every limit.

- Graphs: at most 32 nodes and 128 edges; IDs/labels at most 64 Unicode code points; finite coordinates within ±1000000; weights within ±10000. Dijkstra/A* reject negative weights; Bellman–Ford/Floyd/Prim/Kruskal retain negative-weight semantics. Start/target must exist when supplied. Preserve parallel edges and directedness where supported; validate matrix shape for matrix-form inputs. Floyd matrix at most 16×16. Defaults inferred by existing builders remain valid; no hidden coordinate normalization may change A* heuristic input.
- BST/AVL: at most 32 initial values and 64 operations, integer values within ±1000000. Preserve duplicate-key semantics. Validate operation names and required operands.
- Trie: at most 32 words, each at most 32 runes; search/prefix at most 32 runes. Preserve Unicode rune semantics and empty-word behavior supported by direct API.
- Segment tree: 1–64 array values within ±1000000; require `0 <= queryL <= queryR < n` and update index within range. Sum fits 32-bit at these bounds.
- Union-find: n in [1,64], at most 128 operations, all referenced indices in [0,n). Known operations only. Direct Trace's skipped-invalid-operation behavior remains separately tested.
- Linear/binary search: at most 128 integers within ±1000000, target in same range. Binary input must be nondecreasing; reject unsorted input rather than secretly sorting. Preserve first-match behavior of linear search and actual binary duplicate-match semantics.
- KMP: text at most 256 bytes and pattern at most 64 bytes. Preserve actual byte algorithm; non-ASCII byte cells use `0xNN` with original escaped string shown separately. Preserve direct empty-pattern behavior and show its real result rather than synthesizing a match animation.

These bounds protect finite browser traces and arithmetic for the displayed lessons; they are implementation choices proposed for review, not previously existing guarantees. Do not broaden input support or add generic validation/configuration layers.
