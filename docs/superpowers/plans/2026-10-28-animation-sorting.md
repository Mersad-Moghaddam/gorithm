# Sorting Lesson Animation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `executing-plans` for assigned task execution, or `subagent-driven-development` when parent explicitly delegates. Steps use checkboxes. Execute only assigned algorithm; shared playback belongs to parent owner. No commits without authorization.

**Goal:** Teach bubble, selection, insertion, merge, quick, heap, countingsort, radixsort, and bucketsort through algorithm-specific, identity-preserving SVG snapshots.

**Architecture:** Each algorithm owns typed trace state and one pure SVG renderer. Root animation plan defines shared motion/data-attribute contract and integration; this companion supplies algorithm-specific semantics and geometry. No scene framework, new dependencies, or application redesign.

**Tech Stack:** Go 1.24+, Go/WASM, existing `trace.Step`, SVG, standard-library tests.

**Spec:** `docs/superpowers/specs/2026-10-28-algorithm-animations-design.md`. Shared integration contract: `docs/superpowers/plans/2026-10-28-algorithm-animations.md` (authoritative for attributes, timing, ownership, and dispatch).

## Global constraints and current evidence

- Source read: all 27 `lesson.go`, `trace.go`, and `algorithm.go` files listed below, plus `web/renderers/array.go`. Parent supplied Verify graph evidence: project `home-mersad-Projects-gorithm`, root `/home/mersad/Projects/gorithm`, generation `2026-10-27T07:56:32Z`, Build/Meta/State/Render search fully paginated (226 results), HEAD `49815a34b5cd2f77321f254c54de80d7bbe8e53e`. Local coverage recheck: these paths plus `web/wasm/player.go` all `no_recorded_issue` / `metadata_match`; best-effort, not completeness proof. Starting worktree clean.
- All nine currently declare `Renderer: "array"`, export `func Trace(initial []int) []trace.Step`, and expose `func Sort(arr []int) []int` returning sorted copy. Each `Build(input json.RawMessage) ([]trace.Step, error)` uses default unless JSON decodes as nonempty `[]int`; malformed JSON, `null`, and `[]` silently use default. Distinguish direct `Trace(nil)` tests from Build tests. Do not claim current Build rejects invalid input.
- `ArrayState` has `Values, Comparing, Swapped, Sorted []int` and `Current int`. Existing renderer selects colors by slot/status, lacks identity and auxiliary-storage geometry, clamps negative bars to minimum height, and supports typed ArrayState or map fallback. New embedded state does not satisfy its type assertion: parent must add explicit per-lesson dispatch atomically with trace migration. Do not change generic array rendering for unrelated lessons.
- Parent verified `RenderStep` replaces whole SVG and playback currently defaults to 500ms. Existing `array.go` emits inline 250ms height/y and 200ms fill/stroke transitions, which cannot preserve nodes across replacement. Shared playback owner implements identity motion, cancellation, 1800ms normal operation cadence, and 750ms swap arcs. Workers supply geometry/data attributes only; never edit player, controls, shell, themes, or scheduling.
- Final data snapshots authoritative. No cosmetic trace steps, artificial pause frames, or fake internal operations. Preserve existing LineID/action order and metrics unless specific discrepancy is proven and parent approves correction. Source remains algorithm-specific. No pure algorithm changes merely to suit animation.
- Identity starts at original input index, independent of value and current slot. Integer IDs move with values; never recover duplicate identity by matching values. Fill palette cycles ID modulo 5: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`. Comparison uses gold `#ffe29a` outline/bracket; confirmed result uses green `#9bf0bb` check/outline. Preserve identity fill. Labels include value and original identity when duplicates need disambiguation; color alone insufficient.
- Every proposed state below lives in its own renderer file, embeds `ArrayState`, and adds explicitly typed fields. Clone every mutable slice, including nested buffers and IDs, at capture. Fields not yet meaningful use `-1` indices/IDs and explicit booleans, not accidental zero-index highlights. No algorithm state recovered by parsing explanations, Variables strings, or earlier frames.
- Memory copies are real: array and auxiliary buffer can hold same logical identity simultaneously. Use root contract's distinct representation key for each storage location plus logical identity for transfer matching. Show copies as copies; insertion's conceptual hole must not erase real underlying duplicate memory value. Occupancy masks distinguish zero values from unwritten storage.
- Render functions return SVG fragments, matching existing renderer convention. Root contract supplies attribute names. Required geometry: logical identity, representation/storage role, source/destination slot centers, bounding boxes, operation kind, and arc control points only for genuine two-item swaps. Interpolate presentation, never mutate trace. Seek/reverse/reduced motion render final snapshot directly.
- Main sorting row uses height-encoded bars matching approved colorful reference, with zero baseline and signed values extending below it. Equal-size tiles are reserved for auxiliary storage and compact secondary heap array. Scale min/max once over the input, not per frame; zero retains visible baseline marker and label. Keep logical geometry within canvas bounds, allow stage-local scrolling to preserve readable labels at 320px. Escape derived text/attributes. No JavaScript dependencies. References below to main-row “tiles” mean these identity-bearing bars.

