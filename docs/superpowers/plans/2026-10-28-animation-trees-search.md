# Tree and Search Lesson Animations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` for assigned tasks. Parent may dispatch independent tasks with `superpowers:subagent-driven-development`. Steps use checkbox syntax. No commits without explicit authorization.

**Goal:** Teach BST, AVL, trie, segment tree, union-find, linear search, binary search, and KMP through immutable, algorithm-specific SVG snapshots.

**Architecture:** Each algorithm owns one typed lesson state and one pure renderer in `web/renderers/<id>_lesson.go`, with tests in `<id>_lesson_test.go`. Existing trace generators emit these states; parent owns shared playback, animation interpolation, and renderer dispatch. Preserve current design and Go algorithms; no generic scene manager.

**Tech Stack:** Go 1.24+, Go/WASM, stdlib, SVG. No dependencies.

**Spec / integration plan:** `docs/superpowers/plans/2026-10-28-algorithm-animations.md`. This child plan records source-backed decisions for its eight assigned lessons; parent must reconcile shared animation contracts before implementation.

## Global constraints

- Normal operation duration **1800ms**, movement portion **750ms**. Renderers emit destination geometry, not timers. Parent supplies movement between stable entity IDs, cancellation, reverse/seek, and reduced-motion behavior. An isolated snapshot must remain complete when interpolation is disabled.
- Entity palette, in stable identity order: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`. Gold `#ffe29a` means compare; green `#9bf0bb` means checked/done. Keep identity fill; show semantic state with outlines, badges, and explicit text. Green checked does not imply match: label `≠`, `checked`, or `found` distinctly.
- Preserve shell, themes, controls, and existing project design. Parent edits `web/wasm/player.go` and shared playback/dispatch; child workers do not. Keep existing `Meta().Renderer` values until parent switches dispatch by lesson ID or negotiated keys.
- State types live in `package renderers`, avoiding import cycles: algorithms already import renderers. Each function returns SVG **fragment**, matching existing renderers. Signature is `func Render<ID>LessonSVG(step trace.Step, width, height int) string`, with exact names below.
- Every snapshot owns every mutable slice/map, including embedded state and nested slices. `internal/trace/trace.go` only shallow-copies `Variables` and stores `State` directly. Clone variables' nested arrays too; never retain algorithm node pointers in exported state.
- Preserve deterministic iteration and immutable IDs. Tree IDs describe allocated node identity, never value, rank, position, or color. Duplicates must not allocate phantom nodes. BST replacement copies value into existing target ID; AVL rotation preserves all node IDs.
- SVG text and attribute strings use `html.EscapeString`. IDs are generated safe tokens; never use dataset strings as raw markup, ID, style, or URL. Existing tree/table/graph string interpolation is not a safe escaping model.
- No animation inferred from prose or `Variables` at render time. Typed fields carry phase, active IDs, bounds, operands, and results. Explicit booleans distinguish pending results from false/zero. Absent numeric index is `-1`; absent node ID is `""`.
- Source IDs below are current `Lesson.Source()` IDs. Repeated snapshots may share ID. Never attach fabricated helper-source IDs. Preserve final output parity with `algorithm.go` subject to documented Trace defaults/normalization.
- Geometry contract: use inner SVG with `viewBox="0 0 W H"`, viewport `width`/`height`, `preserveAspectRatio="xMidYMid meet"`; all geometry finite for nonpositive caller dimensions (return readable empty-state text). Logical W/H grow to fit minimum spacing, rather than overlap/clamp nodes. Parent decides local pan/zoom for large inputs; default examples must be legible at 320px. Default text 14 logical px, numbers 16, dark text `#101820` on bright badges. Wrap explanations in existing explanation pane, not long SVG lines.
- Prefix DOM IDs by lesson, e.g. `bst-node-1`; stable `data-entity` on movable groups, with root plan's `data-x`, `data-y`, and motion metadata. Edges key by child ID where parent can change. Fixed per-entity palette index derives from allocation ordinal/index, not sorted position. No CSS-only transition claim when parent replaces SVG markup.

## Evidence and existing limitations

Parent supplied Verify graph discovery: project `home-mersad-Projects-gorithm`, generation `2026-10-27T07:56:32Z`, HEAD `49815a34b5cd2f77321f254c54de80d7bbe8e53e`, 226 Build/Meta/Render/State matches fully paginated. Child confirmed project root via `list_projects`, read all eight `lesson.go`, `trace.go`, `algorithm.go` files and `internal/trace/{types,trace}.go`, plus `web/renderers/{tree,graph,array,table}.go`. Coverage recheck for these 30 paths reported `no_recorded_issue`, `metadata_match`, generation `2026-10-28T05:35:08Z`; best-effort, not completeness proof. Initial working tree clean.

Current tree renderer ignores width/height for node layout; array renderer uses magnitude bars unsuitable for search identity; trie state omits terminal flags; segment-tree state omits intervals; union-find graph stores union-request edges rather than parent links; KMP shows precomputed LPS only. These dedicated renderers address those specific teaching gaps.

All eight current Builds return nil error and often silently use defaults on malformed/empty input. Record this behavior; do not silently change input contracts in animation work. Parent owns any input-validation policy change. In particular binary search assumes sorted input without checking it; show precondition and flag unsorted input truthfully, never secretly sort it.

## Review focus

1. BST successor-copy duplicate values and AVL double rotations: node identity survives, old frames remain unchanged (Tasks 1–2).
2. Trie rune indices versus KMP byte offsets: Unicode labels and markup-like input remain correct and escaped (Tasks 3, 8).
3. Empty/direct Trace input differs from Build fallback: test both boundaries without inventing empty-input UI support (all tasks).
4. Union-find compression inside union/connected and segment-tree partial query returns: displayed structure equals actual snapshot, not historical request edges or accumulated double counts (Tasks 4–5).
5. Tiny viewport, long/skewed trees, and terminal/out-of-range pointers: finite geometry, unique IDs, no fake match or absent-index access (all renderer tests; parent browser checks).

