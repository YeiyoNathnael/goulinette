package rules

import "strings"

var level0Rules = []string{
	"FMT-01", "FMT-02", "FUN-02", "FUN-04", "ERR-04", "ERR-05", "ERR-07", "TYP-03", "TYP-04", "TYP-05",
	"CTL-04", "CTX-02", "CTX-03", "RES-01", "SAF-01", "SAF-02", "CER-02", "CER-03", "CON-03",
}

var level1Rules = []string{
	"FUN-01", "ERR-01", "ERR-02", "ERR-06", "ERR-08", "TYP-06", "STR-01", "STR-02", "STR-04", "CTL-02",
	"CTL-03", "VAR-01", "VAR-04", "NAM-01", "NAM-02", "NAM-06", "NAM-07", "DOC-04", "IMP-01", "IMP-02",
	"CTX-01", "CON-01", "CON-02", "CER-01", "RES-02", "TST-02",
}

var level2Rules = []string{
	"FMT-03", "FUN-03", "ERR-03", "TYP-01", "TYP-02", "TYP-07", "STR-03", "VAR-02", "VAR-03", "NAM-03",
	"NAM-04", "NAM-05", "DOC-01", "DOC-02", "DOC-03", "DOC-05", "CTL-01", "SLC-01", "IMP-03", "MAG-01",
	"MAG-02", "TST-01", "TST-03",
}

var level3Rules = []string{
	"LIM-01", "LIM-02", "LIM-03", "LIM-04", "CTX-04",
}

func RulesForLevel(level int) map[string]struct{} {
	out := make(map[string]struct{})
	if level < 0 {
		return out
	}
	addRuleIDs(out, level0Rules)
	if level >= 1 {
		addRuleIDs(out, level1Rules)
	}
	if level >= 2 {
		addRuleIDs(out, level2Rules)
	}
	if level >= 3 {
		addRuleIDs(out, level3Rules)
	}
	return out
}

func addRuleIDs(out map[string]struct{}, ids []string) {
	for _, id := range ids {
		out[strings.ToUpper(id)] = struct{}{}
	}
}
