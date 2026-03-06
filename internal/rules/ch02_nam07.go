package rules

import (
	"go/ast"
	"strings"

	"goulinette/internal/diag"
)

type nam07Rule struct{}

func NewNAM07() Rule {
	return nam07Rule{}
}

func (nam07Rule) ID() string {
	return "NAM-07"
}

func (nam07Rule) Chapter() int {
	return 2
}

func (nam07Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		pkg := strings.ToLower(pf.File.Name.Name)
		if pkg == "" {
			continue
		}

		for _, named := range exportedDeclNames(pf.File) {
			lname := strings.ToLower(named.Name.Name)
			if strings.HasPrefix(lname, pkg) || strings.HasSuffix(lname, pkg) {
				pos := pf.FSet.Position(named.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "NAM-07",
					Severity: diag.SeverityError,
					Message:  "exported identifier must not stutter the package name",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "remove package-name prefix/suffix from exported identifier",
				})
			}
		}
	}

	return diagnostics, nil
}

type namedIdent struct {
	Name *ast.Ident
}

func exportedDeclNames(file *ast.File) []namedIdent {
	out := make([]namedIdent, 0)
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil && d.Name.IsExported() {
				out = append(out, namedIdent{Name: d.Name})
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name != nil && s.Name.IsExported() {
						out = append(out, namedIdent{Name: s.Name})
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n != nil && n.IsExported() {
							out = append(out, namedIdent{Name: n})
						}
					}
				}
			}
		}
	}
	return out
}