## Shared test and handoff contract

Each task adds trace assertions to its existing `internal/algorithms/<id>/algorithm_test.go` (read before editing) and dedicated renderer tests in external `package renderers_test` when importing algorithm packages, to avoid cycles. Use stdlib `testing`, `encoding/xml`, `reflect`, `strings` only. Wrap returned fragment in `<svg xmlns="http://www.w3.org/2000/svg">` for XML parsing. Test behavior and geometry rather than exact giant SVG strings.

Every task's tests must verify: every emitted `LineID` exists in `Lesson.Source()`; typed state present on every frame; repeated rendering deterministic; rendering never mutates snapshot; early snapshot unchanged after later snapshots/input mutation; unique entity IDs; edges reference existing nodes; comparison/done have readable non-color cues; no NaN/Inf/negative dimensions at 320×420 and 800×480. Test missing/wrong state produces readable fallback without panic. XML decoding must reject neither escaped quotes nor `<>&` in labels. Color checks inspect representative entities, not every snapshot for all seven colors.

Per-task tests below are required in addition to these shared assertions. Run targeted checks once red and once green. Parent runs `go test ./...`, `make build`, race checks for changed trace state, browser interaction/animation checks, and verifies cancellation before claiming integration complete. Local Docker-backed `go` wrapper may need explicit container `GOOS=js GOARCH=wasm`; see repository AGENTS.md. Documentation preparation itself does not run product tests.

---

## Task 1 — BST (`bst`)

**Files:** Read `internal/algorithms/bst/algorithm.go`, `internal/algorithms/bst/lesson.go`, `web/renderers/tree.go`. Modify `internal/algorithms/bst/trace.go`; tests in `internal/algorithms/bst/algorithm_test.go`. Create `web/renderers/bst_lesson.go` and `web/renderers/bst_lesson_test.go`.

**Existing interface:** `func Trace(initial []int, ops []Operation) []trace.Step`. `Operation` is `{Type string json:"type"; Value int json:"value"}`, types `insert`, `search`, `delete`; unknown type currently ignored. Build accepts `{initial: []int, ops: []Operation}` or bare `[]int`; nonempty fields replace respective defaults. Default initial `[50,30,70,20,40,60,80]`, operations insert 35, search 40, delete 30. Bare arrays retain default operations. Direct Trace defaults only when both slices empty; initial values are built before first frame, not individually traced.

**Proposed interface:** `func RenderBSTLessonSVG(step trace.Step, width, height int) string`.

```go
type BSTLessonState struct {
    TreeState
    Phase string // init, compare, descend, insert, duplicate, successor, copy, remove, done
    Operation string // insert, search, delete; empty during initial snapshot
    Target int
    ActiveID, ParentID, SuccessorID, ReplacementID string
    PathIDs []string
    LeftIDs, RightIDs []string // aligned with Nodes; explicit child sides
    RemovedID string
    OldValue int
    HasOldValue bool
    Found, ResultKnown bool
}
```

Node IDs retain existing pointer-to-ID registry; `LeftIDs`/`RightIDs` prevent child-side guessing during duplicate-valued successor copy. Clear per-operation fields at operation start. `RemovedID` may name exited entity for parent's transition; it is not an edge endpoint in destination tree.

**Visualization/geometry:** Ordered tree plus target chip at `(32,24)` and operation badge. Assign in-order rank slots with minimum 64px pitch; `W=max(800,64*(N+1))`, `x=64*(rank+1)+(W-64*(N+1))/2`, `y=88+76*depth`, radius 22. `H=max(440,152+76*maxDepth)`. In-order traversal follows explicit child IDs, not values. Highlight path edges gold; checked nodes green outline. Put `nil` landing marker on absent requested child side, 32px horizontally and 76px below parent. Successor value travels as separate labeled copy chip; do not move successor node into target's identity. Render deleted node only via parent exit transition, destination omits it.

**Teaching sequence / example:**
1. `L1`: show fully built default initial tree; disclose initialization was prebuilt.
2. Insert 35: `L1`, compare `L4`, left `L6` at 50, right `L7` at 30, left `L6` at 40, allocate `L5`, settle `L8`.
3. Search 40: `L10`, inspect `L13`, branches `L16`/`L17`, match `L15`; missing target ends `L14` with visible nil marker.
4. Delete 30: `L19`, `L22`, left `L24`, select `L26`, find successor along `L29` (40 then 35), copy 35 at `L30`, recurse `L31`, remove leaf at `L27`, finish `L33`.
5. One-child deletion uses `L27` or `L28`; absent node uses `L23`; duplicate insert uses `L8` without insertion movement.

**Required assertions:** `TestBSTLessonSuccessorIdentity` checks final in-order `[20,35,40,50,60,70,80]`, target's original ID changes value 30→35, successor's distinct ID disappears, and temporary equal-valued nodes at `L30` have distinct IDs. `TestBSTLessonDeleteCases` covers leaf, left-only/right-only root, absent target, deletion to empty, and empty initial with explicit insert operation. Root/parent replacement must already be committed in post-removal snapshot: current recursive tracer records removal before caller reconnects, so add a post-relink snapshot at same `L27`/`L28` or `L33` with correct topology. `TestBSTLessonDuplicate` checks unchanged node count/IDs on insert existing key; negative/zero keys and monotone 20-node tree have valid geometry.

- [ ] Add named trace tests and renderer geometry/escaping tests; run `go test ./internal/algorithms/bst ./web/renderers` and observe missing typed API failures.
- [ ] Define `BSTLessonState` and `RenderBSTLessonSVG` in dedicated renderer file with geometry above.
- [ ] Emit complete cloned state from BST tracer, including post-relink deletion frames and stable successor-copy identity.
- [ ] Run `gofmt` on changed Go files, rerun targeted command; require PASS.
- [ ] Hand parent renderer name/state, test output, and two-child deletion frame indices; do not edit shared dispatch.

