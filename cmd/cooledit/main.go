package main

import (
	"log"
	"os"

	"cooledit/internal/app"
)

func main() {
	var path string
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	if err := app.Run(path); err != nil {
		log.Fatal(err)
	}
}
