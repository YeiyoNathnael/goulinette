package rules

import "go/ast"

func collectDeclaredIdents(file *ast.File) []*ast.Ident {
	out := make([]*ast.Ident, 0)
	ast.Inspect(file, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id == nil || id.Obj == nil {
			return true
		}
		out = append(out, id)
		return true
	})
	return out
}
