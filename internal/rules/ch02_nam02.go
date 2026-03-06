package rules

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam02Rule struct{}

func NewNAM02() Rule {
	return nam02Rule{}
}

func (nam02Rule) ID() string {
	return "NAM-02"
}

func (nam02Rule) Chapter() int {
	return 2
}

func (nam02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}

			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if isAllCapsStyle(name.Name) {
						pos := pf.FSet.Position(name.Pos())
						diagnostics = append(diagnostics, diag.Diagnostic{
							RuleID:   "NAM-02",
							Severity: diag.SeverityError,
							Message:  "constant name must not use ALL_CAPS style",
							Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
							Hint:     "use camelCase or PascalCase constant naming",
						})
					}
				}
			}
		}
	}

	return diagnostics, nil
}

func isAllCapsStyle(name string) bool {
	if name == "" {
		return false
	}

	hasLetter := false
	hasLower := false
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