## Review focus and shared worker verification

Each algorithm task includes these checks in addition to its specific assertions:

- [ ] Add one focused algorithm test in listed `lesson_animation_test.go` and renderer test in listed `<id>_lesson_test.go`; use standard library only. Record current LineID/action sequence and metrics for concrete fixture before migration, then assert preserved sequence/metrics afterward. Assert terminal Values equal `Sort(input)` and expected literal output, input unchanged, every LineID belongs to `Lesson.Source()`.
- [ ] Assert snapshots immutable: mutate a later state's values/IDs/buffers in isolated test and ensure earlier snapshot unchanged. Assert final identity mapping contains every original item exactly once in authoritative final array, with stability only where algorithm guarantees it.
- [ ] Render fixture at `320x360` and `960x540`: valid parseable SVG fragment wrapped in `<svg>`, finite nonnegative box dimensions, no out-of-range dereferences, consistent logical IDs, readable empty state, expected labels and palette. Test real zero versus unfilled slot wherever occupancy exists.
- [ ] Run `go test ./internal/algorithms/<id> ./web/renderers` (substitute assigned ID), then `go test ./...` and `make build` after parent dispatch integration. Expected PASS and successful WASM/native build. If local Docker-backed Go wrapper loses GOOS/GOARCH, use repository AGENTS.md explicit-container procedure; report exact commands/outcomes.
- [ ] Hand parent browser cases: pause during arc, seek/back/reset, quick navigation, custom data, reduced motion, all eight themes, 320px layout. Shared owner runs against rebuilt embedded assets. Native tests do not establish browser timing/focus/cancellation correctness.

No implementation tests run for this documentation-only task. Root contract defines motion and dispatch integration; workers must follow it before browser completion claims.

## 1. Bubble (`bubble`)

**Paths:** read `internal/algorithms/bubble/lesson.go`, `internal/algorithms/bubble/algorithm.go`; modify `internal/algorithms/bubble/trace.go`; create `internal/algorithms/bubble/lesson_animation_test.go`, `web/renderers/bubble_lesson.go`, `web/renderers/bubble_lesson_test.go`. Read shared baseline `web/renderers/array.go`. Lesson metadata dispatch change belongs to parent integration agreement.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[29, 10, 14, 37, 13]`.

**New interface:** `func RenderBubbleLessonSVG(step trace.Step, width, height int) string`.

```go
type BubbleLessonState struct {
    ArrayState
    IDs []int
    Pass, Scan, UnsortedEnd int
}
```

**Teaching beats / layout:** One tile row with indices, adjacent-pair bracket, pass label, and sorted-suffix boundary. `L1` initialize; `L2` length; `L3` starts pass with inclusive `UnsortedEnd=n-pass-1`; `L5` gold bracket compares `j,j+1`; existing `L6` post-swap moves both IDs over opposite 750ms arcs; `L8` confirms last item and shrinks suffix; `L10` completion. No synthetic `L4` frame. Normal beats 1800ms; no swap arc when comparison false. Keep lane space above/below row for two crossing paths.

**Before/after:** Current `[3,1,2]` only recolors positional bars. New `L5` shows `3#0 > 1#1`; first `L6` Values `[1,3,2]`, IDs `[1,0,2]`; next swap yields `[1,2,3]`, IDs `[1,2,0]`; `L8` checks ID 0 at slot 2.

- [ ] Add `TestBubbleLessonSnapshots`: `[3,1,2]` final `[1,2,3]`, exactly two swap steps, first `L6` IDs `[1,0,2]`, `L8` sorted suffix includes index 2. `[2,2,1]` final IDs `[2,0,1]` proves equal items never swap; sorted `[1,2,3]` has zero swap steps but retains passes.
- [ ] Capture IDs alongside swaps and Pass/Scan/UnsortedEnd on existing steps; implement renderer and shared checks.
- [ ] Edge assertions: `[]` final empty, `[7]` final `[7]`, `[-2,0,-3]` final `[-3,-2,0]`. Current empty trace marks Sorted `[0]` despite no values: renderer must ignore invalid marker; report discrepancy before changing trace semantics. Current omitted Current defaults to 0; new renderer uses explicit scan fields.
- [ ] Run shared verification and deliver geometry integration notes.

**Copy-ready worker prompt:**
> Implement only section 1 (bubble) of `docs/superpowers/plans/2026-10-28-animation-sorting.md`, including global/shared checks and root motion contract. Read listed current files. Own bubble trace, bubble lesson animation test, and bubble renderer/test only. Add BubbleLessonState and RenderBubbleLessonSVG exactly as specified. Preserve source IDs, values, metrics, duplicate stability, and existing app design. Teach adjacent comparisons, 750ms swap geometry, and suffix growth; normal beats 1800ms owned by playback. No cosmetic frames, generic scene system, dependencies, player edits, pure Sort edits, or commits. Report empty Sorted marker discrepancy, test evidence, and dispatch handoff.

## 2. Selection (`selection`)

