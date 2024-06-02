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

	for {

		_, folderAError := os.Open(FolderLink.FolderA)
		_, folderBError := os.Open(FolderLink.FolderB)

		if folderAError == nil && folderBError == nil {
			checkedFiles := []string{}
			walkItems := []WalkItem{}

			walkItems, checkedFiles = walkCheck(FolderLink, false, FolderLink.FolderA, FolderLink.FolderB, checkedFiles, walkItems)

			walkItems, _ = walkCheck(FolderLink, true, FolderLink.FolderB, FolderLink.FolderA, checkedFiles, walkItems)

			checkedFiles = []string{}
			for _, currentWalkItem := range walkItems {
				checkedFiles = walkItemLinkSumCheck(&currentWalkItem, checkedFiles)
			}

			for key, _ := range linkSums[FolderLink.UUID] {
				if !slices.Contains(checkedFiles, key) {
					fmt.Println("I THINK THIS FILE DOESNT EXIST ANYMORE", checkedFiles, key)
				}
			}
			// fmt.Println(" ")
		} else {
			if folderAError != nil {
				fmt.Println("Error loading links A:", folderAError)
			}
			if folderBError != nil {
				fmt.Println("Error loading links B:", folderBError)
			}

		}
	}

}

func walkItemLinkSumCheck(walkItem *WalkItem, checkedFiles []string) []string {
	folderLink, _, relativePath := walkItem.properties()

	if linkSums[folderLink.UUID] == nil {
		fmt.Print("LinkSum at UUID ", folderLink.UUID, " does not exist... Creating one!\n")
		linkSums[folderLink.UUID] = make(map[string]time.Time)
	}

	status, time := walkItem.syncStatus()
	// fmt.Println(status)

	// fmt.Println("\n", relativePath)
	if status != "ERROR" {

		//For when the files are not in sync, keep trying to copy until it reports synced
		for status != "INSYNC" {
			walkItem.quickSync()
			status, time = walkItem.syncStatus()
		}

		//This means that the value @ relative path is not up to date
		if linkSums[folderLink.UUID][relativePath].IsZero() {
			// fmt.Println("linksum reporting zero")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println("added to linksum", relativePath, linkSums[folderLink.UUID][relativePath])

		}

		if linkSums[folderLink.UUID][relativePath] != time {
			// fmt.Println("linksum time not the same")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println("updated linksum", relativePath, linkSums[folderLink.UUID][relativePath])

		}

	}

	checkedFiles = append(checkedFiles, relativePath)
	return checkedFiles

}
