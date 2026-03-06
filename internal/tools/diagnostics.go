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

// ParseOutputDiagnostics converts the line-oriented text output of an external
// tool (e.g. staticcheck, errcheck) into a slice of diag.Finding values.
// It expects each line to follow the convention "file:line:col: message" or
// "file:line: message"; lines that don't parse as positional diagnostics are
// returned as findings without position information so that no output is
// silently dropped. Empty lines and lines beginning with '#' are ignored.
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