**Paths:** read `internal/algorithms/selection/lesson.go`, `internal/algorithms/selection/algorithm.go`; modify `internal/algorithms/selection/trace.go`; create `internal/algorithms/selection/lesson_animation_test.go`, `web/renderers/selection_lesson.go`, `web/renderers/selection_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[64, 25, 12, 22, 11]`.

**New interface:** `func RenderSelectionLessonSVG(step trace.Step, width, height int) string`.

```go
type SelectionLessonState struct {
    ArrayState
    IDs []int
    Pass, Scan, MinIndex, SortedPrefix int
}
```

**Teaching beats / layout:** One row; left prefix separated from scan range, independent `i`, `j`, and minimum flags. `L1,L2` setup; `L3` pass; `L4` candidate minimum; `L6` gold comparison; `L7` move minimum flag, not item; `L10` swap endpoints only if ActionSwap (750ms wide arcs), otherwise label already minimum; `L11` extend prefix check; `L12` done. 1800ms normal beats. Reserve arc clearance across whole unsorted range, not adjacent-only geometry.

**Before/after:** `[3,1,2]` starts minimum `3#0`; `L7` points to `1#1` without moving tiles; `L10` becomes `[1,3,2]` IDs `[1,0,2]`; final `[1,2,3]`. Current renderer conflates candidate and current slot and cannot retain candidate during scan.

- [ ] Add `TestSelectionLessonSnapshots`: above first swap and terminal values; `[1,2]` `L10` is ActionSet with unchanged IDs and no arc. `[2,2,1]` final IDs `[2,1,0]` demonstrates instability, not falsely stable equal colors.
- [ ] Snapshot min/pass/scan/prefix fields and ID swaps; implement labeled scan/minimum renderer and shared checks.
- [ ] Edge assertions: `[]` final empty, `[7]` unchanged, `[-1,0,-2]` final `[-2,-1,0]`, `[2,2]` unchanged IDs; do not add swaps for self-selection. Pure Sort does self-assignment while trace records no-op: same result, no reason to change algorithm.
- [ ] Run shared verification and geometry handoff.

**Copy-ready worker prompt:**
> Implement only section 2 (selection) and shared checks in `docs/superpowers/plans/2026-10-28-animation-sorting.md`; read root contract and listed source. Own selection trace/test plus selection renderer/test. Add SelectionLessonState and RenderSelectionLessonSVG. Show minimum flag separately from scan pointer, prefix growth, long-distance 750ms swap geometry, and no-op selection without fake arc; normal 1800ms timing belongs to playback. Preserve unstable identity ordering, source, metrics, pure Sort, and app design. No player edits, cosmetic frames, dependencies, generic framework, or commits. Return tests and dispatch needs.

## 3. Insertion (`insertion`)

**Paths:** read `internal/algorithms/insertion/lesson.go`, `internal/algorithms/insertion/algorithm.go`; modify `internal/algorithms/insertion/trace.go`; create `internal/algorithms/insertion/lesson_animation_test.go`, `web/renderers/insertion_lesson.go`, `web/renderers/insertion_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[12, 11, 13, 5, 6]`.

**New interface:** `func RenderInsertionLessonSVG(step trace.Step, width, height int) string`.

```go
type InsertionLessonState struct {
    ArrayState
    IDs []int // actual memory-slot identities; copies can repeat during shift
    Pass, Scan, KeyValue, KeyID, Hole, PrefixEnd int
    KeyHeld bool
}
```

**Teaching beats / layout:** Array row plus explicit held-key shelf, conceptual vacancy overlay, sorted-prefix bracket. `L1,L2` initialize prefix; `L3` choose next slot; `L4` copy key to shelf; `L5` start backward cursor; `L6` compare key to current item or show `j < 0` boundary without dereference; `L7` copy rightward; `L8` move cursor; `L10` write held key; `L12` done. Shifts are not exchanges despite current ActionSwap/IncSwaps; use 1800ms one-way copy geometry, never 750ms two-item swap arc. Label real duplicate memory value under conceptual hole, keeping snapshot truthful.

**Before/after:** `[3,1,2]` at first `L7` really contains `[3,3,2]`; key shelf contains `1#1`, hole at 0. Current bars appear to lose 1. New renderer explains copy, shows slot IDs `[0,0,2]` plus held ID 1, then `L10` yields `[1,3,2]` IDs `[1,0,2]`.

- [ ] Add `TestInsertionLessonSnapshots`: first shift above, `KeyHeld=true`, KeyValue 1, KeyID 1, Hole 0; first insert `[1,3,2]`; final `[1,2,3]` IDs `[1,2,0]`. `[2,2,1]` final IDs `[2,0,1]` proves stable equal ordering.
- [ ] Capture key identity before overwrites, clone slot IDs after copies, clear held status on insertion; implement renderer and shared checks without changing shift metrics.
- [ ] Edge assertions: `[]`, `[7]`, sorted `[1,2,3]` with no shifts; reverse `[3,2,1]` final `[1,2,3]` and boundary `j=-1`; `[-1,-3,0]` final `[-3,-1,0]`. Prefix bracket means locally sorted, not globally final until completion.
- [ ] Run shared verification; explicitly hand off shift-versus-swap operation discriminator.

