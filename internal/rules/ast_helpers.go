package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"sync"
)

type parsedFile struct {
	Path string
	FSet *token.FileSet
	File *ast.File
}

var (
	parsedFilesCacheMu sync.RWMutex = sync.RWMutex{}
	parsedFilesCache   sync.Map     = sync.Map{}
)

func parseFiles(paths []string) ([]parsedFile, error) {
	cacheKey := parsedFilesCacheKey(paths)

	parsedFilesCacheMu.RLock()
	if cached, ok := parsedFilesCache.Load(cacheKey); ok {
		parsedFilesCacheMu.RUnlock()
		if files, ok := cached.([]parsedFile); ok {
			return files, nil
		}
	}
	parsedFilesCacheMu.RUnlock()

	parsed := make([]parsedFile, 0, len(paths))
	for _, path := range paths {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, parsedFile{Path: path, FSet: fset, File: file})
	}

	parsedFilesCacheMu.Lock()
	parsedFilesCache.Store(cacheKey, parsed)
	parsedFilesCacheMu.Unlock()

	return parsed, nil
}

func parsedFilesCacheKey(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	cpy := append([]string(nil), paths...)
	sort.Strings(cpy)
	return strings.Join(cpy, "\n")
}

func clearParseFilesCache() {
	parsedFilesCacheMu.Lock()
	parsedFilesCache.Range(func(key, value any) bool {
		parsedFilesCache.Delete(key)
		return true
	})
	parsedFilesCacheMu.Unlock()
}
