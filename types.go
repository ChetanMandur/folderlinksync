package main

import (
	"os"
	"path/filepath"
	"time"
)

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	Name    string `json:"name"`
	FolderA string `json:"folderA"`
	FolderB string `json:"folderB"`
	//0 = no changes/idle state, 1 = changes detected, 2 = approved sync, 3 = cancel changes (folder B has changes, folder A reject changes and makes B in sync with A)
	status          int
	UUID            int    `json:"uuid"`
	FolderRecursion bool   `json:"folderRecursion"` // whether or not we want to enter/copy folders
	SyncType        string `json:"syncType"`        //'normal' = bidirectional, 'specific-exts' = only looks for files in specified format, creates 1:1 copy with different file ext
	////Additional properties for 'specific-exts' type
	//File format used for each folder, respectively
	SpecificExts_FolderA       string `json:"specificExts_FolderA"`
	SpecificExts_FolderB       string `json:"specificExts_FolderB"`
	SpecificExts_CaseSensitive bool   `json:"specificExts_CaseSensitive"`

	//In/Ex-clude Files
	FileInclude []string `json:"fileInclude"`
	FileExclude []string `json:"fileExclude"`
	//In/Ex-clude Folders
	FolderInclude []string `json:"folderInclude"`
	FolderExclude []string `json:"folderExclude"`
	//In/Ex-clude Extensions
	ExtInclude []string `json:"extInclude"`
	ExtExclude []string `json:"extExclude"`
}
type WalkItem struct {
	FolderLink   *FolderLink
	Inverse      bool
	RelativePath string
	source       string
	destination  string
}

func (w *WalkItem) properties() (FolderLink *FolderLink, Inverse bool, RelativePath string) {
	return w.FolderLink, w.Inverse, w.RelativePath
}

func (w *WalkItem) itemPaths() (source string, destination string) {
	folderLink, inverse, relativePath := w.properties()

	if w.source == "" || w.destination == "" {
		//Building the paths
		if !inverse {
			source = folderLink.FolderA + relativePath
			destination = folderLink.FolderB + relativePath
			if folderLink.SyncType == "specific-exts" {
				source += "." + folderLink.SpecificExts_FolderA
				destination += "." + folderLink.SpecificExts_FolderB
			}
		} else {
			source = folderLink.FolderB + relativePath
			destination = folderLink.FolderA + relativePath
			if folderLink.SyncType == "specific-exts" {
				source += "." + folderLink.SpecificExts_FolderB
				destination += "." + folderLink.SpecificExts_FolderA
			}
		}

		//Clean the compiled paths
		w.source = filepath.Clean(source)
		w.destination = filepath.Clean(destination)

	}

	return w.source, w.destination

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) syncStatus() (status string, mostRecentTime time.Time) {
	// _, _, relativePath := w.properties()
	source, destination := w.itemPaths()

	//Opening the files (YOLO for the errors on source, im hoping no errors cuz it was read w/o error during its walk find)
	sourceFile, _ := os.Open(source)
	sourceFileStat, _ := sourceFile.Stat()
	sourceFileModTime := sourceFileStat.ModTime()

	//Opening destination AND checking for error, cuz error means file does not exist (aka something important)
	destinationFile, destinationError := os.Open(destination)
	// fmt.Println("\n" + relativePath)
	if destinationError == nil {
		destinationFileStat, destinationError := destinationFile.Stat()

		if destinationError == nil {
			// No errors means we are in! We grab the mod time for the destination file and begin our work of finding the most up-to date version
			destinationFileModTime := destinationFileStat.ModTime()

			if sourceFileModTime.After(destinationFileModTime) {

				// fmt.Println("File is more recent in source (", source, sourceFileModTime, ")")
				return "SOURCE=RECENT", sourceFileModTime

			} else if sourceFileModTime.Before(destinationFileModTime) {

				// fmt.Println("File is more recent at destination (", destination, destinationFileModTime, ")")
				return "DESTINATION=RECENT", destinationFileModTime

			} else {
				// fmt.Println("Files are in sync!")
				return "INSYNC", sourceFileModTime
			}

		} else {
			// fmt.Println("Error loading file at destionation ", destination)
			return "ERROR", time.Time{}
		}
	} else {
		// fmt.Println("File does not exist at destination.(", destination, ") Copying and Creating!")
		return "DESTINATION!=EXIST", sourceFileModTime
	}

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) quickSync() {
	source, destination := w.itemPaths()
	status, time := w.syncStatus()
	switch status {
	case "SOURCE=RECENT", "DESTINATION!=EXIST":
		copySync(source, destination, time)
	case "DESTINATION=RECENT":
		copySync(destination, source, time)

	}

}