**Copy-ready worker prompt:**
> Implement only section 3 (insertion), shared checks, and root contract from `docs/superpowers/plans/2026-10-28-animation-sorting.md`. Own insertion trace/test and insertion renderer/test. Add InsertionLessonState and RenderInsertionLessonSVG; preserve held-key identity through real duplicate memory slots and conceptual hole. Existing ActionSwap means shift here: emit one-way copy geometry, normal 1800ms, no fake exchange. Preserve stable order, metrics, source, pure algorithm, and app shell. No player edits, cosmetic frames, generic framework, dependencies, or commits. Report tests and motion discriminator needs.

## 4. Merge (`merge`)

**Paths:** read `internal/algorithms/merge/lesson.go`, `internal/algorithms/merge/algorithm.go`; modify `internal/algorithms/merge/trace.go`; create `internal/algorithms/merge/lesson_animation_test.go`, `web/renderers/merge_lesson.go`, `web/renderers/merge_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[38, 27, 43, 3, 9, 82, 10]`.

**New interface:** `func RenderMergeLessonSVG(step trace.Step, width, height int) string`.

```go
type MergeLessonRange struct { Left, Mid, Right int }
type MergeLessonState struct {
    ArrayState
    IDs, Aux, AuxIDs []int
    AuxFilled []bool
    Stack []MergeLessonRange
    Left, Mid, Right, I, J, K, CopyIndex int
}
```

**Teaching beats / layout:** Main row with active left/right brackets, recursion breadcrumb stack, auxiliary row with explicit empty cells, `i/j/k` pointers. `L1` init; `L2,L3` base check/return; `L5` split; `L6,L7` recursive descent; `L8,L10` merge setup; `L11,L12` allocate/init pointers; `L14` compare; `L15` left wins including equal; `L17` right wins; `L21,L22` drain remainder; `L23` copy back; `L24` locally merged; `L9` globally done. Keep pre-increment cursor semantics at emitted copy steps. Existing trace has no `L13/L19` emissions: do not add them for animation.

All transfers use 1800ms source-to-aux/copy-back geometry, not swap arcs. Stack snapshots identify recursion range without reconstructing history. Snapshot fields must explicitly clear old auxiliary buffer when entering different invocation; no stale nested-call buffer.

**Before/after:** `[2,1]` `L17` leaves main `[2,1]` while aux becomes `[1,empty]`, AuxIDs `[1,-1]`; `L21` gives aux `[1,2]`; first `L23` main `[1,1]` is real intermediate memory, not disappearance; final main `[1,2]` IDs `[1,0]`. Current array renderer hides aux throughout.

- [ ] Add `TestMergeLessonSnapshots`: above aux occupancy, pointer/range and copy-back assertions. `[2,1,2,1]` final `[1,1,2,2]`, IDs `[1,3,0,2]`; equal comparison takes `L15`. `[3,1,2]` final `[1,2,3]`; root range 0..2 uses mid 1.
- [ ] Add trace-owned parallel identity buffers and cloned stack/aux masks; implement renderer with separate storage keys and shared checks.
- [ ] Edge assertions: `[]`, `[7]`, `[-2,0,-3]` final `[-3,-2,0]`; `[0,0]` filled zero distinguishable from allocation. Local merged check not global Sorted claim. Aux snapshots remain unchanged after later recursive merges.
- [ ] Run shared verification and hand parent inter-row transfer geometry.

**Copy-ready worker prompt:**
> Implement section 4 (merge) and shared checks in `docs/superpowers/plans/2026-10-28-animation-sorting.md`, after root contract. Own merge trace/test and merge renderer/test. Add MergeLessonRange, MergeLessonState, RenderMergeLessonSVG exactly. Show recursive range, real aux occupancy, stable left tie, drain and copy-back using 1800ms one-way transfers. Preserve actual duplicate memory during copy-back, source step order/metrics, pure Sort, and app design. No cosmetic frames, player edits, generic scenes, dependencies, or commits. Return tests and storage-key geometry handoff.

## 5. Quick (`quick`)

**Paths:** read `internal/algorithms/quick/lesson.go`, `internal/algorithms/quick/algorithm.go`; modify `internal/algorithms/quick/trace.go`; create `internal/algorithms/quick/lesson_animation_test.go`, `web/renderers/quick_lesson.go`, `web/renderers/quick_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[10, 80, 30, 90, 40, 50, 70]`.

**New interface:** `func RenderQuickLessonSVG(step trace.Step, width, height int) string`.

```go
type QuickLessonRange struct { Low, High int }
type QuickLessonState struct {
    ArrayState
    IDs []int
    Stack []QuickLessonRange
    Low, High, Boundary, Scan, PivotIndex, PivotID, PivotValue int
    PivotActive bool
}
```

