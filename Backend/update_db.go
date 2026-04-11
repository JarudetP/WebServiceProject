package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
)

func main() {
	dsn := "host=localhost port=5435 user=admin password=admin1234 dbname=game_data_platform sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}

	res, err := db.Exec("UPDATE games SET current_players = (20000 + floor(random() * 30001))::INT")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := res.RowsAffected()
	fmt.Printf("Updated %d games\n", rows)
}
