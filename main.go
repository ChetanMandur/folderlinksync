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

// var linkSums = make(map[string][]linkSumItem) //linkSums[folderLinkUUID][relativePath]=modTime
var linkSums = make(map[int]map[string]time.Time)

// Global wait group cuz thats how im structuring things
var wg sync.WaitGroup

func main() {
	alive := true
	delay := 5
	delayInc := 3
	delayMax := 59
	for alive {
		if launchSequence() {
			delay = 5
			for _, currentFolderLink := range folderLinks.FolderLinks {
				if currentFolderLink.ContinuousSync {
					wg.Add(1)
					go pollingRoutine(&currentFolderLink)
				}
			}
			wg.Wait()

			alive = false
		} else {
			fmt.Println("Launch failed! Retrying in", delay, "seconds...")
			time.Sleep(time.Duration(delay) * time.Second)
			if delay >= delayMax {
				alive = false
			}
			delay += delayInc
		}
	}
	fmt.Println("Closing... goodbye!")
	time.Sleep(time.Second * 5)
}

func launchSequence() (succcess bool) {
	fmt.Println("---LAUNCH SEQUENCE---")
	// Start webserver
	// Need to figure out a way to make this constantly running

	// Parse JSONs and load into memory

	// // links.json will hold all A <-> B links, here A and B are both paths to a folder/directory
	// // // Paths to both A and B
	linksJson, linksError := os.ReadFile("links3.json")

	if linksError != nil {
		fmt.Println("Error! Cannot find 'links.json' in program directory!")
		return false
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
		// fmt.Println(current)
		// fmt.Println(current.FolderRecursion, current.SyncType, current.SpecificExts_FolderA, current.SpecificExts_FolderB, current.SpecificExts_CaseSensitive)
		// fmt.Println(current.FileExclude, current.FileInclude, current.ExtExclude, current.ExtInclude, current.FolderExclude, current.FolderInclude)
	}

	fmt.Println(" ")

	firstBootSync()
	fmt.Println("================Launch Sequence Complete!")
	return true
}

func firstBootSync() {
	fmt.Println("---FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		if !currentFolderLink.SyncOnStartup {
			continue
		}

		fmt.Println("\n=======", currentFolderLink.Name, "|", currentFolderLink.UUID, "=======")

		_, folderAError := os.Open(currentFolderLink.FolderA)
		_, folderBError := os.Open(currentFolderLink.FolderB)

		if folderAError != nil || folderBError != nil {
			if folderAError != nil {
				fmt.Println("Error loading links A:", folderAError)
			}
			if folderBError != nil {
				fmt.Println("Error loading links B:", folderBError)
			}
			continue
		}

		checkedFiles := []string{}
		walkItems := []WalkItem{}
		fmt.Println("Folder A: ", currentFolderLink.FolderA, "\nFolder B:", currentFolderLink.FolderB)

		// fmt.Println("\nWALKING FOLDER ", currentFolderLink.FolderA, "-> ("+currentFolderLink.FolderB+")")
		walkingError := WalkCheck(&currentFolderLink, false, &checkedFiles, &walkItems)
		if walkingError != nil {
			continue
		}
		// fmt.Println("\nWALKING FOLDER ", currentFolderLink.FolderB, "-> ("+currentFolderLink.FolderA+")")
		walkingError = WalkCheck(&currentFolderLink, true, &checkedFiles, &walkItems)

		if walkingError != nil {
			continue
		}

		// fmt.Println("HELLO WORLD\n", len(walkItems), len(checkedFiles), checkedFiles)
		for _, currentWalkItem := range walkItems {
			currentWalkItem.quickSync()
		}

		fmt.Println(" ")
	}
}
