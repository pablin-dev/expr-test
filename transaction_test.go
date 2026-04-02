package router

import (
	"router/internal/rules"
	"testing"
)

type TestCase struct {
	payment rules.Payment
	want    bool
	name    string
}

func TestRuleEvaluations(t *testing.T) {
	// Initialize the expr client
	client := NewClient()

	rulesFiles := []string{
		"examples/rules/rule.yaml",
		"examples/rules/rule2.yaml",
		"examples/rules/rule3.yaml",
	}

	for _, rulePath := range rulesFiles {
		rule, err := ParseYAML(rulePath)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", rulePath, err)
		}

		generatedExpr := rule.Conditions.ToExpr()

		// Verify that generatedExpr matches the yaml expr field semantically.
		// We normalize the YAML expr field operators to match expr library syntax.
		normalizedYamlExpr := rules.NormalizeExprString(rule.Expr)

		program, err := client.Compile(generatedExpr)
		if err != nil {
			t.Fatalf("Failed to compile generated expr for %s: %v", rulePath, err)
		}

		yamlProgram, err := client.Compile(normalizedYamlExpr)
		if err != nil {
			t.Fatalf("Failed to compile YAML expr for %s: %v", rulePath, err)
		}

		var testCases []TestCase

		switch rule.ID {
		case "complex_rules_with_multiple_operators":
			testCases = []TestCase{
				{
					name: "Valid: Amount > 1000, Currency GBP, Country USA, User false",
					payment: rules.Payment{
						Amount:   1500,
						Currency: "GBP",
						Country:  "USA",
						User:     false,
					},
					want: true,
				},
				{
					name: "Invalid: Amount exactly 1000 (gt 1000 required)",
					payment: rules.Payment{
						Amount:   1000,
						Currency: "GBP",
						Country:  "USA",
						User:     false,
					},
					want: false,
				},
				{
					name: "Valid: Not Guest (Member), Country USA",
					payment: rules.Payment{
						Amount:   2000,
						Currency: "USD",
						UserType: "Member",
						Country:  "USA",
						User:     false,
					},
					want: true,
				},
				{
					name: "Invalid: User is true (not(User eq true) required)",
					payment: rules.Payment{
						Amount:   2000,
						Currency: "GBP",
						Country:  "USA",
						User:     true,
					},
					want: false,
				},
			}

		case "vip_high_value_uk":
			testCases = []TestCase{
				{
					name: "Valid: VIP User, User flag true",
					payment: rules.Payment{
						UserType: "VIP",
						User:     true,
					},
					want: true,
				},
				{
					name: "Valid: High Amount UK, User flag true",
					payment: rules.Payment{
						Amount:  6000,
						Country: "UK",
						User:    true,
					},
					want: true,
				},
				{
					name: "Invalid: High Amount non-UK",
					payment: rules.Payment{
						Amount:  6000,
						Country: "FR",
						User:    true,
					},
					want: false,
				},
				{
					name: "Invalid: VIP but User flag false",
					payment: rules.Payment{
						UserType: "VIP",
						User:     false,
					},
					want: false,
				},
			}

		case "euro_small_non_usa_admin":
			testCases = []TestCase{
				{
					name: "Valid: EUR Currency, Country Germany",
					payment: rules.Payment{
						Currency: "EUR",
						Country:  "DE",
					},
					want: true,
				},
				{
					name: "Valid: Small Amount, Country Canada",
					payment: rules.Payment{
						Amount:  50,
						Country: "CA",
					},
					want: true,
				},
				{
					name: "Valid: USA but Admin and User true",
					payment: rules.Payment{
						Currency: "EUR",
						Country:  "USA",
						UserType: "Admin",
						User:     true,
					},
					want: true,
				},
				{
					name: "Invalid: USA, not Admin",
					payment: rules.Payment{
						Amount:  10,
						Country: "USA",
					},
					want: false,
				},
			}
		}

		t.Run(rule.ID, func(t *testing.T) {
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Evaluation using generated expr
					got, err := client.Run(program, tc.payment)
					if err != nil {
						t.Fatalf("Run failed: %v", err)
					}
					if got != tc.want {
						t.Errorf("got %v, want %v for rules.Payment %+v", got, tc.want, tc.payment)
					}

					// Verification using original YAML expr (normalized)
					gotYaml, err := client.Run(yamlProgram, tc.payment)
					if err != nil {
						t.Fatalf("YAML Run failed: %v", err)
					}
					if gotYaml != got {
						t.Errorf("Semantic mismatch! generated result %v != YAML expr result %v", got, gotYaml)
					}
				})
			}
		})
	}
}
