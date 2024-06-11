package main

import (
	"fmt"
	"os"
	"slices"
	"time"
)

// linkSums is what holds all directory info (name+lastModified)
func linkSumRoutine(FolderLink *FolderLink) {
	defer wg.Done()

	fmt.Println("\n=======", FolderLink.Name, "|", FolderLink.UUID, "=======")
	fmt.Println("Folder A: ", FolderLink.FolderA, "\nFolder B:", FolderLink.FolderB)

	var folderHistory []string

	for {

		_, folderAError := os.Stat(FolderLink.FolderA)
		_, folderBError := os.Stat(FolderLink.FolderB)

		if !os.IsNotExist(folderAError) && !os.IsNotExist(folderBError) {
			checkedFiles := []string{}
			walkItems := []WalkItem{}

			walkCheck(FolderLink, false, FolderLink.FolderA, FolderLink.FolderB, &checkedFiles, &walkItems)

			walkCheck(FolderLink, true, FolderLink.FolderB, FolderLink.FolderA, &checkedFiles, &walkItems)

			for _, currentWalkItem := range walkItems {
				walkItemLinkSumCheck(&currentWalkItem, &folderHistory)
			}
			folderHistory = checkedFiles

		} else {
			if os.IsNotExist(folderAError) {
				fmt.Println("Folder A does not exist:", folderAError)
			}
			if os.IsNotExist(folderBError) {
				fmt.Println("Folder B does not exist", folderBError)
			}

		}
	}

}

func walkItemFileRemover(walkItem *WalkItem) {
	source, destination := walkItem.itemPaths()

	os.Remove(source)
	os.Remove(destination)

}

func walkItemLinkSumCheck(walkItem *WalkItem, folderHistory *[]string) {
	folderLink, _, relativePath := walkItem.properties()

	if linkSums[folderLink.UUID] == nil {
		fmt.Print("LinkSum at UUID ", folderLink.UUID, " does not exist... Creating one!\n")
		linkSums[folderLink.UUID] = make(map[string]time.Time)
	}

	status, time := walkItem.syncStatus()

	errorCheck := false
	if len(status) > 6 {
		errorCheck = status[0:7] == "ERROR::"
	}

	fileDeleted := false

	if status == "DESTINATION!=EXIST" {
		// fmt.Println("DESTINATION!=EXIST")
		if slices.Contains(*folderHistory, relativePath) {
			fmt.Println(relativePath, "I think this file was deleted at destination", "\n")
			walkItemFileRemover(walkItem)
			fileDeleted = true
		} else {
			fmt.Println(relativePath, "New file that is not yet at destination")
		}

	}

	if status == "SOURCE!=EXIST" {
		// fmt.Println("SOURCE!=EXIST")
		if slices.Contains(*folderHistory, relativePath) {
			fmt.Println(relativePath, "I think this file was deleted at source", "\n")
			walkItemFileRemover(walkItem)

			fileDeleted = true

		} else {
			fmt.Println(relativePath, "New file that is not yet at source")
		}
	}

	// fmt.Println("\n", relativePath)
	if !errorCheck && !fileDeleted {

		//For when the files are not in sync, keep trying to copy until it reports synced
		for status != "INSYNC" {
			// fmt.Println("NOT INSYNC", status)
			walkItem.quickSync()
			status, time = walkItem.syncStatus()
		}

		//This means that the value @ relative path is not up to date
		if linkSums[folderLink.UUID][relativePath].IsZero() {
			// fmt.Println("linksum reporting zero")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println("added to linksum", relativePath, linkSums[folderLink.UUID][relativePath], "\n")

		}

		if linkSums[folderLink.UUID][relativePath] != time {
			// fmt.Println("linksum time not the same")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println("updated linksum", relativePath, linkSums[folderLink.UUID][relativePath], "\n")

		}

	} else if errorCheck {
		fmt.Println("Error hit!", status, "\n")
	}

}
