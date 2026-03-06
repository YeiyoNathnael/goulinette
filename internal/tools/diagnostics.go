package tools

import (
	"strconv"
	"strings"

	"goulinette/internal/diag"
)

func ParseOutputDiagnostics(output, ruleID, tool string, severity diag.Severity) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
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

func parseSingleLine(line, ruleID, tool string, severity diag.Severity) diag.Diagnostic {
	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		return diag.Diagnostic{
			RuleID:   ruleID,
			Severity: severity,
			Message:  line,
			Tool:     tool,
		}
	}

	file := strings.TrimSpace(parts[0])
	lineNo, lineErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if lineErr != nil {
		return diag.Diagnostic{RuleID: ruleID, Severity: severity, Message: line, Tool: tool}
	}

	colNo := 0
	messageIdx := 2
	if c, colErr := strconv.Atoi(strings.TrimSpace(parts[2])); colErr == nil {
		colNo = c
		messageIdx = 3
	}

	message := strings.TrimSpace(strings.Join(parts[messageIdx:], ":"))
	if message == "" {
		message = line
	}

	return diag.Diagnostic{
		RuleID:   ruleID,
		Severity: severity,
		Message:  message,
		Pos:      diag.Position{File: file, Line: lineNo, Col: colNo},
		Tool:     tool,
	}
}
