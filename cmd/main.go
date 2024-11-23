package main

import (
	"log"
	"neawn-backend/internal/app"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := app.New()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
