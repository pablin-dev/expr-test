package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// ConditionNode represents a recursive structure that can hold a group of conditions (AND, OR, NOT)
// or a single base condition.
type ConditionNode struct {
	And       []ConditionNode `yaml:"and,omitempty"`
	Or        []ConditionNode `yaml:"or,omitempty"`
	Not       []ConditionNode `yaml:"not,omitempty"`
	Condition *Condition      `yaml:"condition,omitempty"`
}

// Condition represents a single attribute-value comparison.
type Condition struct {
	Attribute string `yaml:"attribute" validate:"required"`
	Operator  string `yaml:"operator" validate:"required"`
	Value     any    `yaml:"value" validate:"required"` // Value can be string, number, boolean, etc.
}

// Helper function to check if a string represents a number.
func isNumber(s string) bool {
	if s == "" {
		return false
	}
	var dotSeen bool
	for i, char := range s {
		if char == '.' {
			if dotSeen {
				return false
			}
			dotSeen = true
		} else if char < '0' || char > '9' {
			if i == 0 && char == '-' { // Allow leading minus sign
				continue
			}
			return false
		}
	}
	return true
}

var unquotedAttributes = map[string]bool{
	"Currency": true,
}

// ToExpr converts the ConditionNodeConfig structure into an expr-compatible string.
func (n *ConditionNode) ToExpr() string {
	if n == nil {
		return ""
	}

	// Handle base condition
	if n.Condition != nil {
		cond := n.Condition
		// Map your custom operators to Expr/Go operators
		opMap := map[string]string{
			"eq":  "==",
			"gt":  ">",
			"lt":  "<",
			"gte": ">=",
			"lte": "<=",
		}

		op, ok := opMap[cond.Operator]
		if !ok {
			op = cond.Operator // Fallback if already converted
		}

		val := cond.Value
		// Format value: Handle Booleans, Identifiers (Constants), and Strings
		if s, ok := val.(string); ok {
			switch {
			case s == "true" || s == "false":
				// Boolean literals should not be quoted
				val = s
			case unquotedAttributes[cond.Attribute]:
				// If the attribute is 'Currency', we want: Currency == GBP
				// This allows Expr to match the type of your Go constant.
				val = s
			default:
				// Regular strings (e.g., 'Guest') must be quoted
				val = fmt.Sprintf("'%s'", s)
			}
		}

		return fmt.Sprintf("(%s %s %v)", cond.Attribute, op, val)
	}

	// Handle logical operators (AND)
	if len(n.And) > 0 {
		var parts []string
		for _, child := range n.And {
			if expr := child.ToExpr(); expr != "" {
				parts = append(parts, expr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		if len(parts) == 1 {
			return parts[0]
		}
		return "(" + strings.Join(parts, " and ") + ")"
	}

	// Handle logical operators (OR)
	if len(n.Or) > 0 {
		var parts []string
		for _, child := range n.Or {
			if expr := child.ToExpr(); expr != "" {
				parts = append(parts, expr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		if len(parts) == 1 {
			return parts[0]
		}
		return "(" + strings.Join(parts, " or ") + ")"
	}

	// Handle logical operators (NOT)
	if len(n.Not) > 0 {
		var parts []string
		for _, child := range n.Not {
			if expr := child.ToExpr(); expr != "" {
				parts = append(parts, expr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		inner := strings.Join(parts, " and ")
		return "not(" + inner + ")"
	}

	return ""
}

// // NormalizeExprString replaces common rule operators with their expression language equivalents.
// func NormalizeExprString(expr string) string {
// 	replacements := map[string]string{
// 		" eq ":  " == ",
// 		" gt ":  " > ",
// 		" lt ":  " < ",
// 		" gte ": " >= ",
// 		" lte ": " <= ",
// 	}
// 	normalizedExpr := expr
// 	for old, new := range replacements {
// 		normalizedExpr = strings.ReplaceAll(normalizedExpr, old, new)
// 	}
// 	return normalizedExpr
// }

func NormalizeExprString(expr string) string {
	// 1. First, replace the operators
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

	// 2. Fix the quoting logic
	// This Regex looks for: Attribute == Value
	// It captures the Attribute in ${1} and the Value in ${2}
	re := regexp.MustCompile(`(\w+)\s+==\s+([\w']+)`)

	return re.ReplaceAllStringFunc(normalized, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		attr := submatches[1]
		val := submatches[2]

		// If it's already quoted, or it's a boolean/number, leave it alone
		if strings.HasPrefix(val, "'") || val == "true" || val == "false" {
			return match
		}

		// If the attribute is NOT a constant (like Currency), it needs quotes
		// We use the same 'unquotedAttributes' logic here
		if !unquotedAttributes[attr] {
			return fmt.Sprintf("%s == '%s'", attr, val)
		}

		// Otherwise (like Currency == GBP), leave it as an identifier
		return match
	})
}