**Copy-ready agent prompt:**
```text
Implement Task 1 (BST) in docs/superpowers/plans/2026-10-28-animation-trees-search.md, including its global/test contracts. Read internal/algorithms/bst/{lesson,trace,algorithm,algorithm_test}.go and web/renderers/tree.go. Own only internal/algorithms/bst/trace.go, internal/algorithms/bst/algorithm_test.go, web/renderers/bst_lesson.go, web/renderers/bst_lesson_test.go. Add BSTLessonState and RenderBSTLessonSVG(step trace.Step, width, height int) string. Preserve Trace API/defaults; teach insert/search/delete, stable IDs through successor value-copy and duplicate values, accurate post-relink snapshots. Implement stated geometry, source IDs, tests and palette; clone state and SVG-escape strings. Parent owns playback/dispatch: 1800ms operations, 750ms movement. No dependencies, generic manager, shared edits, or commit. Run targeted tests; return files, evidence, integration handoff and blockers.
```

## Task 2 — AVL (`avl`)

**Files:** Read `internal/algorithms/avl/algorithm.go`, `internal/algorithms/avl/lesson.go`, `web/renderers/tree.go`. Modify `internal/algorithms/avl/trace.go`, `internal/algorithms/avl/algorithm_test.go`. Create `web/renderers/avl_lesson.go`, `web/renderers/avl_lesson_test.go`.

**Existing interface:** `func Trace(values []int) []trace.Step`. Build accepts `{values: []int}` or bare `[]int`; nonempty array required to override defaults. Default `[10,20,30,40,50,25]`. Direct Trace also substitutes default for empty slice. Duplicate insertion ignored; only insertion is traced, despite native AVL also supporting Search.

**Proposed interface:** `func RenderAVLLessonSVG(step trace.Step, width, height int) string`.

```go
type AVLLessonState struct {
    TreeState
    Phase string // insert, compare, height, balance, rotation, done
    Target int
    ActiveID, PivotID, NewRootID, TransferID string
    PathIDs, LeftIDs, RightIDs []string
    Heights, Balances []int // aligned with Nodes
    Rotation, Case string // left/right; LL/RR/LR/RL, empty when absent
    RotationPart int // 0, 1, 2
}
```

**Visualization/geometry:** In-order slots as BST, pitch 88, `W=max(800,88*(N+1))`, centered slots, `y=96+104*depth`, radius 22, `H=max(440,184+104*maxDepth)`. Two-line `h=…`, `bf=…` badge under each node; null child height 0 in active balance equation banner. During rotation show pivot→new-root curved directional arc above pair and transfer subtree with dashed outline. Node IDs/colors unchanged; destination coordinates follow new depth. RotationCase/Part supplies `LR 1/2` and `LR 2/2`, not a single instantaneous double rotation.

**Teaching sequence / example:** `L1` begin insert; `L2` allocate height 1; `L1` compare and `L3`/`L4` descend; duplicate `L5`; unwind `L6` update height, `L7` compare balance. Default insertion of 30 produces `bf(10)=-2`, RR `L9`; show left rotation setup `L22`, rewiring **new repeated/existing-source frame `L23`**, height result `L24`. LL is `L8`, `L18`, **new frame `L19`**, `L20`. LR uses `L10`, `L11` before each of its two rotations; RL uses `L13`, `L14`. `L16` settles balanced subtree and final tree. Default final root 30, in-order `[10,20,25,30,40,50]`, height 3. Rewire before recording `L19`/`L23`; update heights before `L20`/`L24` so source/state agree.

**Required assertions:** `TestAVLLessonRotationCases` uses `[30,20,10]` LL, `[10,20,30]` RR, `[30,10,20]` LR, `[10,30,20]` RL: root 20, three unique persistent IDs, height 2, terminal balance magnitude ≤1. Assert LR/RL contain two ordered rotation stages with an observable intermediate topology. `TestAVLLessonTransferSubtree` uses `[30,20,40,10,25,5]`: right rotation transfers 25 to old pivot's left without losing ID. For each height/balance stage, compare typed values with structural child heights; at rewire stages heights may still be stale and phase must say pending recomputation. Duplicate `[5,5,5]` creates one node. Default/empty Trace parity and deep snapshot cloning required.

- [ ] Add named rotation/identity assertions and renderer badge/arc tests; run `go test ./internal/algorithms/avl ./web/renderers` red.
- [ ] Define state/renderer and geometry; retain pointer registry, explicitly snapshot child IDs, heights and balances.
- [ ] Split rotation rewiring and height-update frames at real `L19`/`L23` and `L20`/`L24`; expose each double-rotation stage.
- [ ] Format changed Go files and run targeted command green.
- [ ] Hand parent four case fixture frame indices and renderer API; no shared dispatch changes.

**Copy-ready agent prompt:**
```text
Implement Task 2 (AVL) and global/test contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Read internal/algorithms/avl/{lesson,trace,algorithm,algorithm_test}.go and web/renderers/tree.go. Own avl trace.go/algorithm_test.go plus web/renderers/avl_lesson.go and avl_lesson_test.go only. Add AVLLessonState and RenderAVLLessonSVG(step trace.Step, width, height int) string. Preserve insertion/default semantics. Teach heights/balance, LL/RR/LR/RL, each rotation rewire and height update at actual Source IDs; preserve IDs and transfer subtrees. Use specified geometry/palette, immutable state, escaping, targeted tests. Parent owns 1800ms/750ms playback and dispatch. No shared edits/deps/managers/commit. Return tests, handoff, blockers.
```

## Task 3 — Trie (`trie`)

**Files:** Read `internal/algorithms/trie/algorithm.go`, `internal/algorithms/trie/lesson.go`, `web/renderers/tree.go`. Modify `internal/algorithms/trie/trace.go`, `internal/algorithms/trie/algorithm_test.go`. Create `web/renderers/trie_lesson.go`, `web/renderers/trie_lesson_test.go`.

