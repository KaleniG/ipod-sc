package libs

import (
	"strings"
	"unicode"
)

func sanitize(name string) string {
	runes := []rune(name)
	out := make([]rune, 0, len(runes))

	for i, r := range runes {
		if !strings.ContainsRune(`<>:"/\|?*`, r) {
			out = append(out, r)
			continue
		}

		var prev, next rune
		hasPrev := i > 0
		hasNext := i < len(runes)-1

		if hasPrev {
			prev = runes[i-1]
		}
		if hasNext {
			next = runes[i+1]
		}

		// Replace with '_' only if both neighbors exist and are not spaces.
		if hasPrev && hasNext && !unicode.IsSpace(prev) && !unicode.IsSpace(next) {
			out = append(out, '_')
		}
		// Otherwise, omit the illegal character.
	}

	result := string(out)

	// Remove leading/trailing spaces and dots.
	result = strings.Trim(result, " .")

	// Collapse consecutive spaces into a single space.
	result = strings.Join(strings.Fields(result), " ")

	return result
}
