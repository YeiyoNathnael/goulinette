package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc04Rule struct{}

func NewDOC04() Rule {
	return doc04Rule{}
}

func (doc04Rule) ID() string {
	return "DOC-04"
}

func (doc04Rule) Chapter() int {
	return 9
}

func (doc04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, target := range collectExportedDocTargets(pf.File) {
			if target.Doc != nil {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "DOC-04",
				Severity: diag.SeverityError,
				Message:  "all exported symbols must have a doc comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add a // comment that starts with " + target.Name,
			})
		}
	}

	return diagnostics, nil
}
