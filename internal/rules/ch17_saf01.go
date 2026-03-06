package rules

import (
	"go/ast"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type saf01Rule struct{}

const saf01Chapter = 17

// NewSAF01 returns the SAF01 rule implementation.
func NewSAF01() Rule {
	return saf01Rule{}
}

// ID returns the rule identifier.
func (saf01Rule) ID() string {
	return ruleSAF01
}

// Chapter returns the chapter number for this rule.
func (saf01Rule) Chapter() int {
	return saf01Chapter
}

// Run executes this rule against the provided context.
func (saf01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleSAF01,
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
