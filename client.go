package main

import (
	"log"
	"os"

	"github.com/BrandonWade/contact"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
}

func main() {
	host := os.Getenv("TRACE_SERVER_HOST")
	conn := contact.NewConnection()
	conn.Dial(host)
}
