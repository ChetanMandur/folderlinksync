package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	copy "github.com/otiai10/copy"
)

type WalkItem struct {
	FolderLink   *FolderLink
	RelativePath string
	FileAPath    string
	FileBPath    string
}

func (w *WalkItem) stillAlive() (output bool) {

	return w.FolderLink.stillAlive() == 0

}

func (w *WalkItem) itemsExist() (output bool) {
	fileA, fileB := w.itemPaths()

	//CHECK IF FOLDER LINK ROOT STILL EXISTS
	_, fileAError := os.Stat(fileA)
	_, fileBError := os.Stat(fileB)

	fileAExists := !os.IsNotExist(fileAError)
	fileBExists := !os.IsNotExist(fileBError)

	if !fileAExists || !fileBExists {
		return false
	}

	return true

}

func (w *WalkItem) properties() (FolderLink *FolderLink, RelativePath string) {
	return w.FolderLink, w.RelativePath
}

func (w *WalkItem) itemPaths() (fileAPath string, fileBPath string) {
	folderLink, relativePath := w.properties()

	if w.FileAPath == "" || w.FileBPath == "" {
		//Building the paths

		fileAPath = folderLink.FolderA + relativePath
		fileBPath = folderLink.FolderB + relativePath
		if folderLink.SyncType == "specific-exts" {
			fileAPath += "." + folderLink.SpecificExts_FolderA
			fileBPath += "." + folderLink.SpecificExts_FolderB
		}

		//Clean the compiled paths
		w.FileAPath = filepath.Clean(fileAPath)
		w.FileBPath = filepath.Clean(fileBPath)

	}

	return w.FileAPath, w.FileBPath

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) syncStatus() (status string, recentTime time.Time) {
	fileAPath, fileBPath := w.itemPaths()

	if !w.stillAlive() {
		return "ERROR::FOLDERLINK_NOT_ALIVE", recentTime
	}

	fileA, fileAError := os.Stat(fileAPath)
	fileB, fileBError := os.Stat(fileBPath)

	fileAExists := !os.IsNotExist(fileAError)
	fileBExists := !os.IsNotExist(fileBError)

	if fileAExists && fileAError != nil {
		return "ERROR::FILEA_error_exists:FILEA_exists", recentTime
	}

	if fileBExists && fileBError != nil {
		return "ERROR::FILEB_error_exists:FILEB_exists", recentTime
	}

	if fileAExists && !fileBExists {

		return "FILEB!=EXIST", w.FolderLink.roundedModTime(fileA)
	}

	if !fileAExists && fileBExists {
		return "FILEA!=EXIST", w.FolderLink.roundedModTime(fileB)

	}

	if !fileAExists && !fileBExists {
		fmt.Println(fileAPath, fileBPath)
		return "ERROR::FILEA_doesnt_exist:FILEB_doesnt_exist", recentTime
	}

	if fileAExists && fileBExists {
		fileAModTime := w.FolderLink.roundedModTime(fileA)
		fileBModTime := w.FolderLink.roundedModTime(fileB)

		// No errors means we are in! We grab the mod time for the destination file and begin our work of finding the most up-to date version

		if fileAModTime.After(fileBModTime) {

			// fmt.Println("File is more recent in source (", source, sourceFileModTime, ")")
			return "FILEA=RECENT", fileAModTime

		} else if fileAModTime.Before(fileBModTime) {

			// fmt.Println("File is more recent at destination (", destination, destinationFileModTime, ")")
			return "FILEB=RECENT", fileBModTime

		} else {
			// fmt.Println("Files are in sync!")
			return "INSYNC", fileAModTime
		}

	}

	return "ERROR::END_OF_CASES", recentTime

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) quickSync() (status string, recentTime time.Time) {
	status, recentTime = w.syncStatus()

	switch status {
	case "FILEA=RECENT", "FILEB!=EXIST":
		w.copySync(false, recentTime)
	case "FILEB=RECENT", "FILEA!=EXIST":
		w.copySync(true, recentTime)
	}

	return status, recentTime

}

// Copies file to destination, and updates the file's mod time to the given mod time
func (w *WalkItem) copySync(inverse bool, recentTime time.Time) (success bool) {

	//True= actually copies False = doesnt copy; both print the paths it was gonna use (to -> from)
	yes := true

	var source, destination string

	if !inverse {
		source, destination = w.itemPaths()
	} else {
		destination, source = w.itemPaths()
	}

	if !w.stillAlive() {
		return false
	}

	sourceItem, sourceErr := os.Stat(source)
	destinationItem, destinationErr := os.Stat(destination)

	if !recentTime.Equal(w.FolderLink.roundedModTime(sourceItem)) {
		return false
	}

	var destinationModTime time.Time

	if sourceErr != nil {
		return false
	}

	if destinationErr == nil {
		destinationModTime = w.FolderLink.roundedModTime(destinationItem)
	}

	if !w.stillAlive() {
		return false
	}

	if yes {
		copy.Copy(source, destination)
		// os.Chtimes(source, recentTime, recentTime)
		w.FolderLink.fileModDelay()
		os.Chtimes(destination, recentTime, recentTime)
	} else {
		fmt.Print("😈")
	}

	fmt.Print("Copy!\n  ", source, "\n  (", recentTime, ")\n  --->\n  ", destination, "\n  (", destinationModTime, ")\n")

	return w.itemsExist()
}

func (w *WalkItem) pollingCheck() (pollingError bool) {
	_, relativePath := w.properties()

	status, recentTime := w.syncStatus()

	errorCheck := errorCheck(status)

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
		if !w.stillAlive() {
			return true
		}
		// fmt.Println("DESTINATION!=EXIST")
		if !w.itemsExist() {
			if slices.Contains(*w.FolderLink.folderHistory, relativePath) {
				fmt.Println(relativePath, "File was deleted @ Folder B")
				w.itemsRemover()
				fileDeleted = true
			} else {
				fmt.Println(relativePath, "New file created @ Folder A")
			}
		}

	}

	if status == "FILEA!=EXIST" {
		if !w.stillAlive() {
			return true
		}
		// fmt.Println("SOURCE!=EXIST")
		if !w.itemsExist() {
			if slices.Contains(*w.FolderLink.folderHistory, relativePath) {
				fmt.Println(relativePath, "deleted @ Folder A")
				w.itemsRemover()

				fileDeleted = true

			} else {
				fmt.Println(relativePath, "created @ Folder B")

			}
		}
	}

	if fileDeleted {
		fmt.Println()
		return false
	}

	//For when the files are not in sync, keep trying to copy until it reports synced
	for status != "INSYNC" {
		if !w.stillAlive() {
			return true
		}
		// fmt.Println("NOT INSYNC", status)
		status, recentTime = w.quickSync()
	}

	w.linkSumUpdate(recentTime)

	return false
}

func (w *WalkItem) linkSumUpdate(recentTime time.Time) (success bool) {
	folderLink, relativePath := w.properties()

	if linkSums[folderLink.UUID] == nil {
		fmt.Print("LinkSum at UUID ", folderLink.UUID, " does not exist... Creating one!\n")
		linkSums[folderLink.UUID] = make(map[string]time.Time)
	}

	status, time := w.syncStatus()

	if status != "INSYNC" {
		return false
	}

	if time != recentTime {
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

func (w *WalkItem) itemsRemover() {
	source, destination := w.itemPaths()

	if w.stillAlive() {
		os.Remove(source)
		os.Remove(destination)
	}

}
