package report

import (
	"io"

	"goulinette/internal/diag"
)

func Print(w io.Writer, format string, result diag.Result) {
	if format == "json" {
		printJSON(w, result)
		return
	}
	printText(w, result)
}
