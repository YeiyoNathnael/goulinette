package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
	"goulinette/internal/diag"
)

type cer02Rule struct{}

func NewCER02() Rule {
	return cer02Rule{}
}

func (cer02Rule) ID() string {
	return "CER-02"
}

func (cer02Rule) Chapter() int {
	return 12
}

func (cer02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		concreteErrorTypes := collectPackageConcreteErrorTypes(pkg)
		if len(concreteErrorTypes) == 0 {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				gd, ok := n.(*ast.GenDecl)
				if !ok || gd.Tok != token.VAR {
					return true
				}

				for _, spec := range gd.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range vs.Names {
						if name == nil || name.Name == "_" {
							continue
						}

						obj, ok := pkg.TypesInfo.Defs[name].(*types.Var)
						if !ok || !typeInSet(obj.Type(), concreteErrorTypes) {
							continue
						}

						pos := pkg.Fset.Position(name.Pos())
						diagnostics = append(diagnostics, diag.Diagnostic{
							RuleID:   "CER-02",
							Severity: diag.SeverityError,
							Message:  "custom error variables must be declared as error interface, not concrete error types",
							Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
							Hint:     "declare variable as error (e.g., var " + name.Name + " error)",
						})
					}
				}

				return true
			})
		}
	}

	return diagnostics, nil
}

func collectPackageConcreteErrorTypes(pkg *packages.Package) []types.Type {
	result := make([]types.Type, 0)
	if pkg == nil || pkg.Types == nil || pkg.Types.Scope() == nil {
		return result
	}

	seen := make(map[string]bool)
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		if isConcreteErrorType(tn.Type()) {
			key := types.TypeString(tn.Type(), nil)
			if !seen[key] {
				seen[key] = true
				result = append(result, tn.Type())
			}
		}

		ptr := types.NewPointer(tn.Type())
		if isConcreteErrorType(ptr) {
			key := types.TypeString(ptr, nil)
			if !seen[key] {
				seen[key] = true
				result = append(result, ptr)
			}
		}
	}

	return result
}

func typeInSet(t types.Type, set []types.Type) bool {
	for _, candidate := range set {
		if types.Identical(types.Unalias(t), types.Unalias(candidate)) {
			return true
		}
	}
	return false
}
