package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/radovskyb/watcher"
)
var sourcePath string = "path/to/.m2/on/wsl"
var destinationPath string = "/path/to/.m2/on/windows"

func main() {
	watcherService := watcher.New()


    log.Printf("Adding [%s] to the watcher service. This might take a while depending on the depth of the source directory\n", sourcePath)
	if err := watcherService.AddRecursive(sourcePath); err != nil {
		log.Fatalln(err)
	} else {
        log.Println("Watcher has been initialized")
    }

    go listenForEvents(watcherService)

	if err := watcherService.Start(time.Millisecond * 10); err != nil {
		log.Fatalln(err)
	} 
}

func listenForEvents(watcherService *watcher.Watcher) {
    log.Println("Started listening to events...")
    for {
        select { 
        case event := <-watcherService.Event:	
            handleEvent(event)
        case err := <-watcherService.Error:
            log.Fatalln(err)
        case <-watcherService.Closed:
            return
        }
    }
}

func handleEvent(event watcher.Event) {
    switch event.Op {
        case watcher.Write:
            if !event.IsDir() {
                go copyFile(event)
            }
        default:
            // do nothing
    }
}

func copyFile(event watcher.Event) {
    sourceReader, err := os.Open(event.Path)
    handleError("While reading file " + event.Path, err)
    defer sourceReader.Close()

    destinationFilePath := destinationPath + strings.TrimPrefix(event.Path, sourcePath)
    prepareDestinationDirectories(destinationFilePath)

    destinationWriter, err := os.Create(destinationFilePath)
    handleError("While creating a writer for " + destinationFilePath, err)
    defer destinationWriter.Close()

    _, err = io.Copy(destinationWriter, sourceReader)
    handleError("While copying into " + destinationFilePath, err)

    log.Printf("Copied [%s] to [%s]\n", event.Path, destinationFilePath)
}

func prepareDestinationDirectories(destinationPath string) {
    dirPath := filepath.Dir(destinationPath)
    err := os.MkdirAll(dirPath, os.ModePerm)
    handleError("While preparing directories for " + destinationPath, err)
}

func handleError(message string, err error) {
    if err != nil {
        log.Printf("ERROR: %s: %s\n", message, err)
    }
}
