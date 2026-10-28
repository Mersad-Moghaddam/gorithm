# DP, backtracking, and Hanoi lesson animations

Implementation supplement to [algorithm animation root plan](2026-10-28-algorithm-animations.md). Scope: `fibonacci`, `coinchange`, `knapsack`, `lcs`, `nqueens`, `sudoku`, `hanoitower`. This document proposes implementation; no runtime changes have been made.

## Evidence and ownership

Read current `lesson.go`, `trace.go`, and `algorithm.go` for all seven packages, plus `web/renderers/table.go`, `web/renderers/hanoi.go`, `internal/trace/types.go`, and `internal/trace/trace.go`. Parent supplied Verify-tier structural discovery: project `home-mersad-Projects-gorithm`, generation `2026-10-27T07:56:32Z`, HEAD `49815a34b5cd2f77321f254c54de80d7bbe8e53e`, 226 fully paginated Build/Meta/Render/State results. Local coverage check for these 25 existing files returned `no_recorded_issue` / `metadata_match`; this is best-effort, not completeness proof. Current source fallback, rather than graph age, grounds details below.

Parent owns shared playback, motion lifecycle, dispatch, and common shell/CSS integration. Individual implementers own only their package trace payload, dedicated renderer, and focused tests. Coordinate any `lesson.go` validation change with parent; do not silently replace input behavior while adding presentation. `algorithm.go` is reference/oracle, not animation code. Send parent renderer symbol, concrete state type, supported lesson ID, and any required integration hook. Do not concurrently edit `web/wasm/player.go`, controls, themes, or shared renderer dispatch.

## Shared implementation contract

- New renderer and typed payload live together in `web/renderers/<id>_lesson.go`; new renderer tests in `web/renderers/<id>_lesson_test.go`. Exact exported names appear below. Return SVG fragments, matching existing renderer convention; parent owns outer SVG and accessible naming.
- Pure Go/WASM and SVG, standard library only. Keep existing project layout, controls, glass surfaces, theme tokens, and dark visualization canvas. No generic animation manager, dependency, fake trace, or renderer-side algorithm rerun.
- Normal-speed meaningful operation occupies **1800ms**, with **750ms** movement inside that interval, not 1800ms plus 750ms. Parent owns speed scaling, elapsed time, cancellation, reduced-motion, and stable-node/transition integration. Renderers return final snapshot geometry plus stable identities/explicit motion metadata agreed with parent. Do not rely on CSS transitions surviving whole-SVG replacement. Seeking, reverse, reset, and reduced-motion show selected snapshot immediately; no intermediate visual state can change algorithm data or source highlight.
- Identity palette: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`. Assign by stable item/index identity, not comparison outcome. Comparison is gold `#ffe29a`; final verified result gets green `#9bf0bb` outline/check. Keep labels and patterned/outlined status cues; color alone is insufficient. Computed intermediate values are not final success.
- Every phase corresponds to actual `Step.LineID` and operation. Multiple existing frames may share a line. Add metadata to current snapshots; do not invent comparisons, selected items, recursive calls, or solution paths absent from execution. New trace events, if genuinely needed, must record real execution and use actual Source IDs; default plan preserves event order and metrics.
- `trace.Builder.AddStep` stores state directly and only shallow-copies variables. Clone all mutable slices, nested rows, stack frames, and embedded `TableState` slices before recording. Arrays such as `[9][9]int` copy by value. Render only selected snapshot; no hidden state carried across calls.
- Proposed `Phase string` fields below use documented finite local values, set by trace code, never parsed from explanation text. Index fields use `-1` when inactive; booleans disambiguate absent candidate/result from valid zero. Embedded `TableState` is compatibility/debug evidence, not permission to call unsafe generic table rendering for user strings. Escape all text and attributes with `html.EscapeString`; never interpolate dataset strings raw.
- Geometry: use passed `width,height`; padding 16px, 12px section gap, top explanation/formula band about 56px, footer about 44px. Desktop can place evidence cards beside board; narrow views stack them. Keep numbers readable, with locally scrollable matrix viewport negotiated with parent rather than shrinking 30-column tables to illegible pixels. Clamp drawable dimensions for tiny/zero inputs; emit readable empty/fallback state, never NaN, Inf, negative rectangle sizes, or panic. Check 320/390/768/1440px; no document overflow.

### Common acceptance checklist (each algorithm)

- [ ] Read current local sources and unfamiliar diffs before edits; preserve algorithm order, result, metrics, inputs, and Source IDs.
- [ ] Add typed snapshot to real trace emission sites; retain existing source/variables/metrics alignment and update existing tests that assert concrete state type deliberately.
- [ ] Add dedicated pure renderer with stable object IDs, escaped labels, explicit candidate/result text, and readable responsive geometry.
- [ ] Add focused native renderer tests plus `internal/algorithms/<id>/lesson_animation_test.go` for trace/oracle parity, snapshot independence, and Source-ID membership. Tests must exercise real traces, not only hand-built snapshots.
- [ ] Give parent dispatch handoff; shared playback owner integrates. Verify 1800ms operation/750ms movement, rapid pause/resume, seeking backwards, reset, route change, final step, and reduced motion using freshly rebuilt WASM.
- [ ] Run `gofmt` on changed Go files, `go test ./internal/algorithms/<id> ./web/renderers`, then required `go test ./...` and `make build` with working Go/WASM toolchain. Run browser checks after parent integration. Do not claim browser success from native tests.

## 1. Fibonacci: two remembered numbers make next number

