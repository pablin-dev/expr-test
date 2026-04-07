package rules

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestConditionValidation(t *testing.T) {
	v := validator.New()

	tests := []struct {
		name      string
		condition Condition
		wantErr   bool
	}{
		{
			name: "Valid Condition",
			condition: Condition{
				Attribute: "user.age",
				Operator:  "gt",
				Value:     30,
			},
			wantErr: false,
		},
		{
			name: "Missing Attribute",
			condition: Condition{
				Operator: "eq",
				Value:    "active",
			},
			wantErr: true,
		},
		{
			name: "Missing Operator",
			condition: Condition{
				Attribute: "user.status",
				Value:     "active",
			},
			wantErr: true,
		},
		{
			name: "Missing Value", // Value is `any`, so nil is a possibility
			condition: Condition{
				Attribute: "user.email",
				Operator:  "eq",
				// Value is nil by default
			},
			wantErr: true,
		},
		{
			name: "Value as string",
			condition: Condition{
				Attribute: "user.name",
				Operator:  "eq",
				Value:     "Alice",
			},
			wantErr: false,
		},
		{
			name: "Value as boolean",
			condition: Condition{
				Attribute: "user.isAdmin",
				Operator:  "eq",
				Value:     true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("Condition validation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"Positive integer", "123", true},
		{"Negative integer", "-45", true},
		{"Zero", "0", true},
		{"Positive float", "1.23", true},
		{"Negative float", "-1.23", true},
		{"Float ending with dot", "123.", true},
		{"Float starting with dot", ".123", true},
		{"Empty string", "", false},
		{"Alphabetical string", "abc", false},
		{"Alphanumeric", "1a2", false},
		{"Multiple dots", "1.2.3", false},
		{"Multiple minus", "--1", false},
		{"Positive sign", "+1", false}, // Current isNumber doesn't handle '+', so this is false
		{"String ending with minus", "10-", false},
		{"Float with exponent", "1e3", false}, // Not handled by current isNumber
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumber(tt.s); got != tt.want {
				t.Errorf("isNumber(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestConditionNodeToExpr(t *testing.T) {
	tests := []struct {
		name     string
		node     ConditionNode
		expected string
	}{
		// Base Condition Tests
		{
			name: "Base Condition - Greater Than Integer",
			node: ConditionNode{
				Condition: &Condition{Attribute: "age", Operator: "gt", Value: 30},
			},
			expected: "(age > 30)",
		},
		{
			name: "Base Condition - Equal String",
			node: ConditionNode{
				Condition: &Condition{Attribute: "name", Operator: "eq", Value: "Alice"},
			},
			expected: "(name == 'Alice')",
		},
		{
			name: "Base Condition - Equal Boolean True",
			node: ConditionNode{
				Condition: &Condition{Attribute: "isActive", Operator: "eq", Value: true},
			},
			expected: "(isActive == true)",
		},
		{
			name: "Base Condition - Less Than Integer",
			node: ConditionNode{
				Condition: &Condition{Attribute: "count", Operator: "lt", Value: 10},
			},
			expected: "(count < 10)",
		},
		{
			name: "Base Condition - Greater Than Equal Float",
			node: ConditionNode{
				Condition: &Condition{Attribute: "score", Operator: "gte", Value: 90.5},
			},
			expected: "(score >= 90.5)",
		},
		{
			name: "Base Condition - Equal Boolean False",
			node: ConditionNode{
				Condition: &Condition{Attribute: "hasDiscount", Operator: "eq", Value: false},
			},
			expected: "(hasDiscount == false)",
		},
		{
			name: "Base Condition - Less Than Equal Number String",
			node: ConditionNode{
				Condition: &Condition{Attribute: "version", Operator: "lte", Value: "1.0"},
			},
			expected: "(version <= '1.0')",
		},

		// Logical Operator Tests
		{
			name: "AND - Two Conditions",
			node: ConditionNode{
				And: []ConditionNode{
					{Condition: &Condition{Attribute: "age", Operator: "gt", Value: 30}},
					{Condition: &Condition{Attribute: "city", Operator: "eq", Value: "NY"}},
				},
			},
			expected: "((age > 30) and (city == 'NY'))",
		},
		{
			name: "OR - Two Conditions",
			node: ConditionNode{
				Or: []ConditionNode{
					{Condition: &Condition{Attribute: "status", Operator: "eq", Value: "active"}},
					{Condition: &Condition{Attribute: "status", Operator: "eq", Value: "pending"}},
				},
			},
			expected: "((status == 'active') or (status == 'pending'))",
		},
		{
			name: "NOT - One Condition",
			node: ConditionNode{
				Not: []ConditionNode{
					{Condition: &Condition{Attribute: "isExpired", Operator: "eq", Value: true}},
				},
			},
			expected: "not((isExpired == true))",
		},
		{
			name: "Nested AND",
			node: ConditionNode{
				And: []ConditionNode{
					{
						And: []ConditionNode{
							{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
							{Condition: &Condition{Attribute: "b", Operator: "eq", Value: 2}},
						},
					},
					{Condition: &Condition{Attribute: "c", Operator: "eq", Value: 3}},
				},
			},
			expected: "(((a == 1) and (b == 2)) and (c == 3))",
		},
		{
			name: "Nested OR",
			node: ConditionNode{
				Or: []ConditionNode{
					{
						Or: []ConditionNode{
							{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
							{Condition: &Condition{Attribute: "b", Operator: "eq", Value: 2}},
						},
					},
					{Condition: &Condition{Attribute: "c", Operator: "eq", Value: 3}},
				},
			},
			expected: "(((a == 1) or (b == 2)) or (c == 3))",
		},
		{
			name: "Mixed AND/OR",
			node: ConditionNode{
				And: []ConditionNode{
					{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
					{
						Or: []ConditionNode{
							{Condition: &Condition{Attribute: "b", Operator: "eq", Value: 2}},
							{Condition: &Condition{Attribute: "c", Operator: "eq", Value: 3}},
						},
					},
				},
			},
			expected: "((a == 1) and ((b == 2) or (c == 3)))",
		},
		{
			name: "NOT with AND",
			node: ConditionNode{
				Not: []ConditionNode{
					{
						And: []ConditionNode{
							{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
							{Condition: &Condition{Attribute: "b", Operator: "eq", Value: 2}},
						},
					},
				},
			},
			expected: "not(((a == 1) and (b == 2)))",
		},
		{
			name:     "Empty ConditionNode",
			node:     ConditionNode{},
			expected: "",
		},
		{
			name:     "Nil ConditionNode",
			node:     ConditionNode{}, // Representing a nil ConditionNode passed by value
			expected: "",
		},
		{
			name: "ConditionNode with nil Condition and empty logical ops",
			node: ConditionNode{
				Condition: nil,
				And:       []ConditionNode{},
				Or:        []ConditionNode{},
				Not:       []ConditionNode{},
			},
			expected: "",
		},
		{
			name: "ConditionNode with single OR clause",
			node: ConditionNode{
				Or: []ConditionNode{
					{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
				},
			},
			expected: "(a == 1)", // Should not wrap single OR clause in parentheses
		},
		{
			name: "ConditionNode with single AND clause",
			node: ConditionNode{
				And: []ConditionNode{
					{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
				},
			},
			expected: "(a == 1)", // Should not wrap single AND clause in parentheses
		},
		{
			name: "ConditionNode with single NOT clause",
			node: ConditionNode{
				Not: []ConditionNode{
					{Condition: &Condition{Attribute: "a", Operator: "eq", Value: 1}},
				},
			},
			expected: "not((a == 1))", // NOT wraps its content in parentheses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.ToExpr(); got != tt.expected {
				t.Errorf("ConditionNode.ToExpr() for %s = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestNormalizeExprString(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"eq replacement", "age eq 30", "age == 30"},
		{"gt replacement", "status gt 100", "status > 100"},
		{"lt replacement", "count lt 5", "count < 5"},
		{"gte replacement", "score gte 90", "score >= 90"},
		{"lte replacement", "value lte 0", "value <= 0"},
		{"String with eq", "name eq 'Alice'", "name == 'Alice'"},
		{"Boolean with eq", "hasError eq true", "hasError == true"},
		{"Multiple replacements", "price gt 10.5 and quantity lt 5", "price > 10.5 and quantity < 5"},
		{"Number literal replacement", "1 eq 1", "1 == 1"},
		{"No replacement needed", "user.id == 123", "user.id == 123"},
		{"Adjacent operators", "a eq b eq c", "a == b == c"}, // Edge case: testing literal replacement
		{"Invalid operator", "invalid_op 5", "invalid_op 5"},
		{"Leading/Trailing spaces", "  age eq 30  ", "  age == 30  "},
		{"Mixed operators and values", "active_users gte 10 and inactive_users lt 5", "active_users >= 10 and inactive_users < 5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeExprString(tt.expr); got != tt.want {
				t.Errorf("NormalizeExprString(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestFromExpr(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    ConditionNode
		wantErr bool
	}{
		{
			name: "Eq operator",
			expr: "(name == 'Alice')",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "name",
					Operator:  "eq",
					Value:     "Alice",
				},
			},
			wantErr: false,
		},
		{
			name: "Gt operator",
			expr: "(age > 30)",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "age",
					Operator:  "gt",
					Value:     "30",
				},
			},
			wantErr: false,
		},
		{
			name: "Lt operator",
			expr: "(count < 10)",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "count",
					Operator:  "lt",
					Value:     "10",
				},
			},
			wantErr: false,
		},
		{
			name: "Gte operator",
			expr: "(score >= 90)",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "score",
					Operator:  "gte",
					Value:     "90",
				},
			},
			wantErr: false,
		},
		{
			name: "Lte operator",
			expr: "(score <= 90)",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "score",
					Operator:  "lte",
					Value:     "90",
				},
			},
			wantErr: false,
		},
		{
			name: "In operator",
			expr: "(status in 'active')",
			want: ConditionNode{
				Condition: &Condition{
					Attribute: "status",
					Operator:  "in",
					Value:     "active",
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid expression",
			expr:    "name eq",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromExpr(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !assert.Equal(t, tt.want, got) {
				t.Errorf("FromExpr() = %v, want %v", got, tt.want)
			}
		})
	}
}

