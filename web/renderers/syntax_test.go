package renderers

import (
	"strings"
	"testing"
)

func TestHighlightGoCode(t *testing.T) {
	code := `func BubbleSort(arr []int) { // sorts in place`
	highlighted := HighlightGoCode(code)

	if !strings.Contains(highlighted, `class="syn-kw"`) {
		t.Errorf("expected syn-kw in %q", highlighted)
	}
	if !strings.Contains(highlighted, `class="syn-fn"`) {
		t.Errorf("expected syn-fn in %q", highlighted)
	}
	if !strings.Contains(highlighted, `class="syn-comment"`) {
		t.Errorf("expected syn-comment in %q", highlighted)
	}
}
