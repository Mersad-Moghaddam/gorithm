# Gorithm living lab redesign

## Intent

Rebuild the frontend presentation and navigation from scratch so visitors feel interested and can quickly find the algorithm they need. User selected the Living Lab mockup: dark green surfaces, lime accents, bold typography, immersive presentation.

## Technical decisions

Builder choice: retain Go/WASM, the existing algorithm registry, deterministic traces, and SVG renderers. Use semantic HTML and CSS for the interface, Go/WASM for interaction, and existing dependencies. No framework migration or external search service is needed. Algorithm implementations and educational content remain the source of truth.

## Discovery screen

- Open on an algorithm library rather than an automatically selected lesson.
- Compact Gorithm header with library navigation and GitHub link.
- Bold introduction: “Make complex click.” Supporting copy explains finding algorithms and watching them work.
- Prominent labeled search, with examples of names and concepts. Search is local and case-insensitive across names, IDs, categories, summaries, and key concepts. All query words must match the searchable text; this is keyword search, not natural-language or semantic search.
- Category filters populated from actual registry categories; every registered algorithm must be reachable. Include an All filter. Search and category selection combine.
- Show an accurate result count. Empty results explain how to recover and offer a clear reset action.
- Responsive algorithm cards show name, category, short existing summary, time complexity, and an explicit open action. Use small category illustrations as visual cues without presenting them as live executions.
- Preserve query and category while viewing a lesson and returning to the library.
- Support direct lesson links using a URL fragment containing the registered algorithm ID. Browser back/forward restores library or lesson views. Unknown IDs recover to the library with a readable message.
- Search shortcut `/` focuses search outside editable controls. All functions also remain available through visible controls.

## Learning workspace

- Open a selected lesson in a dedicated workspace; clearly labeled back-to-library action restores discovery context.
- Show name, category, existing summary, and time/space complexity above the visualization.
- Prioritize the SVG canvas and concise step narration. Place playback controls and timeline directly below the canvas, without covering visual content.
- Preserve play/pause, previous/next, restart, speed, timeline seeking, shuffle, custom datasets, presets, metrics, variable snapshots, source highlighting, and documentation.
- Code and Learn use accessible tab controls. Desktop layout pairs the canvas with an inspector; narrow layouts stack these panels without squeezing code or controls offscreen.
- Custom input uses accessible dialog semantics, contained keyboard focus, Escape dismissal, focus restoration, and readable validation feedback. Existing dataset validation stays intact.
- Disable controls when runtime or lesson data is unavailable. Loading and failure states communicate real status and allow reload after a loading failure.

## Visual system and motion

- Ink-green backgrounds, slightly lighter green panels, pale text, muted readable secondary text, and lime primary actions. Use amber/error accents only for meaningful states.
- Bold, tightly set display typography with readable body text and monospace code; retain available fonts with system fallbacks rather than adding dependencies.
- Clear borders and generous spacing establish hierarchy. Avoid crowded pills, decorative status lights, and competing glow effects.
- Short CSS entrance and hover transitions make navigation feel responsive. Never delay search or input for animation.
- Preserve deterministic alignment between visualization, highlighted source, and step narration. Animate supported state changes without interpolating misleading execution states. Renderer DOM strategy must be inspected before promising object movement across frames.
- Honor `prefers-reduced-motion`; remove decorative and transition motion while retaining explicit step playback. No automatic playback on discovery or lesson entry.

## Accessibility and responsive behavior

- Keyboard-operable links, buttons, filters, tabs, playback, and dialog, with visible focus and descriptive accessible names.
- Active categories and selected tabs expose their state. Search result count announces changes politely; per-frame playback updates do not flood live regions.
- Manage focus when switching library/workspace, returning to the originating card where possible. Inputs must not trigger global playback shortcuts.
- Maintain readable contrast, comfortable touch targets, and usable layout at mobile, tablet, and desktop widths. Keep code scrolling local rather than overflowing the page.

## Scope boundaries

This is a complete frontend redesign of discovery and learning flows, not a new algorithm collection. No accounts, favorites, AI recommendations, new backend, or new animation framework. Do not invent algorithm metadata such as difficulty ratings or unsupported complexity guarantees.

## Verification and delivery

- Check catalog coverage, keyword matching, combined filters, no-results recovery, direct links, back/forward, and preserved discovery state.
- Check playback, source alignment, custom input errors, tabs, keyboard navigation, reduced motion, loading failure, and mobile layout.
- Run existing Go tests, compile WASM and server, and leave a runnable regression check for new non-trivial search/navigation behavior.
- Inspect rendered desktop and mobile screens when browser tooling is available; disclose any verification limitation.
- Review changes, commit approved implementation, and push the current branch to its configured remote, as requested. Exclude local browser-companion files from commits.

## Approval

Visual direction and full scope approved by user on 2026-10-26.
