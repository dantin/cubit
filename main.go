package main

import (
	"log"
	"os"

	"github.com/dantin/cubit/app"
)

func main() {
	instance := app.New(os.Stdout, os.Args)
	if err := instance.Run(); err != nil {
		log.Fatal(err)
	}
}
