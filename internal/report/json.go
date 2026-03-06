package report

import (
	"encoding/json"
	"io"

	"goulinette/internal/diag"
)

func printJSON(w io.Writer, result diag.Result) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}
