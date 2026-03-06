package main

import (
	"context"
	"os"

	"goulinette/internal/app"
	"goulinette/internal/config"
)

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}

	a := app.New(cfg)
	os.Exit(a.Run(context.Background()))
}
