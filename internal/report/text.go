package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiRed    = "\x1b[31m"
	ansiYellow = "\x1b[33m"
	ansiGray   = "\x1b[90m"

	chapterFMT = 1
	chapterNAM = 2
	chapterVAR = 3
	chapterCTL = 4
	chapterFUN = 5
	chapterERR = 6
	chapterTYP = 7
	chapterSTR = 8
	chapterDOC = 9
	chapterSLC = 10
	chapterCON = 11
	chapterCER = 12
	chapterLIM = 13
	chapterCTX = 14
	chapterIMP = 15
	chapterRES = 16
	chapterSAF = 17
	chapterMAG = 18
	chapterTST = 19

	defaultChapterColor = "\x1b[38;5;81m"
)

var chapterColors = map[int]string{
	chapterFMT: "\x1b[38;5;39m",
	chapterNAM: "\x1b[38;5;208m",
	chapterVAR: "\x1b[38;5;45m",
	chapterCTL: "\x1b[38;5;171m",
	chapterFUN: "\x1b[38;5;220m",
	chapterERR: "\x1b[38;5;196m",
	chapterTYP: "\x1b[38;5;51m",
	chapterSTR: "\x1b[38;5;99m",
	chapterDOC: "\x1b[38;5;34m",
	chapterSLC: "\x1b[38;5;214m",
	chapterCON: "\x1b[38;5;129m",
	chapterCER: "\x1b[38;5;44m",
	chapterLIM: "\x1b[38;5;177m",
	chapterCTX: defaultChapterColor,
	chapterIMP: "\x1b[38;5;203m",
	chapterRES: "\x1b[38;5;141m",
	chapterSAF: "\x1b[38;5;40m",
	chapterMAG: "\x1b[38;5;201m",
	chapterTST: "\x1b[38;5;93m",
}

var ruleChapters = map[string]int{
	"FMT": chapterFMT,
	"NAM": chapterNAM,
	"VAR": chapterVAR,
	"CTL": chapterCTL,
	"FUN": chapterFUN,
	"ERR": chapterERR,
	"TYP": chapterTYP,
	"STR": chapterSTR,
	"DOC": chapterDOC,
	"SLC": chapterSLC,
	"CON": chapterCON,
	"CER": chapterCER,
	"LIM": chapterLIM,
	"CTX": chapterCTX,
	"IMP": chapterIMP,
	"RES": chapterRES,
	"SAF": chapterSAF,
	"MAG": chapterMAG,
	"TST": chapterTST,
}

func printText(w io.Writer, result diag.Result) {
	for _, d := range result.Diagnostics {
		severityLabel := colorizeSeverity(d.Severity)
		ruleLabel := colorizeRuleID(d.RuleID)
		if d.Pos.File == "" {
			fmt.Fprintf(w, "%s [%s] %s\n", severityLabel, ruleLabel, d.Message)
			continue
		}
		if d.Pos.Line > 0 {
			fmt.Fprintf(w, "%s%s:%d:%d%s: %s [%s] %s\n", ansiGray, d.Pos.File, d.Pos.Line, d.Pos.Col, ansiReset, severityLabel, ruleLabel, d.Message)
			continue
		}
		fmt.Fprintf(w, "%s%s%s: %s [%s] %s\n", ansiGray, d.Pos.File, ansiReset, severityLabel, ruleLabel, d.Message)
	}

	for _, runtimeErr := range result.RuntimeErrs {
		fmt.Fprintf(w, "runtime: %s\n", runtimeErr)
	}
}

func colorizeRuleID(ruleID string) string {
	chapter := chapterForRuleID(ruleID)
	color, ok := chapterColors[chapter]
	if !ok {
		color = defaultChapterColor
	}
	return ansiBold + color + ruleID + ansiReset
}

func chapterForRuleID(ruleID string) int {
	prefix, _, _ := strings.Cut(ruleID, "-")
	chapter, ok := ruleChapters[prefix]
	if !ok {
		return 0
	}
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
