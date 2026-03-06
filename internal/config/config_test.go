package config

import "testing"

func TestParseFlags_LevelValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantLvl int
	}{
		{name: "default level is 1", args: []string{}, wantErr: false, wantLvl: 1},
		{name: "level 0 allowed", args: []string{"--level", "0"}, wantErr: false, wantLvl: 0},
		{name: "level 3 allowed", args: []string{"--level", "3"}, wantErr: false, wantLvl: 3},
		{name: "negative level rejected", args: []string{"--level", "-1"}, wantErr: true},
		{name: "level above max rejected", args: []string{"--level", "4"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
