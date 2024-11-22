package main

import (
	"log"
	"neawn-backend/internal/app"
)

func main() {
	app := app.New()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
