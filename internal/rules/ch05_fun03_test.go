package rules

import (
	"go/types"
	"testing"
)

// TestShouldWarnInterfaceReturnTyped verifies that FUN-03 detects functions
// whose return type is a concrete type implementing an interface when the
// caller would benefit from the broader interface type.
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
