package main

import (
	"flag"
	"log"

	"cooledit/internal/app"
	"cooledit/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v (using defaults)", err)
		cfg = config.Default()
	}

	// CLI flags override config file
	mouse := flag.Bool("mouse", cfg.UI.MouseEnabled, "Enable mouse support")
	lineNumbers := flag.Bool("line-numbers", cfg.Editor.LineNumbers, "Show line numbers")
	flag.Parse()

	path := flag.Arg(0)

	// Pass config and overrides to app
	if err := app.Run(path, *mouse, *lineNumbers, cfg); err != nil {
		log.Fatal(err)
	}
}