**Existing interface:** `func Trace(words []string, searchWord string, prefix ...string) []trace.Step`. Build accepts `{words: []string, search: string, prefix: string}` or bare `[]string`; nonempty fields replace defaults independently. Default words `["go","golang","good","god"]`, search `"gold"`, prefix `"go"`. Direct Trace defaults empty words/search, takes only first variadic prefix when nonempty, otherwise `"go"`; no way to disable prefix via empty string. Individual words can be empty: inserts terminal marker on root. Algorithm traverses runes, but current loop index variables are UTF-8 byte offsets.

**Proposed interface:** `func RenderTrieLessonSVG(step trace.Step, width, height int) string`.

```go
type TrieLessonState struct {
    TreeState
    Phase, Operation, Word string // operation insert/search/startsWith
    Runes []rune
    RuneIndex, ByteOffset int
    ActiveID, NextID string
    PathIDs, TerminalIDs []string
    Found, ResultKnown bool
    Created bool
}
```

**Visualization/geometry:** Root labeled `root`, child circles show character, connecting edges label rune. Double ring + `end` badge persists for every terminal node, not just current word. Rune strip at y=30, active rune gold; consumed prefix green underline. Sort children by rune using existing `getChildren`. Leaf slots `x=48+72*k`, parent midpoint of first/last child's x; center extent within `W=max(800,96+72*(leafCount-1))`; `y=112+80*depth`, radius 20; `H=max(480,184+80*maxDepth)`. Missing edge draws dashed 48px stub with escaped missing rune, never fake allocated node. Operation result explicitly distinguishes `path exists` from `word ends here`.

**Teaching sequence / example:** Insert each default word: `L1` begin, `L2` root, `L3` rune, `L4` edge lookup, `L5` create or `L6` reuse, `L7` advance, `L9` terminal flag, `L10` complete. Search `gold`: `L11`/`L12`, repeated `L13`/`L14`/`L16` for g,o,l, then missing d at `L15`. Prefix `go`: `L20`/`L21`, `L22`/`L23`/`L25`, `L27` true. Searching `gol` follows path but `L18` false, despite prefix query `gol` true. Keep terminal markers for go/golang/good/god throughout search.

**Required assertions:** `TestTrieLessonTerminalVsPrefix` proves `gol` search false/prefix true; default gold false/go true, shared g/o node IDs reused across all inserts. `TestTrieLessonRunesAndEscaping` uses words `["é猫","é","<&\"",""]`, search `"é猫"`, prefix `"é"`: displayed rune indices 0,1 correspond byte offsets 0,2; root terminal for empty word; malformed markup never appears raw. Duplicate word creates no extra IDs; search absent first edge shows missing stub; empty/default behavior preserved. IDs not assigned by rune label (many branches may contain same rune).

- [ ] Add named trace assertions and renderer persistent-terminal/XML tests; run `go test ./internal/algorithms/trie ./web/renderers` red.
- [ ] Define state/renderer, rune strip, persistent terminal rings and deterministic leaf layout.
- [ ] Populate typed state on insertion/search/prefix frames; clone rune/path/terminal slices, distinguish rune and byte indices.
- [ ] Format files and rerun targeted command green.
- [ ] Hand parent rune and missing-edge fixtures plus renderer name.

**Copy-ready agent prompt:**
```text
Implement Task 3 (Trie) and shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Read internal/algorithms/trie/{lesson,trace,algorithm,algorithm_test}.go and web/renderers/tree.go. Own trie trace.go/algorithm_test.go plus web/renderers/trie_lesson.go and trie_lesson_test.go. Add TrieLessonState and RenderTrieLessonSVG(step trace.Step, width, height int) string. Preserve Trace variadic/default semantics. Teach edge reuse/create, persistent terminal flags, failed word versus successful prefix, rune index versus byte offset. Stable IDs, sorted children, specified geometry/palette, immutable snapshots and SVG escaping mandatory. Parent owns playback/dispatch 1800ms/750ms. No shared edits/deps/managers/commit. Run targeted tests; return handoff and blockers.
```

## Task 4 — Segment tree (`segmenttree`)

**Files:** Read `internal/algorithms/segmenttree/algorithm.go`, `internal/algorithms/segmenttree/lesson.go`, `web/renderers/tree.go`. Modify `internal/algorithms/segmenttree/trace.go`, `internal/algorithms/segmenttree/algorithm_test.go`. Create `web/renderers/segmenttree_lesson.go`, `web/renderers/segmenttree_lesson_test.go`.

**Existing interface:** `func Trace(arr []int, ql, qr, updIdx, updVal int) []trace.Step`. Build accepts `{arr: []int, query: []int, update: {index: int, value: int}}` with nonempty arr or bare `[]int`. Default arr `[1,3,5,7,9,11]`, query `[1,4]`, update index 2 value 6. Custom arr without exactly two query entries uses full range; omitted update still uses 2→6. Empty direct Trace resets all parameters to defaults. Trace clamps ql below 0 and qr above n−1, resets reversed range to full array; invalid update index becomes 0. Native Query/Update have different out-of-bounds semantics: preserve trace behavior and label normalized values, do not imply strict rejection.

**Proposed interface:** `func RenderSegmentTreeLessonSVG(step trace.Step, width, height int) string`.

```go
type SegmentTreeLessonNode struct {
    ID, ParentID string
    L, R, Sum int
    Built bool
}
type SegmentTreeLessonState struct {
    Nodes []SegmentTreeLessonNode
    Values []int
    Phase, Overlap string // build/query/update; none/partial/full or empty
    ActiveID string
    QL, QR, UpdateIndex, UpdateValue int
    AcceptedIDs []string
    Result, LeftResult, RightResult int
    ResultKnown, LeftKnown, RightKnown bool
    PathIDs []string
}
```

