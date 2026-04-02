package router

import (
	"fmt"
	"router/internal/rules"

	// Added import for rules package

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

type RuleEnv struct {
	rules.Payment
	GBP rules.Pepe
	CAD rules.Pepe
	EUR rules.Pepe
}

// NewClient creates a new expr client.
// It takes Options to configure the client, such as the environment type for expressions.
func NewClient() *Client {
	return &Client{Env: RuleEnv{}}
}

// Compile compiles an expression string into an expr program.
// It uses the environment type defined in the client's options.
func (c *Client) Compile(expression string) (*vm.Program, error) {
	env := expr.Env(c.Env)
	program, err := expr.Compile(expression, env)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
	}
	return program, nil
}

// Run executes a compiled expr program with the given environment data.
// It returns the result of the execution or an error if execution fails.
func (c *Client) Run(program *vm.Program, p rules.Payment) (any, error) {
	envData := BaseRuleEnv
	envData.Payment = p
	result, err := expr.Run(program, envData)
	if err != nil {
		return nil, fmt.Errorf("failed to run expression: %w", err)
	}
	return result, nil
}
