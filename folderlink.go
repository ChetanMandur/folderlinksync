package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	//// Info not given by JSON
	folderHistory *[]string

	//// General info (Bare minimum)
	Name          string `json:"name"`
	FolderA       string `json:"folderA"`
	FolderB       string `json:"folderB"`
	UUID          int    `json:"uuid"`
	SyncDirection string `json:"syncDirection"`

	//// General syncing settings
	FolderRecursion bool   `json:"folderRecursion"` // whether or not we want to enter/copy folders
	SyncOnStartup   bool   `json:"syncOnStartup"`
	ContinuousSync  bool   `json:"continuousSync"`
	SyncType        string `json:"syncType"` //'normal' = bidirectional, 'specific-exts' = only looks for files in specified format, creates 1:1 copy with different file ext

	//// Sync comparison settings
	FileTimeRound int `json:"fileTimeRound"` // (time.Seconds) File modification times are rounded by these many seconds (can reduce constant updates, will increase chance of missed data)
	FileModDelay  int `json:"fileModDelay"`  // (time.Seconds) The amount of time the program waits before modying destination copy properties

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

func (f *FolderLink) stillAlive() (output int) {
	//CHECK IF FOLDER LINK ROOTS STILL EXISTS
	_, folderAError := os.Stat(f.FolderA)
	_, folderBError := os.Stat(f.FolderB)

	folderAExists := !os.IsNotExist(folderAError)
	folderBExists := !os.IsNotExist(folderBError)

	if !folderAExists {
		output = 1
	} else if !folderBExists {
		output = 2
	} else if !folderAExists && !folderBExists {
		output = 3
	}

	return output

}

func (f *FolderLink) roundedModTime(input os.FileInfo) (output time.Time) {
	inputModTime := input.ModTime()

	if f.FileTimeRound == 0 {
		return inputModTime
	} else {
		return inputModTime.Round(time.Second * time.Duration(f.FileTimeRound))
	}
}

func (f *FolderLink) fileModDelay() {
	if f.FileModDelay != 0 {
		time.Sleep(time.Second * time.Duration(f.FileModDelay))
	}

}

/*
The goal of this function is to walk through a directory, and return a list of files found in the directory.
The function takes in account acceptions/exclusions listed in the related FolderLink.
The list is of type WalkItem, which consists of a file's relative path, related FolderLink, whether it is 'inverse' or not (aka whether it is folder B or A in the Link)
*/
func (f *FolderLink) walkCheck(inverse bool, checkedFiles *[]string, walkItems *[]WalkItem) (output error) {
	var source string
	// var destination string

	if !inverse {
		source = f.FolderA
	} else {
		source = f.FolderB
	}

	filepath.Walk(source, func(filePath string, file os.FileInfo, err error) error {

		if err != nil {
			output = err
			return filepath.SkipAll
		}

		fileName := file.Name()
		fileIsDir := file.IsDir()

		if fileIsDir {
			fileIsRoot := (filepath.Base(filePath) == filepath.Base(source))

			if !fileIsRoot {
				if (!folderFilterCheck(f, fileName)) || (!f.FolderRecursion) {
					return filepath.SkipDir
				}
			}

		} else {
			fileNameShort, fileNameExt := FileNameSplitter(fileName)

			if !fileFilterCheck(f, inverse, fileNameShort, fileNameExt) {
				return nil
			}

			extensionlessRelativePath := false
			if f.SyncType != "normal" {
				extensionlessRelativePath = true
			}

			relativePath := relativePathFromRoot(source, filePath, extensionlessRelativePath)

			if !slices.Contains(*checkedFiles, relativePath) {
				*walkItems = append(*walkItems, WalkItem{FolderLink: f, RelativePath: relativePath})
				*checkedFiles = append(*checkedFiles, relativePath)
				// fmt.Println("Added", relativePath)

			} /* else {
				fmt.Println("Already checked!", relativePath)
			} */

		}

		return nil

	})

	return output
}

func (f *FolderLink) quickWalk() (checkedFiles []string, walkItems []WalkItem) {
	if f.stillAlive() != 0 {
		return nil, nil
	}

	var walkError error
	switch f.SyncDirection {
	case "bi":
		{
			walkError = f.walkCheck(false, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}
			walkError = f.walkCheck(true, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}

		}
	case "ab":
		{
			walkError = f.walkCheck(false, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}
		}
	case "ba":
		{
			walkError = f.walkCheck(true, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}
		}
	}

	return checkedFiles, walkItems
}

func (f *FolderLink) quickSync() (success bool) {
	if f.stillAlive() != 0 {
		return false
	}

	checkedFiles, walkItems := f.quickWalk()

	if checkedFiles == nil && walkItems == nil {
		return false
	}

	for _, currentItem := range walkItems {
		currentItem.quickSync()
	}
	return true
}

// This is the routine that runs when we want continuous watching
func (f *FolderLink) pollingRoutine() {
	defer wg.Done()

	fmt.Println("\n=======", f.Name, "|", f.UUID, "=======")
	fmt.Println("Folder A: ", f.FolderA, "\nFolder B:", f.FolderB)

	var errorHit bool

	for {
		stillAlive := f.stillAlive()

		if stillAlive == 0 {
			if errorHit {
				fmt.Println("Reconnected to", f.Name, f.UUID)
			}
			errorHit = false

			checkedFiles, walkItems := f.quickWalk()

			if checkedFiles == nil && walkItems == nil {
				continue
			}

			pollingError := false
			for _, currentWalkItem := range walkItems {
				pollingError = currentWalkItem.pollingCheck()
				if pollingError {
					break
				}
			}
			if !pollingError {
				f.folderHistory = &checkedFiles

			}

		} else {
			if stillAlive == 1 && !errorHit {
				fmt.Println("Folder A does not exist")
			} else if stillAlive == 2 && !errorHit {
				fmt.Println("Folder B does not exist")
			} else if stillAlive == 3 && !errorHit {
				fmt.Println("Both folders do not exist")
			}
			errorHit = true
			f.folderHistory = nil

		}
	}

}
