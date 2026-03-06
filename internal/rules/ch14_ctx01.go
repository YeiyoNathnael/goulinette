package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"goulinette/internal/diag"
)

type ctx01Rule struct{}

func NewCTX01() Rule {
	return ctx01Rule{}
}

func (ctx01Rule) ID() string {
	return "CTX-01"
}

func (ctx01Rule) Chapter() int {
	return 14
}

func (ctx01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Name == nil {
						continue
					}
					obj, ok := pkg.TypesInfo.Defs[d.Name].(*types.Func)
					if !ok {
						continue
					}
					sig, ok := obj.Type().(*types.Signature)
					if !ok {
						continue
					}
					diagnostics = append(diagnostics, ctx01DiagnosticsForSignature(pkg.Fset, sig, d.Type, d.Name.Pos())...)

				case *ast.GenDecl:
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						iface, ok := ts.Type.(*ast.InterfaceType)
						if !ok || iface.Methods == nil {
							continue
						}

						for _, m := range iface.Methods.List {
							ft, ok := m.Type.(*ast.FuncType)
							if !ok {
								continue
							}
							t := pkg.TypesInfo.TypeOf(ft)
							sig, ok := t.(*types.Signature)
							if !ok {
								continue
							}

							pos := m.Type.Pos()
							if len(m.Names) > 0 {
								pos = m.Names[0].Pos()
							}
							diagnostics = append(diagnostics, ctx01DiagnosticsForSignature(pkg.Fset, sig, ft, pos)...)
						}
					}
				}
			}

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.FuncLit)
				if !ok {
					return true
				}
				t := pkg.TypesInfo.TypeOf(lit.Type)
				sig, ok := t.(*types.Signature)
				if !ok {
					return true
				}
				diagnostics = append(diagnostics, ctx01DiagnosticsForSignature(pkg.Fset, sig, lit.Type, lit.Type.Func)...)
				return true
			})
		}
	}

	return diagnostics, nil
}

func ctx01DiagnosticsForSignature(fset *token.FileSet, sig *types.Signature, ft *ast.FuncType, fallbackPos token.Pos) []diag.Diagnostic {
	if sig == nil || sig.Params() == nil {
		return nil
	}

	diags := make([]diag.Diagnostic, 0)
	for i := 0; i < sig.Params().Len(); i++ {
		p := sig.Params().At(i)
		if !isStrictContextType(p.Type()) {
			continue
		}
		if i == 0 {
			continue
		}

		posToken := fallbackPos
		if field := funcParamFieldByIndex(ft, i); field != nil {
			posToken = field.Type.Pos()
		}
		pos := fset.Position(posToken)
		diags = append(diags, diag.Diagnostic{
			RuleID:   "CTX-01",
			Severity: diag.SeverityError,
			Message:  "context.Context must be the first parameter",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "move context.Context to the first parameter position",
		})
	}

	return diags
}

func funcParamFieldByIndex(ft *ast.FuncType, index int) *ast.Field {
	if ft == nil || ft.Params == nil || index < 0 {
		return nil
	}

	count := 0
	for _, field := range ft.Params.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		if index < count+n {
			return field
		}
		count += n
	}

	return nil
}
