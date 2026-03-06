package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSLC01(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "empty literal warning",
			source: `package sample
func f() {
	s := []int{}
	_ = s
}
`,
			wantCount: 1,
		},
		{
			name: "non-empty literal ignored",
			source: `package sample
func f() {
	s := []int{1}
	_ = s
}
`,
			wantCount: 0,
		},
		{
			name: "array literal ignored",
			source: `package sample
func f() {
	a := [0]int{}
	_ = a
}
`,
			wantCount: 0,
		},
		{
			name: "justified for json",
			source: `package sample
func f() {
	// require [] for JSON output
	s := []string{}
	_ = s
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			file := filepath.Join(dir, "sample.go")
			if err := os.WriteFile(file, []byte(tc.source), 0o644); err != nil {
				t.Fatalf("write sample: %v", err)
			}

			diags, err := NewSLC01().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run SLC-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
