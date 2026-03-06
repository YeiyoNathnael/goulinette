package rules

import "strings"

const (
	minLevel        = 0
	defaultLevelOne = 1
	defaultLevelTwo = 2
	maxLevelThree   = 3
)

var level0Rules = []string{
	ruleFMT01, ruleFMT02, ruleFUN02, ruleFUN04, ruleERR04, ruleERR05, ruleERR07, ruleTYP03, ruleTYP04, ruleTYP05,
	ruleCTL04, ruleCTX02, ruleCTX03, ruleRES01, ruleSAF01, ruleSAF02, ruleCER02, ruleCER03, ruleCON03,
}

var level1Rules = []string{
	ruleFUN01, ruleERR01, ruleERR02, ruleERR06, ruleERR08, ruleTYP06, ruleSTR01, ruleSTR02, ruleSTR04, ruleCTL02,
	ruleCTL03, ruleVAR01, ruleVAR04, ruleNAM01, ruleNAM02, ruleNAM06, ruleNAM07, ruleDOC04, ruleIMP01, ruleIMP02,
	ruleCTX01, ruleCON01, ruleCON02, ruleCER01, ruleRES02, ruleTST02,
}

var level2Rules = []string{
	ruleFMT03, ruleFUN03, ruleERR03, ruleTYP01, ruleTYP02, ruleTYP07, ruleSTR03, ruleVAR02, ruleVAR03, ruleNAM03,
	ruleNAM04, ruleNAM05, ruleDOC01, ruleDOC02, ruleDOC03, ruleDOC05, ruleCTL01, ruleSLC01, ruleIMP03, ruleMAG01,
	ruleMAG02, ruleTST01, ruleTST03,
}

var level3Rules = []string{
	ruleLIM01, ruleLIM02, ruleLIM03, ruleLIM04, ruleCTX04,
}

// IDsForLevel returns the set of rule IDs that are active at the given
// strictness level. Level 0 enables a baseline of high-value, low-noise
// rules; each higher level (1, 2, 3) adds progressively more thorough
// checks. Levels below 0 return an empty set; levels above 3 are treated
// as 3. The returned map uses upper-case IDs as keys.
func IDsForLevel(level int) map[string]struct{} {
	out := make(map[string]struct{})
	if level < minLevel {
		return out
	}
	addRuleIDs(out, level0Rules)
	if level >= defaultLevelOne {
		addRuleIDs(out, level1Rules)
	}
	if level >= defaultLevelTwo {
		addRuleIDs(out, level2Rules)
	}
	if level >= maxLevelThree {
		addRuleIDs(out, level3Rules)
	}
	return out
}

func addRuleIDs(out map[string]struct{}, ids []string) {
	for _, id := range ids {
		out[strings.ToUpper(id)] = struct{}{}
	}
}
