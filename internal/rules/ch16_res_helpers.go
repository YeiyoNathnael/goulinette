package rules

import (
	"go/token"
	"go/types"
)

func closerInterfaceType() *types.Interface {
	errObj := types.Universe.Lookup("error")
	if errObj == nil {
		return nil
	}
	errType := errObj.Type()
	results := types.NewTuple(types.NewVar(token.NoPos, nil, "", errType))
	sig := types.NewSignatureType(nil, nil, nil, nil, results, false)
	closeFn := types.NewFunc(token.NoPos, nil, "Close", sig)
	iface := types.NewInterfaceType([]*types.Func{closeFn}, nil)
	iface.Complete()
	return iface
}

func isCloserType(t types.Type) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	closer := closerInterfaceType()
	if closer == nil {
		return false
	}

	if types.Implements(t, closer) {
		return true
	}
	if _, ok := t.(*types.Pointer); ok {
		return false
	}
	return types.Implements(types.NewPointer(t), closer)
}

func hasCloserBodyField(t types.Type) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}

	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return false
	}

	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		if field == nil || field.Name() != "Body" {
			continue
		}
		if isCloserType(field.Type()) {
			return true
		}
	}

	return false
}
