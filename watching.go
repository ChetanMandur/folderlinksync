package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func WatcherFactory() {
	watcher, watcherError := fsnotify.NewWatcher()

	if watcherError != nil {
		fmt.Println("error creating a new watcher", watcherError)
		os.Exit(1)
	}

	defer watcher.Close()

	go WatchersLoop(watcher)

	for _, currentFolderLink := range folderLinks.FolderLinks {
		watcherError = watcher.Add(currentFolderLink.FolderA)
		if watcherError != nil {
			os.Exit(1)
		}
		watcherError = watcher.Add(currentFolderLink.FolderB)
		if watcherError != nil {
			os.Exit(1)
		}
	}
	<-make(chan struct{})

}

func EventParser(event fsnotify.Event) Event {
	temp := strings.Split(event.Name, "\\")
	op := event.Op
	fileName := temp[len(temp)-1]
	path := ""
	for i := 0; i < len(temp)-1; i++ {
		path += temp[i]
	}
	isDirectory := !strings.Contains(fileName, ".")
	output := Event{fileName: fileName, path: path, isDirectory: isDirectory, op: op}
	return output
}

func WatchersLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		// Read from Errors.
		case err, ok := <-watcher.Errors:
			fmt.Println(err)

			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

		// Read from Events.
		case e, ok := <-watcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// Just print the event nicely aligned, and keep track how many
			// events we've seen
			event := EventParser(e)
			uuids := FolderUUIDFinder(event.path)
			fmt.Println(event.op, event.fileName, event.path, event.isDirectory, uuids)
		}
	}
}

func FolderUUIDFinder(path string) []int {
	var output []int
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

	for _, currentFolderLink := range folderLinks.FolderLinks {
		currentPathA := nonAlphanumericRegex.ReplaceAllString(currentFolderLink.FolderA, "")
		currentPathB := nonAlphanumericRegex.ReplaceAllString(currentFolderLink.FolderB, "")
		if path == currentPathA || path == currentPathB {
			output = append(output, currentFolderLink.UUID)
		}
	}
	return output
}
