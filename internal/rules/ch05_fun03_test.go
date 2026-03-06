package rules

import (
	"go/types"
	"testing"
)

// TestShouldWarnInterfaceReturnTyped documents this exported function.
func TestShouldWarnInterfaceReturnTyped(t *testing.T) {
	iface := types.NewInterfaceType(nil, nil)
	iface.Complete()

	withNoParams := types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(types.NewParam(0, nil, "v", iface)), false)
	if shouldWarnInterfaceReturnTyped(withNoParams) {
		t.Fatalf("should not warn when function has no params")
	}

	withParams := types.NewSignatureType(nil, nil, nil, types.NewTuple(types.NewParam(0, nil, "x", types.Typ[types.Int])), types.NewTuple(types.NewParam(0, nil, "v", iface)), false)
	if !shouldWarnInterfaceReturnTyped(withParams) {
		t.Fatalf("should warn when function has params and returns interface")
	}
}
