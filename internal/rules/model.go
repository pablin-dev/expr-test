package rules

// RuleConfig represents the structure of a rule defined in a YAML file,
// intended for direct mapping using YAML tags.
type RuleConfig struct {
	ID         string        `yaml:"id" validate:"required"`
	Expr       string        `yaml:"expr"`
	Conditions ConditionNode `yaml:"conditions" validate:"required"`
}
