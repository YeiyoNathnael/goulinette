package rules

import (
	"context"
	"strings"
	"time"

	"goulinette/internal/diag"
	"goulinette/internal/tools"
)

type fmt01Rule struct{}

func NewFMT01() Rule {
	return fmt01Rule{}
}

func (fmt01Rule) ID() string {
	return "FMT-01"
}

func (fmt01Rule) Chapter() int {
	return 1
}

func (fmt01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	if len(ctx.Files) == 0 {
		return nil, nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, chunk := range chunkFiles(ctx.Files, 200) {
		args := append([]string{"-l"}, chunk...)
		output, err := tools.Run(context.Background(), 2*time.Minute, "gofmt", args...)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(output, "\n") {
			file := strings.TrimSpace(line)
			if file == "" {
				continue
			}
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "FMT-01",
				Severity: diag.SeverityError,
				Message:  "file is not formatted with gofmt",
				Pos:      diag.Position{File: file},
				Hint:     "run gofmt -w on this file",
				Tool:     "gofmt",
			})
		}
	}

	return diagnostics, nil
}

func chunkFiles(files []string, size int) [][]string {
	if size <= 0 {
		size = 100
	}
	out := make([][]string, 0, (len(files)+size-1)/size)
	for start := 0; start < len(files); start += size {
		end := start + size
		if end > len(files) {
			end = len(files)
		}
		out = append(out, files[start:end])
	}
	return out
}
