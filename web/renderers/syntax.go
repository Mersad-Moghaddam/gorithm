package renderers

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	keywordRegex = regexp.MustCompile(`\b(break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var|nil|true|false)\b`)
	typeRegex    = regexp.MustCompile(`\b(int|int8|int16|int32|int64|uint|uint8|uint16|uint32|uint64|uintptr|float32|float64|complex64|complex128|byte|rune|string|bool|error|any)\b`)
	funcCallRegex = regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*\(`)
	numberRegex  = regexp.MustCompile(`\b(\d+)\b`)
	stringRegex  = regexp.MustCompile(`("[^"]*")`)
	commentRegex = regexp.MustCompile(`(//.*$)`)
)

// HighlightGoCode takes raw Go code and outputs HTML with syntax coloring classes.
func HighlightGoCode(code string) string {
	// First extract comments to protect them
	commentMatch := commentRegex.FindStringIndex(code)
	codePart := code
	commentPart := ""
	if commentMatch != nil {
		codePart = code[:commentMatch[0]]
		commentPart = code[commentMatch[0]:]
	}

	escaped := html.EscapeString(codePart)

	// Replace strings first
	escaped = stringRegex.ReplaceAllString(escaped, `<span class="syn-str">$1</span>`)

	// Keywords
	escaped = keywordRegex.ReplaceAllString(escaped, `<span class="syn-kw">$1</span>`)

	// Types
	escaped = typeRegex.ReplaceAllString(escaped, `<span class="syn-type">$1</span>`)

	// Function calls
	escaped = funcCallRegex.ReplaceAllString(escaped, `<span class="syn-fn">$1</span>(`)

	// Numbers
	escaped = numberRegex.ReplaceAllString(escaped, `<span class="syn-num">$1</span>`)

	if commentPart != "" {
		escaped += fmt.Sprintf(`<span class="syn-comment">%s</span>`, html.EscapeString(commentPart))
	}

	// Preserve spaces
	escaped = strings.ReplaceAll(escaped, "    ", "&nbsp;&nbsp;&nbsp;&nbsp;")
	return escaped
}
