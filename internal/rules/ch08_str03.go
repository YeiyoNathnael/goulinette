package rules

import (
	"go/ast"
	"go/types"
	"strings"

	"goulinette/internal/diag"

	"golang.org/x/tools/go/packages"
)

type str03Rule struct{}

func NewSTR03() Rule {
	return str03Rule{}
}

func (str03Rule) ID() string {
	return "STR-03"
}

func (str03Rule) Chapter() int {
	return 8
}

func (str03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		localIfaces := collectLocalInterfaces(pkg)
		for _, syntaxFile := range pkg.Syntax {
			for _, decl := range syntaxFile.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name == nil || fn.Recv == nil {
					continue
				}
				if !strings.HasPrefix(fn.Name.Name, "Get") && !strings.HasPrefix(fn.Name.Name, "Set") {
					continue
				}

				fobj, ok := pkg.TypesInfo.Defs[fn.Name].(*types.Func)
				if !ok {
					continue
				}
				sig, ok := fobj.Type().(*types.Signature)
				if !ok || sig.Recv() == nil {
					continue
				}
				recvNamed := receiverNamedType(sig.Recv().Type())
				if recvNamed == nil {
					continue
				}

				if methodUsedByAnyInterface(recvNamed, fn.Name.Name, localIfaces) {
					continue
				}

				pos := pkg.Fset.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "STR-03",
					Severity: diag.SeverityWarning,
					Message:  "getter/setter methods should be avoided unless required by an interface",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer direct field access or behavior-focused methods",
				})
			}
		}
	}

	return diagnostics, nil
}

func collectLocalInterfaces(pkg *packages.Package) []*types.Interface {
	ifaces := make([]*types.Interface, 0)
	if pkg == nil || pkg.Types == nil || pkg.Types.Scope() == nil {
		return ifaces
	}

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			continue
		}
		ifaces = append(ifaces, iface.Complete())
	}

	return ifaces
}

func receiverNamedType(t types.Type) *types.Named {
	if t == nil {
		return nil
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, _ := t.(*types.Named)
	return named
}

func methodUsedByAnyInterface(recvNamed *types.Named, methodName string, ifaces []*types.Interface) bool {
	if recvNamed == nil {
		return false
	}
	ptrRecv := types.NewPointer(recvNamed)

	for _, iface := range ifaces {
		if iface == nil {
			continue
		}
		hasMethod := false
		for i := 0; i < iface.NumMethods(); i++ {
			if iface.Method(i).Name() == methodName {
				hasMethod = true
				break
			}
		}
		if !hasMethod {
			continue
		}
		if types.Implements(recvNamed, iface) || types.Implements(ptrRecv, iface) {
			return true
		}
	}
	return false
}
