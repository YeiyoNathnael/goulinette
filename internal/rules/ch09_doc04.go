package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc04Rule struct{}

const (
	doc04Chapter = 9
)

// NewDOC04 returns the DOC04 rule implementation.
func NewDOC04() Rule {
	return doc04Rule{}
}

// ID returns the rule identifier.
func (doc04Rule) ID() string {
	return ruleDOC04
}

// Chapter returns the chapter number for this rule.
func (doc04Rule) Chapter() int {
	return doc04Chapter
}

// Run executes this rule against the provided context.
func (doc04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, target := range collectExportedDocTargets(pf.File) {
			if target.Doc != nil {
				continue
			}

			pos := pf.FSet.Position(target.Pos)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleDOC04,
				Severity: diag.SeverityError,
				Message:  "all exported symbols must have a doc comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add a // comment that starts with " + target.Name,
			})
		}
	}

	return diagnostics, nil
}
