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
	// Helper variables used within the main loop

	// Used to keep the main-loop alive
	alive := true
	// Indicates the delay before retrying the launch
	delay := 5
	// When a launch fails, increase the delay by this amount
	delayInc := 3
	// When the delay reaches this length, give up and terminate the app
	delayMax := 59

	// Main loop for the app
	for alive {
		// If the launch was a success
		if launchSequence() {
			// Reset the delay counter
			delay = 5
			// Iterate through all the links
			for _, currentFolderLink := range folderLinks.FolderLinks {
				// If 'ContinuousSync` is enabled, launch the pollingRoutine
				// Does this make sense? Should I even keep this parameter
				if currentFolderLink.ContinuousSync {
					// Add an increment to the WaitGroup to signify that a goroutine has been added
					wg.Add(1)
					// Begin goroutine that does the polling for
					go currentFolderLink.pollingRoutine()
				}
			}
			// Wait for all goroutines to end
			wg.Wait()

			// Once all goroutines are closed, end main loop
			alive = false
		} else { // If the launch failed

			// Wait a `delay` amount of seconds
			fmt.Println("Launch failed! Retrying in", delay, "seconds...")
			time.Sleep(time.Duration(delay) * time.Second)

			// if the delay is at the max, end the main loop
			if delay >= delayMax {
				alive = false
			}

			// Increment the delay by designated amount
			delay += delayInc
		}
	}
	// Closing messages before shutting down
	fmt.Println("Closing... goodbye!")
	time.Sleep(time.Second * 5)
}

// Initiates the launch of the app
// Used to load data into memory
//
// Used to also launch services:
//   - [firstBootSync]
func launchSequence() (succcess bool) {
	fmt.Println("---LAUNCH SEQUENCE---")
	// Start webserver
	// 	- Need to figure out a way to make this constantly running

	// Parse JSONs and load into memory
	// 	- links.json will hold all A <-> B links, here A and B are both paths to a folder/directory
	// 		- Paths to both A and B

	// This block was done to add in-line documentation to the variable names
	var (
		// The actual data retrieved from the JSON
		linksJson []byte

		// Any error generated from grabbing the JSON
		linksError error
	)

	linksJson, linksError = os.ReadFile("userFiles/links.json")

	// Ends early if there was an error
	// No matter the error, will spit out a generic error stating the `links.json` cannot be found
	if linksError != nil {
		fmt.Println("Error! Cannot find 'links.json' in program directory!")
		return false
	}

	// fmt.Println("Opened 'links.json'!")

	// The JSON file is decode and placed into the global `folderLinks` variable
	json.Unmarshal(linksJson, &folderLinks)

	// Iterate through all the `links` and display its contents
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

	// Launch the `firstBootSync` sequence
	firstBootSync()

	// Once everything has ran without error, this means that the launch sequence completed without error
	// Return `true` to signifiy a successful boot
	fmt.Println("================Launch Sequence Complete!")
	return true
}

// Used to sync the folders after a cold boot
//
// Essentially, this syncs all the `links` once, and provides CLI printouts
func firstBootSync() {
	fmt.Println("---FIRST BOOT SYNC---")

	// Iterate through all the links
	for _, currentFolderLink := range folderLinks.FolderLinks {

		// Checks if the file should be allowed to sync on startup
		// NOTE: This is dumb, `sync on startup` should be enabled no matter what
		// Maybe this should be replaced with a 'paused' state? works better, and can be integrated into polling
		if !currentFolderLink.SyncOnStartup {
			continue
		}

		fmt.Println("\n=======", currentFolderLink.Name, "|", currentFolderLink.UUID, "=======")

		// Checks if the folders within the link exist
		stillAlive := currentFolderLink.stillAlive()
		if stillAlive != 0 {
			switch stillAlive {
			case 1:
				fmt.Println("Error loading folder A")
			case 2:
				fmt.Println("Error loading folder B")
			case 3:
				fmt.Println("Error loading both folders")
			}
			continue
		}

		fmt.Println("Folder A: ", currentFolderLink.FolderA, "\nFolder B:", currentFolderLink.FolderB)

		// If the folders exist, initiate a sync between the folders
		currentFolderLink.quickSync()

		fmt.Println(" ")
	}
}
