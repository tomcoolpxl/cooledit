package main

import (
	"flag"
	"log"

	"cooledit/internal/app"
)

func main() {
	mouse := flag.Bool("mouse", false, "Enable mouse support")
	lineNumbers := flag.Bool("line-numbers", false, "Show line numbers")
	goToLine := flag.Bool("go-to-line", false, "Enable Go To Line shortcut (Ctrl+G)")
	flag.Parse()

	path := flag.Arg(0)

	if err := app.Run(path, *mouse, *lineNumbers, *goToLine); err != nil {
		log.Fatal(err)
	}
}
