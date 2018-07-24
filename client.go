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
	"sync"

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

	syncDir = os.Getenv("SYNC_DIR")
	if syncDir == "" {
		log.Fatal("error reading sync directory")
	}
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
	synth.TrimPaths(localFiles, syncDir)
	localFiles = append(localFiles, &synth.File{})

	for _, file := range localFiles {
		conn.WriteJSON(file)
	}

	newFiles := []synth.File{}
	for {
		file := synth.File{}
		conn.ReadJSON(&file)

		if file.IsEmpty() {
			break
		}

		file.Path = filepath.ToSlash(file.Path)
		newFiles = append(newFiles, file)
	}

	if len(newFiles) > 0 {
		promptDownload(&newFiles)
	}
}

// promptDownload - prompt the user to download all new files from the server
func promptDownload(files *[]synth.File) {
	for _, file := range *files {
		fmt.Println(file.Path)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Download %d new file(s)? [y/n]: ", len(*files))

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
func downloadFiles(files *[]synth.File) {
	var wg sync.WaitGroup
	wg.Add(len(*files))

	for _, file := range *files {
		params := make(map[string]string)
		params["file"] = file.Path

		conn := contact.NewConnection(bufferSize)
		conn.Dial(serverHost, "/download", params)

		go saveFile(conn, file.Path, &wg)
	}

	wg.Wait()
}

// saveFile - read a file from the server and save it to disk
func saveFile(conn *contact.Connection, file string, wg *sync.WaitGroup) {
	defer wg.Done()
	fullPath := syncDir + file

	// Create the file and any parent directories
	filePtr, err := synth.CreateFile(fullPath)
	if err != nil {
		log.Printf("error creating file %s\n", fullPath)
		return
	}
	defer filePtr.Close()

	buffer := bufio.NewWriter(filePtr)
	for {
		// Get the next file block from the server
		_, data, err := conn.Read()
		if err != nil {
			log.Printf("error reading file %s contents from connection\n", fullPath)
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
			log.Printf("error writing file %s contents to disk\n", fullPath)
			return
		}
	}

	buffer.Flush()
	fmt.Printf("File %s saved to disk.\n", file)
}
