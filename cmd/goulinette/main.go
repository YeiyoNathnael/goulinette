package main

import (
	"context"
	"fmt"
	"os"

	"github.com/YeiyoNathnael/goulinette/internal/app"
	"github.com/YeiyoNathnael/goulinette/internal/config"
	"github.com/YeiyoNathnael/goulinette/internal/rules"
	"github.com/YeiyoNathnael/goulinette/internal/version"
)

const appName = "goulinette"

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}

	if cfg.PrintVersion {
		fmt.Println(appName, version.Current)
		os.Exit(0)
	}

	if cfg.ExplainRule != "" {
		fmt.Println(rules.PrintExplain(cfg.ExplainRule))
		os.Exit(0)
	}

	a := app.New(cfg)
	os.Exit(a.Run(context.Background()))
}
