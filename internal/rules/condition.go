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
