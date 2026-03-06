package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type parsedFile struct {
	Path string
	FSet *token.FileSet
	File *ast.File
}

func parseFiles(paths []string) ([]parsedFile, error) {
	parsed := make([]parsedFile, 0, len(paths))
	for _, path := range paths {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, parsedFile{Path: path, FSet: fset, File: file})
	}
	return parsed, nil
}
