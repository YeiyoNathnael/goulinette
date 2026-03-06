package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type con01Rule struct{}

const (
	con01Chapter                      = 11
	con01MessageNoConcurrencyExposure = "public APIs must not expose channels or mutexes"
	con01HintExposes                  = " exposes "
	syncPkgPath                       = "sync"
	syncMutexType                     = "Mutex"
	syncRWMutexType                   = "RWMutex"
)

// NewCON01 returns the CON01 rule implementation.
func NewCON01() Rule {
	return con01Rule{}
}

// ID returns the rule identifier.
func (con01Rule) ID() string {
	return ruleCON01
}

// Chapter returns the chapter number for this rule.
func (con01Rule) Chapter() int {
	return con01Chapter
}

// Run executes this rule against the provided context.
func (con01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				diagnostics = append(diagnostics, con01DiagnosticsForDecl(pkg.Fset, pkg.TypesInfo, decl)...)
			}
		}
	}

	return diagnostics, nil
}

func con01DiagnosticsForDecl(fset *token.FileSet, info *types.Info, decl ast.Decl) []diag.Finding {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return con01DiagnosticsForFuncDecl(fset, info, d)
	case *ast.GenDecl:
		return con01DiagnosticsForGenDecl(fset, info, d)
	default:
		return nil
	}
}

func con01DiagnosticsForFuncDecl(fset *token.FileSet, info *types.Info, d *ast.FuncDecl) []diag.Finding {
	if d == nil || d.Name == nil || !d.Name.IsExported() {
		return nil
	}
	fnObj, ok := info.Defs[d.Name].(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := fnObj.Type().(*types.Signature)
	if !ok {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	diagnostics = append(diagnostics, con01ParamDiagnostics(fset, sig)...)
	diagnostics = append(diagnostics, con01ResultDiagnostics(fset, d, sig)...)
	return diagnostics
}

func con01ParamDiagnostics(fset *token.FileSet, sig *types.Signature) []diag.Finding {
	if sig.Params() == nil {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for i := 0; i < sig.Params().Len(); i++ {
		v := sig.Params().At(i)
		if bad, detail := containsForbiddenConcurrencyType(v.Type(), map[types.Type]bool{}); bad {
			pos := fset.Position(v.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleCON01,
				Severity: diag.SeverityError,
				Message:  con01MessageNoConcurrencyExposure,
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "parameter " + v.Name() + con01HintExposes + detail + "; hide concurrency primitives behind behavior",
			})
		}
	}

	return diagnostics
}

func con01ResultDiagnostics(fset *token.FileSet, d *ast.FuncDecl, sig *types.Signature) []diag.Finding {
	if sig.Results() == nil {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for i := 0; i < sig.Results().Len(); i++ {
		v := sig.Results().At(i)
		if bad, detail := containsForbiddenConcurrencyType(v.Type(), map[types.Type]bool{}); bad {
			pos := fset.Position(d.Name.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleCON01,
				Severity: diag.SeverityError,
				Message:  con01MessageNoConcurrencyExposure,
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "return value exposes " + detail + "; return a behavior-focused type instead",
			})
		}
	}

	return diagnostics
}

func con01DiagnosticsForGenDecl(fset *token.FileSet, info *types.Info, d *ast.GenDecl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, spec := range d.Specs {
		diagnostics = append(diagnostics, con01DiagnosticsForTypeSpec(fset, info, spec)...)
	}
	return diagnostics
}

func con01DiagnosticsForTypeSpec(fset *token.FileSet, info *types.Info, spec ast.Spec) []diag.Finding {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok || ts.Name == nil || !ts.Name.IsExported() {
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
		if bad, detail := containsForbiddenConcurrencyType(field.Type(), map[types.Type]bool{}); bad {
			pos := fset.Position(field.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleCON01,
				Severity: diag.SeverityError,
				Message:  con01MessageNoConcurrencyExposure,
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "field " + field.Name() + con01HintExposes + detail + "; keep synchronization internal",
			})
		}
	}

	return diagnostics
}

func containsForbiddenConcurrencyType(t types.Type, seen map[types.Type]bool) (bool, string) {
	if t == nil {
		return false, ""
	}
	if visited, ok := seen[t]; ok && visited {
		return false, ""
	}
	seen[t] = true

	switch tt := t.(type) {
	case *types.Chan:
		return true, "a channel"

	case *types.Pointer:
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Named:
		if isSyncMutexNamed(tt) {
			return true, "a sync mutex"
		}
		if bad, detail := containsForbiddenConcurrencyType(tt.Underlying(), seen); bad {
			return true, detail
		}
		return false, ""

	case *types.Alias:
		if bad, detail := containsForbiddenConcurrencyType(types.Unalias(tt), seen); bad {
			return true, detail
		}
		return false, ""

	case *types.Struct:
		return containsForbiddenInStruct(tt, seen)

	case *types.Array:
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Slice:
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Map:
		return containsForbiddenInMap(tt, seen)

	case *types.Signature:
		return containsForbiddenInSignature(tt, seen)
	default:
		return false, ""
	}
}

func containsForbiddenInStruct(st *types.Struct, seen map[types.Type]bool) (bool, string) {
	for i := 0; i < st.NumFields(); i++ {
		if bad, detail := containsForbiddenConcurrencyType(st.Field(i).Type(), seen); bad {
			return true, detail
		}
	}
	return false, ""
}

func containsForbiddenInMap(mt *types.Map, seen map[types.Type]bool) (bool, string) {
	if bad, detail := containsForbiddenConcurrencyType(mt.Key(), seen); bad {
		return true, detail
	}
	return containsForbiddenConcurrencyType(mt.Elem(), seen)
}

func containsForbiddenInSignature(sig *types.Signature, seen map[types.Type]bool) (bool, string) {
	if bad, detail := containsForbiddenInTuple(sig.Params(), seen); bad {
		return true, detail
	}
	return containsForbiddenInTuple(sig.Results(), seen)
}

func containsForbiddenInTuple(tuple *types.Tuple, seen map[types.Type]bool) (bool, string) {
	if tuple == nil {
		return false, ""
	}
	for i := 0; i < tuple.Len(); i++ {
		if bad, detail := containsForbiddenConcurrencyType(tuple.At(i).Type(), seen); bad {
			return true, detail
		}
	}
	return false, ""
}

func isSyncMutexNamed(named *types.Named) bool {
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	pkgPath := named.Obj().Pkg().Path()
	if pkgPath != syncPkgPath {
		return false
	}
	name := named.Obj().Name()
	return name == syncMutexType || name == syncRWMutexType
}
