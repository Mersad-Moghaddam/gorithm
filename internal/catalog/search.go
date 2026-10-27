package catalog

import "strings"

// Matches finds every query word in lesson metadata, optionally within a category.
func Matches(meta Meta, query string, category Category) bool {
	if category != "" && meta.Category != category {
		return false
	}
	text := strings.ToLower(strings.Join([]string{meta.ID, meta.Name, string(meta.Category), meta.Documentation.Summary, strings.Join(meta.Documentation.KeyConcepts, " ")}, " "))
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
}
