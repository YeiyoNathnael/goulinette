package report

import (
	"io"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

// Print writes a formatted analysis report to w. When format is "json" the
// output is a single JSON object; any other value (including the default
// "text") produces the ANSI colour-coded human-readable text format.
func Print(w io.Writer, format string, result diag.Result) {
	if format == "json" {
		printJSON(w, result)
		return
	}
	printText(w, result)
}
