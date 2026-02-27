// Package slug provides URL-safe slug generation.
package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var multiHyphen = regexp.MustCompile(`-{2,}`)

// Generate creates a URL-safe slug from a name.
// It lowercases, normalises Unicode, replaces non-alphanumeric characters with hyphens,
// and collapses consecutive hyphens.
func Generate(name string) string {
	// Normalise Unicode (NFD) and strip combining marks (accents)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalised, _, _ := transform.String(t, name)

	lower := strings.ToLower(normalised)

	var b strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	result := multiHyphen.ReplaceAllString(b.String(), "-")
	return strings.Trim(result, "-")
}