### Existing contract and exact files

- Read/reference: `internal/algorithms/fibonacci/lesson.go`, `internal/algorithms/fibonacci/algorithm.go` (`Compute(n int) (int, []int)`), `web/renderers/table.go`.
- Modify: `internal/algorithms/fibonacci/trace.go` — exact signature `func Trace(n int) []trace.Step`.
- Add: `web/renderers/fibonacci_lesson.go`, `web/renderers/fibonacci_lesson_test.go`, `internal/algorithms/fibonacci/lesson_animation_test.go`.
- Default `{"n": 8}`. Build accepts object with `n *int` or direct integer, including zero/negative; ordinary malformed input falls back to 8. Current JSON `null` falls through direct-int decoding to zero. No upper bound. Trace and Compute return zero for `n <= 0`.
- Export `func RenderFibonacciLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type FibonacciLessonState struct {
    TableState
    N, Index int
    Values []int
    Known []bool
    Sources [2]int // i-2, i-1; {-1,-1} outside recurrence
    Phase string // allocate, base, read, write, done
    Result int
    HasResult bool
}
```

Keep `Known[i] == false` until L9 even though current trace computes local `dp[i]` before its L8 highlight. Do not leak this future value into visible state. Copy numeric table and known mask at each emission.

### Scene and source sequence

One horizontal indexed tile rail at y≈100, 48px cells with 8px gaps; wrap into rows for modest `n`, preserve indexed connectors across rows. Below rail, two operand cards and one result card occupy equal thirds. Identity follows index modulo five. Curved gold connectors label `i-2` and `i-1`; only operand copies move to formula, originals remain stored.

| Source ID | Existing event / scene |
| --- | --- |
| L2, L3 | Nonpositive branch: condition then single `F(0)=0` return card; no `dp[1]`. |
| L5 | Allocate n+1 empty indexed slots; do not display future zeros as computed. |
| L6, L7 | Reveal base 0 and 1 separately. |
| L8 | Gold-ring two known sources, target still blank; show `F(i)=F(i-1)+F(i-2)`. |
| L9 | Move operand copies over 750ms, reveal sum in target, retain values in source tiles. |
| L11 | Final index/result check; earlier values stay ordinary stored values. |

Beginner example `n=5`: tiles become `0,1,1,2,3,5`; at i=5, sources 3 and 4 contain 2 and 3, so `2+3=5`. Default n=8 finishes at 21. Explain reuse rather than exponential recursive tree: this implementation is bottom-up DP.

### Specific checks and steps

- [ ] Trace final result equals `Compute(n)` for -1,0,1,2,5,8. n=1 emits no L8/L9 recurrence and finishes 1.
- [ ] Each L8 has both source indices less than i and known, with target unknown; L9 changes only newly computed target's value/known status. Recurrence holds at every L9.
- [ ] Earlier snapshots remain unchanged after later writes and caller mutation; renderer can render frames in reverse order identically.
- [ ] Large n layout retains readable values. Existing `int` arithmetic can overflow, especially WASM's width: parent must define accepted safe range before promising arbitrary n accuracy; do not switch to big integers silently.

### Copy-ready prompt

```text
Implement only fibonacci lesson animation from docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md section 1 and shared contract, under root plan docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/fibonacci/{lesson,trace,algorithm}.go and web/renderers/table.go. Keep Trace(n int) []trace.Step, Compute semantics, existing input behavior, L2/L3/L5/L6/L7/L8/L9/L11 events and metrics. Add FibonacciLessonState and RenderFibonacciLessonSVG(step trace.Step,width,height int) string in web/renderers/fibonacci_lesson.go; add fibonacci_lesson_test.go and internal/algorithms/fibonacci/lesson_animation_test.go. Use known-mask snapshots, two-source sum geometry, n=5 result 5 and default result 21; test nonpositive and n=1 paths, immutability, source membership, escaped SVG and small dimensions. Pure Go/SVG; no deps. Use shared identity palette, gold compare, green final check; normal operation 1800ms with 750ms movement, parent-owned timing. Do not edit playback/dispatch/shared styles. Report renderer/state handoff and overflow/input-range blocker. Run scoped tests, go test ./..., make build; report actual results. No commit.
```

## 2. Coin Change: compare one more coin against best count

### Existing contract and exact files

- Read/reference: `internal/algorithms/coinchange/lesson.go`, `internal/algorithms/coinchange/algorithm.go` (`MinCoins(coins []int, amount int) (int, []int)`), `web/renderers/table.go`.
- Modify: `internal/algorithms/coinchange/trace.go` — `func Trace(coins []int, amount int) []trace.Step`.
- Add: `web/renderers/coinchange_lesson.go`, `web/renderers/coinchange_lesson_test.go`, `internal/algorithms/coinchange/lesson_animation_test.go`.
- Default `{"coins": [1, 2, 5], "amount": 11}`. Build accepts `coins []int` when nonempty and `amount *int` independently; omitted/empty coins retain defaults, omitted amount retains 11. Bad JSON retains defaults. Negative amount returns -1. Zero/negative denominations are not validated here; negative coins can index beyond dp.
- Export `func RenderCoinChangeLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type CoinChangeLessonState struct {
    TableState
    Coins []int
    Amount, Index, CoinIndex, SourceIndex int
    DP []int // amount+1 means unreachable, not a real count
    Initialized bool
    CanTake, HasCandidate bool
    Previous, Candidate int
    Phase string // init, base, infinity, compare, update, done
    Result int
    HasResult bool
}
```

