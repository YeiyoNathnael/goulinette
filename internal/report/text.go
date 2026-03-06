package report

import (
	"fmt"
	"io"
	"strings"

	"goulinette/internal/diag"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiRed    = "\x1b[31m"
	ansiYellow = "\x1b[33m"
	ansiGray   = "\x1b[90m"
)

var chapterColors = map[int]string{
	1:  "\x1b[38;5;39m",
	2:  "\x1b[38;5;208m",
	3:  "\x1b[38;5;45m",
	4:  "\x1b[38;5;171m",
	5:  "\x1b[38;5;220m",
	6:  "\x1b[38;5;196m",
	7:  "\x1b[38;5;51m",
	8:  "\x1b[38;5;99m",
	9:  "\x1b[38;5;34m",
	10: "\x1b[38;5;214m",
	11: "\x1b[38;5;129m",
	12: "\x1b[38;5;44m",
	13: "\x1b[38;5;177m",
	14: "\x1b[38;5;81m",
	15: "\x1b[38;5;203m",
	16: "\x1b[38;5;141m",
	17: "\x1b[38;5;40m",
	18: "\x1b[38;5;201m",
	19: "\x1b[38;5;93m",
}

var ruleChapters = map[string]int{
	"FMT": 1,
	"NAM": 2,
	"VAR": 3,
	"CTL": 4,
	"FUN": 5,
	"ERR": 6,
	"TYP": 7,
	"STR": 8,
	"DOC": 9,
	"SLC": 10,
	"CON": 11,
	"CER": 12,
	"LIM": 13,
	"CTX": 14,
	"IMP": 15,
	"RES": 16,
	"SAF": 17,
	"MAG": 18,
	"TST": 19,
}

func printText(w io.Writer, result diag.Result) {
	for _, d := range result.Diagnostics {
		severityLabel := colorizeSeverity(d.Severity)
		ruleLabel := colorizeRuleID(d.RuleID)
		if d.Pos.File == "" {
			_, _ = fmt.Fprintf(w, "%s [%s] %s\n", severityLabel, ruleLabel, d.Message)
			continue
		}
		if d.Pos.Line > 0 {
			_, _ = fmt.Fprintf(w, "%s%s:%d:%d%s: %s [%s] %s\n", ansiGray, d.Pos.File, d.Pos.Line, d.Pos.Col, ansiReset, severityLabel, ruleLabel, d.Message)
			continue
		}
		_, _ = fmt.Fprintf(w, "%s%s%s: %s [%s] %s\n", ansiGray, d.Pos.File, ansiReset, severityLabel, ruleLabel, d.Message)
	}

	for _, runtimeErr := range result.RuntimeErrs {
		_, _ = fmt.Fprintf(w, "runtime: %s\n", runtimeErr)
	}
}

func colorizeRuleID(ruleID string) string {
	chapter := chapterForRuleID(ruleID)
	color, ok := chapterColors[chapter]
	if !ok {
		color = "\x1b[38;5;81m"
	}
	return ansiBold + color + ruleID + ansiReset
}

func chapterForRuleID(ruleID string) int {
	prefix, _, _ := strings.Cut(ruleID, "-")
	chapter, _ := ruleChapters[prefix]
	return chapter
}

func colorizeSeverity(severity diag.Severity) string {
	text := string(severity)
	switch severity {
	case diag.SeverityError:
		return ansiBold + ansiRed + text + ansiReset
	case diag.SeverityWarning:
		return ansiYellow + text + ansiReset
	default:
		return text
	}
}
