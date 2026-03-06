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

// Runner documents this exported type.
type Runner struct {
	cfg config.Settings
}

// New documents this exported function.
func New(cfg config.Settings) Runner {
	return Runner{cfg: cfg}
}

// Run documents this exported method.
func (r Runner) Run(_ context.Context) int {
	result := diag.Result{}
	rules.ResetCaches()

	files, err := discovery.GoFiles(r.cfg.Root)
	if err != nil {
		result.RuntimeErrs = append(result.RuntimeErrs, err.Error())
		report.Print(os.Stdout, r.cfg.Format, result)
		return result.ExitCode(r.cfg.WarningsAsErrors)
	}

	ruleCtx := rules.Context{
		Root:        r.cfg.Root,
		Files:       files,
		StrictTools: r.cfg.StrictTools,
	}

	includeRules := r.cfg.Rules
	if includeRules == nil {
		includeRules = rules.IDsForLevel(r.cfg.Level)
	}

	selected := rules.Select(rules.Registry(), r.cfg.Chapters, includeRules, r.cfg.DisableRules)
	for _, rule := range selected {
		ds, runErr := rule.Run(ruleCtx)
		if runErr != nil {
			result.RuntimeErrs = append(result.RuntimeErrs, fmt.Sprintf("%s: %v", rule.ID(), runErr))
			continue
		}
		result.Diagnostics = append(result.Diagnostics, ds...)
	}

	diag.Sort(result.Diagnostics)
	report.Print(os.Stdout, r.cfg.Format, result)
	return result.ExitCode(r.cfg.WarningsAsErrors)
}
