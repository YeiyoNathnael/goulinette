package rules

import "go/types"

func isSyncNamedType(t types.Type, name string) bool {
	t = types.Unalias(t)
	named, ok := t.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == name
}

func containsSyncMutexValue(t types.Type, seen map[types.Type]bool) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	if seen[t] {
		return false
	}
	seen[t] = true

	if isSyncNamedType(t, "Mutex") || isSyncNamedType(t, "RWMutex") {
		return true
	}

	switch tt := t.(type) {
	case *types.Pointer:
		return false
	case *types.Named:
		return containsSyncMutexValue(tt.Underlying(), seen)
	case *types.Struct:
		for i := 0; i < tt.NumFields(); i++ {
			if containsSyncMutexValue(tt.Field(i).Type(), seen) {
				return true
			}
		}
	}

	return false
}

func containsWaitGroupValue(t types.Type, seen map[types.Type]bool) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	if seen[t] {
		return false
	}
	seen[t] = true

	if isSyncNamedType(t, "WaitGroup") {
		return true
	}

	switch tt := t.(type) {
	case *types.Pointer:
		return false
	case *types.Named:
		return containsWaitGroupValue(tt.Underlying(), seen)
	case *types.Struct:
		for i := 0; i < tt.NumFields(); i++ {
			if containsWaitGroupValue(tt.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Array:
		return containsWaitGroupValue(tt.Elem(), seen)
	}

	return false
}

func receiverBaseNamed(recv types.Type) (*types.Named, bool) {
	if recv == nil {
		return nil, false
	}
	recv = types.Unalias(recv)
	if ptr, ok := recv.(*types.Pointer); ok {
		named, _ := types.Unalias(ptr.Elem()).(*types.Named)
		return named, true
	}
	named, _ := recv.(*types.Named)
	return named, false
}
