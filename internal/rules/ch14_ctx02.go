package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctx02Rule struct{}

const (
	ctx02Chapter = 14
)

// NewCTX02 returns the CTX02 rule implementation.
func NewCTX02() Rule {
	return ctx02Rule{}
}

// ID returns the rule identifier.
func (ctx02Rule) ID() string {
	return ruleCTX02
}

// Chapter returns the chapter number for this rule.
func (ctx02Rule) Chapter() int {
	return ctx02Chapter
}

// Run executes this rule against the provided context.
func (ctx02Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			diagnostics = append(diagnostics, ctx02DiagnosticsForDecls(pkg.Fset, pkg.TypesInfo, file.Decls)...)
		}
	}

	return diagnostics, nil
}

func ctx02DiagnosticsForDecls(fset *token.FileSet, info *types.Info, decls []ast.Decl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		diagnostics = append(diagnostics, ctx02DiagnosticsForGenDecl(fset, info, gd)...)
	}
	return diagnostics
}

func ctx02DiagnosticsForGenDecl(fset *token.FileSet, info *types.Info, gd *ast.GenDecl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, spec := range gd.Specs {
		diagnostics = append(diagnostics, ctx02DiagnosticsForSpec(fset, info, spec)...)
	}
	return diagnostics
}

func ctx02DiagnosticsForSpec(fset *token.FileSet, info *types.Info, spec ast.Spec) []diag.Finding {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok || ts.Name == nil {
		return nil
	}

	obj, ok := info.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return nil
	}

	st, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		if !isContextTypeOrPointer(field.Type()) {
			continue
		}

		pos := fset.Position(field.Pos())
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleCTX02,
			Severity: diag.SeverityError,
			Message:  "context.Context must not be stored in struct fields",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "pass context explicitly via function parameters instead of storing it",
		})
	}

	return diagnostics
}
