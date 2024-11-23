package main

import (
	"fmt"
	"log"
	"neawn-backend/internal/app"
	"neawn-backend/internal/db"

	"github.com/go-redis/redis"
)

func main() {
	db.Init()
	defer db.Close()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	err := rdb.Set("foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get("foo").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("foo", val)

	defer rdb.Close()

	app := app.New()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
