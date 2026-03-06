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
		NewVAR02(),
		NewVAR03(),
		NewVAR04(),
		NewCTL01(),
		NewCTL02(),
		NewCTL03(),
		NewCTL04(),
		NewFUN01(),
		NewFUN02(),
		NewFUN03(),
		NewFUN04(),
		NewERR01(),
		NewERR02(),
		NewERR03(),
		NewERR04(),
		NewERR05(),
		NewERR06(),
		NewERR07(),
		NewERR08(),
		NewTYP01(),
		NewTYP02(),
		NewTYP03(),
		NewTYP04(),
		NewTYP05(),
		NewTYP06(),
		NewTYP07(),
		NewSTR01(),
		NewSTR02(),
		NewSTR03(),
		NewSTR04(),
		NewDOC01(),
		NewDOC02(),
		NewDOC03(),
		NewDOC04(),
		NewDOC05(),
		NewSLC01(),
		NewCON01(),
		NewCON02(),
		NewCON03(),
		NewCER01(),
		NewCER02(),
		NewCER03(),
		NewLIM01(),
		NewLIM02(),
		NewLIM03(),
		NewLIM04(),
		NewCTX01(),
		NewCTX02(),
		NewCTX03(),
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
