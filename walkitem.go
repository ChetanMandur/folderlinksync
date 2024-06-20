package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	folderLink, _ := w.properties()

	//CHECK IF FOLDER LINK ROOTS STILL EXISTS
	_, folderAError := os.Stat(folderLink.FolderA)
	_, folderBError := os.Stat(folderLink.FolderB)

	folderAExists := !os.IsNotExist(folderAError)
	folderBExists := !os.IsNotExist(folderBError)

	if !folderAExists || !folderBExists {
		return false
	}

	return true

}

func (w *WalkItem) filesExists() (output bool) {
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
func (w *WalkItem) syncStatus() (status string, time time.Time) {
	fileAPath, fileBPath := w.itemPaths()

	if !w.stillAlive() {
		return "ERROR::FOLDERLINK_NOT_ALIVE", time
	}

	fileA, fileAError := os.Stat(fileAPath)
	fileB, fileBError := os.Stat(fileBPath)

	fileAExists := !os.IsNotExist(fileAError)
	fileBExists := !os.IsNotExist(fileBError)

	if fileAExists && fileAError != nil {
		return "ERROR::FILEA_error_exists:FILEA_exists", time
	}

	if fileBExists && fileBError != nil {
		return "ERROR::FILEB_error_exists:FILEB_exists", time
	}

	if fileAExists && !fileBExists {
		return "FILEB!=EXIST", fileA.ModTime()
	}

	if !fileAExists && fileBExists {
		return "FILEA!=EXIST", fileB.ModTime()

	}

	if !fileAExists && !fileBExists {
		fmt.Println(fileAPath, fileBPath)
		return "ERROR::FILEA_doesnt_exist:FILEB_doesnt_exist", time
	}

	if fileAExists && fileBExists {
		fileAModTime := fileA.ModTime()
		fileBModTime := fileB.ModTime()

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

	return "ERROR::END_OF_CASES", time

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) quickSync() (status string) {
	status, _ = w.syncStatus()

	switch status {
	case "FILEA=RECENT", "FILEB!=EXIST":
		w.copySync(false)
	case "FILEB=RECENT", "FILEA!=EXIST":
		w.copySync(true)
	}

	return status

}

// Copies file to destination, and updates the file's mod time to the given mod time
func (w *WalkItem) copySync(inverse bool) (success bool) {

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

	var sourceModTime time.Time
	var destinationModTime time.Time

	if sourceErr != nil {
		return false
	}

	sourceModTime = sourceItem.ModTime()

	if destinationErr == nil {
		destinationModTime = destinationItem.ModTime()
	}

	if !w.stillAlive() {
		return false
	}

	if yes {
		copy.Copy(source, destination)
		os.Chtimes(source, sourceModTime, sourceModTime)
		os.Chtimes(destination, sourceModTime, sourceModTime)
	} else {
		fmt.Print("😈")
	}

	fmt.Print("Copy!\n  ", source, "\n  (", sourceModTime, ")\n  --->\n  ", destination, "\n  (", destinationModTime, ")\n")

	return w.filesExists()
}
