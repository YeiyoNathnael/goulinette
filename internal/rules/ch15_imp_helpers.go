package rules

import (
	"os"
	"path/filepath"
	"strings"
)

type importClass int

const (
	importClassStd importClass = iota
	importClassThirdParty
	importClassInternal
)

func readModulePath(root string) string {
	if root == "" {
		return ""
	}
	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, "module "))
	}
	return ""
}

func classifyImportPath(path, modulePath string) importClass {
	path = strings.Trim(path, "\"")
	if modulePath != "" && (path == modulePath || strings.HasPrefix(path, modulePath+"/")) {
		return importClassInternal
	}
	first := path
	if idx := strings.Index(first, "/"); idx >= 0 {
		first = first[:idx]
	}
	if strings.Contains(first, ".") {
		return importClassThirdParty
	}
	return importClassStd
}

func defaultImportName(path string) string {
	path = strings.Trim(path, "\"")
	path = strings.TrimSuffix(path, "/")
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		path = path[idx+1:]
	}
	return path
}