Populate `CoinIndex` directly from existing iteration (change range binding, not order); identity remains stable even for duplicate denominations. Sentinel never becomes a pile of real coins. `HasCandidate` false for oversized coin or unreachable predecessor; still preserve original comparison and arithmetic semantics.

### Scene and source sequence

Top denomination bank: labeled circles, stable colors by input index. Middle amount rail: indexed cells, `∞` for unreachable, `0` base. Bottom comparison panel: current best at left, predecessor `dp[i-c]` plus one coin at right; two labeled lanes merge only for actual update. Width 52px per cell; wrap/viewport as shared contract. Use counts, not unbounded thousands of circle nodes.

| Source ID | Scene |
| --- | --- |
| L2 | Invalid negative amount, return -1, no amount-sized allocation. |
| L1 | Input bank and empty rail. |
| L3 | Mark dp[0]=0. |
| L4 | Set positive amounts to `∞`. |
| L7 | Gold comparison: if coin>i, show `does not fit` and no predecessor arrow. Otherwise connect i-c to i and show candidate/current; unreachable source stays `∞`. |
| L8 | Only strict improvement writes target; show `min(old, candidate)=new`. A tie emits no fabricated write. |
| L12 / L13 | Unreachable -1 / verified minimum result. |

Beginner example coins `[1,3,4]`, amount 6: at amount 6, coin 1 proposes `dp[5]+1=3`, coin 3 improves to `dp[3]+1=2`, coin 4 proposes `dp[2]+1=3`, retained best 2. This demonstrates why greedy 4+1+1 is worse than 3+3. Current algorithm returns count/table, not reconstructed coin list; example decomposition must be explanatory, not a claimed traced output. Default minimum is 3 (5+5+1).

### Specific checks and steps

- [ ] Compare final result and table to `MinCoins` for `[1,3,4],6`, `[1,2,5],11`, `[2],3` (-1), amount 0 (0), negative amount (-1), direct Trace with empty coins and positive amount (-1).
- [ ] Every L8 has predecessor i-c in range and strict improvement over prewrite snapshot; no L8 on tie or too-large coin. Gold comparison never changes dp.
- [ ] Copy input coins into every snapshot or immutable trace-owned input; external caller mutation cannot alter recorded denominations.
- [ ] Duplicate positive coins preserve input order and count semantics; coin 0 cannot improve a nonnegative dp cell. Negative denomination rejection and allocation limits need explicit parent validation decision; never run unsafe negative-coin trace in renderer tests and call it supported.

### Copy-ready prompt

```text
Implement only coinchange section 2 and shared contract in docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md; root plan docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/coinchange/{lesson,trace,algorithm}.go and web/renderers/table.go. Preserve Trace(coins []int, amount int) []trace.Step, unbounded minimum-count recurrence, strict improvement, L1/L2/L3/L4/L7/L8/L12/L13 and metrics. Add CoinChangeLessonState plus RenderCoinChangeLessonSVG(step trace.Step,width,height int) string in web/renderers/coinchange_lesson.go, matching renderer test, and internal/algorithms/coinchange/lesson_animation_test.go. Show bank, amount rail, predecessor+1 versus old count; distinguish infinity/oversized/tie. Test [1,3,4],6=2, default=3, unreachable and zero/negative amounts, snapshot copies and source IDs. No fabricated coin reconstruction. Go/SVG no deps; identity palette, gold compare, green final check; parent timing 1800ms operation/750ms movement. Parent owns playback/dispatch/styles. Report negative-coin and large-allocation validation blockers; do not change semantics silently. Run focused tests, go test ./..., make build and report outcomes. No commit.
```

## 3. 0/1 Knapsack: include once or leave out

### Existing contract and exact files

- Read/reference: `internal/algorithms/knapsack/lesson.go`, `internal/algorithms/knapsack/algorithm.go` (`Solve(weights, values []int, capacity int) (int, [][]int)`), `web/renderers/table.go`.
- Modify: `internal/algorithms/knapsack/trace.go` — `func Trace(weights, values []int, capacity int) []trace.Step`.
- Add: `web/renderers/knapsack_lesson.go`, `web/renderers/knapsack_lesson_test.go`, `internal/algorithms/knapsack/lesson_animation_test.go`.
- Default `{"weights": [2, 3, 4, 5], "values": [3, 4, 5, 6], "capacity": 5}`. Custom object applies only when weights nonempty, lengths equal, and capacity supplied (`*int`); otherwise whole default retained. Direct Trace/Solve use shorter input length and clamp negative capacity to 0. Negative weights are unsafe; no bound in Build.
- Export `func RenderKnapsackLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type KnapsackLessonState struct {
    TableState
    Weights, Values []int
    DP [][]int
    Known [][]bool
    Capacity, Item, Column int // Item is DP row, 1-based
    Sources [2][2]int // exclude, include; {-1,-1} when absent
    Fits bool
    Exclude, Include, Chosen int
    Phase string // init, allocate, item, compare, write, done
    Result int
    HasResult bool
}
```

Base row known; uncomputed rows unknown. Preserve previously computed rows at L5 item-highlight: current generic helper with activeRow=-1 hides those values, but actual DP already contains them. Typed known mask fixes presentation without inventing execution. `Chosen` numeric value does not claim a selected subset or a tie-break absent from `max`.

### Scene and source sequence