**Teaching beats / layout:** One row, active range bracket, `< pivot` region ending at i, scanned `>= pivot` region, unscanned region, pivot badge tied to identity. `L1` init; `L2` base check; `L3,L8` partition entry; `L9` last-slot pivot; `L10` boundary `low-1`; `L11,L12` inspect/compare; `L13` advance boundary; `L14` swap; `L17` place pivot; `L18` return index; `L4,L5` recurse; `L6` singleton confirmed; `L7` done. Empty ranges get textual base case, no negative-index node.

1800ms compare/region beats; 750ms actual swap arcs. Trace counts self-swaps at `L14` and `L17`; preserve metrics/steps, emit stationary/no-op presentation instead of crossing identical coordinates. Reset pivot state on new call; final pivot badge follows ID to sorted position.

**Before/after:** `[3,1,2]` pivot `2#2`; scanning 3 leaves boundary -1; scanning 1 gives first `L14` `[1,3,2]` IDs `[1,0,2]`; `L17` gives `[1,2,3]` IDs `[1,2,0]` and green check at 1. Current display cannot show simultaneous pivot and partition boundary.

- [ ] Add `TestQuickLessonSnapshots`: above range, boundary and pivot identity; `[1,2]` retains two self-swap actions without motion; `[2,2,1]` final `[1,2,2]` IDs `[2,0,1]`, no general stability promise.
- [ ] Track recursive range/pivot/IDs in snapshots, preserve strict `<` and sorted-set ordering; implement renderer and shared checks.
- [ ] Edge assertions: `[]`, `[7]`, equal `[2,2,2]` final unchanged values but IDs `[2,0,1]`; `[-1,-3,0]` final `[-3,-1,0]`; ascending/descending fixtures terminate and keep out-of-range base-case ranges safe.
- [ ] Run shared verification; hand parent self-swap suppression and broad arc geometry.

**Copy-ready worker prompt:**
> Implement only section 5 (quick), shared checks, and root contract from `docs/superpowers/plans/2026-10-28-animation-sorting.md`. Own quick trace/test plus quick renderer/test. Add QuickLessonRange, QuickLessonState, RenderQuickLessonSVG. Preserve Lomuto strict comparison, identity-based pivot, recursive range, and self-swap trace metrics; emit arcs only for distinct slots (750ms), normal beats 1800ms. No algorithm substitutions, cosmetic frames, player edits, app redesign, dependencies, frameworks, or commits. Return exact test evidence and dispatch/motion handoff.

## 6. Heap (`heap`)

**Paths:** read `internal/algorithms/heap/lesson.go`, `internal/algorithms/heap/algorithm.go`; modify `internal/algorithms/heap/trace.go`; create `internal/algorithms/heap/lesson_animation_test.go`, `web/renderers/heap_lesson.go`, `web/renderers/heap_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[12, 11, 13, 5, 6, 7]`.

**New interface:** `func RenderHeapLessonSVG(step trace.Step, width, height int) string`.

```go
type HeapLessonState struct {
    ArrayState
    IDs []int
    HeapSize, Root, LeftChild, RightChild, Largest, ExtractIndex int
    Building bool
}
```

**Teaching beats / layout:** Indexed array rail below complete-binary-tree projection; tree slot i has children `2*i+1`, `2*i+2`, only within HeapSize. ID colors/labels match rail, with distinct tree/array representation keys. Tree edges connect slots, not item identities. `L1,L2` init; `L3,L4` bottom-up build; `L12` subtree candidates; `L13,L14` child comparisons; `L15` promote winner; `L16` exchange; `L17` descend; `L6` extraction; `L7` root/end swap, retire end from tree; `L8` reduced heap; `L10` done. At compare snapshots Largest represents pre-comparison candidate; next emitted state reflects branch outcome.

1800ms candidate/heap-size beats; 750ms exchanges on both tree and rail, with tree endpoints computed from depth/position and rail endpoints from slots. Sorted suffix remains visible but outside active tree; never claim entire tree is a heap during bottom-up construction or before reheapification.

**Before/after:** `[2,3,1]` first build `L13` compares root 2 and child 3; `L16` gives `[3,2,1]` IDs `[1,0,2]`; first extraction `L7` gives `[1,2,3]` IDs `[2,0,1]`, HeapSize 2 and sorted slot 2. Current bars conceal parent/child relation.

- [ ] Add `TestHeapLessonSnapshots`: above values/IDs/size; parent-child edges only for indices below HeapSize; final `[1,2,3]` IDs `[2,0,1]`. `[2,2,1]` final `[1,2,2]` IDs `[2,1,0]`, no stability guarantee.
- [ ] Thread parallel IDs and heap metadata through existing `heapifyTrace` calls; clone state; implement two-projection renderer and shared checks.
- [ ] Edge assertions: `[]`, `[7]`, two nodes `[2,1]` final `[1,2]` with no right-child edge, `[-1,-3,-2]` final `[-3,-2,-1]`; wide tree remains within viewport and labels not clipped.
- [ ] Run shared verification; hand parent simultaneous representation motion geometry.

