package rules

import (
	"go/ast"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam07Rule struct{}

const nam07Chapter = 2

// NewNAM07 returns the NAM07 rule implementation.
func NewNAM07() Rule {
	return nam07Rule{}
}

// ID returns the rule identifier.
func (nam07Rule) ID() string {
	return ruleNAM07
}

// Chapter returns the chapter number for this rule.
func (nam07Rule) Chapter() int {
	return nam07Chapter
}

// Run executes this rule against the provided context.
func (nam07Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		pkg := strings.ToLower(pf.File.Name.Name)
		if pkg == "" {
			continue
		}

		for _, named := range exportedDeclNames(pf.File) {
			lname := strings.ToLower(named.Name.Name)
			if strings.HasPrefix(lname, pkg) || strings.HasSuffix(lname, pkg) {
				pos := pf.FSet.Position(named.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleNAM07,
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
		out = append(out, exportedDeclNamesForDecl(decl)...)
	}
	return out
}

func exportedDeclNamesForDecl(decl ast.Decl) []namedIdent {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return exportedNamesForFuncDecl(d)
	case *ast.GenDecl:
		return exportedNamesForGenDecl(d)
	default:
		return nil
	}
}

func exportedNamesForFuncDecl(d *ast.FuncDecl) []namedIdent {
	if d == nil || d.Name == nil || !d.Name.IsExported() {
		return nil
	}
	return []namedIdent{{Name: d.Name}}
}

func exportedNamesForGenDecl(d *ast.GenDecl) []namedIdent {
	out := make([]namedIdent, 0)
	for _, spec := range d.Specs {
		out = append(out, exportedNamesForSpec(spec)...)
	}
	return out
}

func exportedNamesForSpec(spec ast.Spec) []namedIdent {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Name == nil || !s.Name.IsExported() {
			return nil
		}
		return []namedIdent{{Name: s.Name}}
	case *ast.ValueSpec:
		out := make([]namedIdent, 0)
		for _, n := range s.Names {
			if n != nil && n.IsExported() {
				out = append(out, namedIdent{Name: n})
			}
		}
		return out
	default:
		return nil
	}
}
