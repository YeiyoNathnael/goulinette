package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctx01Rule struct{}

const (
	ctx01Chapter = 14
)

// NewCTX01 returns the CTX01 rule implementation.
func NewCTX01() Rule {
	return ctx01Rule{}
}

// ID returns the rule identifier.
func (ctx01Rule) ID() string {
	return ruleCTX01
}

// Chapter returns the chapter number for this rule.
func (ctx01Rule) Chapter() int {
	return ctx01Chapter
}

// Run executes this rule against the provided context.
func (ctx01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			diagnostics = append(diagnostics, ctx01DiagnosticsForDecls(pkg.Fset, pkg.TypesInfo, file.Decls)...)
			diagnostics = append(diagnostics, ctx01DiagnosticsForFuncLits(pkg.Fset, pkg.TypesInfo, file)...)
		}
	}

	return diagnostics, nil
}

func ctx01DiagnosticsForDecls(fset *token.FileSet, info *types.Info, decls []ast.Decl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range decls {
		diagnostics = append(diagnostics, ctx01DiagnosticsForDecl(fset, info, decl)...)
	}
	return diagnostics
}

func ctx01DiagnosticsForDecl(fset *token.FileSet, info *types.Info, decl ast.Decl) []diag.Finding {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return ctx01DiagnosticsForFuncDecl(fset, info, d)
	case *ast.GenDecl:
		return ctx01DiagnosticsForInterfaceDecl(fset, info, d)
	default:
		return nil
	}
}

func ctx01DiagnosticsForFuncDecl(fset *token.FileSet, info *types.Info, d *ast.FuncDecl) []diag.Finding {
	if d == nil || d.Name == nil {
		return nil
	}
	obj, ok := info.Defs[d.Name].(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return nil
	}
	return ctx01DiagnosticsForSignature(fset, sig, d.Type, d.Name.Pos())
}

func ctx01DiagnosticsForInterfaceDecl(fset *token.FileSet, info *types.Info, d *ast.GenDecl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, spec := range d.Specs {
		diagnostics = append(diagnostics, ctx01DiagnosticsForTypeSpec(fset, info, spec)...)
	}
	return diagnostics
}

func ctx01DiagnosticsForTypeSpec(fset *token.FileSet, info *types.Info, spec ast.Spec) []diag.Finding {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil
	}
	iface, ok := ts.Type.(*ast.InterfaceType)
	if !ok || iface.Methods == nil {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, m := range iface.Methods.List {
		ft, ok := m.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		t := info.TypeOf(ft)
		sig, ok := t.(*types.Signature)
		if !ok {
			continue
		}

		pos := m.Type.Pos()
		if len(m.Names) > 0 {
			pos = m.Names[0].Pos()
		}
		diagnostics = append(diagnostics, ctx01DiagnosticsForSignature(fset, sig, ft, pos)...)
	}

	return diagnostics
}

func ctx01DiagnosticsForFuncLits(fset *token.FileSet, info *types.Info, file *ast.File) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.FuncLit)
		if !ok {
			return true
		}
		t := info.TypeOf(lit.Type)
		sig, ok := t.(*types.Signature)
		if !ok {
			return true
		}
		diagnostics = append(diagnostics, ctx01DiagnosticsForSignature(fset, sig, lit.Type, lit.Type.Func)...)
		return true
	})
	return diagnostics
}

func ctx01DiagnosticsForSignature(fset *token.FileSet, sig *types.Signature, ft *ast.FuncType, fallbackPos token.Pos) []diag.Finding {
	if sig == nil || sig.Params() == nil {
		return nil
	}

	diags := make([]diag.Finding, 0)
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
		diags = append(diags, diag.Finding{
			RuleID:   ruleCTX01,
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

	var count int
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
