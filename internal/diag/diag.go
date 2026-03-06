package diag

import "sort"

// Severity documents this exported type.
type Severity string

const (
	// SeverityError marks a diagnostic as an error.
	SeverityError Severity = "error"
	// SeverityWarning marks a diagnostic as a warning.
	SeverityWarning Severity = "warning"
)

// Position documents this exported type.
type Position struct {
	File string `json:"file"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

// Finding documents this exported type.
type Finding struct {
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Pos      Position `json:"position"`
	Hint     string   `json:"hint,omitempty"`
	Tool     string   `json:"tool,omitempty"`
}

// Result documents this exported type.
type Result struct {
	Diagnostics []Finding `json:"diagnostics"`
	RuntimeErrs []string  `json:"runtimeErrors,omitempty"`
}

// ExitCode documents this exported method.
func (r Result) ExitCode(warningsAsErrors bool) int {
	var hasError bool
	var hasWarning bool
	for _, d := range r.Diagnostics {
		if d.Severity == SeverityError {
			hasError = true
		}
		if d.Severity == SeverityWarning {
			hasWarning = true
		}
	}
	if hasError || (warningsAsErrors && hasWarning) {
		return 1
	}
	if len(r.RuntimeErrs) > 0 {
		return 2
	}
	return 0
}

// Sort documents this exported function.
func Sort(diags []Finding) {
	sort.SliceStable(diags, func(i, j int) bool {
		if diags[i].Pos.File != diags[j].Pos.File {
			return diags[i].Pos.File < diags[j].Pos.File
		}
		if diags[i].Pos.Line != diags[j].Pos.Line {
			return diags[i].Pos.Line < diags[j].Pos.Line
		}
		if diags[i].Pos.Col != diags[j].Pos.Col {
			return diags[i].Pos.Col < diags[j].Pos.Col
		}
		return diags[i].RuleID < diags[j].RuleID
	})
}
