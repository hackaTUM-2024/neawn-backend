package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func Init() {
	var err error
	connString := "postgres://postgres:mysecretpassword@localhost:5433/postgres?sslmode=disable"

	Conn, err = pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	fmt.Println("Connected to database")
}

func Close() {
	Conn.Close(context.Background())
}
