package rules

import "strings"

// Registry returns every rule known to goulinette, grouped by category
// (style, behavior, architecture, reliability, test & magic-value rules).
// Rules are returned in a stable order that mirrors chapter ordering.
func Registry() []Rule {
	rules := make([]Rule, 0)
	rules = append(rules, registryStyleRules()...)
	rules = append(rules, registryBehaviorRules()...)
	rules = append(rules, registryArchitectureRules()...)
	rules = append(rules, registryReliabilityRules()...)
	rules = append(rules, registryTestAndMagicRules()...)
	return rules
}

func registryStyleRules() []Rule {
	return []Rule{
		NewFMT01(), NewFMT02(), NewFMT03(),
		NewNAM01(), NewNAM02(), NewNAM03(), NewNAM04(), NewNAM05(), NewNAM06(), NewNAM07(),
		NewVAR01(), NewVAR02(), NewVAR03(), NewVAR04(),
		NewCTL01(), NewCTL02(), NewCTL03(), NewCTL04(),
	}
}

func registryBehaviorRules() []Rule {
	return []Rule{
		NewFUN01(), NewFUN02(), NewFUN03(), NewFUN04(),
		NewERR01(), NewERR02(), NewERR03(), NewERR04(), NewERR05(), NewERR06(), NewERR07(), NewERR08(),
		NewTYP01(), NewTYP02(), NewTYP03(), NewTYP04(), NewTYP05(), NewTYP06(), NewTYP07(),
		NewSTR01(), NewSTR02(), NewSTR03(), NewSTR04(),
	}
}

func registryArchitectureRules() []Rule {
	return []Rule{
		NewDOC01(), NewDOC02(), NewDOC03(), NewDOC04(), NewDOC05(),
		NewSLC01(),
		NewCON01(), NewCON02(), NewCON03(),
		NewCER01(), NewCER02(), NewCER03(),
	}
}

func registryReliabilityRules() []Rule {
	return []Rule{
		NewLIM01(), NewLIM02(), NewLIM03(), NewLIM04(),
		NewCTX01(), NewCTX02(), NewCTX03(), NewCTX04(),
		NewIMP01(), NewIMP02(), NewIMP03(),
		NewRES01(), NewRES02(),
		NewSAF01(), NewSAF02(),
	}
}

func registryTestAndMagicRules() []Rule {
	return []Rule{NewMAG01(), NewMAG02(), NewTST01(), NewTST02(), NewTST03()}
}

// Select returns the subset of all that should run given the active filters.
// A rule is excluded if its ID appears in disableRules, its chapter is not
// in chapters (when the set is non-nil), or its ID is not in includeRules
// (when that set is non-nil). All comparisons are case-insensitive.
func Select(all []Rule, chapters map[int]struct{}, includeRules map[string]struct{}, disableRules map[string]struct{}) []Rule {
	selected := make([]Rule, 0, len(all))
	for _, rule := range all {
		id := strings.ToUpper(rule.ID())
		if disableRules != nil {
			if _, disabled := disableRules[id]; disabled {
				continue
			}
		}

		if chapters != nil {
			if _, ok := chapters[rule.Chapter()]; !ok {
				continue
			}
		}

		if includeRules != nil {
			if _, ok := includeRules[id]; !ok {
				continue
			}
		}

		selected = append(selected, rule)
	}
	return selected
}
