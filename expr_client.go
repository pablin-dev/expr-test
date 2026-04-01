package router

import (
	"fmt"

	"router/internal/rules" // Added import for rules package

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Options for the expr client.
type Options struct {
	// EnvType is the type used for the expression environment.
	// Defaults to Payment if not provided.
	EnvType interface{}
}

// Client is a client for evaluating expressions.
type Client struct {
	options Options
}

// NewClient creates a new expr client.
// It takes Options to configure the client, such as the environment type for expressions.
func NewClient(opts Options) *Client {
	// Default EnvType to Payment if not provided.
	if opts.EnvType == nil {
		opts.EnvType = rules.Payment{} // Changed to rules.Payment{}
	}
	return &Client{options: opts}
}

// Compile compiles an expression string into an expr program.
// It uses the environment type defined in the client's options.
func (c *Client) Compile(expression string) (*vm.Program, error) {
	env := expr.Env(c.options.EnvType)
	program, err := expr.Compile(expression, env)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
	}
	return program, nil
}

// Run executes a compiled expr program with the given environment data.
// It returns the result of the execution or an error if execution fails.
func (c *Client) Run(program *vm.Program, envData interface{}) (interface{}, error) {
	result, err := expr.Run(program, envData)
	if err != nil {
		return nil, fmt.Errorf("failed to run expression: %w", err)
	}
	return result, nil
}
