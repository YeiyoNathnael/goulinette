package rules

import (
	"go/ast"
	"go/types"
)

const errorTypeName = "error"

func builtinErrorType() types.Type {
	obj := types.Universe.Lookup(errorTypeName)
	if obj == nil {
		return nil
	}
	return obj.Type()
}

func isErrorInterfaceType(t types.Type) bool {
	if t == nil {
		return false
	}
	errType := builtinErrorType()
	if errType == nil {
		return false
	}
	return types.Identical(types.Unalias(t), types.Unalias(errType))
}

func isConcreteErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	if isErrorInterfaceType(t) {
		return false
	}
	if _, ok := types.Unalias(t).Underlying().(*types.Interface); ok {
		return false
	}

	errType := builtinErrorType()
	if errType == nil {
		return false
	}
	errIface, ok := types.Unalias(errType).Underlying().(*types.Interface)
	if !ok {
		return false
	}

	if types.Implements(t, errIface) {
		return true
	}

	if _, ok := t.(*types.Pointer); ok {
		return false
	}

	if types.Implements(types.NewPointer(t), errIface) {
		return true
	}

	return false
}

func functionResultFieldByIndex(fn *ast.FuncDecl, index int) *ast.Field {
	if fn == nil || fn.Type == nil || fn.Type.Results == nil || index < 0 {
		return nil
	}

	var count int
	for _, field := range fn.Type.Results.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		if index < count+n {
			return field
		}
		count += n
	}

	return nil
}
