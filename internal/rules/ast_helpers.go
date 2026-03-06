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
	parsedFilesCacheMu sync.RWMutex
	parsedFilesCache   = make(map[string][]parsedFile)
)

func parseFiles(paths []string) ([]parsedFile, error) {
	cacheKey := parsedFilesCacheKey(paths)

	parsedFilesCacheMu.RLock()
	if cached, ok := parsedFilesCache[cacheKey]; ok {
		parsedFilesCacheMu.RUnlock()
		return cached, nil
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
	parsedFilesCache[cacheKey] = parsed
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
	parsedFilesCache = make(map[string][]parsedFile)
	parsedFilesCacheMu.Unlock()
}
