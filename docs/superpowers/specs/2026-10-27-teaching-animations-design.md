# Teaching animations for all 40 algorithms

## Approved direction

User approved interactive teaching mockups on 2026-10-27: bubble-sort objects move with their values, BFS exposes its queue, gradient descent plots data and its changing fitted line. Audience: curious beginners learning through playful visual data. Success: learners can explain what changed, why the algorithm chose it, and how the result follows.

## Scope and constraints

- Review and improve all 40 registered algorithms, with one dedicated subagent per algorithm as requested.
- Retain Go/WASM, deterministic trace snapshots, existing controls, source synchronization, and SVG; add no runtime dependencies.
- Each lesson shows its goal, meaningful intermediate decisions, changing data, and a truthful final result. Narration explains concrete values rather than merely naming operations.
- Shared renderers provide consistent visual vocabulary; use algorithm-specific scenes where generic bars/tables obscure the concept.
- Preserve object identity during continuous adjacent steps. Seeking, replay, reverse stepping, and lesson changes must always resolve to the exact selected snapshot; cancel stale motion.
- Text labels, shape/line cues, visible focus, responsive layouts, and reduced-motion support accompany color. No autoplay on lesson entry.
- Trace/source correctness, custom-input validation, and immutable historical snapshots must survive changes.
- Do not claim convergence, shortest paths, sortedness, or success before the algorithm establishes them.

## Teaching scenes

- Sorting: moving values, compared pair, sorted region, pivot/boundaries, merge inputs, heap structure, counts or buckets and digit passes where relevant.
- Searching and strings: target, checked/excluded regions, boundaries, text/pattern alignment, and KMP fallback.
- Graphs: directed/weighted edges, current worklist, discovery order, distance updates, accepted/rejected edges, components, and final paths/forests where relevant.
- Trees: search path, insertion, rotations, prefix progression, queried intervals, and union-find parent changes.
- Backtracking/recursion: actual boards or pegs, attempted moves, conflict reasons, backtracking, and current recursive context.
- Dynamic programming: labeled subproblems, source cells and dependency arrows, concrete recurrence, chosen values, and reconstructed result where supported.
- AI: minimax choices/pruning; cluster assignments and centroids; regression points/line/errors; Q-learning environment/action/reward; perceptron decision boundary; genetic population/selection/mutation/fitness.

## Execution and delivery

Central owner implements shared contracts and playback. Each algorithm subagent owns its algorithm folder and, if needed, its uniquely named renderer/test files; shared changes return to central owner. Each agent reports teaching gap, change, correctness evidence, and limitations. Work proceeds in bounded batches to avoid conflicts.

Run Go tests, race checks where supported, WASM/server builds, plus browser checks of animation identity, cancellation, keyboard controls, reduced motion, and mobile readability. Review all 40 lessons and changes before committing and pushing current configured branch. Preserve existing untracked documents and local companion files; stage only this task's files. Disclose unavailable browser checks.

## Builder choices

Keep existing visual shell. Prefer actual movement and explicit data structures over decorative animation or new gamification. Default examples remain small enough to follow. Shared visual contracts are additive to existing trace state, not a second algorithm implementation.
