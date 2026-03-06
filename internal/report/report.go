package report

import (
	"io"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

func Print(w io.Writer, format string, result diag.Result) {
	if format == "json" {
		printJSON(w, result)
		return
	}
	printText(w, result)
}
