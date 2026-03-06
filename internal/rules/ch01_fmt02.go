package rules

import (
	"context"
	"time"

	"goulinette/internal/diag"
	"goulinette/internal/tools"
)

type fmt02Rule struct{}

func NewFMT02() Rule {
	return fmt02Rule{}
}

func (fmt02Rule) ID() string {
	return "FMT-02"
}

func (fmt02Rule) Chapter() int {
	return 1
}

func (fmt02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	output, err := tools.RunInDir(context.Background(), 3*time.Minute, ctx.Root, "go", "vet", "./...")
	if err != nil && output == "" {
		return nil, err
	}

	return tools.ParseOutputDiagnostics(output, "FMT-02", "go vet", diag.SeverityError), nil
}
