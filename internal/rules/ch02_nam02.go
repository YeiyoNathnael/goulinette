package rules

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam02Rule struct{}

const nam02Chapter = 2

// NewNAM02 returns the NAM02 rule implementation.
func NewNAM02() Rule {
	return nam02Rule{}
}

// ID returns the rule identifier.
func (nam02Rule) ID() string {
	return ruleNAM02
}

// Chapter returns the chapter number for this rule.
func (nam02Rule) Chapter() int {
	return nam02Chapter
}

// Run executes this rule against the provided context.
func (nam02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		diagnostics = append(diagnostics, nam02DiagnosticsForDecls(pf)...)
	}

	return diagnostics, nil
}

func nam02DiagnosticsForDecls(pf parsedFile) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range pf.File.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			diagnostics = append(diagnostics, nam02DiagnosticsForSpec(pf, spec)...)
		}
	}
	return diagnostics
}

func nam02DiagnosticsForSpec(pf parsedFile, spec ast.Spec) []diag.Finding {
	valueSpec, ok := spec.(*ast.ValueSpec)
	if !ok {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, name := range valueSpec.Names {
		if !isAllCapsStyle(name.Name) {
			continue
		}
		pos := pf.FSet.Position(name.Pos())
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleNAM02,
			Severity: diag.SeverityError,
			Message:  "constant name must not use ALL_CAPS style",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "use camelCase or PascalCase constant naming",
		})
	}

	return diagnostics
}

func isAllCapsStyle(name string) bool {
	if name == "" {
		return false
	}

	var hasLetter bool
	var hasLower bool
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				hasLower = true
			}
		}
	}

	if !hasLetter || hasLower {
		return false
	}

	if strings.ContainsRune(name, '_') {
		return true
	}

	return name == strings.ToUpper(name)
}
