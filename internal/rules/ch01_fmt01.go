package rules

import (
	"context"
	"strings"
	"time"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
	"github.com/YeiyoNathnael/goulinette/internal/tools"
)

type fmt01Rule struct{}

const (
	fmt01Chapter        = 1
	fmt01ToolName       = "gofmt"
	fmt01ListArg        = "-l"
	fmt01ChunkSize      = 200
	fmt01TimeoutMinutes = 2
	fmt01DefaultChunk   = 100
)

// NewFMT01 returns the FMT01 rule implementation.
func NewFMT01() Rule {
	return fmt01Rule{}
}

// ID returns the rule identifier.
func (fmt01Rule) ID() string {
	return ruleFMT01
}

// Chapter returns the chapter number for this rule.
func (fmt01Rule) Chapter() int {
	return fmt01Chapter
}

// Run executes this rule against the provided context.
func (fmt01Rule) Run(ctx Context) ([]diag.Finding, error) {
	if len(ctx.Files) == 0 {
		return nil, nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, chunk := range chunkFiles(ctx.Files, fmt01ChunkSize) {
		args := append([]string{fmt01ListArg}, chunk...)
		output, err := tools.Run(context.Background(), fmt01TimeoutMinutes*time.Minute, fmt01ToolName, args...)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(output, "\n") {
			file := strings.TrimSpace(line)
			if file == "" {
				continue
			}
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleFMT01,
				Severity: diag.SeverityError,
				Message:  "file is not formatted with gofmt",
				Pos:      diag.Position{File: file},
				Hint:     "run gofmt -w on this file",
				Tool:     fmt01ToolName,
			})
		}
	}

	return diagnostics, nil
}

func chunkFiles(files []string, size int) [][]string {
	if size <= 0 {
		size = fmt01DefaultChunk
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
