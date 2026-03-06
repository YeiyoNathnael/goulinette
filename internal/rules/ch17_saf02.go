package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type saf02Rule struct{}

func NewSAF02() Rule {
	return saf02Rule{}
}

func (saf02Rule) ID() string {
	return "SAF-02"
}

func (saf02Rule) Chapter() int {
	return 17
}

func (saf02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
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
			ast.Inspect(file, func(n ast.Node) bool {
				switch s := n.(type) {
				case *ast.GoStmt:
					diagnostics = append(diagnostics, saf02CheckGoStmt(s, pkg.TypesInfo, pkg.Fset)...)
				case *ast.AssignStmt:
					diagnostics = append(diagnostics, saf02CheckAssignStmt(s, pkg.TypesInfo, pkg.Fset)...)
				case *ast.DeclStmt:
					diagnostics = append(diagnostics, saf02CheckDeclStmt(s, pkg.TypesInfo, pkg.Fset)...)
				case *ast.ReturnStmt:
					diagnostics = append(diagnostics, saf02CheckReturnStmt(s, pkg.TypesInfo, pkg.Fset)...)
				}
				return true
			})
		}
	}

	return diagnostics, nil
}

func saf02CheckGoStmt(gs *ast.GoStmt, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if gs == nil || gs.Call == nil || info == nil {
		return nil
	}

	call := gs.Call
	if call.Fun == nil {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, arg := range call.Args {
		argType := info.TypeOf(arg)
		if !containsCopySensitiveValue(argType, map[types.Type]bool{}) {
			continue
		}
		pos := fset.Position(arg.Pos())
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "SAF-02",
			Severity: diag.SeverityError,
			Message:  "copy-sensitive sync/noCopy value must not be passed by value in goroutine launch",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "pass pointers for sync/noCopy-bearing values",
		})
	}

	return diagnostics
}

func saf02CheckAssignStmt(as *ast.AssignStmt, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if as == nil || info == nil {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for i, rhs := range as.Rhs {
		if i >= len(as.Lhs) {
			continue
		}
		lhs := as.Lhs[i]
		if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
			continue
		}

		lhsType := info.TypeOf(lhs)
		rhsType := info.TypeOf(rhs)
		if !containsCopySensitiveValue(lhsType, map[types.Type]bool{}) || !containsCopySensitiveValue(rhsType, map[types.Type]bool{}) {
			continue
		}

		pos := fset.Position(as.Pos())
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "SAF-02",
			Severity: diag.SeverityError,
			Message:  "copy-sensitive sync/noCopy value must not be copied by assignment",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "store and pass sync/noCopy-bearing values by pointer",
		})
	}

	return diagnostics
}

func saf02CheckDeclStmt(ds *ast.DeclStmt, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if ds == nil || info == nil {
		return nil
	}

	gd, ok := ds.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, rhs := range vs.Values {
			if i >= len(vs.Names) {
				continue
			}
			if vs.Names[i] != nil && vs.Names[i].Name == "_" {
				continue
			}

			lhsType := info.TypeOf(vs.Names[i])
			rhsType := info.TypeOf(rhs)
			if !containsCopySensitiveValue(lhsType, map[types.Type]bool{}) || !containsCopySensitiveValue(rhsType, map[types.Type]bool{}) {
				continue
			}

			pos := fset.Position(vs.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "SAF-02",
				Severity: diag.SeverityError,
				Message:  "copy-sensitive sync/noCopy value must not be copied in variable initialization",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use pointer values for sync/noCopy-bearing data",
			})
		}
	}

	return diagnostics
}

func saf02CheckReturnStmt(rs *ast.ReturnStmt, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if rs == nil || info == nil {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, result := range rs.Results {
		rt := info.TypeOf(result)
		if !containsCopySensitiveValue(rt, map[types.Type]bool{}) {
			continue
		}

		pos := fset.Position(result.Pos())
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "SAF-02",
			Severity: diag.SeverityError,
			Message:  "copy-sensitive sync/noCopy value must not be returned by value",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "return pointer types for sync/noCopy-bearing values",
		})
	}

	return diagnostics
}
