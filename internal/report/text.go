package report

import (
	"fmt"
	"io"

	"goulinette/internal/diag"
)

func printText(w io.Writer, result diag.Result) {
	for _, d := range result.Diagnostics {
		if d.Pos.File == "" {
			_, _ = fmt.Fprintf(w, "%s [%s] %s\n", d.Severity, d.RuleID, d.Message)
			continue
		}
		if d.Pos.Line > 0 {
			_, _ = fmt.Fprintf(w, "%s:%d:%d: %s [%s] %s\n", d.Pos.File, d.Pos.Line, d.Pos.Col, d.Severity, d.RuleID, d.Message)
			continue
		}
		_, _ = fmt.Fprintf(w, "%s: %s [%s] %s\n", d.Pos.File, d.Severity, d.RuleID, d.Message)
	}

	for _, runtimeErr := range result.RuntimeErrs {
		_, _ = fmt.Fprintf(w, "runtime: %s\n", runtimeErr)
	}
}
