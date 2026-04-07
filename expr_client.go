package router

import (
	"fmt"
	"router/internal/rules"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Client is a client for evaluating expressions.
type Client struct {
	Env RuleEnv
}

var BaseRuleEnv = RuleEnv{
	GBP: rules.GBP,
	CAD: rules.CAD,
	EUR: rules.EUR,
}

type BlacklistData struct {
	Mail     []string
	Name     []string
	Balance  []float64
	IsActive []bool
}

type RuleEnv struct {
	rules.Payment
	GBP         rules.Pepe
	CAD         rules.Pepe
	EUR         rules.Pepe
	Blacklisted BlacklistData `expr:"blacklisted"`
}

// NewClient creates a new expr client.
func NewClient() *Client {
	return &Client{Env: RuleEnv{}}
}

// Compile compiles an expression string into an expr program.
func (c *Client) Compile(expression string) (*vm.Program, error) {
	program, err := expr.Compile(expression, expr.Env(RuleEnv{}))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
	}
	return program, nil
}

// Run executes a compiled expr program with the given environment data.
func (c *Client) Run(program *vm.Program, p rules.Payment, b map[string][]any) (any, error) {
	var bData BlacklistData
	for _, v := range b["Mail"] {
		bData.Mail = append(bData.Mail, v.(string))
	}
	for _, v := range b["Name"] {
		bData.Name = append(bData.Name, v.(string))
	}
	for _, v := range b["Balance"] {
		bData.Balance = append(bData.Balance, v.(float64))
	}
	for _, v := range b["IsActive"] {
		bData.IsActive = append(bData.IsActive, v.(bool))
	}

	envData := RuleEnv{
		Payment:     p,
		GBP:         BaseRuleEnv.GBP,
		CAD:         BaseRuleEnv.CAD,
		EUR:         BaseRuleEnv.EUR,
		Blacklisted: bData,
	}

	result, err := expr.Run(program, envData)
	if err != nil {
		return nil, fmt.Errorf("failed to run expression: %w", err)
	}
	return result, nil
}
