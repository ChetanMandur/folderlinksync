package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// initalize global variables (folderLinks, linkGroups)
var folderLinks FolderLinks
var groups FolderGroups

func main() {
	// ON LAUNCH
	launchSequence()

	for true {
		// for _, currFLink := range folderLinks {
		// 	// folderLinks is a collection of all the links from links.json, link is a singular item from that collection

		// 	//First check the linkSums, and check if there are any differences
		// 	var linkSumACheck bool = linkSumCheck(currFLink.pathA, currFLink.linkSumA)
		// 	var linkSumBCheck bool = linkSumCheck(currFLink.pathB, currFLink.linkSumB)

		// }
		fmt.Println("Hello World!")
		break
	}

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
		fmt.Println("Loaded Folder Link #" + string(i))
	}

	// // group UUID (if it is apart of a group)
	// // // groups.json will hold basic info of all groups in the system
	// // // name and UUID of group

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
