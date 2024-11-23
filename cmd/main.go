package main

import (
	"log"
	"neawn-backend/internal/app"
	"neawn-backend/internal/db"
)

func main() {
	db.Init()
	defer db.Close()

	app := app.New()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