**Visualization/geometry:** Explicit interval/sum boxes, not generic circular value tree. Array row at y=44; pitch 76, `W=max(800,64+76*n)`, array cell center `x=32+76*(i+0.5)` centered in W. Interval node x is midpoint of leaf centers L/R, y=152+92*depth, box 68×52 with `[L,R]` above `sum`. `H=max(500,244+92*maxDepth)`. Unbuilt nodes display `?`, never zero as computed sum. Query bracket spans inclusive cell edges; full overlap green checked badge, disjoint dashed `+0`, partial gold split. Accepted-cover strip below tree lists disjoint intervals and sums; ancestor return must not add accepted descendants again. Update marks one leaf and only ancestors; show `left + right = sum` at active parent. IDs remain current `node-<heap index>` across build/query/update.

**Teaching sequence / example:** Build `L1`/`L2`, recurse `L4`, leaf `L5`, split `L6`, left/right `L7`/`L8`, combine `L9`, return `L10`, complete `L3` root 36. Query `L11`/`L12`, inspect `L14`; full `L15`, disjoint `L16`, partial `L17`, children `L18`/`L19`, combine `L20`, result `L13` =24. Accepted cover `[1,1]` sum3, `[2,2]` sum5, `[3,4]` sum16. Update `L21`/`L22`, inspect `L24`, compare `L26`, branch `L27`/`L28`, set leaf `L25` from5 to6, recompute `L29`, settle `L30`, finish `L23` root37. Do not display a second query result as executed: no post-update query exists in current trace.

**Required assertions:** `TestSegmentTreeLessonCoverAndUpdate` checks accepted intervals disjoint and union exactly [1,4], sum24; final values `[1,3,6,7,9,11]`, root37; only covering ancestors change. `TestSegmentTreeLessonBuildKnown` distinguishes actual zero leaf from unbuilt sum. `TestSegmentTreeLessonNormalization` covers one-element `[0]`, negative numbers, non-power-of-two size5, reversed query, clamped query, invalid update index; compare with native algorithm using **normalized** parameters. Root completion sums equal full array sum; recursive subtotals include known flags so pending zero is not treated as result.

- [ ] Add named interval/normalization tests plus renderer box/cover geometry assertions; run `go test ./internal/algorithms/segmenttree ./web/renderers` red.
- [ ] Define typed node/state and renderer; preserve node heap IDs and leaf-aligned geometry.
- [ ] Track built nodes, accepted query cover, per-call return operands, and cloned array/path state directly in tracer.
- [ ] Format changed files; targeted command green.
- [ ] Hand parent build/query/update fixture frames and renderer API.

**Copy-ready agent prompt:**
```text
Implement Task 4 (Segment Tree) plus shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Own internal/algorithms/segmenttree/trace.go, algorithm_test.go, web/renderers/segmenttree_lesson.go, segmenttree_lesson_test.go; read lesson.go/algorithm.go/tree renderer first. Add SegmentTreeLessonState and RenderSegmentTreeLessonSVG(step trace.Step, width, height int) string. Preserve exact Trace defaults/normalization. Teach interval build, full/partial/disjoint query cover without double counting, leaf update and ancestor sums. Use typed bounds/Built/return-known fields, stable heap IDs, specified geometry/palette, immutable snapshots and escaping. Parent handles shared playback/dispatch, 1800ms/750ms. No shared edits/deps/managers/commit. Run targeted tests; return integration evidence/blockers.
```

## Task 5 — Union-find (`unionfind`)

**Files:** Read `internal/algorithms/unionfind/algorithm.go`, `internal/algorithms/unionfind/lesson.go`, `web/renderers/graph.go`. Modify `internal/algorithms/unionfind/trace.go`, `internal/algorithms/unionfind/algorithm_test.go`. Create `web/renderers/unionfind_lesson.go`, `web/renderers/unionfind_lesson_test.go`.

**Existing interface:** `func Trace(n int, ops []Operation) []trace.Step`; `Operation` fields `Op string json:"op"`, `I int json:"i"`, `J int json:"j,omitempty"`. Build object `{n: int, operations: []Operation}`; defaults n5, unions (0,1),(1,2),(3,4),(0,4), finds2,4, connected(0,3). Only positive n/nonempty operations override respective defaults. Trace n≤0 returns zero steps; invalid indices/unknown ops skipped. Empty ops with positive n emits initial/final only. Existing graph edges are requested union pairs, not actual parent pointers; replacing those with parent forest is essential.

**Proposed interface:** `func RenderUnionFindLessonSVG(step trace.Step, width, height int) string`.

```go
type UnionFindLessonState struct {
    Parent, Rank []int
    Phase, Operation string // init, walk, roots, compare, link, compress, done
    I, J, RootI, RootJ, Active int
    Path []int
    ChangedNode, OldParent, NewParent int
    Components int
    Connected, ResultKnown, Redundant bool
}
```

IDs derive from element index `unionfind-node-<i>`; no GraphState embedding because it would retain misleading historical edges. Root rank is not current tree height after compression; label `rank`, never `height`.

**Visualization/geometry:** Parent forest with upward directed edges child→parent; root `parent[i]=i` shown as `root` badge, no self-loop. Sort component roots and children by element index. Recursive subtree span = sum child spans, minimum one leaf; leaf pitch76, extra76 gap between components, W at least800; place parent at midpoint of child extent, y=88+80*depth, radius22. H includes parent/rank table below deepest node: table y=168+80*maxDepth, rows index/parent/rank of height32. Table columns width56, expand W to at least64+56*n. Compression shows old parent edge dashed during parent transition, destination solid edge uses current Parent; green outline follows visited path. Entity fill stays index-based even when representative changes. Component count decrements only on successful union.