Item cards display `item #`, weight, value in top rail. Matrix rows mean first i items, columns capacity c; 40–48px cells, labeled zero row and capacity zero. Side/bottom fork: blue exclude path from `(i-1,c)`, coral include path from `(i-1,c-w)` plus value. Gold outline compares both; result writes target. Both arrows originate in previous row, making 0/1 reuse rule visible. Capacity bar depicts current capacity and current item's weight, not an invented packed subset.

| Source ID | Scene |
| --- | --- |
| L1, L3 | Initialize, allocate matrix and reveal zero-item base row. |
| L5 | Select item card; preserve computed prior rows. |
| L8 | Check fit; if fits compare include/exclude, otherwise show blocked include. |
| L9 | Write max(exclude,include) from previous row. |
| L11 | Copy exclude when weight>capacity. |
| L15 | Check bottom-right optimum. |

Beginner default: first item weight 2/value 3 yields `dp[1][2]=3`; second weight 3/value 4 at capacity 5 compares exclude 3 with `dp[1][2]+4=7`; write 7. Later items cannot beat 7. Two items with weights 2+3 explain 7, but no animated subset backtracking: current trace only computes optimum table/value.

### Specific checks and steps

- [ ] Final table/result equals Solve, default=7; `[2],[3],4` returns 3, not 6 (item cannot repeat).
- [ ] Each L9 equals `max(previousRow[c], previousRow[c-w]+v)`; every source row is i-1. L11 copies above. On ties show equal score, no invented inclusion decision.
- [ ] Test capacity 0, negative capacity normalization, direct mismatched arrays, empty arrays, zero-weight positive-value item (can contribute once even at capacity 0), negative values (empty subset remains 0).
- [ ] Deep-copy rows and known masks; after later rows, earlier snapshots and previously computed row visibility remain stable.
- [ ] Parent resolves negative weights, huge capacity, and integer-overflow input policy before expanding supported inputs.

### Copy-ready prompt

```text
Implement knapsack only, section 3/shared contract of docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md and root docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/knapsack/{lesson,trace,algorithm}.go plus web/renderers/table.go. Keep Trace(weights, values []int, capacity int) []trace.Step, 0/1 previous-row recurrence, normalization and L1/L3/L5/L8/L9/L11/L15 metrics. Add KnapsackLessonState, RenderKnapsackLessonSVG(step trace.Step,width,height int) string, web/renderers/knapsack_lesson.go and knapsack_lesson_test.go, internal/algorithms/knapsack/lesson_animation_test.go. Show item weight/value cards, matrix, include/exclude fork; retain known rows during L5. Do not invent packed subset or tie decision. Test default 7, one item cannot repeat, zero weight/capacity, mismatched direct inputs, deep-copy snapshots, source IDs and SVG geometry. Use no deps, shared colors, 1800ms operation/750ms movement handled by parent. Parent owns dispatch/playback/styles. Report negative-weight/range validation blockers. Run focused tests, go test ./..., make build; report actual outcomes. No commit.
```

## 4. LCS: match diagonal, otherwise keep longer prefix

### Existing contract and exact files

- Read/reference: `internal/algorithms/lcs/lesson.go`, `internal/algorithms/lcs/algorithm.go` (`Solve(s1, s2 string) (string, [][]int)`), `web/renderers/table.go`.
- Modify: `internal/algorithms/lcs/trace.go` — `func Trace(s1, s2 string) []trace.Step`.
- Add: `web/renderers/lcs_lesson.go`, `web/renderers/lcs_lesson_test.go`, `internal/algorithms/lcs/lesson_animation_test.go`.
- Default `{"s1": "AGGTAB", "s2": "GXTXAYB"}`. Object accepts s1/s2 if either nonempty, else text/pattern if either nonempty. If object unmarshal fails, string array of length >=2 uses first two. Object with both empty strings retains defaults; `["",""]` reaches empty-pair trace. Implementation compares **bytes**, not Unicode runes.
- Export `func RenderLCSLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type LCSLessonState struct {
    TableState
    S1, S2 string
    DP [][]int
    Filled [][]bool
    Row, Col int
    From, To [2]int // explicit backtrack movement endpoints
    Sources [][2]int
    Matches [][2]int // matched original string byte indices, stable order
    Sequence string
    Phase string // allocate, compare, write, backtrack, match, up, left, done
    Match bool
}
```

Copy matrix, filled mask, source/match slices. Store full current sequence in every backtrack snapshot; variables do not currently carry it in every event. At L17 snapshot is still at matched `(bi,bj)` before decrement; `To` is `(bi-1,bj-1)`. L19/L21 current table highlight already moves but Variables i/j still hold origin; explicit endpoints resolve this without parsing prose.

### Scene and source sequence

Two character ribbons above matrix; each tile keyed by string side and **byte index**, so repeated G characters remain distinct. Matrix includes empty-prefix row/column. Formula panel compares `diagonal+1` or `max(top,left)`. Backtrack cursor traverses actual chosen cells; matching character copies enter reconstructed sequence at left because traversal prepends. Use 40px cells and local viewport for large matrices. Display non-ASCII bytes with escaped numeric `0xNN` labels rather than pretending rune indexing.

| Source ID | Scene |
| --- | --- |
| L3 | Allocate, zero boundary only known. |
| L7 | Gold-ring characters; matching highlights diagonal, mismatch top+left. |
| L8 / L10 | Write diagonal+1 / maximum. |
| L14 | Start cursor bottom-right, empty reconstruction ribbon. |
| L16 | Compare characters at cursor. |
| L17 | Prepend actual matching byte, record matched indices; diagonal movement endpoints. |
| L19 / L21 | Move up on top>=left, otherwise left. Never choose arbitrary tie direction. |
| L24 | Final sequence and length check. |

