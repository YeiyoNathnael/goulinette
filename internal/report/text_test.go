package report

import (
	"bytes"
	"strings"
	"testing"

	"goulinette/internal/diag"
)

func TestPrintTextSeverityColors(t *testing.T) {
	result := diag.Result{
		Diagnostics: []diag.Diagnostic{
			{RuleID: "TYP-03", Severity: diag.SeverityError, Message: "err", Pos: diag.Position{File: "a.go", Line: 1, Col: 1}},
			{RuleID: "CTX-03", Severity: diag.SeverityWarning, Message: "warn", Pos: diag.Position{File: "a.go", Line: 2, Col: 1}},
		},
	}

	var buf bytes.Buffer
	printText(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "\x1b[1m\x1b[31merror\x1b[0m") {
		t.Fatalf("expected bold red error severity in output, got: %q", out)
	}
	if !strings.Contains(out, "\x1b[33mwarning\x1b[0m") {
		t.Fatalf("expected yellow-colored warning severity in output, got: %q", out)
	}
	if !strings.Contains(out, "\x1b[1m\x1b[38;5;51mTYP-03\x1b[0m") {
		t.Fatalf("expected fixed chapter-colored TYP-03 rule id in output, got: %q", out)
	}
	if !strings.Contains(out, "\x1b[1m\x1b[38;5;81mCTX-03\x1b[0m") {
		t.Fatalf("expected fixed chapter-colored CTX-03 rule id in output, got: %q", out)
	}

	if colorizeRuleID("TYP-03") == colorizeRuleID("CTX-03") {
		t.Fatalf("expected different fixed chapter colors for different rule groups")
	}
}
