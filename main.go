package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// initalize global variables (folderLinks, linkGroups)
var folderLinks FolderLinks

// var linkSums = make(map[string][]linkSumItem)
var linkSums = make(map[string]map[string]time.Time)

func main() {
	// ON LAUNCH
	launchSequence()
	// for {
	// 	WatcherFactory()
	// }
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
	fmt.Println("---LAUNCH SEQUENCE---")
	// Start webserver
	// Need to figure out a way to make this constantly running

	// Parse JSONs and load into memory

	// // links.json will hold all A <-> B links, where A and B are both paths to a folder/directory
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

	// initalizeLinkSum()
	// firstBootSync()
	firstBootSync()
	fmt.Println("Launch Sequence Complete!")
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