Beginner `s1="ABC", s2="AC"`: `dp[1][1]=1` for A; B/C mismatch preserves 1; `dp[3][2]=dp[2][1]+1=2` for C. Backtracking records C then prepends A, producing `AC`, length 2. Default produces `GTAB`, length 4.

### Specific checks and steps

- [ ] Match Solve's matrix and exact sequence, including tie fixture `("AB","BA")` which returns `"A"` under up-on-tie.
- [ ] Empty direct inputs, one empty, no match, identical strings, repeats `("AAA","AA")`, default and beginner examples.
- [ ] Every matched pair has equal original bytes and final increasing indices in both strings; final length equals bottom-right dp. Backtrack movement has required endpoint delta and L17 prepends exactly one byte.
- [ ] Dataset strings `<`, `>`, `&`, quotes cannot create SVG elements/attributes; parse generated fragment wrapped in SVG using `encoding/xml`. Byte semantics preserved for multibyte input, even if result is not valid UTF-8; display byte labels explicitly and report limitation rather than silently converting to runes.
- [ ] No uncomputed cell displayed as zero except base boundary; deep snapshots and reverse rendering stay independent.

### Copy-ready prompt

```text
Implement lcs only per section 4/shared contract of docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md, root docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/lcs/{lesson,trace,algorithm}.go and web/renderers/table.go. Preserve Trace(s1, s2 string) []trace.Step, byte semantics, up-on-tie and L3/L7/L8/L10/L14/L16/L17/L19/L21/L24. Add LCSLessonState and RenderLCSLessonSVG(step trace.Step,width,height int) string in web/renderers/lcs_lesson.go, lcs_lesson_test.go and internal/algorithms/lcs/lesson_animation_test.go. Implement character ribbons, DP dependency arrows, explicit backtrack From/To and actual prepend reconstruction. Test ABC/AC=AC, default GTAB, AB/BA=A, empties/repeats and oracle parity; deep-copy state, validate Source IDs, escape user labels and show byte identity honestly. No deps; shared palette and parent 1800ms operation/750ms movement. Parent owns playback/dispatch/styles. Report byte/large-input limits. Run scoped tests, go test ./..., make build; report evidence. No commit.
```

## 5. N-Queens: try, explain conflict, place, undo

### Existing contract and exact files

- Read/reference: `internal/algorithms/nqueens/lesson.go`, `internal/algorithms/nqueens/algorithm.go` (`Solve(n int) [][]int`), `web/renderers/table.go`.
- Modify: `internal/algorithms/nqueens/trace.go` — `func Trace(n int) []trace.Step`.
- Add: `web/renderers/nqueens_lesson.go`, `web/renderers/nqueens_lesson_test.go`, `internal/algorithms/nqueens/lesson_animation_test.go`.
- Default `{"n":4}`; accepts positive object n or direct positive integer; Build caps at 8, invalid/nonpositive falls back to 4. Direct Trace normalizes n<=0 to 4 but does not cap; Solve(n<=0) returns nil. Search records **all** solutions, not first only.
- Export `func RenderNQueensLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type NQueensLessonState struct {
    TableState
    N int
    Board []int // column per row, -1 empty; committed queens only
    Row, Col int
    Candidate [2]int // {-1,-1} when absent
    Conflict [2]int // first conflicting committed queen, {-1,-1} absent
    ConflictKind string // column, diagonal, or empty
    Depth, Solutions int
    Phase string // init, base-check, try, conflict, safe, place, descend, remove, solution, done
    LastSolution []int // copied only from actual L7 board; retained for labelled thumbnail
}
```

Candidate ghost must not be committed into Board. Conflict records first earlier row encountered by current loop, not invented exhaustive conflict checks. Depth follows real recursion. `LastSolution` is explicitly past solution, never substituted for live board.

### Scene and source sequence

Square chessboard, side `min(availableWidth,availableHeight)`, centered; coordinate labels, alternating neutral squares, row identities for queens. Candidate outlined queen token distinct from solid placed queen. Gold line joins candidate to recorded conflicting queen, with `same column` or `|Δcol|=Δrow` label. Side/bottom row stack shows current depth and solution count; optional single thumbnail labeled `last recorded solution`. At narrow widths board above explanatory stack.

| Source ID | Scene |
| --- | --- |
| L2 | Empty board initialized to -1. |
| L6 | Check row==n; no speculative solution. |
| L10 | Candidate ghost in next column. |
| L21 | Show first conflict ray and reject candidate. |
| L24 | Safe comparison result, ready to place. |
| L12 | Commit solid queen. |
| L13 | Descend one row. |
| L7 | Record solution, green check and increment count; this ActionDone is **not end of trace**. |
| L14 | Remove actual queen after recursive return, including after a solution. |
| L1 | Exhaustion complete; live board empty, total solutions displayed. |

Beginner n=4: row0 col0 conflicts with row1 col0 (same column) and row1 col1 (diagonal); eventually first valid board `[1,3,0,2]`, then `[2,0,3,1]`; count 2. Explain why backing out is necessary. Do not freeze first solution or repaint final live board as solved.

### Specific checks and steps

