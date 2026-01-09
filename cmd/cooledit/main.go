package main

import (
	"fmt"
	"log"
	"os"

	"cooledit/internal/app"
	"cooledit/internal/config"

	flag "github.com/spf13/pflag"
)

const version = "0.1.0"

func main() {
	// Define flags
	showVersion := flag.BoolP("version", "v", false, "Show version information")
	configPath := flag.StringP("config", "c", "", "Path to config file")
	lineNumbers := flag.BoolP("line-numbers", "l", false, "Show line numbers")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cooledit - A terminal text editor\n\n")
		fmt.Fprintf(os.Stderr, "Usage: cooledit [OPTIONS] [FILE]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("cooledit version %s\n", version)
		os.Exit(0)
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.LoadFrom(*configPath)
		if err != nil {
			log.Fatalf("Error: Failed to load config from %s: %v", *configPath, err)
		}
	} else {
		cfg, err = config.Load()
		if err != nil {
			log.Printf("Warning: Failed to load config: %v (using defaults)", err)
			cfg = config.Default()
		}
	}

	// CLI flags override config file
	if *lineNumbers {
		cfg.Editor.LineNumbers = *lineNumbers
	}

	path := flag.Arg(0)

	// Pass config and overrides to app
	if err := app.Run(path, cfg.Editor.LineNumbers, cfg); err != nil {
		log.Fatal(err)
	}
}
