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

	for {
		_, folderAError := os.Open(FolderLink.FolderA)
		_, folderBError := os.Open(FolderLink.FolderB)

		if folderAError == nil && folderBError == nil {
			checkedFiles := []string{}
			walkItems := []WalkItem{}
			fmt.Println("Folder A: ", FolderLink.FolderA, "\nFolder B:", FolderLink.FolderB)

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
			fmt.Println(" ")
		} else {
			if folderAError != nil {
				fmt.Println("Error loading links A:", folderAError)
			} else if folderBError != nil {
				fmt.Println("Error loading links B:", folderBError)
			}

		}
	}

}

func walkItemLinkSumCheck(walkItem *WalkItem, checkedFiles []string) []string {
	//Splits the struct into its own variables

	//Paths for file @ each location
	// _, _ := walkItem.buildPaths()
	folderLink, _, relativePath := walkItem.properties()

	if linkSums[folderLink.UUID] == nil {
		fmt.Print("LinkSum at UUID does not exist")
		linkSums[folderLink.UUID] = make(map[string]time.Time)
		fmt.Print("... Creating one!\n")
	}

	status, time := walkItem.syncStatus()
	// fmt.Println(status)

	// fmt.Println("\n", relativePath)
	if status != "ERROR" {
		for status != "INSYNC" {

			walkItem.quickSync()
			checkedFiles = append(checkedFiles, relativePath)
			status = "INSYNC" //DELETE THIS LATER CUZ THIS IS ONLY FOR DEBUGGING (im not copying anything over rn so it will never change)
		}

		if linkSums[folderLink.UUID][relativePath].IsZero() {
			// fmt.Println("linksum reporting zero")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println(linkSums[folderLink.UUID][relativePath])

		}

		if linkSums[folderLink.UUID][relativePath] != time {
			fmt.Println("linksum time not the same")
			linkSums[folderLink.UUID][relativePath] = time
			fmt.Println(linkSums[folderLink.UUID][relativePath])

		}

	}

	checkedFiles = append(checkedFiles, relativePath)
	return checkedFiles

}
