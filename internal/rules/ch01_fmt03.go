package rules

import (
	"context"
	"time"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
	"github.com/YeiyoNathnael/goulinette/internal/tools"
)

type fmt03Rule struct{}

const (
	fmt03ToolName    = "staticcheck"
	fmt03AllPackages = "./..."
	fmt03TimeoutMin  = 3
)

// NewFMT03 returns the FMT03 rule implementation.
func NewFMT03() Rule {
	return fmt03Rule{}
}

// ID returns the rule identifier.
func (fmt03Rule) ID() string {
	return ruleFMT03
}

// Chapter returns the chapter number for this rule.
func (fmt03Rule) Chapter() int {
	return 1
}

// Run executes this rule against the provided context.
func (fmt03Rule) Run(ctx Context) ([]diag.Finding, error) {
	output, err := tools.RunInDir(context.Background(), fmt03TimeoutMin*time.Minute, ctx.Root, fmt03ToolName, fmt03AllPackages)
	if err != nil {
		if output == "" {
			if ctx.StrictTools {
				return nil, err
			}
			return []diag.Finding{{
				RuleID:   ruleFMT03,
				Severity: diag.SeverityWarning,
				Message:  err.Error(),
				Hint:     "install staticcheck or run with --strict-tools",
				Tool:     fmt03ToolName,
			}}, nil
		}
	}

	return tools.ParseOutputDiagnostics(output, ruleFMT03, fmt03ToolName, diag.SeverityWarning), nil
}
