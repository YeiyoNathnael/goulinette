package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc02Rule struct{}

const (
	doc02Chapter = 9
)

// NewDOC02 returns the DOC02 rule implementation.
func NewDOC02() Rule {
	return doc02Rule{}
}

// ID returns the rule identifier.
func (doc02Rule) ID() string {
	return ruleDOC02
}

// Chapter returns the chapter number for this rule.
func (doc02Rule) Chapter() int {
	return doc02Chapter
}

// Run executes this rule against the provided context.
func (doc02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, target := range collectExportedDocTargets(pf.File) {
			if !isBlockDocComment(target.Doc) {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleDOC02,
				Severity: diag.SeverityError,
				Message:  "doc comments must use // line comments, not block comments",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "replace /* */ documentation comments with // lines",
			})
		}
	}

	return diagnostics, nil
}
