package tools

import (
	"strconv"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	minDiagnosticParts   = 3
	diagnosticMessageIdx = 2
	partsWithColumn      = 3
	messageIdxWithColumn = 3
)

// ParseOutputDiagnostics documents this exported function.
func ParseOutputDiagnostics(output, ruleID, tool string, severity diag.Severity) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		d := parseSingleLine(line, ruleID, tool, severity)
		diagnostics = append(diagnostics, d)
	}
	return diagnostics
}

func parseSingleLine(line, ruleID, tool string, severity diag.Severity) diag.Finding {
	parts := strings.Split(line, ":")
	if len(parts) < minDiagnosticParts {
		return diag.Finding{
			RuleID:   ruleID,
			Severity: severity,
			Message:  line,
			Tool:     tool,
		}
	}

	file := strings.TrimSpace(parts[0])
	lineNo, lineErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if lineErr != nil {
		return diag.Finding{RuleID: ruleID, Severity: severity, Message: line, Tool: tool}
	}

	var colNo int
	messageIdx := diagnosticMessageIdx
	if c, colErr := strconv.Atoi(strings.TrimSpace(parts[partsWithColumn-1])); colErr == nil {
		colNo = c
		messageIdx = messageIdxWithColumn
	}

	message := strings.TrimSpace(strings.Join(parts[messageIdx:], ":"))
	if message == "" {
		message = line
	}

	return diag.Finding{
		RuleID:   ruleID,
		Severity: severity,
		Message:  message,
		Pos:      diag.Position{File: file, Line: lineNo, Col: colNo},
		Tool:     tool,
	}
}