**Copy-ready worker prompt:**
> Implement section 6 (heap) and shared checks of `docs/superpowers/plans/2026-10-28-animation-sorting.md`; read root contract first. Own heap trace/test and heap renderer/test. Add HeapLessonState and RenderHeapLessonSVG. Show synchronized array/tree identities, bounded child indices, build versus extraction, shrinking HeapSize, 750ms swap endpoints, 1800ms normal beats. Preserve trace metrics, pure Sort, source IDs, and existing app. No shared player edits, fake frames, generic framework, dependencies, or commits. Return tests and dual-representation geometry handoff.

## 7. Counting Sort (`countingsort`)

**Paths:** read `internal/algorithms/countingsort/lesson.go`, `internal/algorithms/countingsort/algorithm.go`; modify `internal/algorithms/countingsort/trace.go`; create `internal/algorithms/countingsort/lesson_animation_test.go`, `web/renderers/countingsort_lesson.go`, `web/renderers/countingsort_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[4, 2, 2, 8, 3, 3, 1]`.

**New interface:** `func RenderCountingsortLessonSVG(step trace.Step, width, height int) string`.

```go
type CountingsortLessonState struct {
    ArrayState
    Input, InputIDs, Counts, Output, OutputIDs []int
    OutputFilled []bool
    Max, Scan, CountIndex, Target int
    Phase string // init, maximum, tally, prefix, output, done
}
```

**Teaching beats / layout:** Three labeled rows: immutable input, value-indexed counts, output occupancy. Counts never use item identity colors: they are counters, not input objects. `L1,L2,L3` init/base; `L4` early done; `L6,L7,L8,L9` max scan; `L12` allocate counts; `L13,L14` tally; `L16,L17` prefix; `L19` allocate output; `L20` backward input cursor; `L21` transfer to `count[value]-1`; `L22` decrement; `L24` done. Show prefix meaning “number <= value”; after decrement label remaining insertion boundary rather than frequency. All transfers/prefix updates 1800ms; no swap arcs.

Current ArrayState switches Values from input to zero-filled output at `L19`; preserve that legacy field behavior but typed Input/Output keep both visible. Do not parse count arrays from Variables or interpret unwritten zeros as items.

**Before/after:** `[2,1,2]` counts after tally `[0,1,2]`, prefix `[0,1,3]`. First backward `L21` places `2#2` at output slot 2, mask `[false,false,true]`; `L22` counts `[0,1,2]`; final output `[1,2,2]`, IDs `[1,0,2]`. Current bars lose original input and display phantom zero entries.

- [ ] Add `TestCountingsortLessonSnapshots`: above counts, prefix, placement before decrement, final IDs; `[0,2,0]` final `[0,0,2]` IDs `[0,2,1]`; zero is visible only when filled.
- [ ] Capture cloned counts/input/output/IDs/masks and phase at each existing step; implement renderer and shared checks. Count row may wrap or use existing local viewport for large valid ranges; never silently omit nonzero counts.
- [ ] Edge assertions: direct `[]`, `[7]`, `[0,0]` final unchanged with zero maximum; sparse `[0,8,0]` final `[0,0,8]` with count indices 0..8 represented. Domain blocker: negative multi-element inputs such as `[-1,2]` panic by negative count indexing; huge maxima risk allocation. Root validation owner must define bounded nonnegative accepted domain and rejection tests before exposure; do not quietly change Sort or invent limit here.
- [ ] Run shared verification; hand parent validation blocker and inter-row placement geometry.

**Copy-ready worker prompt:**
> Implement only section 7 (countingsort) and shared checks in `docs/superpowers/plans/2026-10-28-animation-sorting.md`, following root contract. Own countingsort trace/test and countingsort renderer/test. Add CountingsortLessonState and RenderCountingsortLessonSVG. Show immutable input, frequencies/prefix boundaries, real output occupancy, stable backward placement at 1800ms; no swap arcs. Preserve source, metrics, pure algorithm, existing design. Report negative/large-range input blocker to validation owner; no guessed bounds, player edits, extra frames, generic scene framework, dependencies, or commits. Return tests and geometry handoff.

## 8. Radix Sort (`radixsort`)

**Paths:** read `internal/algorithms/radixsort/lesson.go`, `internal/algorithms/radixsort/algorithm.go`; modify `internal/algorithms/radixsort/trace.go`; create `internal/algorithms/radixsort/lesson_animation_test.go`, `web/renderers/radixsort_lesson.go`, `web/renderers/radixsort_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[170, 45, 75, 90, 802, 24, 2, 66]`.

**New interface:** `func RenderRadixsortLessonSVG(step trace.Step, width, height int) string`.

```go
type RadixsortLessonState struct {
    ArrayState
    IDs, Output, OutputIDs []int
    OutputFilled []bool
    Counts [10]int
    Exp, MaxValue, Scan, Digit, Target, CopyIndex int
    Phase string // init, maximum, tally, prefix, output, copy, pass, done
}
```

