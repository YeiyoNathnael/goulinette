package diag

import "sort"

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Position struct {
	File string `json:"file"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

type Diagnostic struct {
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Pos      Position `json:"position"`
	Hint     string   `json:"hint,omitempty"`
	Tool     string   `json:"tool,omitempty"`
}

type Result struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	RuntimeErrs []string     `json:"runtimeErrors,omitempty"`
}

func (r Result) ExitCode(warningsAsErrors bool) int {
	hasError := false
	hasWarning := false
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

func Sort(diags []Diagnostic) {
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
