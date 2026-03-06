package diag

import "sort"

// Severity classifies the impact level of a diagnostic finding.
// The two supported values are SeverityError and SeverityWarning.
type Severity string

const (
	// SeverityError marks a diagnostic as an error.
	SeverityError Severity = "error"
	// SeverityWarning marks a diagnostic as a warning.
	SeverityWarning Severity = "warning"
)

// Position identifies the exact location in source code that triggered a
// finding. File is a path relative to the analysis root; Line and Col are
// 1-based. Col is omitted from JSON output when it is zero.
type Position struct {
	File string `json:"file"`
	Line int    `json:"line,omitempty"`
	Col  int    `json:"col,omitempty"`
}

// Finding represents a single diagnostic produced by a rule. RuleID
// identifies which rule fired; Pos records the source location; Message
// is the human-readable explanation; Hint and Tool carry optional
// remediation guidance and the name of any external tool that was invoked.
type Finding struct {
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Pos      Position `json:"position"`
	Hint     string   `json:"hint,omitempty"`
	Tool     string   `json:"tool,omitempty"`
}

// Result aggregates everything produced by a single analysis run.
// Diagnostics holds the rule findings; RuntimeErrs records non-fatal
// errors encountered while setting up or running individual rules.
type Result struct {
	Diagnostics []Finding `json:"diagnostics"`
	RuntimeErrs []string  `json:"runtimeErrors,omitempty"`
}

// ExitCode maps the result to a process exit code:
//   - 0: no findings (or only warnings with warningsAsErrors = false)
//   - 1: at least one error-severity finding, or any warning when warningsAsErrors is true
//   - 2: at least one runtime error and no diagnostic findings
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

// Sort orders diags in a deterministic, human-friendly sequence:
// ascending by file path, then line, then column, then rule ID.
// It uses a stable sort so equal elements preserve their original order.
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
