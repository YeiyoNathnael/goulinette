package rules

import (
	"go/ast"
	"go/types"

	"goulinette/internal/diag"
)

type con01Rule struct{}

func NewCON01() Rule {
	return con01Rule{}
}

func (con01Rule) ID() string {
	return "CON-01"
}

func (con01Rule) Chapter() int {
	return 11
}

func (con01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Name == nil || !d.Name.IsExported() {
						continue
					}
					fnObj, ok := pkg.TypesInfo.Defs[d.Name].(*types.Func)
					if !ok {
						continue
					}
					sig, ok := fnObj.Type().(*types.Signature)
					if !ok {
						continue
					}

					if sig.Params() != nil {
						for i := 0; i < sig.Params().Len(); i++ {
							v := sig.Params().At(i)
							if bad, detail := containsForbiddenConcurrencyType(v.Type(), map[types.Type]bool{}); bad {
								pos := pkg.Fset.Position(v.Pos())
								diagnostics = append(diagnostics, diag.Diagnostic{
									RuleID:   "CON-01",
									Severity: diag.SeverityError,
									Message:  "public APIs must not expose channels or mutexes",
									Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
									Hint:     "parameter " + v.Name() + " exposes " + detail + "; hide concurrency primitives behind behavior",
								})
							}
						}
					}

					if sig.Results() != nil {
						for i := 0; i < sig.Results().Len(); i++ {
							v := sig.Results().At(i)
							if bad, detail := containsForbiddenConcurrencyType(v.Type(), map[types.Type]bool{}); bad {
								pos := pkg.Fset.Position(d.Name.Pos())
								diagnostics = append(diagnostics, diag.Diagnostic{
									RuleID:   "CON-01",
									Severity: diag.SeverityError,
									Message:  "public APIs must not expose channels or mutexes",
									Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
									Hint:     "return value exposes " + detail + "; return a behavior-focused type instead",
								})
							}
						}
					}

				case *ast.GenDecl:
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok || ts.Name == nil || !ts.Name.IsExported() {
							continue
						}

						obj, ok := pkg.TypesInfo.Defs[ts.Name].(*types.TypeName)
						if !ok {
							continue
						}

						st, ok := obj.Type().Underlying().(*types.Struct)
						if !ok {
							continue
						}

						for i := 0; i < st.NumFields(); i++ {
							field := st.Field(i)
							if bad, detail := containsForbiddenConcurrencyType(field.Type(), map[types.Type]bool{}); bad {
								pos := pkg.Fset.Position(field.Pos())
								diagnostics = append(diagnostics, diag.Diagnostic{
									RuleID:   "CON-01",
									Severity: diag.SeverityError,
									Message:  "public APIs must not expose channels or mutexes",
									Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
									Hint:     "field " + field.Name() + " exposes " + detail + "; keep synchronization internal",
								})
							}
						}
					}
				}
			}
		}
	}

	return diagnostics, nil
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
		for i := 0; i < tt.NumFields(); i++ {
			if bad, detail := containsForbiddenConcurrencyType(tt.Field(i).Type(), seen); bad {
				return true, detail
			}
		}
		return false, ""

	case *types.Array:
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Slice:
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Map:
		if bad, detail := containsForbiddenConcurrencyType(tt.Key(), seen); bad {
			return true, detail
		}
		return containsForbiddenConcurrencyType(tt.Elem(), seen)

	case *types.Signature:
		if tt.Params() != nil {
			for i := 0; i < tt.Params().Len(); i++ {
				if bad, detail := containsForbiddenConcurrencyType(tt.Params().At(i).Type(), seen); bad {
					return true, detail
				}
			}
		}
		if tt.Results() != nil {
			for i := 0; i < tt.Results().Len(); i++ {
				if bad, detail := containsForbiddenConcurrencyType(tt.Results().At(i).Type(), seen); bad {
					return true, detail
				}
			}
		}
		return false, ""
	}

	return false, ""
}

func isSyncMutexNamed(named *types.Named) bool {
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	pkgPath := named.Obj().Pkg().Path()
	if pkgPath != "sync" {
		return false
	}
	name := named.Obj().Name()
	return name == "Mutex" || name == "RWMutex"
}
