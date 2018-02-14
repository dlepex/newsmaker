// Package strext provides additional string utility functions
package strext

import (
	"strings"
	"unicode"
)

// IsBlank is faster (see benchmarks) and cleaner alternative to strings.TrimSpace(s)==""
func IsBlank(s string) bool {
	if len(s) == 0 {
		return true
	}
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// SplitAndTrimSpace splits string with strings.Split and then for each substring it trims spaces.
// Empty substrings are also deleted from result.
func SplitAndTrimSpace(s, sep string) []string {
	a := strings.Split(s, sep)
	r := a[:0]
	// Here we filter slice in-place
	for _, w := range a {
		w := strings.TrimSpace(w)
		if w != "" {
			r = append(r, w)
		}
	}
	if len(r) == 0 {
		// Let gc do its job
		return nil
	}
	return r
}
