package config

import "testing"

const levelFlag = "--level"

// TestParseFlagsLevelValidation documents this exported function.
func TestParseFlagsLevelValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantLvl int
	}{
		{name: "default level is 1", args: nil, wantErr: false, wantLvl: 1},
		{name: "level 0 allowed", args: []string{levelFlag, "0"}, wantErr: false, wantLvl: 0},
		{name: "level 3 allowed", args: []string{levelFlag, "3"}, wantErr: false, wantLvl: maxStrictLevel},
		{name: "negative level rejected", args: []string{levelFlag, "-1"}, wantErr: true},
		{name: "level above max rejected", args: []string{levelFlag, "4"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			cfg, err := ParseFlags(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Level != tc.wantLvl {
				t.Fatalf("expected level %d, got %d", tc.wantLvl, cfg.Level)
			}
		})
	}
}
