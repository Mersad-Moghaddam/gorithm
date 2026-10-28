package renderers

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestCoinChangeScene(t *testing.T) {
	s := CoinChangeLessonState{Coins: []int{1, 3, 4}, Amount: 6, DP: []int{0, 1, 2, 1, 1, 2, 3}, Initialized: true, Phase: "compare", Index: 6, CoinIndex: 1, SourceIndex: 3, CanTake: true, HasCandidate: true, Previous: 3, Candidate: 2}
	for _, width := range []int{0, 320, 390, 768, 1440} {
		svg := RenderCoinChangeLessonSVG(trace.Step{State: s}, width, 450)
		for _, token := range []string{`data-lesson="coinchange"`, `data-scene-width=`, `data-scene-height=`, `data-min-width=`, `data-entity="coin:1"`, `data-entity="candidate"`, `data-from-entity="coin:1"`, `dp[3] + 1 = 2`, `Current best: 3`, `#ffe29a`} {
			if !strings.Contains(svg, token) { t.Errorf("width %d missing %s", width, token) }
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		for { _, err := d.Token(); if err == io.EOF { break }; if err != nil { t.Fatal(err) } }
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "+Inf") || strings.Contains(svg, "<svg") { t.Fatal("invalid fragment") }
	}
	s.Phase = `<script>`
	if svg := RenderCoinChangeLessonSVG(trace.Step{State: s}, 320, 0); strings.Contains(svg, "<script>") || !strings.Contains(svg, "&lt;script&gt;") { t.Fatal("phase not escaped") }
	s.Phase, s.CanTake, s.HasCandidate = "compare", false, false
	if svg := RenderCoinChangeLessonSVG(trace.Step{State: s}, 320, 450); !strings.Contains(svg, "does not fit") || strings.Contains(svg, `data-entity="candidate"`) { t.Fatal("oversized coin fabricated candidate") }
	s.CanTake = true
	if svg := RenderCoinChangeLessonSVG(trace.Step{State: s}, 320, 450); !strings.Contains(svg, "unreachable") { t.Fatal("missing infinity evidence") }
}