- [ ] n=1 has one solution; n=2,3 zero; n=4 exactly two L7 snapshots matching Solve order and boards. Final L1 board all -1, count 2.
- [ ] Every committed placement satisfies prior-row safety; every L21 points to existing queen with equal column or equal absolute deltas. Candidate changes alone never change Board.
- [ ] Each L14 removes exactly its row's prior queen and preserves others; historical solution snapshots/LastSolution survive later removals.
- [ ] Test Build cap separately from direct Trace semantics; no false parity assertion for nonpositive Solve versus Trace. Bound fixture sizes to avoid exponential test load.
- [ ] Parent playback stops only at trace end, not first `ActionDone`; check continuing past first L7 in browser.

### Copy-ready prompt

```text
Implement nqueens only, section 5/shared contract of docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md and root docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/nqueens/{lesson,trace,algorithm}.go and web/renderers/table.go. Preserve Trace(n int) []trace.Step, all-solution search, Build cap and L2/L6/L10/L21/L24/L12/L13/L7/L14/L1. Add NQueensLessonState, RenderNQueensLessonSVG(step trace.Step,width,height int) string in web/renderers/nqueens_lesson.go, matching renderer test and internal/algorithms/nqueens/lesson_animation_test.go. Render committed board, separate candidate, first real conflict, row depth and actual solution count. Test n=1/2/3/4, exact two n=4 solutions, valid placements, removals, snapshot isolation, empty final board and Source IDs. L7 ActionDone is intermediate; send this warning to parent playback owner. No fake solution board or stopped-first search. No deps; shared palette, parent-owned 1800ms operation/750ms movement. Do not edit shared playback/dispatch/styles. Run focused/full tests and make build; report outcomes. No commit.
```

## 6. Sudoku: show why each candidate fits or fails

### Existing contract and exact files

- Read/reference: `internal/algorithms/sudoku/lesson.go`, `internal/algorithms/sudoku/algorithm.go` (`Solve(board [9][9]int) ([9][9]int, bool)` and `IsValid`), `web/renderers/table.go`.
- Modify: `internal/algorithms/sudoku/trace.go` — `func Trace(board [9][9]int) []trace.Step`.
- Add: `web/renderers/sudoku_lesson.go`, `web/renderers/sudoku_lesson_test.go`, `internal/algorithms/sudoku/lesson_animation_test.go`.
- Default object below; Build accepts `{"board": ...}` via `*[9][9]int` or direct fixed array. Go JSON fixed-array decoding pads short arrays and discards extra elements; current schema does not enforce exactly 9×9. Malformed data generally retains CanonicalBoard; JSON null direct-array path can produce all-zero board. Trace validates givens then uses first row-major empty cell, candidates 1..9, returns first solution.

```json
{"board":[[5,3,0,0,7,8,9,1,2],[6,7,2,1,9,5,3,4,8],[0,9,8,3,4,2,5,6,7],[8,5,9,7,6,1,4,2,3],[4,2,6,8,0,3,7,9,1],[7,1,3,9,2,4,8,5,6],[9,6,1,5,3,7,0,8,4],[2,8,7,4,1,9,6,3,5],[3,4,5,2,8,6,1,0,0]]}
```

- Export `func RenderSudokuLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type SudokuLessonState struct {
    TableState
    Board [9][9]int
    Given [9][9]bool
    Row, Col, Candidate, Depth int
    HasCandidate bool
    Conflicts [][2]int
    ConflictKinds []string // parallel entries: row, column, box
    Phase string // init, select, try, reject, place, descend, remove, exhausted, done
    Status string // searching, solved, invalid, unsatisfiable
}
```

Given mask from initial nonzero cells, never current board. Derive conflict evidence at actual validity check against pre-placement board using same row/column/box rules; evidence is explanation of one check, not extra traced constraint-propagation algorithm. Deduplicate coordinates if listing all blockers, with deterministic order and combined labels, or record first blocker with one kind. Fix choice consistently in tests. Preserve `IsValid` result; no solving in renderer. Thread depth/given metadata through existing private trace recursion, not public Trace signature.

### Scene and source sequence

9×9 square, thick boundaries every third cell; at 320px use about 28–30px cells and legible 14px digits. Givens bold with lock/dot cue; tentative digits colored with distinct outline. Candidate 1..9 strip below, wrapping when needed. Gold row/column/box outlines show validation scope; conflict cell marker plus textual reason. Tiny candidate ghost does not replace Board value before L8. Compact breadcrumb shows recursion depth, not an unbounded tree.

| Source ID | Scene |
| --- | --- |
| L1 | Original board and givens. |
| L15 ActionDone | Invalid initial board, or unsatisfiable after exhaustive search; explicit Status distinguishes them. |
| L2 | First row-major empty cell selected. |
| L6 | Try next candidate, validation scopes highlighted. |
| L7 | Rejection evidence, committed board unchanged. This line is rejection-only in current trace. |
| L8 | Valid candidate committed. |
| L9 | Descend to next empty cell. |
| L12 | Undo only current guessed digit to zero. |
| L15 ActionPop | All nine candidates exhausted, not terminal success. |
| L4 | All cells filled successfully, green board check. |

Beginner default first blank `(0,2)`: try 1 (row contains 1), 2 (row contains 2), 3 (row contains 3), then 4 fits row/column/box; L8 writes 4. Next blank `(0,3)` accepts 6. Final first row `5,3,4,6,7,8,9,1,2`; default seven blanks all filled. Show numeric constraints rather than implying advanced propagation absent from code.

### Specific checks and steps

