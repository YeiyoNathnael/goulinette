package rules

import (
	"context"
	"time"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
	"github.com/YeiyoNathnael/goulinette/internal/tools"
)

type fmt02Rule struct{}

const (
	fmt02ToolName      = "go"
	fmt02VetSubcommand = "vet"
	fmt02AllPackages   = "./..."
	fmt02TimeoutMin    = 3
	fmt02ToolDisplay   = "go vet"
)

// NewFMT02 returns the FMT02 rule implementation.
func NewFMT02() Rule {
	return fmt02Rule{}
}

// ID returns the rule identifier.
func (fmt02Rule) ID() string {
	return ruleFMT02
}

// Chapter returns the chapter number for this rule.
func (fmt02Rule) Chapter() int {
	return 1
}

// Run executes this rule against the provided context.
func (fmt02Rule) Run(ctx Context) ([]diag.Finding, error) {
	output, err := tools.RunInDir(context.Background(), fmt02TimeoutMin*time.Minute, ctx.Root, fmt02ToolName, fmt02VetSubcommand, fmt02AllPackages)
	if err != nil && output == "" {
		return nil, err
	}

	return tools.ParseOutputDiagnostics(output, ruleFMT02, fmt02ToolDisplay, diag.SeverityError), nil
}
