package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BrandonWade/contact"
	"github.com/BrandonWade/synth"
	"github.com/joho/godotenv"
)

var (
	bufferSize int
	syncDir    string
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

	syncDir = os.Getenv("TEST_DIR")
}

func main() {
	host := os.Getenv("TRACE_SERVER_HOST")
	conn := contact.NewConnection(bufferSize)

	conn.Dial(host, "/sync")
	defer conn.Close()

	// Get the list of files from the filesystem
	localFiles, err := synth.Scan(syncDir)
	if err != nil {
		log.Fatal("error retrieving local file list")
	}

	// Add an empty element to the end of the list to indicate the end
	localFiles = synth.TrimPaths(localFiles, syncDir)
	localFiles = append(localFiles, "")

	for _, file := range localFiles {
		conn.Write(file)
	}

	newFiles := []string{}
	for {
		_, msg, _ := conn.Read()

		data := string(msg)
		if data == "" {
			break
		}

		path := filepath.ToSlash(data)
		newFiles = append(newFiles, path)
	}
}
