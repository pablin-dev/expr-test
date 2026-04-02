package router

import (
	"fmt"
	"router/internal/rules"
	"testing"

	"github.com/expr-lang/expr"
	// Import the internal rules package
)

func TestPaymentEvaluation(t *testing.T) {
	// Initialize the expr client
	client := NewClient()

	// The expected expression from rule.yaml
	// expr: ((Amount gt 1000) and ((Currency eq 'GBP') or (Currency eq 'CAD') or not((UserType eq 'Guest'))) and ((Country eq 'USA') and not((User eq true))))

	// Let's define the expression directly for the test first to ensure expr works as expected.
	// We'll use the operators supported by expr.
	programStr := "((Amount > 1000) and ((Currency == 'GBP') or (Currency == 'CAD') or not((UserType == 'Guest'))) and ((Country == 'USA') and not((User == true))))"

	program, err := client.Compile(programStr)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	tests := []struct {
		payment rules.Payment
		want    bool
	}{
		{
			payment: rules.Payment{
				Amount:   1500,
				Currency: "GBP",
				UserType: "Guest",
				Country:  "USA",
				User:     false,
			},
			want: true, // (1500 > 1000) AND ('GBP' == 'GBP' OR ...) AND ('USA' == 'USA' AND NOT false)
		},
		{
			payment: rules.Payment{
				Amount:   500,
				Currency: "GBP",
				UserType: "Guest",
				Country:  "USA",
				User:     false,
			},
			want: false, // Amount too low
		},
		{
			payment: rules.Payment{
				Amount:   1500,
				Currency: "USD",
				UserType: "Guest",
				Country:  "USA",
				User:     false,
			},
			want: false, // Currency not in list and UserType is Guest (not Guest would be true)
		},
		{
			payment: rules.Payment{
				Amount:   1500,
				Currency: "USD",
				UserType: "Member",
				Country:  "USA",
				User:     false,
			},
			want: true, // Amount > 1000, not Guest, Country USA, User false
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			got, err := expr.Run(program, tt.payment)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpressionConsistency(t *testing.T) {
	// Initialize the expr client
	client := NewClient()

	ruleFiles := []string{
		"examples/rules/rule.yaml",
		"examples/rules/rule2.yaml",
		"examples/rules/rule3.yaml",
	}

	for _, path := range ruleFiles {
		t.Run(path, func(t *testing.T) {
			rule, err := ParseYAML(path)
			if err != nil {
				t.Fatalf("ParseYAML failed: %v", err)
			}

			// Use the extracted helper function to normalize the expression
			generated := rule.Conditions.ToExpr()
			normalizedYamlExpr := rules.NormalizeExprString(rule.Expr) // Call the new helper function

			t.Logf("Generated: %s", generated)
			t.Logf("Expected (Normalized): %s", normalizedYamlExpr)

			// Functional comparison: run both against diverse inputs to ensure they are semantically identical.
			progGen, err := client.Compile(generated)
			if err != nil {
				t.Fatalf("Failed to compile generated expr: %v", err)
			}
			progYaml, err := client.Compile(normalizedYamlExpr)
			if err != nil {
				t.Fatalf("Failed to compile normalized YAML expr: %v", err)
			}

			testInputs := []rules.Payment{
				{Amount: 1500, Currency: "GBP", Country: "USA", User: false},
				{Amount: 500, Currency: "EUR", Country: "DE", User: true},
				{Amount: 100, UserType: "VIP", Country: "UK", User: true},
				{Amount: 0, UserType: "Guest", Country: "USA", User: false},
				{Amount: 10000, Currency: "CAD", UserType: "Admin", User: true},
			}

			for i, p := range testInputs {
				resGen, _ := expr.Run(progGen, p)
				resYaml, _ := expr.Run(progYaml, p)
				if resGen != resYaml {
					t.Errorf("Mismatch for input %d (rules.Payment %+v): Generated returned %v, YAML expr returned %v", i, p, resGen, resYaml)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	// Initialize the expr client
	client := NewClient()

	// Rule from rule2.yaml: (UserType eq 'VIP' or (Amount gte 5000 and Country eq 'UK')) and not(User eq false)
	// Equivalent: (UserType == 'VIP' or (Amount >= 5000 and Country == 'UK')) and User == true
	rule2Str := "((UserType == 'VIP') or ((Amount >= 5000) and (Country == 'UK'))) and not((User == false))"
	program, err := client.Compile(rule2Str)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	tests := []struct {
		name    string
		payment rules.Payment
		want    bool
	}{
		{
			name: "VIP User",
			payment: rules.Payment{
				UserType: "VIP",
				User:     true,
			},
			want: true,
		},
		{
			name: "VIP User but False User Flag",
			payment: rules.Payment{
				UserType: "VIP",
				User:     false,
			},
			want: false,
		},
		{
			name: "UK High Amount",
			payment: rules.Payment{
				Amount:   5000,
				Country:  "UK",
				UserType: "Member",
				User:     true,
			},
			want: true,
		},
		{
			name: "UK Just Below Threshold",
			payment: rules.Payment{
				Amount:  4999.99,
				Country: "UK",
				User:    true,
			},
			want: false,
		},
		{
			name: "Zero Amount VIP",
			payment: rules.Payment{
				Amount:   0,
				UserType: "VIP",
				User:     true,
			},
			want: true,
		},
		{
			name: "Negative Amount (Invalid but edge)",
			payment: rules.Payment{
				Amount:   -1,
				UserType: "VIP",
				User:     true,
			},
			want: true,
		},
		{
			name: "Empty Strings",
			payment: rules.Payment{
				UserType: "",
				Country:  "",
				User:     true,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expr.Run(program, tt.payment)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule3Evaluation(t *testing.T) {
	// Initialize the expr client
	client := NewClient()

	// Rule 3: (Amount lt 100 or Currency eq 'EUR') and (not(Country eq 'USA') or (UserType eq 'Admin' and User eq true))
	rule, _ := ParseYAML("examples/rules/rule3.yaml")
	generatedExpr := rule.Conditions.ToExpr()
	program, _ := client.Compile(generatedExpr)

	tests := []struct {
		name    string
		payment rules.Payment
		want    bool
	}{
		{
			name: "Small Amount, Non-USA",
			payment: rules.Payment{
				Amount:  50,
				Country: "FR",
			},
			want: true,
		},
		{
			name: "Large Amount, EUR, Admin",
			payment: rules.Payment{
				Amount:   1000,
				Currency: "EUR",
				Country:  "USA",
				UserType: "Admin",
				User:     true,
			},
			want: true,
		},
		{
			name: "Large Amount, Non-EUR, Non-USA",
			payment: rules.Payment{
				Amount:   1000,
				Currency: "USD",
				Country:  "UK",
			},
			want: false, // (1000 < 100 or USD == EUR) is false
		},
		{
			name: "Small Amount, USA, Not Admin",
			payment: rules.Payment{
				Amount:   10,
				Country:  "USA",
				UserType: "Guest",
			},
			want: false, // (10 < 100) is true, but (NOT USA or (Guest and false)) is false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := client.Run(program, tt.payment)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
