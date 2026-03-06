package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type doc05Rule struct{}

const (
	doc05Chapter = 9
)

// NewDOC05 returns the DOC05 rule implementation.
func NewDOC05() Rule {
	return doc05Rule{}
}

// ID returns the rule identifier.
func (doc05Rule) ID() string {
	return ruleDOC05
}

// Chapter returns the chapter number for this rule.
func (doc05Rule) Chapter() int {
	return doc05Chapter
}

// Run executes this rule against the provided context.
func (doc05Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	packageVars := collectPackageVarNames(parsed)
	diagnostics := make([]diag.Finding, 0)

	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "init" {
				continue
			}

			if isImmutableInitBody(fn.Body, packageVars) {
				continue
			}

			pos := pf.FSet.Position(fn.Name.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleDOC05,
				Severity: diag.SeverityWarning,
				Message:  "init() should be avoided except for immutable setup",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "move side effects to explicit initialization code",
			})
		}
	}

	return diagnostics, nil
}
