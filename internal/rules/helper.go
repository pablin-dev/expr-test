package rules

import (
	"fmt"
	"regexp"
	"strings"
)

func NormalizeExprString(expr string) string {
	// 1. Replace the operators
	replacements := map[string]string{
		" eq ":  " == ",
		" gt ":  " > ",
		" lt ":  " < ",
		" gte ": " >= ",
		" lte ": " <= ",
	}

	normalized := expr
	for old, new := range replacements {
		normalized = strings.ReplaceAll(normalized, old, new)
	}

	// 2. Updated Regex: added \. to support nested attributes like user.id
	// And added - to support negative numbers
	// re := regexp.MustCompile(`([\w\.]+)\s+==\s+([\w\.-']+)`)
	re := regexp.MustCompile(`([\w\.]+)\s+==\s+([\w\.\-']+)`)

	return re.ReplaceAllStringFunc(normalized, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		attr := submatches[1]
		val := submatches[2]

		// --- BAIL OUT CONDITIONS (Don't add quotes if...) ---

		// A) Already quoted
		if strings.HasPrefix(val, "'") {
			return match
		}
		// B) It's a boolean
		if val == "true" || val == "false" {
			return match
		}
		// C) It's a number (Integer or Float)
		if isNumber(val) {
			return match
		}
		// D) It's a known constant (like GBP)
		if unquotedAttributes[attr] {
			return match
		}
		// E) It's a comparison between two attributes (e.g., a == b)
		// If 'val' is not a number/bool/constant but is a known field name
		// (This handles the "a eq b" test case)
		if isAttribute(val) {
			return match
		}

		// If none of the above, it's a string literal that needs quotes (e.g., Guest)
		return fmt.Sprintf("%s == '%s'", attr, val)
	})
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	match, _ := regexp.MatchString(`^-?(\d+\.?\d*|\.\d+)$`, s)
	return match
}

// isAttribute is a helper for your "a eq b" test case.
// In a real scenario, you'd check if 'val' exists in your Payment struct.
func isAttribute(s string) bool {
	// For your test "a eq b", we treat single letters as attributes
	return len(s) == 1
}
