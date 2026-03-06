package app

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/YeiyoNathnael/goulinette/internal/config"
	"github.com/YeiyoNathnael/goulinette/internal/diag"
	"github.com/YeiyoNathnael/goulinette/internal/discovery"
	"github.com/YeiyoNathnael/goulinette/internal/report"
	"github.com/YeiyoNathnael/goulinette/internal/rules"
	"github.com/YeiyoNathnael/goulinette/internal/suppress"
)

// Runner orchestrates a full analysis run: it discovers Go source files,
// selects applicable rules for the configured level and chapters, executes
// each rule, and writes a report to stdout.
type Runner struct {
	cfg config.Settings
}

// New constructs a Runner from the provided settings.
func New(cfg config.Settings) Runner {
	return Runner{cfg: cfg}
}

// Run executes the analysis pipeline and returns an exit code suitable for
// os.Exit. It discovers Go files under the configured root, dispatches all
// selected rules to a pool of up to MaxWorkers goroutines, collects their
// findings concurrently, then sorts and formats the results. Returns:
//   - 0 on clean (no errors, or no warnings when warnings-as-errors is off)
//   - 1 when at least one error-severity finding is present (or any warning
//     when WarningsAsErrors is set)
//   - 2 when a runtime error prevented normal analysis
func (r Runner) Run(ctx context.Context) int {
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

	type ruleResult struct {
		ruleID   string
		findings []diag.Finding
		err      error
	}

	workers := r.cfg.MaxWorkers
	if workers < 1 {
		workers = 1
	}

	// results is sized to hold one entry per rule so workers never block on
	// send. The main goroutine drains it by count after all workers exit,
	// which avoids needing a separate closer goroutine.
	work := make(chan rules.Rule, len(selected))
	results := make(chan ruleResult, len(selected))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case rule, ok := <-work:
					if !ok {
						return
					}
					ds, runErr := rule.Run(ruleCtx)
					results <- ruleResult{ruleID: rule.ID(), findings: ds, err: runErr}
				}
			}
		}()
	}

	for _, rule := range selected {
		work <- rule
	}
	close(work)
	wg.Wait()

	for len(results) > 0 {
		res, ok := <-results
		if !ok {
			break
		}
		if res.err != nil {
			result.RuntimeErrs = append(result.RuntimeErrs, fmt.Sprintf("%s: %v", res.ruleID, res.err))
			continue
		}
		result.Diagnostics = append(result.Diagnostics, res.findings...)
	}

	diag.Sort(result.Diagnostics)
	result.Diagnostics = suppress.Filter(result.Diagnostics)
	report.Print(os.Stdout, r.cfg.Format, result)
	return result.ExitCode(r.cfg.WarningsAsErrors)
}
