package rules

import (
	"go/token"
	"go/types"
	"testing"
)

// TestIsLocalStructPointerType verifies that isLocalStructPointerType
// identifies *T expressions where T is a struct type defined in the same
// package, and rejects pointers to builtins, interfaces, or external types.
func TestIsLocalStructPointerType(t *testing.T) {
	localPkg := types.NewPackage("example.com/app", "app")
	foreignPkg := types.NewPackage("example.com/other", "other")

	localStruct := types.NewNamed(types.NewTypeName(token.NoPos, localPkg, "Config", nil), types.NewStruct(nil, nil), nil)
	foreignStruct := types.NewNamed(types.NewTypeName(token.NoPos, foreignPkg, "Client", nil), types.NewStruct(nil, nil), nil)

	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{name: "local struct ptr", typ: types.NewPointer(localStruct), want: true},
		{name: "foreign struct ptr", typ: types.NewPointer(foreignStruct), want: false},
		{name: "local primitive ptr", typ: types.NewPointer(types.Typ[types.Int]), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got := isLocalStructPointerType(tc.typ, localPkg)
			if got != tc.want {
				t.Fatalf("isLocalStructPointerType() = %v, want %v", got, tc.want)
			}
		})
	}
}
