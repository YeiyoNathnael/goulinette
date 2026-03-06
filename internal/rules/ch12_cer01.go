package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type cer01Rule struct{}

const (
	cer01Chapter = 12
)

// NewCER01 returns the CER01 rule implementation.
func NewCER01() Rule {
	return cer01Rule{}
}

// ID returns the rule identifier.
func (cer01Rule) ID() string {
	return ruleCER01
}

// Chapter returns the chapter number for this rule.
func (cer01Rule) Chapter() int {
	return cer01Chapter
}

// Run executes this rule against the provided context.
func (cer01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				diagnostics = append(diagnostics, cer01DiagnosticsForDecl(pkg.Fset, pkg.TypesInfo, decl)...)
			}
		}
	}

	return diagnostics, nil
}

func cer01DiagnosticsForDecl(fset *token.FileSet, info *types.Info, decl ast.Decl) []diag.Finding {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Name == nil {
		return nil
	}

	obj, ok := info.Defs[fn.Name].(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok || sig.Results() == nil {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
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

		pos := fset.Position(posToken)
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleCER01,
			Severity: diag.SeverityError,
			Message:  "functions must return error interface, not concrete custom error types",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "change the return type to error and return concrete errors as error values",
		})
	}

	return diagnostics
}
