package rules

import (
	"go/ast"
	"go/types"

	"goulinette/internal/diag"
)

type saf01Rule struct{}

func NewSAF01() Rule {
	return saf01Rule{}
}

func (saf01Rule) ID() string {
	return "SAF-01"
}

func (saf01Rule) Chapter() int {
	return 17
}

func (saf01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		if pkg == nil || pkg.TypesInfo == nil {
			continue
		}

		for _, file := range pkg.Syntax {
			if file == nil {
				continue
			}

			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
					continue
				}

				recvExpr := fn.Recv.List[0].Type
				recvType := pkg.TypesInfo.TypeOf(recvExpr)
				named, isPtr := receiverBaseNamed(recvType)
				if named == nil {
					continue
				}

				if isPtr {
					continue
				}

				if !containsCopySensitiveValue(named, map[types.Type]bool{}) {
					continue
				}

				name := "method"
				if fn.Name != nil && fn.Name.Name != "" {
					name = fn.Name.Name
				}
				pos := pkg.Fset.Position(fn.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "SAF-01",
					Severity: diag.SeverityError,
					Message:  "method " + name + " on copy-sensitive type must use pointer receiver",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "change receiver from value to pointer to avoid copying synchronization/noCopy state",
				})
			}
		}
	}

	return diagnostics, nil
}
