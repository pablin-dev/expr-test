package router

import (
	"fmt"
	"io"
	"os"
	"router/internal/rules"

	"gopkg.in/yaml.v3"
)

// ParseYAML reads a YAML file and unmarshals it into a rules.RuleConfig struct.
// It returns a pointer to the populated RuleConfig and an error if any occurs.
func ParseYAML(filePath string) (*rules.RuleConfig, error) {
	yamlFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer yamlFile.Close()

	byteValue, err := io.ReadAll(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var ruleConfig rules.RuleConfig
	err = yaml.Unmarshal(byteValue, &ruleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &ruleConfig, nil
}
