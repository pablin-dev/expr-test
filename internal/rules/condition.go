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
			"in":  "in",
		}

		op, ok := opMap[cond.Operator]
		if !ok {
			op = cond.Operator // Fallback if already converted
		}

		var formattedValue string
		// Special handling for the 'in' operator which expects a list
		if cond.Operator == "in" {
			// We expect cond.Value to be a slice (e.g., []any)
			if sliceVal, ok := cond.Value.([]any); ok {
				var elements []string
				for _, elem := range sliceVal {
					// Format each element within the slice
					switch e := elem.(type) {
					case string:
						// Quote string elements
						elements = append(elements, fmt.Sprintf("'%s'", e))
					case bool:
						// Booleans are represented as is
						elements = append(elements, fmt.Sprintf("%v", e))
					case float64, int, float32, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
						// Numeric types are represented as is
						elements = append(elements, fmt.Sprintf("%v", e))
					default:
						// Fallback for any other types in the slice
						elements = append(elements, fmt.Sprintf("%v", elem))
					}
				}
				formattedValue = fmt.Sprintf("[%s]", strings.Join(elements, ", "))
			} else {
				// If cond.Value is not a slice for 'in' operator, represent it as is (or log an error/warning)
				// For now, just use %v, but ideally this should be an error.
				formattedValue = fmt.Sprintf("%v", cond.Value)
			}
		} else {
			// Standard value formatting for other operators (eq, gt, lt, etc.)
			val := cond.Value
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
			formattedValue = fmt.Sprintf("%v", val)
		}

		return fmt.Sprintf("(%s %s %s)", cond.Attribute, op, formattedValue)
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
