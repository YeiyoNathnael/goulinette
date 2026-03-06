package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

const (
	formatText     = "text"
	formatJSON     = "json"
	flagTimeout    = "timeout"
	defaultWorkers = 4
	maxStrictLevel = 3
	appName        = "goulinette"
)

// Settings holds the fully parsed configuration for a single analysis run.
// All fields are populated by ParseFlags and are read-only once the Runner
// is constructed. Zero values are not valid; use ParseFlags to obtain a
// properly defaulted instance.
type Settings struct {
	Root             string
	Format           string
	Level            int
	Chapters         map[int]struct{}
	Rules            map[string]struct{}
	DisableRules     map[string]struct{}
	WarningsAsErrors bool
	StrictTools      bool
	MaxWorkers       int
	Timeout          time.Duration
	// ExplainRule, when non-empty, names a rule whose rationale should be
	// printed and the process should exit without running analysis.
	ExplainRule string
	// PrintVersion, when true, causes the binary to print its version string
	// and exit without running analysis.
	PrintVersion bool
}

// ParseFlags parses the provided CLI argument slice and returns a validated
// Settings value. It defines all goulinette flags (--root, --format, --level,
// --chapter, --rule, --disable, --warnings-as-errors, --strict-tools,
// --max-workers, --timeout) and validates that --format is "text" or "json"
// and that --level is in the range [0, 3]. An error is returned for any
// unrecognised flag or invalid value.
func ParseFlags(args []string) (Settings, error) {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)

	var chapterCSV string
	var ruleCSV string
	var disableCSV string

	cfg := Settings{}
	fs.StringVar(&cfg.Root, "root", ".", "root directory to scan")
	fs.StringVar(&cfg.Format, "format", formatText, "output format: text|json")
	fs.IntVar(&cfg.Level, "level", 1, "strictness level: 0 (bugs) to 3 (maximum strictness)")
	fs.StringVar(&chapterCSV, "chapter", "", "comma-separated chapter numbers")
	fs.StringVar(&ruleCSV, "rule", "", "comma-separated rule IDs")
	fs.StringVar(&disableCSV, "disable", "", "comma-separated rule IDs to disable")
	fs.BoolVar(&cfg.WarningsAsErrors, "warnings-as-errors", false, "treat warnings as errors")
	fs.BoolVar(&cfg.StrictTools, "strict-tools", false, "fail when required external tools are missing")
	fs.IntVar(&cfg.MaxWorkers, "max-workers", defaultWorkers, "number of rules to run in parallel (default 4)")
	fs.DurationVar(&cfg.Timeout, flagTimeout, 2*time.Minute, "command timeout")
	fs.StringVar(&cfg.ExplainRule, "explain", "", "print rationale for a rule ID and exit (e.g. --explain CTX-01)")
	fs.BoolVar(&cfg.PrintVersion, "version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		return Settings{}, err
	}

	if cfg.Format != formatText && cfg.Format != formatJSON {
		return Settings{}, fmt.Errorf("invalid --format %q (expected text or json)", cfg.Format)
	}
	if cfg.Level < 0 || cfg.Level > maxStrictLevel {
		return Settings{}, fmt.Errorf("invalid --level %d (expected 0..%d)", cfg.Level, maxStrictLevel)
	}

	chapters, err := parseChapters(chapterCSV)
	if err != nil {
		return Settings{}, err
	}

	cfg.Chapters = chapters
	cfg.Rules = parseSet(ruleCSV)
	cfg.DisableRules = parseSet(disableCSV)

	return cfg, nil
}

func parseSet(csv string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, raw := range strings.Split(csv, ",") {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		out[strings.ToUpper(item)] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseChapters(csv string) (map[int]struct{}, error) {
	if strings.TrimSpace(csv) == "" {
		return nil, nil
	}

	out := map[int]struct{}{}
	for _, raw := range strings.Split(csv, ",") {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		var value int
		_, err := fmt.Sscanf(item, "%d", &value)
		if err != nil || value <= 0 {
			return nil, fmt.Errorf("invalid chapter value %q", item)
		}
		out[value] = struct{}{}
	}

	if len(out) == 0 {
		return nil, nil
	}

	return out, nil
}
