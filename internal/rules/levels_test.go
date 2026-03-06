package rules

import "testing"

func TestRulesForLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       int
		mustContain []string
		mustNotHave []string
	}{
		{
			name:        "level 0 has bug rules only",
			level:       0,
			mustContain: []string{"FMT-01", "CON-03", "SAF-02"},
			mustNotHave: []string{"FUN-01", "MAG-01", "LIM-01"},
		},
		{
			name:        "level 1 adds strict idiomatic rules",
			level:       1,
			mustContain: []string{"FMT-01", "FUN-01", "CON-02", "TST-02"},
			mustNotHave: []string{"MAG-01", "LIM-01"},
		},
		{
			name:        "level 2 adds opinionated rules",
			level:       2,
			mustContain: []string{"FMT-03", "MAG-01", "TST-03", "NAM-05"},
			mustNotHave: []string{"LIM-01", "CTX-04"},
		},
		{
			name:        "level 3 adds max strictness rules",
			level:       3,
			mustContain: []string{"LIM-01", "LIM-04", "CTX-04", "MAG-01"},
			mustNotHave: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected := RulesForLevel(tc.level)
			for _, id := range tc.mustContain {
				if _, ok := selected[id]; !ok {
					t.Fatalf("expected %s to be enabled at level %d", id, tc.level)
				}
			}
			for _, id := range tc.mustNotHave {
				if _, ok := selected[id]; ok {
					t.Fatalf("did not expect %s at level %d", id, tc.level)
				}
			}
		})
	}
}
