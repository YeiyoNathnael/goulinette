package rules

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type str04Rule struct{}

func NewSTR04() Rule {
	return str04Rule{}
}

func (str04Rule) ID() string {
	return "STR-04"
}

func (str04Rule) Chapter() int {
	return 8
}

func (str04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			filename := pkg.Fset.Position(syntaxFile.Pos()).Filename
			if isGeneratedSourceFile(filename, syntaxFile) {
				continue
			}

			ast.Inspect(syntaxFile, func(n ast.Node) bool {
				cl, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				if len(cl.Elts) == 0 || allKeyValueElts(cl.Elts) {
					return true
				}
				if !isStructCompositeLit(pkg.TypesInfo.TypeOf(cl)) {
					return true
				}
				if isExternalStructLit(pkg.TypesInfo.TypeOf(cl), pkg.Types) {
					return true
				}

				pos := pkg.Fset.Position(cl.Lbrace)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "STR-04",
					Severity: diag.SeverityError,
					Message:  "struct literals must use named fields",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace positional literal with keyed fields, e.g. T{Field: value}",
				})
				return true
			})
		}
	}

	return diagnostics, nil
}

func allKeyValueElts(elts []ast.Expr) bool {
	if len(elts) == 0 {
		return true
	}
	for _, elt := range elts {
		if _, ok := elt.(*ast.KeyValueExpr); !ok {
			return false
		}
	}
	return true
}

func isStructCompositeLit(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Struct)
	return ok
}

func isExternalStructLit(t types.Type, currentPkg *types.Package) bool {
	if t == nil || currentPkg == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() != currentPkg.Path()
}

func isGeneratedSourceFile(filename string, file *ast.File) bool {
	if strings.HasSuffix(filename, ".pb.go") {
		return true
	}
	if strings.Contains(strings.ToLower(filepath.Base(filename)), "generated") {
		return true
	}
	for _, cg := range file.Comments {
		text := strings.ToLower(cg.Text())
		if strings.Contains(text, "code generated") && strings.Contains(text, "do not edit") {
			return true
		}
	}
	return false
}
