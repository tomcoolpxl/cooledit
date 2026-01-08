package main

import (
	"flag"
	"log"

	"cooledit/internal/app"
)

func main() {
	mouse := flag.Bool("mouse", false, "Enable mouse support")
	flag.Parse()

	path := flag.Arg(0)

	if err := app.Run(path, *mouse); err != nil {
		log.Fatal(err)
	}
}
