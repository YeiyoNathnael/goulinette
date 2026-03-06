package rules

import "testing"

// TestIDsForLevel documents this exported function.
func TestIDsForLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       int
		mustContain []string
		mustNotHave []string
	}{
		{
			name:        "level 0 has bug rules only",
			level:       0,
			mustContain: []string{ruleFMT01, ruleCON03, ruleSAF02},
			mustNotHave: []string{ruleFUN01, ruleMAG01, ruleLIM01},
		},
		{
			name:        "level 1 adds strict idiomatic rules",
			level:       1,
			mustContain: []string{ruleFMT01, ruleFUN01, ruleCON02, ruleTST02},
			mustNotHave: []string{ruleMAG01, ruleLIM01},
		},
		{
			name:        "level 2 adds opinionated rules",
			level:       2,
			mustContain: []string{ruleFMT03, ruleMAG01, ruleTST03, ruleNAM05},
			mustNotHave: []string{ruleLIM01, ruleCTX04},
		},
		{
			name:        "level 3 adds max strictness rules",
			level:       3,
			mustContain: []string{ruleLIM01, ruleLIM04, ruleCTX04, ruleMAG01},
			mustNotHave: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			selected := IDsForLevel(tc.level)
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
