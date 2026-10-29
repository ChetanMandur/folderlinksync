package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	copy "github.com/otiai10/copy"
)

/*
Structure used to represent a file marked for sync

"marked for sync" means that the file is valid within the definition of a `FolderLink`, and can be shared between both folders

Holds a reference to it's origin `FolderLink` and the `RelativePath` to the file (ie. where the file will be located within both `FolderA` and `FolderB`)
*/
type WalkItem struct {
	FolderLink   *FolderLink
	RelativePath string
	FileAPath    string
	FileBPath    string
}

// Check to see if the related/origin `FolderLink` of a `WalkItem` is still alive
func (w *WalkItem) stillAlive() (output bool) {

	return w.FolderLink.stillAlive() == 0

}

// Check to see if the item exists within both folders
//
// ie. If the file exists in both `Folder A` and `Folder B`
func (w *WalkItem) itemsExist() (output bool) {
	fileA, fileB := w.itemPaths()

	_, fileAError := os.Stat(fileA)
	_, fileBError := os.Stat(fileB)

	fileAExists := !os.IsNotExist(fileAError)
	fileBExists := !os.IsNotExist(fileBError)

	if !fileAExists || !fileBExists {
		return false
	}

	return true

}

// Getter function to retrieve the main defining properties of a `WalkItem`
//
// In other words, returns:
//   - The `FolderLink` the `WalkItem` originates from
//   - The `RelativePath` for the file within both folders
func (w *WalkItem) properties() (FolderLink *FolderLink, RelativePath string) {
	return w.FolderLink, w.RelativePath
}

// Getter function to generate and retrieve the full paths of an item in both `Folder A` and `Folder B`
//
// # Unlike the `RelativePath`, these paths represent a complete path from `system root` to the item in each of the folders
//
// This function is creates a 'cache' within the `WalkItem` when first called (since these paths are stagnant).
// So, on subsequent calls, this function simply returns the 'cached' field/property
func (w *WalkItem) itemPaths() (fileAPath string, fileBPath string) {
	folderLink, relativePath := w.properties()

	if w.FileAPath == "" || w.FileBPath == "" { // If the full paths have not been generated yet

		// Build the paths
		fileAPath = folderLink.FolderA + relativePath
		fileBPath = folderLink.FolderB + relativePath

		// If `SyncType` is `specific-exts`, this means that the `relativePath` has no file extension
		// Because of this, we need to re-add the extension, but using the folders specificly defined extension
		if folderLink.SyncType == "specific-exts" {
			fileAPath += "." + folderLink.SpecificExts_FolderA
			fileBPath += "." + folderLink.SpecificExts_FolderB
		}

		// Clean the compiled paths
		w.FileAPath = filepath.Clean(fileAPath)
		w.FileBPath = filepath.Clean(fileBPath)

	}

	return w.FileAPath, w.FileBPath

}

// Outputs the current status of the two files refered to within `WalkItem`
//
// Provides clear "ERROR" print-outs when file/folder retrival cause errors by the system
//
// For non-errors, provides print-outs stating:
//   - If the file exists within both folders, and if not, state the folder(s) the file does not exist in
//   - If the file within each folder and equal to each other, and if not, which one is more recent
//
// Error print-outs begin with "ERROR::", while non-error print-outs do not
func (w *WalkItem) syncStatus() (status string, recentTime time.Time) {
	fileAPath, fileBPath := w.itemPaths()

	if !w.stillAlive() {
		return "ERROR::FOLDERLINK_NOT_ALIVE", recentTime
	}

	fileA, fileAError := os.Stat(fileAPath)
	fileB, fileBError := os.Stat(fileBPath)

	fileAExists := !os.IsNotExist(fileAError)
	fileBExists := !os.IsNotExist(fileBError)

	// Checks for when a file exists but some other error appears
	if fileAExists && fileAError != nil {
		return "ERROR::FILEA_error_exists:FILEA_exists", recentTime
	}
	if fileBExists && fileBError != nil {
		return "ERROR::FILEB_error_exists:FILEB_exists", recentTime
	}

	// Checks for when a file exists only in one folder
	if fileAExists && !fileBExists {

		return "FILEB!=EXIST", w.FolderLink.roundedModTime(fileA)
	}
	if !fileAExists && fileBExists {
		return "FILEA!=EXIST", w.FolderLink.roundedModTime(fileB)

	}

	// Check for when file doesn't exist in both folders
	if !fileAExists && !fileBExists {
		fmt.Println(fileAPath, fileBPath)
		return "ERROR::FILEA_doesnt_exist:FILEB_doesnt_exist", recentTime
	}

	// Checks for when the file exists in both folders
	// Compares files to find the most up-to-date version
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

// Outputs result from `syncStatus()`, but also syncs files if they are found to be out-of-sync
func (w *WalkItem) quickSync() (status string, recentTime time.Time) {
	status, recentTime = w.syncStatus()

	switch status {
	case "FILEA=RECENT", "FILEB!=EXIST":
		w.copySync(false, recentTime) // Syncs file copy from `Folder A` to `Folder B`
	case "FILEB=RECENT", "FILEA!=EXIST":
		w.copySync(true, recentTime) // Syncs file copy from `Folder B` to `Folder A`
	}

	return status, recentTime

}

// Copies file to destination, and updates the file's mod time to the given mod time
func (w *WalkItem) copySync(inverse bool, recentTime time.Time) (success bool) {

	// True= actually copies False = doesnt copy; both print the paths it was gonna use (to -> from)
	// goated debug line LMAO wtf was I on
	yes := true

	var source, destination string

	//
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
