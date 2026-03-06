package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc01Rule struct{}

const (
	doc01Chapter       = 9
	doc01MaxAllowedGap = 3
	doc01MinAllowedGap = 2
)

// NewDOC01 returns the DOC01 rule implementation.
func NewDOC01() Rule {
	return doc01Rule{}
}

// ID returns the rule identifier.
func (doc01Rule) ID() string {
	return ruleDOC01
}

// Chapter returns the chapter number for this rule.
func (doc01Rule) Chapter() int {
	return doc01Chapter
}

// Run executes this rule against the provided context.
func (doc01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		targets := collectExportedDocTargets(pf.File)
		for _, target := range targets {
			if target.Doc != nil {
				continue
			}

			declLine := pf.FSet.Position(target.Pos).Line
			_, endLine := nearestCommentGroupBeforeLine(pf.File, pf.FSet, declLine)
			if endLine < 0 {
				continue
			}

			lineGap := declLine - endLine
			if lineGap < doc01MinAllowedGap || lineGap > doc01MaxAllowedGap {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleDOC01,
				Severity: diag.SeverityError,
				Message:  "doc comments must be directly above declarations with no blank line",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "move the comment so it is immediately above the declaration",
			})
		}
	}

	return diagnostics, nil
}
