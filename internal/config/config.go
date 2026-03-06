package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
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
}

func ParseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("goulinette", flag.ContinueOnError)

	var chapterCSV string
	var ruleCSV string
	var disableCSV string

	cfg := Config{}
	fs.StringVar(&cfg.Root, "root", ".", "root directory to scan")
	fs.StringVar(&cfg.Format, "format", "text", "output format: text|json")
	fs.IntVar(&cfg.Level, "level", 1, "strictness level: 0 (bugs) to 3 (maximum strictness)")
	fs.StringVar(&chapterCSV, "chapter", "", "comma-separated chapter numbers")
	fs.StringVar(&ruleCSV, "rule", "", "comma-separated rule IDs")
	fs.StringVar(&disableCSV, "disable", "", "comma-separated rule IDs to disable")
	fs.BoolVar(&cfg.WarningsAsErrors, "warnings-as-errors", false, "treat warnings as errors")
	fs.BoolVar(&cfg.StrictTools, "strict-tools", false, "fail when required external tools are missing")
	fs.IntVar(&cfg.MaxWorkers, "max-workers", 4, "max analysis workers")
	fs.DurationVar(&cfg.Timeout, "timeout", 2*time.Minute, "command timeout")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if cfg.Format != "text" && cfg.Format != "json" {
		return Config{}, fmt.Errorf("invalid --format %q (expected text or json)", cfg.Format)
	}
	if cfg.Level < 0 || cfg.Level > 3 {
		return Config{}, fmt.Errorf("invalid --level %d (expected 0..3)", cfg.Level)
	}

	chapters, err := parseChapters(chapterCSV)
	if err != nil {
		return Config{}, err
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
