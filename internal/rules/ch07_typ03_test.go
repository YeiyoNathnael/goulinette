package rules

import (
	"go/types"
	"testing"
)

// TestHasMeaningfulZeroValue documents this exported function.
func TestHasMeaningfulZeroValue(t *testing.T) {
	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{name: "bool", typ: types.Typ[types.Bool], want: true},
		{name: "int", typ: types.Typ[types.Int], want: true},
		{name: "struct", typ: types.NewStruct(nil, nil), want: true},
		{name: "slice", typ: types.NewSlice(types.Typ[types.Int]), want: false},
		{name: "map", typ: types.NewMap(types.Typ[types.String], types.Typ[types.Int]), want: false},
		{name: "pointer", typ: types.NewPointer(types.Typ[types.Int]), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			if got := hasMeaningfulZeroValue(tc.typ); got != tc.want {
				t.Fatalf("hasMeaningfulZeroValue() = %v, want %v", got, tc.want)
			}
		})
	}
}
