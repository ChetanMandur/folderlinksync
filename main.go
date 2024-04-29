package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

// initalize global variables (folderLinks, linkGroups)
var folderLinks FolderLinks
var folderGroups FolderGroups

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

func WatchersLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		// Read from Errors.
		case err, ok := <-watcher.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			fmt.Println(err)

		// Read from Events.
		case e, ok := <-watcher.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// Just print the event nicely aligned, and keep track how many
			// events we've seen
			fmt.Println(e)
		}
	}
}

func main() {
	// ON LAUNCH
	launchSequence()
	WatcherFactory()

	// for {
	// 	for _, currentFolderLink := range folderLinks.FolderLinks {
	// 		// folderLinks is a collection of all the links from links.json, link is a singular item from that collection

	// 		//First check the linkSums, and check if there are any differences
	// 		// var linkSumACheck bool = linkSumCheck(currFLink.pathA, currFLink.linkSumA)
	// 		// var linkSumBCheck bool = linkSumCheck(currFLink.pathB, currFLink.linkSumB)

	// 	}

	// }

}

func launchSequence() {
	// Start webserver
	// Need to figure out a way to make this constantly running

	// Parse JSONs and load into memory

	// // links.json will hold all A <-> B links, where A and B are both paths to a folder/directory
	// // // Paths to both A and B
	linksJson, linksError := os.ReadFile("links.json")

	if linksError != nil {
		fmt.Println("Error! Cannot find 'links.json' in program directory!")
	}

	fmt.Println("Opened 'links.json'!")

	json.Unmarshal(linksJson, &folderLinks)

	for i := 0; i < len(folderLinks.FolderLinks); i++ {
		fmt.Println("Loaded Folder Link #" + fmt.Sprint(i+1))
		fmt.Println("--Name: " + fmt.Sprint(folderLinks.FolderLinks[i].Name))
		fmt.Println("--UUID: " + fmt.Sprint(folderLinks.FolderLinks[i].UUID))
		if folderLinks.FolderLinks[i].GroupUUID != 0 {
			fmt.Println("--Group UUID: " + fmt.Sprint(folderLinks.FolderLinks[i].GroupUUID))
		}
		fmt.Println("--Folder A: " + folderLinks.FolderLinks[i].FolderA)
		fmt.Println("--Folder B: " + folderLinks.FolderLinks[i].FolderB)
	}

	fmt.Println(" ")

	// // group UUID (if it is apart of a group)
	// // // groups.json will hold basic info of all groups in the system
	// // // name and UUID of group
	groupsJson, groupsError := os.ReadFile("groups.json")

	if groupsError != nil {
		fmt.Println("Error! Cannot find 'groups.json' in program directory!")
	}

	fmt.Println("Opened 'groups.json'!")

	json.Unmarshal(groupsJson, &folderGroups)

	for i := 0; i < len(folderGroups.FolderGroups); i++ {
		fmt.Println("Loaded Folder Group #" + fmt.Sprint(i+1))
		fmt.Println("--Name: " + folderGroups.FolderGroups[i].Name)
		fmt.Println("--UUID: " + fmt.Sprint(folderGroups.FolderGroups[i].UUID))
	}

	// Generate/Check linkSums for all links
	// (linkSum is just the current path/file snapshot for one end of a link, one for A and one for B)
	// (ie. each link has two linkSums)
	// (Used to check for changes on link)

	// // if linkSums doesn't exist, generate it
	// // if linkSums exists, update it
	fmt.Println("Launch sequence")
}

func linkSumCheck(fPath string, linkSum string) bool {
	//Temporarily create new linkSum using fPath
	//Compares temp with given
	if true {
		return true
	} else {
		return false
	}
}
