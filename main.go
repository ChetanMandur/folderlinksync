package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// initalize global variables (folderLinks, linkGroups)
var folderLinks FolderLinks
var folderGroups FolderGroups

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

	// fmt.Println("Opened 'groups.json'!")

	json.Unmarshal(groupsJson, &folderGroups)

	for i := 0; i < len(folderGroups.FolderGroups); i++ {
		fmt.Println("Loaded Folder Group #" + fmt.Sprint(i+1))
		fmt.Println("--Name: " + folderGroups.FolderGroups[i].Name)
		fmt.Println("--UUID: " + fmt.Sprint(folderGroups.FolderGroups[i].UUID))
	}
	fmt.Println(" ")

	// Generate/Check linkSums for all links
	// (linkSum is just the current path/file snapshot for one end of a link, one for A and one for B)
	// (ie. each link has two linkSums)
	// (Used to check for changes on link)

	initalizeLinkSum()
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
