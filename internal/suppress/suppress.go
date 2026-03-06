// Package suppress filters diagnostics that are covered by inline
// suppression directives embedded in source comments.
//
// A suppression directive has the form:
//
// //goulinette:ignore            (suppresses all rules on this line)
// //goulinette:ignore RULE-ID    (suppresses one specific rule)
// //goulinette:ignore R1 R2      (suppresses multiple rules)
//
// The directive may appear on the same line as the flagged code or on
// the line immediately preceding it.
package suppress

import (
	"bufio"
	"os"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const directivePrefix = "//goulinette:ignore"

// Filter returns a copy of findings with every suppressed entry removed.
// A finding is suppressed when its source line, or the line directly
// above it, contains a //goulinette:ignore directive that covers its
// rule ID. Files that cannot be read are left unsuppressed.
func Filter(findings []diag.Finding) []diag.Finding {
	byFile := groupByFile(findings)
	out := make([]diag.Finding, 0, len(findings))
	for file, group := range byFile {
		lines, err := readLines(file)
		if err != nil {
			out = append(out, group...)
			continue
		}
		for _, f := range group {
			if !isSuppressed(f, lines) {
				out = append(out, f)
			}
		}
	}
	return out
}

func groupByFile(findings []diag.Finding) map[string][]diag.Finding {
	out := make(map[string][]diag.Finding, len(findings))
	for _, f := range findings {
		out[f.Pos.File] = append(out[f.Pos.File], f)
	}
	return out
}

func isSuppressed(f diag.Finding, lines []string) bool {
	if f.Pos.Line <= 0 {
		return false
	}
	// Check the finding's own line (index line-1) and the line above (line-2).
	for _, idx := range []int{f.Pos.Line - 1, f.Pos.Line - 2} {
		if idx >= 0 && idx < len(lines) && matchesDirective(lines[idx], f.RuleID) {
			return true
		}
	}
	return false
}

// matchesDirective reports whether the given source line contains a
// goulinette:ignore directive that covers ruleID.
func matchesDirective(line string, ruleID string) bool {
	i := strings.Index(line, directivePrefix)
	if i < 0 {
		return false
	}
	rest := strings.TrimSpace(line[i+len(directivePrefix):])
	if rest == "" {
		// bare directive: suppresses every rule on this line
		return true
	}
	for _, id := range strings.Fields(rest) {
		if strings.EqualFold(id, ruleID) {
			return true
		}
	}
	return false
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
