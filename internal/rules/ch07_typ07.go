package rules

import (
	"go/ast"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ07Rule struct{}

const (
	typ07Chapter           = 7
	typ07AnyTypeName       = "any"
	typ07JSONKeyword       = "json"
	typ07HintJustification = "add inline or preceding comment explaining why any is required"
)

var typ07JustificationKeywords = []string{"because", "justif", "dynamic", "unknown", "generic", typ07JSONKeyword, "reflection", "unmarshal", "arbitrary", "external", "interop"}

// NewTYP07 returns the TYP07 rule implementation.
func NewTYP07() Rule {
	return typ07Rule{}
}

// ID returns the rule identifier.
func (typ07Rule) ID() string {
	return ruleTYP07
}

// Chapter returns the chapter number for this rule.
func (typ07Rule) Chapter() int {
	return typ07Chapter
}

// Run executes this rule against the provided context.
func (typ07Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				diagnostics = append(diagnostics, checkAnyInFuncSignature(pf, d)...)
			case *ast.GenDecl:
				diagnostics = append(diagnostics, checkAnyInStructFields(pf, d)...)
			default:
				// no-op
			}
		}
	}

	return diagnostics, nil
}

func checkAnyInFuncSignature(pf parsedFile, fn *ast.FuncDecl) []diag.Finding {
	out := make([]diag.Finding, 0)
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
			out = append(out, diag.Finding{
				RuleID:   ruleTYP07,
				Severity: diag.SeverityWarning,
				Message:  "any usage in function signature requires justification comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     typ07HintJustification,
			})
		}
	}

	return out
}

func checkAnyInStructFields(pf parsedFile, gd *ast.GenDecl) []diag.Finding {
	out := make([]diag.Finding, 0)
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
			out = append(out, diag.Finding{
				RuleID:   ruleTYP07,
				Severity: diag.SeverityWarning,
				Message:  "any usage in struct fields requires justification comment",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     typ07HintJustification,
			})
		}
	}
	return out
}

func containsAnyType(expr ast.Expr) bool {
	if expr == nil {
		return false
	}

	var found bool
	ast.Inspect(expr, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && id.Name == typ07AnyTypeName {
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
		for _, kw := range typ07JustificationKeywords {
			if strings.Contains(text, kw) {
				return true
			}
		}
	}
	return false
}
