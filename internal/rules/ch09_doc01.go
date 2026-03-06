package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc01Rule struct{}

func NewDOC01() Rule {
	return doc01Rule{}
}

func (doc01Rule) ID() string {
	return "DOC-01"
}

func (doc01Rule) Chapter() int {
	return 9
}

func (doc01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
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
			if lineGap < 2 || lineGap > 3 {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "DOC-01",
				Severity: diag.SeverityError,
				Message:  "doc comments must be directly above declarations with no blank line",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "move the comment so it is immediately above the declaration",
			})
		}
	}

	return diagnostics, nil
}