**Teaching beats / layout:** Input row with active decimal digit underlined, ten counters labeled 0..9, output row, exponent badge. Missing higher digit explicitly 0, not a new leading digit in value label. `L1` init; `L2` max; `L3,L4,L7` pass/helper; `L8,L9` allocation/reset; `L10` tally; `L11` whole prefix accumulation snapshot; `L12,L13` backwards scan/digit; `L14` stable output placement; `L15` decrement; `L17` copy output to input per item; `L5` pass completed; `L6` globally done. One `L11` frame already covers all nine prefix additions: show resulting counts, never fabricate per-counter frames.

1800ms transfers/digit beats, including `L17`: ActionSwap and swaps metric currently mean copy, not two-item exchange. No 750ms arcs for these writes. Arrays reset masks every pass; item ID survives all passes. Counters carry no item IDs. Existing Variables hold only one count at tally and no output array: capture full typed data directly in Trace.

**Before/after:** `[21,11,12]` exp 1 output `[21,11,12]`, IDs `[0,1,2]`; exp 10 output `[11,12,21]`, IDs `[1,2,0]`. Stable tie in ones is visible even though values do not change that pass. Default ones pass result `[170,90,802,2,24,45,75,66]`; final `[2,24,45,66,75,90,170,802]`.

- [ ] Add `TestRadixsortLessonSnapshots`: above pass/final arrays and IDs; at exp 1 counts before prefix have digit 1 count 2 and digit 2 count 1; first reverse placement is `12#2` target 2. `[12,12,2]` final `[2,12,12]` IDs `[2,0,1]`.
- [ ] Capture full counts/output/masks and copied IDs on existing frames; implement digit/counter renderer and shared checks without altering metrics.
- [ ] Edge assertions: `[]`, `[7]`, `[0,0]` final `[0,0]` with no digit passes, `[0,10,1]` final `[0,1,10]`. Negative `[-1,2]` causes negative digit indexing; `exp *= 10` can overflow near machine integer bounds. Root validation owner must bound nonnegative domain consistently on native/WASM and test boundary before enabling such inputs; no pure algorithm fix absent reproduced discrepancy and approval.
- [ ] Run shared verification; hand parent copy-operation discriminator and input-domain blocker.

**Copy-ready worker prompt:**
> Implement section 8 (radixsort) and shared checks in `docs/superpowers/plans/2026-10-28-animation-sorting.md`, reading root contract first. Own radixsort trace/test plus radixsort renderer/test. Add RadixsortLessonState and RenderRadixsortLessonSVG. Capture full decimal counts, exponent, output occupancy, stable IDs across passes; show 1800ms transfers, no swap arc for L17 copies, no synthetic L11 frames. Preserve source/metrics/Sort/design. Report negative and exponent-overflow domain blockers. No player edits, framework, dependencies, or commits. Return tests and handoff geometry.

## 9. Bucket Sort (`bucketsort`)

**Paths:** read `internal/algorithms/bucketsort/lesson.go`, `internal/algorithms/bucketsort/algorithm.go`; modify `internal/algorithms/bucketsort/trace.go`; create `internal/algorithms/bucketsort/lesson_animation_test.go`, `web/renderers/bucketsort_lesson.go`, `web/renderers/bucketsort_lesson_test.go`; baseline `web/renderers/array.go`.

**Existing interface/default:** `func Trace(initial []int) []trace.Step`; JSON `[78, 17, 39, 26, 72, 94, 21, 12, 23, 68]`.

**New interface:** `func RenderBucketsortLessonSVG(step trace.Step, width, height int) string`.

```go
type BucketsortLessonState struct {
    ArrayState
    IDs []int
    Buckets, BucketIDs [][]int
    Min, Max, BucketCount, Scan, ActiveBucket, BucketOffset, WriteIndex int
    Phase string // init, range, scatter, local, gather, done
}
```

**Teaching beats / layout:** Input/output rail, exactly five labeled bucket lanes, range formula `(x-min)*4/(max-min)`, active bucket and gather cursor. `L1,L2` init; `L4` min/max; `L5,L6` five buckets; `L7` input selection; `L8` computed bucket index; `L9` append; `L11` select local bucket; `L12` local insertion sort completed; `L14` reset write index; `L15,L16` bucket/item read; `L17` gather assignment; `L20` return/gathered; `L21` done. Equal range exits after `L4` via `L20,L21`; never divide by zero or invent buckets on that branch.

Bucket lanes use exact integer-hash membership, not equal-width floating intervals incorrectly claiming max shares last interval: with k=5, final bucket only receives max for this formula. Wrap five lanes into rows at phone width while keeping explicit lane indices. Keep geometry for cross-row scatter/gather. Normal transfers 1800ms. Current local insertion loops emit only one before (`L11`) and one after (`L12`) snapshot per bucket: display stable aggregate reorder, not invented inner comparisons/shifts or swap arcs. No 750ms exchange unless trace actually supplies real paired exchange (current bucket trace does not).

