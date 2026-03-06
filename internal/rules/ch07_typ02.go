package rules

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ02Rule struct{}

const typ02Chapter = 7

// NewTYP02 returns the TYP02 rule implementation.
func NewTYP02() Rule {
	return typ02Rule{}
}

// ID returns the rule identifier.
func (typ02Rule) ID() string {
	return ruleTYP02
}

// Chapter returns the chapter number for this rule.
func (typ02Rule) Chapter() int {
	return typ02Chapter
}

// Run executes this rule against the provided context.
func (typ02Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, decl := range syntaxFile.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil || fn.Body == nil {
					continue
				}
				if fn.Recv != nil {
					continue
				}
				if !strings.HasPrefix(fn.Name.Name, "New") {
					continue
				}

				obj, ok := pkg.TypesInfo.Defs[fn.Name].(*types.Func)
				if !ok {
					continue
				}
				sig, ok := obj.Type().(*types.Signature)
				if !ok {
					continue
				}
				if !returnsOnlyErrorOrNothing(sig) {
					continue
				}

				paramName := findStructPointerParamName(fn, pkg.TypesInfo)
				if paramName == "" {
					continue
				}
				if !assignsToParamFields(fn.Body, paramName) {
					continue
				}

				pos := pkg.Fset.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleTYP02,
					Severity: diag.SeverityWarning,
					Message:  "constructor-shaped function should return a struct value instead of populating a pointer parameter",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "return the constructed value directly, e.g., NewX() (X, error)",
				})
			}
		}
	}

	return diagnostics, nil
}

func returnsOnlyErrorOrNothing(sig *types.Signature) bool {
	res := sig.Results()
	if res == nil || res.Len() == 0 {
		return true
	}
	if res.Len() == 1 && isErrorType(res.At(0).Type()) {
		return true
	}
	return false
}

func findStructPointerParamName(fn *ast.FuncDecl, info *types.Info) string {
	if fn.Type == nil || fn.Type.Params == nil {
		return ""
	}
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			if name == nil {
				continue
			}
			t := info.TypeOf(name)
			ptr, ok := t.Underlying().(*types.Pointer)
			if !ok {
				continue
			}
			if _, ok := ptr.Elem().Underlying().(*types.Struct); ok {
				return name.Name
			}
		}
	}
	return ""
}

func assignsToParamFields(body *ast.BlockStmt, paramName string) bool {
	var found bool
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range as.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			id, ok := sel.X.(*ast.Ident)
			if ok && id.Name == paramName {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
