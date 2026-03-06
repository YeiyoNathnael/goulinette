package rules

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type fun03Rule struct{}

const (
	fun03Chapter = 5
	fun03ErrName = "error"
)

// NewFUN03 returns the FUN03 rule implementation.
func NewFUN03() Rule {
	return fun03Rule{}
}

// ID returns the rule identifier.
func (fun03Rule) ID() string {
	return ruleFUN03
}

// Chapter returns the chapter number for this rule.
func (fun03Rule) Chapter() int {
	return fun03Chapter
}

// Run executes this rule against the provided context.
func (fun03Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, decl := range syntaxFile.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil {
					continue
				}
				if !fn.Name.IsExported() {
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
				if shouldSkipConcreteParamCheck(fn, sig) {
					continue
				}

				if shouldWarnConcreteParamsTyped(sig) {
					pos := pkg.Fset.Position(fn.Name.Pos())
					diagnostics = append(diagnostics, diag.Finding{
						RuleID:   ruleFUN03,
						Severity: diag.SeverityWarning,
						Message:  "function accepts concrete types only; consider interface parameters for decoupling",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "accept interfaces where behavior contracts are sufficient",
					})
				}

				if shouldWarnInterfaceReturnTyped(sig) {
					pos := pkg.Fset.Position(fn.Name.Pos())
					diagnostics = append(diagnostics, diag.Finding{
						RuleID:   ruleFUN03,
						Severity: diag.SeverityWarning,
						Message:  "function returns interface type; consider returning concrete type",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "return concrete structs unless polymorphism is required",
					})
				}
			}
		}
	}

	return diagnostics, nil
}

func shouldWarnConcreteParamsTyped(sig *types.Signature) bool {
	params := sig.Params()
	if params == nil || params.Len() == 0 {
		return false
	}

	var hasConcrete bool
	var hasInterface bool
	for i := 0; i < params.Len(); i++ {
		typ := params.At(i).Type()
		if isInterfaceType(typ) {
			hasInterface = true
			continue
		}
		if isLikelyConcreteDomainType(typ) {
			hasConcrete = true
		}
	}

	return hasConcrete && !hasInterface
}

func shouldSkipConcreteParamCheck(fn *ast.FuncDecl, sig *types.Signature) bool {
	if fn == nil || sig == nil {
		return false
	}
	if fn.Recv != nil {
		return true
	}
	if fn.Name != nil && strings.HasPrefix(fn.Name.Name, "New") {
		return true
	}
	return fn.Name != nil && fn.Name.Name == "Run" && sig.Recv() != nil
}

func shouldWarnInterfaceReturnTyped(sig *types.Signature) bool {
	if sig == nil {
		return false
	}
	if sig.Params() == nil || sig.Params().Len() == 0 {
		return false
	}

	results := sig.Results()
	if results == nil {
		return false
	}

	for i := 0; i < results.Len(); i++ {
		typ := results.At(i).Type()
		if isErrorType(typ) {
			continue
		}
		if isInterfaceType(typ) {
			return true
		}
	}

	return false
}

func isInterfaceType(t types.Type) bool {
	_, ok := t.Underlying().(*types.Interface)
	return ok
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup(fun03ErrName)
	if errObj == nil {
		return false
	}
	return types.Identical(t, errObj.Type())
}

func isLikelyConcreteDomainType(t types.Type) bool {
	if t == nil {
		return false
	}

	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			break
		}
		t = ptr.Elem()
	}

	switch u := t.Underlying().(type) {
	case *types.Struct:
		return true
	case *types.Interface:
		return false
	case *types.Basic:
		return false
	case *types.Slice, *types.Array, *types.Map, *types.Chan, *types.Signature:
		return false
	default:
		_ = u
	}

	if named, ok := t.(*types.Named); ok {
		if _, isStruct := named.Underlying().(*types.Struct); isStruct {
			return true
		}
	}

	return false
}
