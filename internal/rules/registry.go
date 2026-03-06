package rules

import "strings"

func Registry() []Rule {
	return []Rule{
		NewFMT01(),
		NewFMT02(),
		NewFMT03(),
		NewNAM01(),
		NewNAM02(),
		NewNAM06(),
		NewNAM07(),
		NewVAR01(),
		NewVAR03(),
	}
}

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
