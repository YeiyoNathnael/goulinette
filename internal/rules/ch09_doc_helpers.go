package rules

import (
	"go/ast"
	"go/token"
	"strings"
)

type docTarget struct {
	Name           string
	Doc            *ast.CommentGroup
	Pos            token.Pos
	PrimaryForDoc3 bool
}

func collectExportedDocTargets(file *ast.File) []docTarget {
	targets := make([]docTarget, 0)
	for _, decl := range file.Decls {
		targets = append(targets, docTargetsForDecl(decl)...)
	}

	return targets
}

func docTargetsForDecl(decl ast.Decl) []docTarget {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return docTargetsForFuncDecl(d)
	case *ast.GenDecl:
		return docTargetsForGenDecl(d)
	default:
		return nil
	}
}

func docTargetsForFuncDecl(d *ast.FuncDecl) []docTarget {
	if d == nil || d.Name == nil || !d.Name.IsExported() {
		return nil
	}
	return []docTarget{{
		Name:           d.Name.Name,
		Doc:            d.Doc,
		Pos:            d.Name.Pos(),
		PrimaryForDoc3: true,
	}}
}

func docTargetsForGenDecl(d *ast.GenDecl) []docTarget {
	targets := make([]docTarget, 0)
	for _, spec := range d.Specs {
		targets = append(targets, docTargetsForSpec(spec, d.Doc)...)
	}
	return targets
}

func docTargetsForSpec(spec ast.Spec, declDoc *ast.CommentGroup) []docTarget {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return docTargetsForTypeSpec(s, declDoc)
	case *ast.ValueSpec:
		return docTargetsForValueSpec(s, declDoc)
	default:
		return nil
	}
}

func docTargetsForTypeSpec(s *ast.TypeSpec, declDoc *ast.CommentGroup) []docTarget {
	if s == nil || s.Name == nil || !s.Name.IsExported() {
		return nil
	}
	doc := s.Doc
	if doc == nil {
		doc = declDoc
	}
	return []docTarget{{
		Name:           s.Name.Name,
		Doc:            doc,
		Pos:            s.Name.Pos(),
		PrimaryForDoc3: true,
	}}
}

func docTargetsForValueSpec(s *ast.ValueSpec, declDoc *ast.CommentGroup) []docTarget {
	firstExportedIndex := firstExportedNameIndex(s)
	if firstExportedIndex == -1 {
		return nil
	}

	doc := s.Doc
	if doc == nil {
		doc = declDoc
	}

	targets := make([]docTarget, 0)
	for i, name := range s.Names {
		if name == nil || !name.IsExported() {
			continue
		}
		targets = append(targets, docTarget{
			Name:           name.Name,
			Doc:            doc,
			Pos:            name.Pos(),
			PrimaryForDoc3: i == firstExportedIndex,
		})
	}

	return targets
}

func firstExportedNameIndex(s *ast.ValueSpec) int {
	if s == nil {
		return -1
	}
	for i, name := range s.Names {
		if name != nil && name.IsExported() {
			return i
		}
	}
	return -1
}

func nearestCommentGroupBeforeLine(file *ast.File, fset *token.FileSet, line int) (*ast.CommentGroup, int) {
	var nearest *ast.CommentGroup
	nearestEndLine := -1

	for _, cg := range file.Comments {
		endLine := fset.Position(cg.End()).Line
		if endLine >= line {
			continue
		}
		if endLine > nearestEndLine {
			nearest = cg
			nearestEndLine = endLine
		}
	}

	return nearest, nearestEndLine
}

func firstDocWord(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	text := strings.TrimSpace(cg.Text())
	if text == "" {
		return ""
	}
	for _, field := range strings.Fields(text) {
		clean := strings.Trim(field, " \t\r\n.,:;!?()[]{}\"'`")
		if clean != "" {
			return clean
		}
	}
	return ""
}

func isBlockDocComment(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.HasPrefix(strings.TrimSpace(c.Text), "/*") {
			return true
		}
	}
	return false
}

func collectPackageVarNames(files []parsedFile) map[string]struct{} {
	vars := make(map[string]struct{})
	for _, pf := range files {
		collectPackageVarNamesFromDecls(vars, pf.File.Decls)
	}
	return vars
}

func collectPackageVarNamesFromDecls(vars map[string]struct{}, decls []ast.Decl) {
	for _, decl := range decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			collectPackageVarNamesFromSpec(vars, spec)
		}
	}
}

func collectPackageVarNamesFromSpec(vars map[string]struct{}, spec ast.Spec) {
	vs, ok := spec.(*ast.ValueSpec)
	if !ok {
		return
	}
	for _, name := range vs.Names {
		if name == nil || name.Name == "_" {
			continue
		}
		vars[name.Name] = struct{}{}
	}
}

func isImmutableExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		return true
	case *ast.BinaryExpr:
		return isImmutableExpr(e.X) && isImmutableExpr(e.Y)
	case *ast.UnaryExpr:
		return isImmutableExpr(e.X)
	case *ast.ParenExpr:
		return isImmutableExpr(e.X)
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			switch v := elt.(type) {
			case *ast.KeyValueExpr:
				if !isImmutableExpr(v.Key) || !isImmutableExpr(v.Value) {
					return false
				}
			default:
				if !isImmutableExpr(v) {
					return false
				}
			}
		}
		return true
	default:
		return false
	}
}

func isImmutableInitBody(body *ast.BlockStmt, packageVars map[string]struct{}) bool {
	if body == nil {
		return true
	}

	for _, stmt := range body.List {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			if s.Tok != token.ASSIGN {
				return false
			}
			if len(s.Lhs) != len(s.Rhs) || len(s.Lhs) == 0 {
				return false
			}
			for i := range s.Lhs {
				id, ok := s.Lhs[i].(*ast.Ident)
				if !ok {
					return false
				}
				if _, ok := packageVars[id.Name]; !ok {
					return false
				}
				if !isImmutableExpr(s.Rhs[i]) {
					return false
				}
			}

		case *ast.EmptyStmt:
			continue

		default:
			return false
		}
	}

	return true
}
