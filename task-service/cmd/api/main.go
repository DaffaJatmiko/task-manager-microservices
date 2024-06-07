package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DaffaJatmiko/task-service/data"
	_ "github.com/go-sql-driver/mysql"
)

const webPort = "80"

type Config struct {
    DB     *sql.DB
    Models data.Models
}

func main() {
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to MySQL!")
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	log.Printf("Starting server on port %s\n", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Println("MySQL not yet ready...")
		} else {
			log.Println("Connected to MySQL!")
			return db
		}

		log.Println("Backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
