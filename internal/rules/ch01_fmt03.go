package rules

import (
	"context"
	"time"

	"goulinette/internal/diag"
	"goulinette/internal/tools"
)

type fmt03Rule struct{}

func NewFMT03() Rule {
	return fmt03Rule{}
}

func (fmt03Rule) ID() string {
	return "FMT-03"
}

func (fmt03Rule) Chapter() int {
	return 1
}

func (fmt03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	output, err := tools.RunInDir(context.Background(), 3*time.Minute, ctx.Root, "staticcheck", "./...")
	if err != nil {
		if output == "" {
			if ctx.StrictTools {
				return nil, err
			}
			return []diag.Diagnostic{{
				RuleID:   "FMT-03",
				Severity: diag.SeverityWarning,
				Message:  err.Error(),
				Hint:     "install staticcheck or run with --strict-tools",
				Tool:     "staticcheck",
			}}, nil
		}
	}

	return tools.ParseOutputDiagnostics(output, "FMT-03", "staticcheck", diag.SeverityWarning), nil
}
