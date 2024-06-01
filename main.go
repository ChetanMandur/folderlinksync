package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// initalize global variables (folderLinks, linkGroups)
var folderLinks FolderLinks

// var linkSums = make(map[string][]linkSumItem)
var linkSums = make(map[int]map[string]time.Time)

//linkSums[folderLinkUUID][relativePath]=modTime

// Global wait group cuz thats how im structuring things
var wg sync.WaitGroup

func continuousCheck(folderLink *FolderLink) {
	defer wg.Done()
	for {

		_, folderAError := os.Open(folderLink.FolderA)
		_, folderBError := os.Open(folderLink.FolderB)

		if folderAError == nil && folderBError == nil {

			checkedFiles := walkCheck(folderLink, false, folderLink.FolderA, folderLink.FolderB, []string(nil))

			walkCheck(folderLink, true, folderLink.FolderB, folderLink.FolderA, checkedFiles)

		}
		if folderAError != nil || folderBError != nil {
			if folderAError != nil {
				fmt.Println("Error loading links A:", folderAError)
			} else if folderBError != nil {
				fmt.Println("Error loading links B:", folderBError)
			}

		}
	}
	// <-make(chan struct{})

}

func main() {
	// ON LAUNCH
	launchSequence()

	// for _, currentFolderLink := range folderLinks.FolderLinks {
	// 	wg.Add(1)
	// 	go continuousCheck(&currentFolderLink)
	// }

	wg.Wait()

}

func launchSequence() {
	fmt.Println("---LAUNCH SEQUENCE---")
	// Start webserver
	// Need to figure out a way to make this constantly running

	// Parse JSONs and load into memory

	// // links.json will hold all A <-> B links, here A and B are both paths to a folder/directory
	// // // Paths to both A and B
	linksJson, linksError := os.ReadFile("links.json")

	if linksError != nil {
		fmt.Println("Error! Cannot find 'links.json' in program directory!")
	}

	// fmt.Println("Opened 'links.json'!")

	json.Unmarshal(linksJson, &folderLinks)

	for i := 0; i < len(folderLinks.FolderLinks); i++ {

		//This is the first time in my life where I realized a pointer was actually needed. Wow
		current := &folderLinks.FolderLinks[i]
		fmt.Println("Loaded Folder Link #" + fmt.Sprint(i+1))
		fmt.Println("--Name: " + fmt.Sprint(current.Name))
		fmt.Println("--UUID: " + fmt.Sprint(current.UUID))
		current.FolderA, _ = filepath.Abs(current.FolderA) //These 2 values werent updating,
		current.FolderB, _ = filepath.Abs(current.FolderB) // I realized the worst
		fmt.Println("--Folder A: " + current.FolderA)
		fmt.Println("--Folder B: " + current.FolderB)
		fmt.Println(current.status, current.FolderRecursion, current.SyncType, current.SpecificExts_FolderA, current.SpecificExts_FolderB, current.SpecificExts_CaseSensitive)
	}

	fmt.Println(" ")

	// Generate/Check linkSums for all links
	// (linkSum is just the current path/file snapshot for one end of a link, one for A and one for B)
	// (ie. each link has two linkSums)
	// (Used to check for changes on link)

	firstBootSync()
	initalizeLinkSum()
	fmt.Println("Launch Sequence Complete!")
}
