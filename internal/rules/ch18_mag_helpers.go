package rules

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

type magAstUnit struct {
	File     *ast.File
	FSet     *token.FileSet
	Filename string
}

type magLiteralOccurrence struct {
	value  string
	pos    token.Position
	isTest bool
}

func collectMagAstUnits(ctx Context) ([]magAstUnit, error) {
	if len(ctx.Files) > 0 {
		parsed, err := parseFiles(ctx.Files)
		if err != nil {
			return nil, err
		}
		units := make([]magAstUnit, 0, len(parsed))
		for _, pf := range parsed {
			units = append(units, magAstUnit{File: pf.File, FSet: pf.FSet, Filename: pf.Path})
		}
		return units, nil
	}

	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	units := make([]magAstUnit, 0)
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		units = append(units, magAstUnitsFromPackage(pkg)...)
	}
	return units, nil
}

func magAstUnitsFromPackage(pkg *packages.Package) []magAstUnit {
	if pkg == nil {
		return nil
	}
	units := make([]magAstUnit, 0, len(pkg.Syntax))
	for i, f := range pkg.Syntax {
		if f == nil || pkg.Fset == nil {
			continue
		}
		filename := ""
		if i < len(pkg.CompiledGoFiles) {
			filename = pkg.CompiledGoFiles[i]
		}
		if filename == "" {
			filename = pkg.Fset.Position(f.Pos()).Filename
		}
		units = append(units, magAstUnit{File: f, FSet: pkg.Fset, Filename: filename})
	}
	return units
}

func astInspectWithStack(root ast.Node, fn func(ast.Node, []ast.Node)) {
	stack := make([]ast.Node, 0, 16)
	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}
		fn(n, stack)
		stack = append(stack, n)
		return true
	})
}

func astDirectParent(stack []ast.Node) ast.Node {
	if len(stack) == 0 {
		return nil
	}
	return stack[len(stack)-1]
}

func astHasAncestor(stack []ast.Node, pred func(ast.Node) bool) bool {
	for i := len(stack) - 1; i >= 0; i-- {
		if pred(stack[i]) {
			return true
		}
	}
	return false
}

func isInConstDecl(stack []ast.Node) bool {
	return astHasAncestor(stack, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		return ok && gd.Tok == token.CONST
	})
}

func isInImportSpec(stack []ast.Node) bool {
	return astHasAncestor(stack, func(n ast.Node) bool {
		_, ok := n.(*ast.ImportSpec)
		return ok
	})
}

func isStructTagLiteral(lit *ast.BasicLit, parent ast.Node) bool {
	field, ok := parent.(*ast.Field)
	if !ok || field.Tag == nil {
		return false
	}
	return field.Tag == lit
}

func isErrorsNewOrFmtErrorf(fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if id.Name == "errors" && sel.Sel.Name == "New" {
		return true
	}
	if id.Name == "fmt" && sel.Sel.Name == "Errorf" {
		return true
	}
	return false
}

func isDirectErrorMessageLiteral(lit *ast.BasicLit, parent ast.Node) bool {
	call, ok := parent.(*ast.CallExpr)
	if !ok || !isErrorsNewOrFmtErrorf(call.Fun) {
		return false
	}
	for _, arg := range call.Args {
		if arg == lit {
			return true
		}
	}
	return false
}

func shortStringLiteral(lit *ast.BasicLit, minLen int) bool {
	if lit == nil || lit.Kind != token.STRING {
		return false
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		raw := strings.Trim(lit.Value, "`\"")
		return len(raw) < minLen
	}
	return len(val) < minLen
}

func isTestFile(path string) bool {
	return strings.HasSuffix(filepath.Base(path), "_test.go")
}

func numericLiteralKey(lit *ast.BasicLit, parent ast.Node) string {
	if lit == nil {
		return ""
	}
	if u, ok := parent.(*ast.UnaryExpr); ok && u.Op == token.SUB && u.X == lit {
		return "-" + lit.Value
	}
	return lit.Value
}
