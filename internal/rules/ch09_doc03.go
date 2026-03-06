package rules

import "goulinette/internal/diag"

type doc03Rule struct{}

func NewDOC03() Rule {
	return doc03Rule{}
}

func (doc03Rule) ID() string {
	return "DOC-03"
}

func (doc03Rule) Chapter() int {
	return 9
}

func (doc03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, target := range collectExportedDocTargets(pf.File) {
			if target.Doc == nil || !target.PrimaryForDoc3 {
				continue
			}

			first := firstDocWord(target.Doc)
			if first == target.Name {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "DOC-03",
				Severity: diag.SeverityError,
				Message:  "doc comments must begin with the exact symbol name",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "start the comment with " + target.Name,
			})
		}
	}

	return diagnostics, nil
}