**Teaching sequence / example:** `L1` initial forest (optionally split allocation `L2`, identity initialization `L3`); union `L11`, call roots `L12`, traced Find `L5`/`L6`, recurse/compress `L7`, root return `L9`; compare equal `L13`, rank lower `L14`, higher `L15`, tie-link/rank increment `L16`, done `L17`. Connected `L19` invokes both traced finds then `L20` compares roots. Default after unions has `parent[4]=3`, `parent[3]=0`; Find(4) visits 4,3,0 and changes parent4 to0 at `L7`, count stays1, rank0 remains2. Final parents all0. Calls to dsu.Find inside union/connected currently compress without per-edge frames: use one local traced-find routine for every caller, preserving native behavior and existing public API.

**Required assertions:** `TestUnionFindLessonCompression` checks above default chain before/after, every destination arrow exactly Parent[i] when nonroot, no union-request edge (0,4) substituted for parent4=3 before compression. `TestUnionFindLessonRanksAndRedundant` covers both unequal rank directions, tie, repeated union, union(i,i), connected false/true; ranks increase only tie, component count unchanged by find/compress/redundant calls. `TestUnionFindLessonNestedFind` uses unions (0,1),(2,3),(0,2), then connected(3,1) and verifies parent3 changes2→0 with `L7` frame. Test n0, n1, invalid i/j (no panic), independent snapshots and deterministic forest layout.

- [ ] Add named rank/compression tests and directed-edge/table renderer tests; run `go test ./internal/algorithms/unionfind ./web/renderers` red.
- [ ] Define typed state/forest renderer; remove dependence on historical union edge list within this tracer.
- [ ] Trace actual parent walk and every changed parent at `L7`, including union/connected finds; clone parent/rank/path snapshots.
- [ ] Format files and run targeted command green.
- [ ] Hand parent compression fixture frames and API; flag any existing tests tied to old GraphState for deliberate migration.

**Copy-ready agent prompt:**
```text
Implement Task 5 (Union-Find) and shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Read unionfind lesson/trace/algorithm/tests and graph renderer. Own internal/algorithms/unionfind/trace.go, algorithm_test.go, web/renderers/unionfind_lesson.go, unionfind_lesson_test.go. Add UnionFindLessonState and RenderUnionFindLessonSVG(step trace.Step, width, height int) string. Preserve Trace/default/invalid-index behavior. Replace historical union-edge teaching with actual parent forest, rank table, component count, and per-edge compression from every Find caller. Use immutable parent/rank/path slices, stable element IDs, exact source IDs/geometry/palette and SVG escaping. Parent handles playback/dispatch 1800ms/750ms. No shared edits/deps/managers/commit. Run targeted tests; return evidence and blockers.
```

## Task 6 — Linear search (`linearsearch`)

**Files:** Read `internal/algorithms/linearsearch/algorithm.go`, `internal/algorithms/linearsearch/lesson.go`, `web/renderers/array.go`. Modify `internal/algorithms/linearsearch/trace.go`, `internal/algorithms/linearsearch/algorithm_test.go`. Create `web/renderers/linearsearch_lesson.go`, `web/renderers/linearsearch_lesson_test.go`.

**Existing interface:** `func Trace(initial []int, target int) []trace.Step`. Build accepts `{arr: []int, array: []int, target: *int}` or bare `[]int`; nonempty arr wins over array; target pointer admits zero. Default array `[15,28,42,8,33,91,55]`, target33; metadata shows bare array only. Empty array does not override Build default; direct Trace supports empty and returns−1. No sorting.

**Proposed interface:** `func RenderLinearSearchLessonSVG(step trace.Step, width, height int) string`.

```go
type LinearSearchLessonState struct {
    ArrayState
    Phase string // init, scan, compare, done
    Target, Index, CheckedThrough, Result int
    Compared, Equal, ResultKnown bool
}
```

**Visualization/geometry:** Equal-size value tiles, identity `linearsearch-item-<original index>`, not magnitude bars. Target badge `(32,30)`. Tile pitch76, width64 height56, `W=max(800,64+76*n)`, first tile x32 plus centering offset; y144. Index below at y224; `i` pointer triangle at y116. Comparison equation centered above active tile at y96. Checked prefix labeled `≠ target`; green outline with small `checked` caption; matched tile `found`. Never recolor unvisited suffix as checked. H=max(320,height). Not-found badge displays `result = -1`; empty state has no pointer.

**Teaching sequence / example:** `L1` target33 and unchanged array; `L2` move i to next tile; `L3` compare gold, show equality outcome; repeat for indices0–4; `L4` green found index4. No match ends `L7`. At failed `L3`, CheckedThrough=i and Equal=false; at pending `L2`, CheckedThrough=i−1 and Compared=false. A match at `L3` may be shown gold with equality text, then done green at `L4`. Pointer movement occurs only when i changes, not as fake value swaps.

**Required assertions:** `TestLinearSearchLessonFirstMatch` target33 result4, exactly5 comparisons; `[7,7,7]`, target7 returns0, one comparison, no checked suffix. `TestLinearSearchLessonMissAndEmpty` `[0,-2,4]` target9 returns−1 with three checked, nil array returns−1 with zero comparisons; target0 found at0. Tile widths/heights equal for negative, zero and positive values; IDs remain index-based for duplicates. Test Build `{target:0}` overrides default target without replacing default array.

- [ ] Add named trace assertions and fixed-tile/pointer renderer tests; run `go test ./internal/algorithms/linearsearch ./web/renderers` red.
- [ ] Define state/renderer and fixed geometry; no sorting/bar scaling.
- [ ] Emit cloned state with explicit compared/equal/checked/result fields at existing source IDs.
- [ ] Format changed files and run targeted command green.
- [ ] Hand parent renderer name and first-match/miss frames.

**Copy-ready agent prompt:**
```text
Implement Task 6 (Linear Search) and shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Own internal/algorithms/linearsearch/trace.go, algorithm_test.go, web/renderers/linearsearch_lesson.go, linearsearch_lesson_test.go; read lesson.go/algorithm.go/array renderer. Add LinearSearchLessonState and RenderLinearSearchLessonSVG(step trace.Step, width, height int) string. Preserve first-match/default/target-zero semantics. Render fixed identity tiles, target, i pointer, compare equation and checked prefix, not magnitude bars. Follow exact fields/source IDs/geometry/palette, immutable snapshots/escaping, named tests. Parent owns 1800ms/750ms playback and dispatch. No shared edits/deps/managers/commit. Return targeted test output/handoff/blockers.
```

