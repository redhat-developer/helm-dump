package test

import (
	"github.com/ghodss/yaml"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	ContextLines = 4
)

func YamlDiff(expected interface{}, actual interface{}) (string, error) {
	yamlActual, err := yaml.Marshal(actual)
	if err != nil {
		return "", err
	}

	yamlExpected, err := yaml.Marshal(expected)
	if err != nil {
		return "", err
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(yamlExpected)),
		B:        difflib.SplitLines(string(yamlActual)),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  ContextLines,
	}

	return difflib.GetUnifiedDiffString(diff)
}
