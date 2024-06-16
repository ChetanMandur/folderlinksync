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
	//// General info (Bare minimum)
	Name    string `json:"name"`
	FolderA string `json:"folderA"`
	FolderB string `json:"folderB"`
	UUID    int    `json:"uuid"`

	//// General syncing settings
	FolderRecursion bool   `json:"folderRecursion"` // whether or not we want to enter/copy folders
	SyncOnStartup   bool   `json:"syncOnStartup"`
	ContinuousSync  bool   `json:"continuousSync"`
	SyncType        string `json:"syncType"` //'normal' = bidirectional, 'specific-exts' = only looks for files in specified format, creates 1:1 copy with different file ext

	////Additional properties for 'specific-exts' type
	//File format used for each folder, respectively
	SpecificExts_FolderA       string `json:"specificExts_FolderA"`
	SpecificExts_FolderB       string `json:"specificExts_FolderB"`
	SpecificExts_CaseSensitive bool   `json:"specificExts_CaseSensitive"`

	//// General (Basic) Filtering rules
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

func (w *WalkItem) StillAlive() (output bool) {
	folderLink, _, _ := w.properties()

	//CHECK IF FOLDER LINK ROOT STILL EXISTS
	_, folderAError := os.Stat(folderLink.FolderA)
	_, folderBError := os.Stat(folderLink.FolderB)

	folderAExists := !os.IsNotExist(folderAError)
	folderBExists := !os.IsNotExist(folderBError)

	if !folderAExists || !folderBExists {
		return false
	}

	return true

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
	source, destination := w.itemPaths()

	if !w.StillAlive() {
		return "ERROR::FOLDERLINK_ERROR", time.Time{}
	}

	sourceFileStat, sourceError := os.Stat(source)
	var sourceFileModTime time.Time

	destinationFileStat, destinationError := os.Stat(destination)
	var destinationFileModTime time.Time

	sourceExists := !os.IsNotExist(sourceError)
	destinationExists := !os.IsNotExist(destinationError)

	if sourceExists {
		sourceFileModTime = sourceFileStat.ModTime()

	}
	if destinationExists {
		destinationFileModTime = destinationFileStat.ModTime()

	}

	if sourceExists && sourceError != nil {
		return "ERROR::SOURCE_ERRORexists:SOURCE_FILEexists", sourceFileModTime
	}

	if destinationExists && destinationError != nil {
		return "ERROR::DESTINATION_ERRORexists:DESTINATION_FILEexists", destinationFileModTime
	}

	if sourceExists && !destinationExists {
		return "DESTINATION!=EXIST", sourceFileModTime
	}

	if !sourceExists && destinationExists {
		return "SOURCE!=EXIST", destinationFileModTime

	}

	if !sourceExists && !destinationExists {
		return "ERROR::SOURCE_FILEdoesnt_exist:DESTINATION_FILEdoesnt_exist", time.Time{}
	}

	if sourceExists && destinationExists {

		// No errors means we are in! We grab the mod time for the destination file and begin our work of finding the most up-to date version

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

	}

	return "ERROR::END_OF_CASES", time.Time{}

}

// Outputs the current status between the two files refered to in WalkItem
func (w *WalkItem) quickSync() {
	source, destination := w.itemPaths()
	status, time := w.syncStatus()

	switch status {
	case "SOURCE=RECENT", "DESTINATION!=EXIST":
		copySync(source, destination, time)
	case "DESTINATION=RECENT", "SOURCE!=EXIST":
		copySync(destination, source, time)
	default:
		// fmt.Println("quick sync default", status, time)

	}

}