- [ ] Default final board equals Solve, has digits 1..9 once per row/column/box, and all original givens unchanged. Already solved valid board emits L1 then L4 without candidate loop.
- [ ] Each L8 changes exactly selected zero cell to valid candidate; each L12 restores guessed cell to zero, never given. L6/L7 do not mutate Board.
- [ ] Invalid duplicate givens and digits outside 0..9 yield explicit invalid L15 terminal result, not green solved state.
- [ ] Unsatisfiable but initially valid fixture: otherwise-zero board with row 0 `[1,2,3,4,5,6,7,8,0]` and `(1,8)=9`. Givens have no duplicates; first empty `(0,8)` needs 9 but column already contains 9. Solve returns false, trace ends unsatisfiable, givens retained.
- [ ] Exercise real rollback using a fixture proven to emit L12; assert fixture emits L12 rather than assuming default does. Candidate conflicts reference actual equal digits in correct row/column/box; deep-copy conflict slices.
- [ ] Parent resolves strict input dimensions and resource limits before treating arbitrary empty/hard puzzles as bounded interactive input. Do not change search order to make demonstration faster.

### Copy-ready prompt

```text
Implement sudoku only per section 6/shared contract of docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md and root docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/sudoku/{lesson,trace,algorithm}.go and web/renderers/table.go. Preserve Trace(board [9][9]int) []trace.Step, row-major first empty, candidate order 1..9, first solution and L1/L2/L6/L7/L8/L9/L12/L15/L4. Add SudokuLessonState, RenderSudokuLessonSVG(step trace.Step,width,height int) string in web/renderers/sudoku_lesson.go, matching renderer test and internal/algorithms/sudoku/lesson_animation_test.go. Show fixed givens, candidate ghost, real conflict evidence, 3x3 boundaries, actual placement/rollback and explicit solved/invalid/unsatisfiable status. Test canonical first placement 4 at (0,2), final oracle board, givens, solved input, invalid givens, unsatisfiable otherwise-zero board with row0=[1,2,3,4,5,6,7,8,0] and (1,8)=9, real L12 fixture, source membership and snapshot independence. No simulated propagation, no altered search. Pure Go/SVG no deps, shared palette; parent timing 1800ms operation/750ms movement. Parent owns playback/dispatch/styles. Report fixed-array JSON dimension/resource limitations. Run focused tests, go test ./..., make build; report actual results. No commit.
```

## 7. Tower of Hanoi: recursive promise, one legal disk move

### Existing contract and exact files

- Read/reference: `internal/algorithms/hanoitower/lesson.go`, `internal/algorithms/hanoitower/algorithm.go` (`Solve(n int) []Move`, `Move{Disk,From,To int}`), `web/renderers/hanoi.go`.
- Modify: `internal/algorithms/hanoitower/trace.go` — `func Trace(n int) []trace.Step`.
- Add: `web/renderers/hanoitower_lesson.go`, `web/renderers/hanoitower_lesson_test.go`, `internal/algorithms/hanoitower/lesson_animation_test.go`.
- Default direct JSON `3`. Build accepts direct integer 3..5 or object aliases `disks`, `n`, `count` in that precedence when in range; otherwise 3. Direct Trace accepts n>=1, normalizes n<1 to 3; Solve(n<=0) returns nil. Existing HanoiState pegs are bottom-to-top slices, already cloned.
- Export `func RenderHanoiTowerLessonSVG(step trace.Step, width, height int) string`.

### Proposed payload

```go
type HanoiTowerCall struct { N, From, To, Aux int }
type HanoiTowerLessonState struct {
    HanoiState
    TotalDisks, MoveCount int
    Calls []HanoiTowerCall
    Phase string // init, base-check, return, left-call, move, right-call, done
    Disk, From, To int // indices for actual move; -1 when inactive
    BeforePegs [3][]int // motion origin for L6 only; cloned
}
```

Record actual active call stack during recursion, including k=0 base calls. L5/L7 show parent plus pending subcall label, not falsely claim child has entered before next L2. Pop real call stack on return without adding fake execution events. L6 snapshots are already post-move; `BeforePegs` captures legal source/target positions before mutation. Renderer can output authoritative final frame alone; parent motion uses explicit before/after, never reverse-engineers `Move` prose.

### Scene and source sequence

Three peg centers at 1/6, 1/2, 5/6 of padded width; baseline near 70% scene height, not existing hardcoded x=180/400/620 and y=360. Disk width increases monotonically with rank within peg lane, height 22–28px depending on available height. Stable disk IDs/rank colors; labels A/B/C and explicit `from/to/aux` badges change with active call, not peg identity. Stack cards below or beside pegs list `(n,from,to,aux)` and current subproblem. Motion L6 lifts top disk above tallest peg, translates horizontally, then lowers to destination over total 750ms; no disk passes through another disk.

| Source ID | Scene |
| --- | --- |
| L1 | All disks on A, target C, auxiliary B. |
| L2 | Actual call frame, check k==0. |
| L3 | Base return; no disk moves. |
| L5 | Explain pending move of k-1 from source to auxiliary; next L2 enters it. |
| L6 | Exactly one legal move, post-move pegs authoritative, move counter increments once. |
| L7 | Pending move k-1 from auxiliary to target. |
| L8 | All disks C, `2^n-1` actual moves checked. |

Beginner n=3, disk numbers smallest=1: `1 A→C`, `2 A→B`, `1 C→B`, `3 A→C`, `1 B→A`, `2 B→C`, `1 A→C`. Seven physical moves; recursion bookkeeping frames are not moves. The central disk-3 move is only safe after first three moves relocate smaller disks.

### Specific checks and steps

