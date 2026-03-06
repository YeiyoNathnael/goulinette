package rules

import (
	"go/ast"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type cer01Rule struct{}

func NewCER01() Rule {
	return cer01Rule{}
}

func (cer01Rule) ID() string {
	return "CER-01"
}

func (cer01Rule) Chapter() int {
	return 12
}

func (cer01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil {
					continue
				}

				obj, ok := pkg.TypesInfo.Defs[fn.Name].(*types.Func)
				if !ok {
					continue
				}
				sig, ok := obj.Type().(*types.Signature)
				if !ok || sig.Results() == nil {
					continue
				}

				for i := 0; i < sig.Results().Len(); i++ {
					res := sig.Results().At(i)
					if !isConcreteErrorType(res.Type()) {
						continue
					}

					field := functionResultFieldByIndex(fn, i)
					posToken := fn.Name.Pos()
					if field != nil {
						posToken = field.Type.Pos()
					}

					pos := pkg.Fset.Position(posToken)
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "CER-01",
						Severity: diag.SeverityError,
						Message:  "functions must return error interface, not concrete custom error types",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "change the return type to error and return concrete errors as error values",
					})
				}
			}
		}
	}

	return diagnostics, nil
}
