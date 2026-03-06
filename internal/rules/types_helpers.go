package rules

import (
	"fmt"
	"sync"

	"golang.org/x/tools/go/packages"
)

var (
	typedPackagesCacheMu sync.RWMutex
	typedPackagesCache   = make(map[string][]*packages.Package)
)

func loadTypedPackages(root string) ([]*packages.Package, error) {
	typedPackagesCacheMu.RLock()
	if cached, ok := typedPackagesCache[root]; ok {
		typedPackagesCacheMu.RUnlock()
		return cached, nil
	}
	typedPackagesCacheMu.RUnlock()

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

	typedPackagesCacheMu.Lock()
	typedPackagesCache[root] = pkgs
	typedPackagesCacheMu.Unlock()

	return pkgs, nil
}

func clearTypedPackagesCache() {
	typedPackagesCacheMu.Lock()
	typedPackagesCache = make(map[string][]*packages.Package)
	typedPackagesCacheMu.Unlock()
}
