package rules

import "go/types"

func isStrictContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	named, ok := t.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "context" && named.Obj().Name() == "Context"
}

func isContextTypeOrPointer(t types.Type) bool {
	if isStrictContextType(t) {
		return true
	}
	ptr, ok := types.Unalias(t).(*types.Pointer)
	if !ok {
		return false
	}
	return isStrictContextType(ptr.Elem())
}
