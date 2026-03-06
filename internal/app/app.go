package app

import (
	"context"
	"fmt"
	"os"

	"github.com/YeiyoNathnael/goulinette/internal/config"
	"github.com/YeiyoNathnael/goulinette/internal/diag"
	"github.com/YeiyoNathnael/goulinette/internal/discovery"
	"github.com/YeiyoNathnael/goulinette/internal/report"
	"github.com/YeiyoNathnael/goulinette/internal/rules"
)

type App struct {
	cfg config.Config
}

func New(cfg config.Config) App {
	return App{cfg: cfg}
}

func (a App) Run(_ context.Context) int {
	result := diag.Result{}
	rules.ResetCaches()

	files, err := discovery.GoFiles(a.cfg.Root)
	if err != nil {
		result.RuntimeErrs = append(result.RuntimeErrs, err.Error())
		report.Print(os.Stdout, a.cfg.Format, result)
		return result.ExitCode(a.cfg.WarningsAsErrors)
	}

	ruleCtx := rules.Context{
		Root:        a.cfg.Root,
		Files:       files,
		StrictTools: a.cfg.StrictTools,
	}

	includeRules := a.cfg.Rules
	if includeRules == nil {
		includeRules = rules.RulesForLevel(a.cfg.Level)
	}

	selected := rules.Select(rules.Registry(), a.cfg.Chapters, includeRules, a.cfg.DisableRules)
	for _, rule := range selected {
		ds, runErr := rule.Run(ruleCtx)
		if runErr != nil {
			result.RuntimeErrs = append(result.RuntimeErrs, fmt.Sprintf("%s: %v", rule.ID(), runErr))
			continue
		}
		result.Diagnostics = append(result.Diagnostics, ds...)
	}

	diag.Sort(result.Diagnostics)
	report.Print(os.Stdout, a.cfg.Format, result)
	return result.ExitCode(a.cfg.WarningsAsErrors)
}
