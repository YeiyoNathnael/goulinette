package rules

import (
	"fmt"
	"sync"

	"golang.org/x/tools/go/packages"
)

var (
	typedPackagesCacheMu sync.RWMutex = sync.RWMutex{}
	typedPackagesCache   sync.Map     = sync.Map{}
)

func loadTypedPackages(root string) ([]*packages.Package, error) {
	typedPackagesCacheMu.RLock()
	if cached, ok := typedPackagesCache.Load(root); ok {
		typedPackagesCacheMu.RUnlock()
		if pkgs, ok := cached.([]*packages.Package); ok {
			return pkgs, nil
		}
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
	typedPackagesCache.Store(root, pkgs)
	typedPackagesCacheMu.Unlock()

	return pkgs, nil
}

func clearTypedPackagesCache() {
	typedPackagesCacheMu.Lock()
	typedPackagesCache.Range(func(key, value any) bool {
		typedPackagesCache.Delete(key)
		return true
	})
	typedPackagesCacheMu.Unlock()
}
