package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam04Rule struct{}

const nam04Chapter = 2

// NewNAM04 returns the NAM04 rule implementation.
func NewNAM04() Rule {
	return nam04Rule{}
}

// ID returns the rule identifier.
func (nam04Rule) ID() string {
	return ruleNAM04
}

// Chapter returns the chapter number for this rule.
func (nam04Rule) Chapter() int {
	return nam04Chapter
}

// Run executes this rule against the provided context.
func (nam04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		diagnostics = append(diagnostics, nam04DiagnosticsForDecls(pf)...)
	}

	return diagnostics, nil
}

func nam04DiagnosticsForDecls(pf parsedFile) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range pf.File.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || (gd.Tok != token.VAR && gd.Tok != token.CONST) {
			continue
		}
		for _, spec := range gd.Specs {
			diagnostics = append(diagnostics, nam04DiagnosticsForSpec(pf, spec)...)
		}
	}
	return diagnostics
}

func nam04DiagnosticsForSpec(pf parsedFile, spec ast.Spec) []diag.Finding {
	vs, ok := spec.(*ast.ValueSpec)
	if !ok {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, n := range vs.Names {
		if n == nil || n.Name == "_" {
			continue
		}
		if len([]rune(n.Name)) > 2 {
			continue
		}
		pos := pf.FSet.Position(n.Pos())
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleNAM04,
			Severity: diag.SeverityWarning,
			Message:  "package-level identifier name is too short to be descriptive",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "use full descriptive names for package-level identifiers",
		})
	}

	return diagnostics
}
