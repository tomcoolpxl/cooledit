package main

import (
	"flag"
	"log"

	"cooledit/internal/app"
)

func main() {
	mouse := flag.Bool("mouse", false, "Enable mouse support")
	lineNumbers := flag.Bool("line-numbers", false, "Show line numbers")
	flag.Parse()

	path := flag.Arg(0)

	if err := app.Run(path, *mouse, *lineNumbers); err != nil {
		log.Fatal(err)
	}
}
