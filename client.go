package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BrandonWade/contact"
	"github.com/BrandonWade/synth"
	"github.com/joho/godotenv"
)

var (
	bufferSize int
	serverHost string
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
	fmt.Printf("Syncing directory %s with server...\n", syncDir)

	serverHost = os.Getenv("TRACE_SERVER_HOST")
	conn := contact.NewConnection(bufferSize)

	conn.Dial(serverHost, "/sync", nil)
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

	promptDownload(newFiles)
}

func promptDownload(files []string) {
	for _, file := range files {
		fmt.Println(file)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Download %d new file(s)? [y/n]: ", len(files))

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error reading download prompt input")
		}

		input := strings.ToLower(response[:len(response)-1])
		if input == "y" || input == "n" {
			if input == "y" {
				downloadFiles(files)
			}

			break
		}
	}
}

func downloadFiles(files []string) {
	for _, file := range files {
		params := make(map[string]string)
		params["file"] = file

		conn := contact.NewConnection(bufferSize)
		conn.Dial(serverHost, "/download", params)
	}
}
