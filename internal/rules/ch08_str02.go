package rules

import (
	"go/ast"
	"strings"

	"goulinette/internal/diag"
)

type str02Rule struct{}

func NewSTR02() Rule {
	return str02Rule{}
}

func (str02Rule) ID() string {
	return "STR-02"
}

func (str02Rule) Chapter() int {
	return 8
}

func (str02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			recv := fn.Recv.List[0]
			if len(recv.Names) == 0 || recv.Names[0] == nil {
				continue
			}
			name := strings.ToLower(recv.Names[0].Name)
			if name != "this" && name != "self" {
				continue
			}

			pos := pf.FSet.Position(recv.Names[0].Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "STR-02",
				Severity: diag.SeverityError,
				Message:  "receiver names this and self are forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use a short receiver abbreviation based on type name",
			})
		}
	}

	return diagnostics, nil
}
