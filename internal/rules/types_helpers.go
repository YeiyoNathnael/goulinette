package rules

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

func loadTypedPackages(root string) ([]*packages.Package, error) {
	mode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedSyntax |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes

	pkgs, err := packages.Load(&packages.Config{
		Mode: mode,
		Dir:  root,
	}, "./...")
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			continue
		}
		return nil, fmt.Errorf("type loading failed for package %s: %s", pkg.PkgPath, pkg.Errors[0].Msg)
	}

	return pkgs, nil
}