## Task 7 — Binary search (`binarysearch`)

**Files:** Read `internal/algorithms/binarysearch/algorithm.go`, `internal/algorithms/binarysearch/lesson.go`, `web/renderers/array.go`. Modify `internal/algorithms/binarysearch/trace.go`, `internal/algorithms/binarysearch/algorithm_test.go`. Create `web/renderers/binarysearch_lesson.go`, `web/renderers/binarysearch_lesson_test.go`.

**Existing interface:** `func Trace(initial []int, target int) []trace.Step`. Build object `{array: []int, arr: []int, target: *int}` or bare `[]int`; nonempty array wins over arr. Default `[3,8,14,22,35,47,59,71,88]`, target47. Empty Build array keeps default; direct Trace handles empty. Neither Build nor Trace checks sortedness or sorts. Duplicates may return any matching index dictated by midpoint, not first occurrence.

**Proposed interface:** `func RenderBinarySearchLessonSVG(step trace.Step, width, height int) string`.

```go
type BinarySearchLessonState struct {
    ArrayState
    Phase string // init, bounds, midpoint, compare, discard, done
    Target, Low, High, Mid, Result int
    PreviousLow, PreviousHigh int
    DiscardLow, DiscardHigh int // inclusive discarded interval; -1 means absent
    SortedInput, ResultKnown bool
}
```

**Visualization/geometry:** Fixed identity tiles, pitch76/width64/height56, y168; `W=max(800,64+76*n)`, H at least360. Use original-index entity IDs `binarysearch-item-<i>`. Low label y124 and high label y272 keep coincident bounds readable; midpoint arrow y148. Active inclusive bracket below tiles y248; discarded regions keep labels at reduced emphasis with `excluded` hatch rather than disappear. Midpoint equation at `(32,64)`, target at `(32,28)`. Bounds crossing low>high shows `empty interval`, no negative-width bracket; out-of-range low=n/high=−1 uses boundary marker outside tile row. Explicit sorted precondition badge uses `slices.IsSorted`; unsorted inputs show `sorted input required` while preserving existing trace/output and never claiming correct exclusion proof.

**Teaching sequence / example:** `L1` target47, `L2` [0,8], `L3` nonempty, `L4` mid4/value35, `L5` equality comparison, `L7` less-than, `L8` low5 discards[0,4]; next mid6/value59, `L9` greater-than, `L10` high5 discards[6,8]; mid5/value47, `L6` found index5. Failure `L3` false then `L13` −1. Set Mid=−1 after bound update while retaining previous bound fields for transition. Do not call excluded indices individually compared.

**Required assertions:** `TestBinarySearchLessonBounds` default intervals `[0,8]`, `[5,8]`, `[5,5]`, mids4,6,5, result5, three comparisons; discard sets exactly[0,4] and[6,8], values never rearranged. `TestBinarySearchLessonBoundaryCases` nil, singleton hit/miss, below-min/above-max miss, negative sorted values, duplicates `[2,2,2]` returns1. Every active midpoint lies within bounds; bracket omitted on crossed bounds; `SortedInput=false` for `[3,1,2]` without input mutation. Build `{target:0}` honored; result parity with native Search for sorted fixtures.

- [ ] Add named bounds/edge tests and renderer crossing-bound tests; run `go test ./internal/algorithms/binarysearch ./web/renderers` red.
- [ ] Define state/renderer with honest sortedness badge, fixed tiles and separate bound labels.
- [ ] Emit explicit previous/current/discard bounds with cloned values at actual source IDs.
- [ ] Format changed files and run targeted command green.
- [ ] Hand parent interval contraction fixture and renderer API.

**Copy-ready agent prompt:**
```text
Implement Task 7 (Binary Search) and shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Own internal/algorithms/binarysearch/trace.go, algorithm_test.go, web/renderers/binarysearch_lesson.go, binarysearch_lesson_test.go; read lesson/algorithm/array renderer first. Add BinarySearchLessonState and RenderBinarySearchLessonSVG(step trace.Step, width, height int) string. Preserve midpoint/duplicate/default semantics; never sort. Teach low/high/mid, inclusive active range, exact discarded intervals, crossed bounds, truthful sorted-input precondition. Use specified source IDs/geometry/palette, stable item IDs, cloned state, escaping and tests. Parent owns playback/dispatch 1800ms/750ms. No shared edits/deps/managers/commit. Return tests, handoff, blockers.
```

## Task 8 — KMP (`kmp`)

**Files:** Read `internal/algorithms/kmp/algorithm.go`, `internal/algorithms/kmp/lesson.go`, `web/renderers/table.go`. Modify `internal/algorithms/kmp/trace.go`, `internal/algorithms/kmp/algorithm_test.go`. Create `web/renderers/kmp_lesson.go`, `web/renderers/kmp_lesson_test.go`.

**Existing interface:** `func Trace(text, pattern string) []trace.Step`. Build `LessonInput{Text string json:"text"; Pattern string json:"pattern"}` requires both nonempty to override; also accepts `[]string` length≥2 using first two entries, including empty strings. Default text `"ABABDABACDABABCABAB"`, pattern `"ABABCABAB"`. Direct Trace empty pattern returns0, text shorter than pattern returns−1. Native Search and ComputeLPS use **bytes**, not runes. Returned index is byte offset.

**Proposed interface:** `func RenderKMPLessonSVG(step trace.Step, width, height int) string`.

```go
type KMPLessonState struct {
    Text, Pattern string
    LPS []int
    Phase string // init, lps, compare, advance, mismatch, fallback, done
    I, J, Shift int
    LPSIndex, PrefixLength, FallbackFrom, FallbackTo int
    LPSKnownThrough int
    MatchIndex int
    ResultKnown bool
}
```

