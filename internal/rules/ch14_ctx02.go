package rules

import (
	"go/ast"
	"go/types"

	"goulinette/internal/diag"
)

type ctx02Rule struct{}

func NewCTX02() Rule {
	return ctx02Rule{}
}

func (ctx02Rule) ID() string {
	return "CTX-02"
}

func (ctx02Rule) Chapter() int {
	return 14
}

func (ctx02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}

				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					obj, ok := pkg.TypesInfo.Defs[ts.Name].(*types.TypeName)
					if !ok {
						continue
					}

					st, ok := obj.Type().Underlying().(*types.Struct)
					if !ok {
						continue
					}

					for i := 0; i < st.NumFields(); i++ {
						field := st.Field(i)
						if !isContextTypeOrPointer(field.Type()) {
							continue
						}

						pos := pkg.Fset.Position(field.Pos())
						diagnostics = append(diagnostics, diag.Diagnostic{
							RuleID:   "CTX-02",
							Severity: diag.SeverityError,
							Message:  "context.Context must not be stored in struct fields",
							Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
							Hint:     "pass context explicitly via function parameters instead of storing it",
						})
					}
				}
			}
		}
	}

	return diagnostics, nil
}
