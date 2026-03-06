package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc02Rule struct{}

func NewDOC02() Rule {
	return doc02Rule{}
}

func (doc02Rule) ID() string {
	return "DOC-02"
}

func (doc02Rule) Chapter() int {
	return 9
}

func (doc02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, target := range collectExportedDocTargets(pf.File) {
			if !isBlockDocComment(target.Doc) {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "DOC-02",
				Severity: diag.SeverityError,
				Message:  "doc comments must use // line comments, not block comments",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "replace /* */ documentation comments with // lines",
			})
		}
	}

	return diagnostics, nil
}
