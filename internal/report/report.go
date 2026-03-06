package report

import (
	"io"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

// Print documents this exported function.
func Print(w io.Writer, format string, result diag.Result) {
	if format == "json" {
		printJSON(w, result)
		return
	}
	printText(w, result)
}
