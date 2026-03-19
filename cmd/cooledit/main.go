// Copyright (C) 2026 Tom Cool
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cooledit/internal/app"
	"cooledit/internal/buildinfo"
	"cooledit/internal/config"

	flag "github.com/spf13/pflag"
)

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
		fmt.Printf("cooledit version %s\n", buildinfo.Version)
		fmt.Printf("Copyright (C) 2026 Tom Cool\n")
		fmt.Printf("License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>\n")
		fmt.Printf("This is free software: you are free to change and redistribute it.\n")
		fmt.Printf("There is NO WARRANTY, to the extent permitted by law.\n")
		os.Exit(0)
	}

	// Load configuration
	var cfg *config.Config
	var err error

	if *configPath != "" {
		// Override ConfigPath so Save() also targets the custom file
		customPath := *configPath
		config.ConfigPath = func() (string, error) {
			if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
				return "", err
			}
			return customPath, nil
		}
		cfg, err = config.LoadFrom(customPath)
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
