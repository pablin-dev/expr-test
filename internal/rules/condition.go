package rules

import (
	"fmt"
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

// ToExpr converts the ConditionNodeConfig structure into an expr-compatible string.
func (n *ConditionNode) ToExpr() string {
	if n == nil {
		return ""
	}

	// Handle base condition
	if n.Condition != nil {
		cond := n.Condition
		op := cond.Operator
		switch op {
		case "eq":
			op = "=="
		case "gt":
			op = ">"
		case "lt":
			op = "<"
		case "gte":
			op = ">="
		case "lte":
			op = "<="
		}

		val := cond.Value
		// Format value: quote strings if they are not boolean literals or numbers
		if s, ok := val.(string); ok {
			if s != "true" && s != "false" {
				val = fmt.Sprintf("'%s'", s)
			}
		}
		return fmt.Sprintf("(%s %s %v)", cond.Attribute, op, val)
	}

	// Handle logical operators (AND, OR, NOT)
	if len(n.And) > 0 {
		var parts []string
		for _, child := range n.And {
			expr := child.ToExpr()
			if expr != "" {
				parts = append(parts, expr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		if len(parts) == 1 {
			return parts[0] // If only one child, return its expression directly
		}
		return "(" + strings.Join(parts, " and ") + ")"
	}

	if len(n.Or) > 0 {
		var parts []string
		for _, child := range n.Or {
			expr := child.ToExpr()
			if expr != "" {
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

	if len(n.Not) > 0 {
		var parts []string
		for _, child := range n.Not {
			expr := child.ToExpr()
			if expr != "" {
				parts = append(parts, expr)
			}
		}
		if len(parts) == 0 {
			return ""
		}
		inner := strings.Join(parts, " and ")
		return "not(" + inner + ")"
	}

	return "" // Should not reach here if struct is properly formed
}

// NormalizeExprString replaces common rule operators with their expression language equivalents.
func NormalizeExprString(expr string) string {
	replacements := map[string]string{
		" eq ":  " == ",
		" gt ":  " > ",
		" lt ":  " < ",
		" gte ": " >= ",
		" lte ": " <= ",
	}
	normalizedExpr := expr
	for old, new := range replacements {
		normalizedExpr = strings.ReplaceAll(normalizedExpr, old, new)
	}
	return normalizedExpr
}