**Before/after:** `[0,3,2,10]` range 0..10 hashes into buckets `[[0,2],[3],[],[],[10]]` after scatter (0 and 2 share bucket 0); local result same; gather `[0,2,3,10]`, IDs `[0,2,1,3]`. For visible local reorder use `[3,2,0,10]`: before bucket 0 `[2,0]` IDs `[1,2]`, after `L12` `[0,2]` IDs `[2,1]`; final `[0,2,3,10]` IDs `[2,1,0,3]`. Current buckets exist only as formatted Variables strings; no lane geometry or identity.

- [ ] Add `TestBucketsortLessonSnapshots`: both scatter and local reorder above, exact hash assignments, final arrays/IDs; `[2,2,0,10]` final `[0,2,2,10]` IDs `[2,0,1,3]` proves stable insertion within bucket and stable gather.
- [ ] Track nested bucket identity slices during existing insertion shifts, capture deep copies at emitted frames, and copy IDs into array during gather. Implement typed bucket renderer without parsing `bucketVars` strings; shared checks apply.
- [ ] Edge assertions: `[]`, `[7]`, `[4,4,4]` unchanged and no allocation/hash steps; `[-3,-1,-2,0]` final `[-3,-2,-1,0]`; empty buckets render empty lanes and gather emits no fictitious values. Intermediate gather memory retains overwritten duplicates truthfully. Extreme integer subtraction/multiplication can overflow: root validation owner must define safe range; do not silently switch algorithm/hash.
- [ ] Run shared verification; report source discrepancy: Source lacks equal-range guard present in both Trace and Sort (L20 presently carries equal-range exit). Parent may authorize explicit algorithm-specific source correction while preserving IDs; worker must not pretend displayed formula executes for equal values.

**Copy-ready worker prompt:**
> Implement only section 9 (bucketsort), shared checks, and root contract in `docs/superpowers/plans/2026-10-28-animation-sorting.md`. Own bucketsort trace/test and bucketsort renderer/test. Add BucketsortLessonState and RenderBucketsortLessonSVG. Show exact five-bucket integer hash, nested identity snapshots, aggregate local sort only at existing L12, and 1800ms scatter/gather transfers. Preserve pure Sort, source IDs, metrics, and app design. Report equal-range Source omission and overflow-domain blocker for parent resolution. No cosmetic inner frames, fabricated swap arcs, string parsing, player edits, framework, dependencies, or commits. Return tests and geometry handoff.

## Resolved boundary decisions and parent integration checklist

The following are explicit implementation decisions, not unresolved worker choices. Root spec's validation requirement authorizes these bounded lesson-input checks; pure `Sort` APIs are not being redesigned.

- Sorting lesson custom input: at most 64 integers. Comparison sorts and bucketsort accept `[-1000000,1000000]`; countingsort accepts `[0,1000]`; radixsort accepts `[0,1000000]`. Validate in each `Lesson.Build` before Trace; malformed/non-array JSON, empty explicit array, out-of-range values, or excessive length return readable error. `nil` input still selects default. This bounds WASM 32-bit bucket arithmetic, radix exponent, count storage, and pedagogical trace size.
- Give each worker ownership of its `lesson.go` for these checks. Assert `Build(nil)` succeeds; `Build([]byte("[]"))` and invalid JSON fail; signed comparison-sort fixture succeeds; `[-1,2]` fails in countingsort/radixsort; 64 values succeed and 65 fail; each numeric bound is accepted and bound+1 rejected. Dataset dialog must retain existing lesson on failure.
- Bubble empty direct Trace: set terminal Sorted to empty when `n==0`, not `[0]`; add regression assertion and preserve nonempty output/metrics.
- Bucket displayed Source: add a uniquely named line ID `L4_equal` immediately after min/max calculation containing `if minVal == maxVal { return arr }`. Re-number displayed line numbers without renaming existing IDs. Equal-range early exit trace references `L4_equal`; test source validation and absence of hash step for `[4,4,4]`.
- All mentions above of “1800ms transfer geometry” mean total operation budget. Actual movement is capped at 750ms and 60% of hold by root contract; never animate for the full hold interval.

These decisions supersede earlier “report blocker / parent must decide” wording and explicit prompts' narrower file lists. Include this section when dispatching every sorting worker.

- [ ] Supply final root-plan path and exact data-attribute/motion contract to workers. Agree representation keys for copies and heap projections; renderer geometry alone cannot animate replaced DOM.
- [ ] Parent owns atomic dispatch integration for all nine renderer functions and matching typed states. Keep unrelated `array` lessons intact; coordinate metadata change only if root uses renderer keys instead of lesson-ID dispatch.
- [ ] Resolve trust-boundary validation separately: current Builds fall back silently, counting/radix accept unsafe negative values, counting allocates by max, radix exponent and bucket arithmetic can overflow. These are observed source constraints, not authorization to rewrite algorithms during presentation work.
- [ ] Decide documented discrepancies: bubble empty Sorted marker, bucket Source missing equal-range guard. Preserve timing-independent snapshot semantics while fixing only proven issues with explicit tests/approval.
- [ ] Shared playback owner verifies 1800ms normal cadence, 750ms genuine swap arcs, cancellation/reduced-motion/seek behavior, and browser source-state synchronization. No local renderer worker may claim those checks from Go tests alone.
