package main

import (
	"log"
	"os"
	"strconv"

	"github.com/BrandonWade/contact"
	"github.com/joho/godotenv"
)

var (
	bufferSize int
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	bufferSize, err = strconv.Atoi(os.Getenv("TRACE_BUFFER_SIZE"))
	if err != nil {
		log.Fatal("error reading buffer size")
	}
}

func main() {
	host := os.Getenv("TRACE_SERVER_HOST")
	conn := contact.NewConnection(bufferSize)

	conn.Dial(host, "/sync")
	defer conn.Close()

	conn.Write("/this/is/a/test/file")
}
