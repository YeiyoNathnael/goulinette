package rules

import (
	"go/ast"
	"strings"

	"goulinette/internal/diag"
)

type typ07Rule struct{}

func NewTYP07() Rule {
	return typ07Rule{}
}

func (typ07Rule) ID() string {
	return "TYP-07"
}

func (typ07Rule) Chapter() int {
	return 7
}

func (typ07Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				diagnostics = append(diagnostics, checkAnyInFuncSignature(pf, d)...)
			case *ast.GenDecl:
				diagnostics = append(diagnostics, checkAnyInStructFields(pf, d)...)
			}
		}
	}

	return diagnostics, nil
}

func checkAnyInFuncSignature(pf parsedFile, fn *ast.FuncDecl) []diag.Diagnostic {
	out := make([]diag.Diagnostic, 0)
	if fn == nil || fn.Type == nil {
		return out
	}

	containers := []*ast.FieldList{fn.Type.TypeParams, fn.Type.Params, fn.Type.Results}
	for _, list := range containers {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			if !containsAnyType(field.Type) {
				continue
			}
			if hasAnyJustification(field.Doc, field.Comment, fn.Doc) {
				continue
			}

			pos := pf.FSet.Position(field.Pos())
			out = append(out, diag.Diagnostic{
				RuleID:   "TYP-07",
				Severity: diag.SeverityWarning,
				Message:  "any usage in function signature requires justification comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add inline or preceding comment explaining why any is required",
			})
		}
	}

	return out
}

func checkAnyInStructFields(pf parsedFile, gd *ast.GenDecl) []diag.Diagnostic {
	out := make([]diag.Diagnostic, 0)
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			continue
		}

		for _, field := range st.Fields.List {
			if !containsAnyType(field.Type) {
				continue
			}
			if hasAnyJustification(field.Doc, field.Comment, gd.Doc, ts.Doc) {
				continue
			}

			pos := pf.FSet.Position(field.Pos())
			out = append(out, diag.Diagnostic{
				RuleID:   "TYP-07",
				Severity: diag.SeverityWarning,
				Message:  "any usage in struct fields requires justification comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add inline or preceding comment explaining why any is required",
			})
		}
	}
	return out
}

func containsAnyType(expr ast.Expr) bool {
	if expr == nil {
		return false
	}

	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && id.Name == "any" {
			found = true
			return false
		}
		return true
	})
	return found
}

func hasAnyJustification(groups ...*ast.CommentGroup) bool {
	for _, g := range groups {
		if g == nil {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(g.Text()))
		if text == "" {
			continue
		}
		for _, kw := range []string{"because", "justif", "dynamic", "unknown", "generic", "json", "reflection", "unmarshal", "arbitrary", "external", "interop"} {
			if strings.Contains(text, kw) {
				return true
			}
		}
	}
	return false
}
