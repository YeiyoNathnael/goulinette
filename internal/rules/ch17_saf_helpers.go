package rules

import "go/types"

const (
	safSyncPkgPath      = "sync"
	safWaitGroupName    = "WaitGroup"
	safLockMethodName   = "Lock"
	safUnlockMethodName = "Unlock"
	safNoCopyName       = "noCopy"
	safNoCopyTypeName   = "NoCopy"
)

var syncCopySensitiveTypeNames = map[string]struct{}{
	"Mutex":          {},
	"RWMutex":        {},
	safWaitGroupName: {},
	"Once":           {},
	"Cond":           {},
	"Map":            {},
}

func isSyncCopySensitiveType(t types.Type) bool {
	t = types.Unalias(t)
	named, ok := t.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	if named.Obj().Pkg().Path() != safSyncPkgPath {
		return false
	}
	_, ok = syncCopySensitiveTypeNames[named.Obj().Name()]
	return ok
}

func hasMethodNamedNoArgsNoResults(t types.Type, name string) bool {
	ms := types.NewMethodSet(t)
	for i := 0; i < ms.Len(); i++ {
		sel := ms.At(i)
		if sel == nil || sel.Obj() == nil || sel.Obj().Name() != name {
			continue
		}
		fn, ok := sel.Obj().(*types.Func)
		if !ok {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok || sig.Params() == nil || sig.Results() == nil {
			continue
		}
		if sig.Params().Len() == 0 && sig.Results().Len() == 0 {
			return true
		}
	}
	return false
}

func hasLockUnlockPair(t types.Type) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	if isNoCopyTypeName(t) {
		return true
	}
	if hasMethodNamedNoArgsNoResults(t, safLockMethodName) && hasMethodNamedNoArgsNoResults(t, safUnlockMethodName) {
		return true
	}
	if _, isPtr := t.(*types.Pointer); isPtr {
		if pt, ok := t.(*types.Pointer); ok && isNoCopyTypeName(pt.Elem()) {
			return true
		}
		return false
	}
	pt := types.NewPointer(t)
	return hasMethodNamedNoArgsNoResults(pt, safLockMethodName) && hasMethodNamedNoArgsNoResults(pt, safUnlockMethodName)
}

func isNoCopyTypeName(t types.Type) bool {
	t = types.Unalias(t)
	named, ok := t.(*types.Named)
	if !ok || named.Obj() == nil {
		return false
	}
	return named.Obj().Name() == safNoCopyName || named.Obj().Name() == safNoCopyTypeName
}

func containsCopySensitiveValue(t types.Type, seen map[types.Type]bool) bool {
	if t == nil {
		return false
	}
	t = types.Unalias(t)
	if visited, ok := seen[t]; ok && visited {
		return false
	}
	seen[t] = true

	if isSyncCopySensitiveType(t) {
		return true
	}

	switch tt := t.(type) {
	case *types.Pointer:
		return false
	case *types.Named:
		return containsCopySensitiveValue(tt.Underlying(), seen)
	case *types.Struct:
		for i := 0; i < tt.NumFields(); i++ {
			field := tt.Field(i)
			if hasLockUnlockPair(field.Type()) {
				return true
			}
			if containsCopySensitiveValue(field.Type(), seen) {
				return true
			}
		}
	case *types.Array:
		return containsCopySensitiveValue(tt.Elem(), seen)
	default:
		return false
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