Do not embed generic TableState; pattern alignment and LPS construction need separate coordinates. State strings immutable; clone LPS at every construction update. Byte labels: printable ASCII as escaped glyph, space `SP`, nonprintable/non-ASCII bytes as uppercase `0xNN`. Thus UTF-8 is not falsely rendered as one character per byte or converted into mojibake. Title says `byte offsets`; algorithm semantics unchanged.

**Visualization/geometry:** Text cells at y96, pattern at y184; each pitch48, width44 height44. Text byte i x=`32+48*i`; pattern byte j x=`32+48*(Shift+j)`. LPS is fixed row at y300, x=`32+48*j`, not shifted with pattern. `W=max(960,64+48*max(len(text),Shift+len(pattern)))`, H at least420. Stable IDs `kmp-text-<i>`, `kmp-pattern-<j>`, `kmp-lps-<j>` keep pattern identities as alignment shifts. Current text/pattern cells gold; matched prefix green underline; fallback arc from old pattern index to LPS lookup and new j; equation `j = lps[oldJ-1] = newJ`, text pointer remains fixed. LPS construction shows pattern-prefix/suffix matching brackets above fixed LPS strip, only known entries displayed. Pending cells `?`, not prefilled zeros; `lps[0]=0` starts known for nonempty pattern.

**Teaching sequence / example:**
1. `L1` input, `L2` lengths. At **existing `L3`**, emit multiple computation snapshots while running equivalent ComputeLPS logic: compare pattern bytes at scan index/prefix length, extend or follow previous LPS, assign entry. Source only lists `lps := computeLPS(pattern)`, so all internal teaching stays `L3`, with explicit helper explanation. No invented helper Source IDs. Final LPS default `[0,0,1,2,0,1,2,3,4]`.
2. `L4` i=j=0; `L5` loop; `L6` compare, `L7` advance; `L9` full match check, `L10` result10 for default.
3. Mismatch `L11`, nonzero j `L12` marks LPS[j−1], `L13` shifts pattern using typed old/new j with unchanged i. Default first mismatch i4,j4 (D versus C), lookup LPS3=2, shift from0 to2. Later fallback j2→0 keeps i4 again. j0 branch `L14`, advance i at `L15`.
4. Exhaustion `L18`, `L19` result−1. Empty pattern follows existing `L1`/`L10`, no LPS indexing. Shorter text follows current early `L1`/`L2`/`L19`; may show computed table but do not invent searched comparisons.

**Required assertions:** `TestKMPLessonLPSConstruction` final LPS equals ComputeLPS for default, `"AAAA"`→`[0,1,2,3]`, `"ABABAC"`→`[0,0,1,2,3,0]`; at each LPS assignment known prefix agrees with independent prefix/suffix check, earlier snapshots retain unknown entries. `TestKMPLessonFallback` default result10; every `L13` keeps i unchanged, sets j to prior LPS[oldJ−1], shift=i−j; LPS row coordinates never move while pattern does. `TestKMPLessonEdgeBytes` checks `("AAAAA","AAA")` result0 (first match only), `("abc","")`0, `("","a")`−1, shorter text, absent pattern, `("é猫é","猫")` byte offset2, markup-like `("x<&\"y","<&\"")` result1 with safe XML. No indexing pattern[j] when j==m or i==n; terminal arrows at end boundaries, not nonexistent cells. Compare search outputs with native Search.

- [ ] Add named LPS/fallback/byte tests and renderer row/escaping assertions; run `go test ./internal/algorithms/kmp ./web/renderers` red.
- [ ] Define state/renderer with byte labels, fixed LPS row, sliding pattern, typed fallback fields.
- [ ] Trace LPS construction at repeated `L3` using existing algorithm logic, keep search Source IDs and early-return semantics, clone LPS snapshots.
- [ ] Format files and run targeted command green.
- [ ] Hand parent fallback and LPS fixtures; identify repeated `L3` frames and byte-label behavior.

**Copy-ready agent prompt:**
```text
Implement Task 8 (KMP) and shared contracts in docs/superpowers/plans/2026-10-28-animation-trees-search.md. Read internal/algorithms/kmp/{lesson,trace,algorithm,algorithm_test}.go and table renderer. Own kmp trace.go/algorithm_test.go plus web/renderers/kmp_lesson.go and kmp_lesson_test.go. Add KMPLessonState and RenderKMPLessonSVG(step trace.Step, width, height int) string. Preserve byte-based Trace/Search, Build defaults/array input and empty-pattern semantics. Teach actual LPS construction with repeated existing L3, text/pattern comparison, fixed LPS lookup and pattern shift with unchanged i. Implement exact fields/geometry/source IDs/palette, stable byte IDs, cloned LPS and safe escaped ASCII/hex labels; run named tests. Parent owns 1800ms/750ms playback/dispatch. No shared edits/deps/managers/commit. Return test evidence, handoff and blockers.
```

## Parent integration gate and blockers

- [ ] Parent confirms stable SVG entity attribute and transition contract before wiring these eight renderer entry points. This plan supplies `data-entity-id`; rename only by coordinated agreement.
- [ ] Parent migrates any integration tests expecting old generic State types, then registers exact eight renderer functions without routing unrelated lessons through them.
- [ ] Parent runs full native tests/build, race tests for mutable trace changes, then real browser checks: first/middle/final, step backward, arbitrary seek, play/pause/play, navigation during movement, custom input, reduced motion, 320px and desktop.
- [ ] Parent checks default examples legible within existing workspace. Large logical geometry needs shared local pan/zoom or scroll policy; fit-to-view alone can shrink long/skewed examples. Child renderers must not hide nodes to force fit.

No documentation blocker. Integration dependencies: parent animation identity protocol and large-canvas viewport policy. Existing permissive input parsing and unvalidated sortedness are recorded compatibility constraints, not claims that trust-boundary validation is complete. No product files or commits belong to this documentation task.