- [ ] Collect only L6 events and compare `(Disk,From,To)` sequence exactly with Solve(n) for n=1..5. Counts 1,3,7,15,31; final A/B empty and C `[n,...,1]`.
- [ ] Each move removes top source only, adds same disk to destination top, and never places larger on smaller; disk multiset conserved in every frame.
- [ ] BeforePegs and postmove Pegs differ by exactly that move; all non-L6 events leave physical board unchanged relative to preceding physical board.
- [ ] Stack call role permutations match L5 and L7 recursion; k=0 does not increment MoveCount. All arrays/slices deep-copied, including BeforePegs and Calls.
- [ ] Test Build alias precedence and out-of-range fallback separately; do not compare normalized nonpositive Trace to nil Solve. Parent must bound direct large n before allocating exponential trace.
- [ ] Parent verifies lift/translate/lower cancellation, reverse seek to adjacent move and mid-call, final static disk order, reduced-motion immediate landing. No autonomous SVG loop.

### Copy-ready prompt

```text
Implement hanoitower only per section 7/shared contract of docs/superpowers/plans/2026-10-28-animation-dp-puzzles.md and root docs/superpowers/plans/2026-10-28-algorithm-animations.md. Read internal/algorithms/hanoitower/{lesson,trace,algorithm}.go and web/renderers/hanoi.go. Preserve Trace(n int) []trace.Step, Build 3..5/aliases, recursive order and L1/L2/L3/L5/L6/L7/L8. Add HanoiTowerLessonState/HanoiTowerCall and RenderHanoiTowerLessonSVG(step trace.Step,width,height int) string in web/renderers/hanoitower_lesson.go, matching renderer test and internal/algorithms/hanoitower/lesson_animation_test.go. Embed HanoiState, clone before/after pegs, capture actual calls and move count. Render responsive three pegs, rank identity, role badges and real call stack; provide before/after lift-translate-lower geometry for parent 750ms movement within 1800ms operation. Test exact Solve moves for n=1..5, 7 moves for n=3, stack roles, legal moves, immutable snapshots, source IDs and small dimensions. No fake moves, no deps, no changed search semantics. Parent owns timing/dispatch/styles; hand off metadata and renderer. Run focused tests, go test ./..., make build and report outcomes. No commit.
```

## Integration gates and known limits

- [ ] Parent accepts seven exact renderer/state symbols and confirms dispatch based on lesson ID rather than generic `table`/`hanoi` alone. Embedded state does not satisfy old Go type assertions automatically; migrate affected assertions/dispatch deliberately.
- [ ] Parent confirms selected source line, explanation, variables, metrics, scene and scrubber stay on same authoritative Step during animation and cancellation. N-Queens intermediate ActionDone cannot end playback.
- [ ] Implement resolved lesson-input policy below in each package's `Lesson.Build`, with explicit boundary tests; no separate validation framework. These are browser lesson bounds, not changes to pure algorithm definitions.
- [ ] Browser acceptance covers library-to-lesson navigation, custom input during play, reverse/seek, reset, final steps, 320px overflow, every theme's label contrast, and reduced motion. Theme review includes all eight current themes. Keep global styling edits with assigned owner.
- [ ] Documentation verification: source signatures/defaults/events checked against files above; run `git diff --check`. This documentation task does not run or claim implementation, native tests, build, or browser checks.

### Resolved validation and ownership decisions

These decisions supersede earlier instructions to ask parent for bounds. Include this section in every individual worker prompt; workers own their `lesson.go` and `lesson_animation_test.go` too. Empty raw input selects defaults; malformed explicit JSON, null, or wrong field types returns error. Preserve supported alternate forms. Reject invalid input before Trace, retaining current dataset in dialog. Test lower/upper accepted boundaries and one beyond each bound.

- Fibonacci: explicit n in [0,46], because F(46)=1836311903 fits signed 32-bit WASM int, F(47) does not. Direct pure Trace nonpositive behavior remains as documented; Build rejects negative n.
- Coin change: amount [0,200], 1–16 coins, each integer [1,200]. Preserve duplicate denominations and input order. Reject negative/zero coins; direct Trace negative-amount tests remain unchanged.
- Knapsack: 1–20 items, equal-length weights/values, weights [0,200], values [0,1000000], capacity [0,200]. Zero weight remains valid. Reject negative weights, mismatched explicit arrays, and negative capacity before allocation; direct Trace normalization tests still document pure API behavior.
- LCS: each input at most 32 bytes; preserve existing byte-based algorithm. Accept valid UTF-8 but label compared units as bytes and render non-ASCII byte values as `0xNN`; never show broken UTF-8 glyphs or claim rune-level LCS. Test ASCII reference and multi-byte input against byte oracle. Escape text.
- N-Queens: explicit n in [1,8]; no silent clamp for out-of-range user input. Intermediate solution `ActionDone` remains a recordable solution, not end-of-playback signal; only final index ends playback.
- Sudoku: decode dynamic rows first; require exactly nine rows and nine integer entries per row, each [0,9], and no conflicting nonzero givens in any row/column/box. Convert to `[9][9]int` only after validation. Retain real unsatisfiable search outcome for valid givens. Test short/long rows, fractional entries, duplicate givens, solved board, and valid unsatisfiable fixture. Conflict witness is deterministic first blocker: row in increasing column, then column in increasing row, then box row-major; record one witness/kind, not variable all-witness enumeration.
- Hanoi: explicit disk count [1,8]; preserve existing valid peg-name/alternate forms and reject invalid peg references. Direct zero-disk Trace remains testable as empty solution; public lesson Build rejects zero. Maximum legal sequence 255 moves bounds browser trace.
