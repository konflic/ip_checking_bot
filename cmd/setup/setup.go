package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {

	DEFAULT_ADMIN := "s_amurai"
	DATABASE_URL := "postgres://postgres:root@localhost:3306"
	db, _ := sql.Open("pgx", DATABASE_URL)

	if err := db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	db.Exec("CREATE TABLE ipbotdb (id SERIAL PRIMARY KEY, username VARCHAR(256), ip_request VARCHAR(256), ip_result TEXT, chat_id VARCHAR(256));")
	db.Exec("CREATE TABLE botadmins (id SERIAL PRIMARY KEY, username VARCHAR(256));")
	db.Exec("INSERT INTO botadmins (username) VALUES ($1);", DEFAULT_ADMIN)

	fmt.Println("Databases crated and default admin added!")

}
