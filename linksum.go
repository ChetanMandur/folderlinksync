package main

import (
	"fmt"
	"os"
	"slices"
	"time"
)

// This is the routine that runs when we want continuous watching
func pollingRoutine(FolderLink *FolderLink) {
	defer wg.Done()

	fmt.Println("\n=======", FolderLink.Name, "|", FolderLink.UUID, "=======")
	fmt.Println("Folder A: ", FolderLink.FolderA, "\nFolder B:", FolderLink.FolderB)

	var folderHistory []string
	var errorHit bool

	for {

		_, folderAError := os.Stat(FolderLink.FolderA)
		_, folderBError := os.Stat(FolderLink.FolderB)

		if !os.IsNotExist(folderAError) && !os.IsNotExist(folderBError) {
			if errorHit {
				fmt.Println("Reconnected to", FolderLink.Name, FolderLink.UUID)
			}
			errorHit = false
			checkedFiles := []string{}
			walkItems := []WalkItem{}

			walkingError := WalkCheck(FolderLink, false, &checkedFiles, &walkItems)

			if walkingError != nil {
				continue
			}

			walkingError = WalkCheck(FolderLink, true, &checkedFiles, &walkItems)

			if walkingError != nil {
				continue
			}

			pollingError := false
			for _, currentWalkItem := range walkItems {
				pollingError = walkItemPollingCheck(&currentWalkItem, &folderHistory)
				if pollingError {
					break
				}
			}
			if !pollingError {
				folderHistory = checkedFiles

			}

		} else {
			if os.IsNotExist(folderAError) && !errorHit {
				fmt.Println("Folder A does not exist:", folderAError)
			}
			if os.IsNotExist(folderBError) && !errorHit {
				fmt.Println("Folder B does not exist", folderBError)
			}
			errorHit = true
			folderHistory = []string{}

		}
	}

}

func walkItemFileRemover(walkItem *WalkItem) {
	source, destination := walkItem.itemPaths()

	if walkItem.stillAlive() {
		os.Remove(source)
		os.Remove(destination)
	}

}

func linkSumUpdate(walkItem *WalkItem) (success bool) {
	folderLink, relativePath := walkItem.properties()

	if linkSums[folderLink.UUID] == nil {
		fmt.Print("LinkSum at UUID ", folderLink.UUID, " does not exist... Creating one!\n")
		linkSums[folderLink.UUID] = make(map[string]time.Time)
	}

	status, time := walkItem.syncStatus()

	if status != "INSYNC" {
		return false
	}

	//This means that the linksum value @ uuid @ relative path does not exist
	if linkSums[folderLink.UUID][relativePath].IsZero() {
		// fmt.Println("linksum reporting zero")
		linkSums[folderLink.UUID][relativePath] = time
		fmt.Println("added to linksum", relativePath, "@", linkSums[folderLink.UUID][relativePath])
		fmt.Println()

	}

	//This means that the linksum value @ uuid @ relative path is not up to date
	if linkSums[folderLink.UUID][relativePath] != time {
		// fmt.Println("linksum time not the same")
		linkSums[folderLink.UUID][relativePath] = time
		fmt.Println("updated linksum:", relativePath, "@", linkSums[folderLink.UUID][relativePath])
		fmt.Println()
	}

	return true

}

func walkItemPollingCheck(walkItem *WalkItem, folderHistory *[]string) (success bool) {
	_, relativePath := walkItem.properties()

	status, _ := walkItem.syncStatus()

	errorCheck := false
	if len(status) > 6 {
		errorCheck = status[0:7] == "ERROR::"
	}

	if errorCheck {
		fmt.Println("Error hit!", status)
		switch status {
		case "ERROR::FOLDERLINK_NOT_ALIVE":
			return true
		default:
			return false
		}
	}

	fileDeleted := false

	if status == "FILEB!=EXIST" {
		if !walkItem.stillAlive() {
			return true
		}
		// fmt.Println("DESTINATION!=EXIST")
		if !walkItem.filesExists() {
			if slices.Contains(*folderHistory, relativePath) {
				fmt.Println(relativePath, "I think this file was deleted at Folder B")
				walkItemFileRemover(walkItem)
				fileDeleted = true
			} else {
				fmt.Println(relativePath, "New file that is not yet at Folder B")
			}
		}

	}

	if status == "FILEA!=EXIST" {
		if !walkItem.stillAlive() {
			return true
		}
		// fmt.Println("SOURCE!=EXIST")
		if !walkItem.filesExists() {
			if slices.Contains(*folderHistory, relativePath) {
				fmt.Println(relativePath, "I think this file was deleted at Folder A")
				walkItemFileRemover(walkItem)

				fileDeleted = true

			} else {
				fmt.Println(relativePath, "New file that is not yet at Folder A")
				fmt.Println()

			}
		}
	}

	if fileDeleted {
		return false
	}

	//For when the files are not in sync, keep trying to copy until it reports synced
	for status != "INSYNC" {
		if !walkItem.stillAlive() {
			return true
		}
		// fmt.Println("NOT INSYNC", status)
		status = walkItem.quickSync()
	}

	linkSumUpdate(walkItem)

	return false
}
