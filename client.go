package main

import (
	"bufio"
	"bytes"
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

	// TODO: This is a hack to let the goroutines finish running before terminating
	for {
	}
}

// promptDownload - prompt the user to download all new files from the server
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

// downloadFiles - download the provided list of files from the server
func downloadFiles(files []string) {
	for _, file := range files {
		params := make(map[string]string)
		params["file"] = file

		conn := contact.NewConnection(bufferSize)
		conn.Dial(serverHost, "/download", params)

		go saveFile(conn, file)
	}
}

// saveFile - read a file from the server and save it to disk
func saveFile(conn *contact.Connection, file string) {
	filePath := syncDir + file

	filePtr, err := os.Create(filePath)
	if err != nil {
		log.Printf("error creating file %s\n", filePath)
	}
	defer filePtr.Close()

	buffer := bufio.NewWriter(filePtr)
	for {
		// Get the next file block from the server
		_, data, err := conn.Read()
		if err != nil {
			log.Printf("error reading file %s contents from connection\n", filePath)
			return
		}

		// An empty message indicates the end of the file
		// NOTE: Is this necessary? Might be able to just close the Connection instead
		if len(data) == 0 {
			break
		}

		// Remove any trailing NUL bytes
		data = bytes.TrimRight(data, "\x00")

		// Write the block to disk
		_, err = buffer.Write(data)
		if err != nil {
			log.Printf("error writing file %s contents to disk\n", filePath)
			return
		}
	}

	buffer.Flush()
	fmt.Printf("File %s saved to disk.\n", file)
}
