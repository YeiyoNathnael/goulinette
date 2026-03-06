package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type doc03Rule struct{}

const (
	doc03Chapter = 9
)

// NewDOC03 returns the DOC03 rule implementation.
func NewDOC03() Rule {
	return doc03Rule{}
}

// ID returns the rule identifier.
func (doc03Rule) ID() string {
	return ruleDOC03
}

// Chapter returns the chapter number for this rule.
func (doc03Rule) Chapter() int {
	return doc03Chapter
}

// Run executes this rule against the provided context.
func (doc03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleDOC03,
				Severity: diag.SeverityError,
				Message:  "doc comments must begin with the exact symbol name",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "start the comment with " + target.Name,
			})
		}
	}

	return diagnostics, nil
}
