package rules

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

type tstFile struct {
	Path       string
	File       *ast.File
	FSet       *token.FileSet
	ImportPath map[string]string
	DotImports map[string]bool
}

type tstFuncInfo struct {
	Node         *ast.FuncDecl
	Name         string
	Path         string
	ImportPath   map[string]string
	TestingParam map[string]bool
	Subject      string
}

func collectTestFiles(ctx Context) ([]tstFile, error) {
	if len(ctx.Files) > 0 {
		parsed, err := parseFiles(ctx.Files)
		if err != nil {
			return nil, err
		}
		out := make([]tstFile, 0)
		for _, pf := range parsed {
			if !isTestFile(pf.Path) {
				continue
			}
			aliases, dots := collectImportAliases(pf.File)
			out = append(out, tstFile{
				Path:       pf.Path,
				File:       pf.File,
				FSet:       pf.FSet,
				ImportPath: aliases,
				DotImports: dots,
			})
		}
		return out, nil
	}

	if ctx.Root == "" {
		return nil, nil
	}

	paths := make([]string, 0)
	err := filepath.WalkDir(ctx.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if isTestFile(path) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	parsed, err := parseFiles(paths)
	if err != nil {
		return nil, err
	}
	out := make([]tstFile, 0, len(parsed))
	for _, pf := range parsed {
		aliases, dots := collectImportAliases(pf.File)
		out = append(out, tstFile{
			Path:       pf.Path,
			File:       pf.File,
			FSet:       pf.FSet,
			ImportPath: aliases,
			DotImports: dots,
		})
	}
	return out, nil
}

func collectImportAliases(file *ast.File) (map[string]string, map[string]bool) {
	aliases := make(map[string]string)
	dots := make(map[string]bool)
	if file == nil {
		return aliases, dots
	}
	for _, imp := range file.Imports {
		if imp == nil || imp.Path == nil {
			continue
		}
		path := strings.Trim(imp.Path.Value, "\"")
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		}
		switch name {
		case "":
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				aliases[parts[len(parts)-1]] = path
			}
		case ".":
			dots[path] = true
		case "_":
			// ignore blank imports
		default:
			aliases[name] = path
		}
	}
	return aliases, dots
}

func isTopLevelTestLike(name string) bool {
	return strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") || strings.HasPrefix(name, "Example")
}

func isRealTest(name string) bool {
	return strings.HasPrefix(name, "Test")
}

func testSubject(name string) string {
	if !isRealTest(name) {
		return ""
	}
	base := strings.TrimPrefix(name, "Test")
	if base == "" {
		return ""
	}
	if idx := strings.Index(base, "_"); idx > 0 {
		return base[:idx]
	}
	runes := []rune(base)
	if len(runes) <= 1 {
		return base
	}
	for i := 1; i < len(runes); i++ {
		if isUpper(runes[i]) && isLower(runes[i-1]) {
			return string(runes[:i])
		}
	}
	return base
}

func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }

func testingParamNames(ft *ast.FuncType, aliases map[string]string) map[string]bool {
	out := make(map[string]bool)
	if ft == nil || ft.Params == nil {
		return out
	}
	for _, p := range ft.Params.List {
		if p == nil || !isTestingParamTypeExpr(p.Type, aliases) {
			continue
		}
		for _, n := range p.Names {
			if n != nil && n.Name != "" {
				out[n.Name] = true
			}
		}
	}
	return out
}

func isTestingParamTypeExpr(expr ast.Expr, aliases map[string]string) bool {
	if expr == nil {
		return false
	}
	switch t := expr.(type) {
	case *ast.StarExpr:
		sel, ok := t.X.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || sel.Sel == nil {
			return false
		}
		alias, _ := aliases[id.Name]
		return alias == "testing" && sel.Sel.Name == "T"
	case *ast.SelectorExpr:
		id, ok := t.X.(*ast.Ident)
		if !ok || t.Sel == nil {
			return false
		}
		alias, _ := aliases[id.Name]
		return alias == "testing" && t.Sel.Name == "TB"
	default:
		return false
	}
}

func callsTestingMethods(body *ast.BlockStmt, params map[string]bool) bool {
	if body == nil || len(params) == 0 {
		return false
	}
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel == nil {
			return true
		}
		recv, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		active, _ := params[recv.Name]
		if !active {
			return true
		}
		switch sel.Sel.Name {
		case "Error", "Errorf", "Fatal", "Fatalf", "Log", "Logf":
			found = true
			return false
		default:
			return true
		}
	})
	return found
}

func firstStmtIsHelper(body *ast.BlockStmt, params map[string]bool) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	expr, ok := body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Helper" || len(call.Args) != 0 {
		return false
	}
	recv, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	active, _ := params[recv.Name]
	return active
}

func isTRunCall(call *ast.CallExpr, params map[string]bool) bool {
	if call == nil {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Run" {
		return false
	}
	recv, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	active, _ := params[recv.Name]
	if !active {
		return false
	}
	return true
}

func loadTypedPackagesWithTests(root string) ([]*packages.Package, error) {
	if root == "" {
		return nil, nil
	}
	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	pkgs, err := packages.Load(&packages.Config{
		Mode:  mode,
		Dir:   root,
		Tests: true,
	}, "./...")
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("type loading failed for package %s: %s", pkg.PkgPath, pkg.Errors[0].Msg)
		}
	}
	return pkgs, nil
}

func isTimeSleepCallTyped(call *ast.CallExpr, info *types.Info) bool {
	if call == nil || info == nil {
		return false
	}
	var obj types.Object
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if fn.Sel != nil {
			obj = info.Uses[fn.Sel]
		}
	case *ast.Ident:
		obj = info.Uses[fn]
	}
	f, ok := obj.(*types.Func)
	if !ok || f.Pkg() == nil {
		return false
	}
	return f.Pkg().Path() == "time" && f.Name() == "Sleep"
}

func isTimeSleepCallAST(call *ast.CallExpr, aliases map[string]string, dots map[string]bool) bool {
	if call == nil {
		return false
	}
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if fn.Sel == nil || fn.Sel.Name != "Sleep" {
			return false
		}
		recv, ok := fn.X.(*ast.Ident)
		if !ok {
			return false
		}
		alias, _ := aliases[recv.Name]
		return alias == "time"
	case *ast.Ident:
		if fn.Name != "Sleep" {
			return false
		}
		dot, _ := dots["time"]
		return dot
	default:
		return false
	}
}

func parseSingleFile(path string) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return file, fset, nil
}
