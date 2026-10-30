package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// An array that holds every 'link' from the `.json`
type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

// Representation of the data in a single `link` found within the `.json`
type FolderLink struct {
	//// Info not given by JSON

	// An array that holds the files parsed through it's previous cycle in `pollingRoutine`
	folderHistory *[]string

	//// General info (Bare minimum)

	// The name of the `link`
	Name string `json:"name"`
	// The first folder in the link
	FolderA string `json:"folderA"`
	// The second folder in the link
	FolderB string `json:"folderB"`
	// The UUID related to the `link`
	UUID int `json:"uuid"`
	// The direction at which the sync occurs
	// 	- `bi` = 'bi-direction', meaning changes are checked on both folders `A` and `B`
	// 	- `ab` = 'A-to-B', meaning changes are checked on folder `A` and reflected into folder `B`
	//  - `ba` = 'B-to-A', meaning changes are checked on folder `B` and reflected into folder
	SyncDirection string `json:"syncDirection"`

	//// General syncing settings

	// Whether or not to check folders within the link paths
	FolderRecursion bool `json:"folderRecursion"`
	// Whether or not the link should get synced on startup
	SyncOnStartup bool `json:"syncOnStartup"`
	// Whether or not to have link syned continuously after startup
	ContinuousSync bool `json:"continuousSync"`
	// The type of type of sync to occur
	// 	- 'normal' = All files are synced
	// 	- 'specific-exts' = Only files of a specific extension are synced (Both folders need an file extension specified, this will make a 1:1 copy no matter the formats; ie. no file conversion is done)
	SyncType string `json:"syncType"`

	//// Sync comparison settings

	// The amount of seconds the `modification time` is rounded by
	//
	// This can reduce constant updates, but will increase the chance of missed data
	FileTimeRound int `json:"fileTimeRound"`

	// The amount of seconds waited AFTER file copy before updating properties
	//
	// @example - File `1` is copied from `Folder A` to `Folder B`, before updating the `modification` time of the copied file, wait `FileModDelay` amount of time
	FileModDelay int `json:"fileModDelay"`

	////Additional properties for 'specific-exts' type
	//File format used for each folder, respectively

	// (Used when `SyncType` is `specific-exts`)
	//
	// Defines the extensions to target in `Folder A`
	SpecificExts_FolderA string `json:"specificExts_FolderA"`
	// (Used when `SyncType` is `specific-exts`)
	//
	// Defines the extensions to target in `Folder B`
	SpecificExts_FolderB string `json:"specificExts_FolderB"`

	// (Used when `SyncType` is `specific-exts`)
	//
	// Defines whether or not the extension types are case-sensitive or not
	SpecificExts_CaseSensitive bool `json:"specificExts_CaseSensitive"`

	//// General (Basic) Filtering rules
	//In/Ex-clude Files

	// Defines the list of file names (ie. files without the extension) to sync
	FileInclude []string `json:"fileInclude"`
	// Defines the list of file names (ie. files without the extension) to ignore/not sync
	FileExclude []string `json:"fileExclude"`

	//In/Ex-clude Folders
	// Defines the list of folder names to sync
	FolderInclude []string `json:"folderInclude"`
	// Defines the list of folder names to ignore/not sync
	FolderExclude []string `json:"folderExclude"`

	//In/Ex-clude Extensions
	// Defines the list of extensions to sync
	ExtInclude []string `json:"extInclude"`
	// Defines the list of extensions to ignore/not sync
	ExtExclude []string `json:"extExclude"`
}

// Returns an interger value stating the status of the `link`
//   - 0 = `Link` is alive and healthy; ie. both folders exist and can be synced
//   - 1 = `Folder A` does not exist
//   - 2 = `Folder B` does not exist
//   - 3 = Both `Folder A` and `Folder B` do not exist
func (f *FolderLink) stillAlive() (output int) {
	//// CHECK IF FOLDER LINK ROOTS STILL EXISTS

	// Grabs the error from retrieving `FileInfo` from both folders `A` and `B`
	_, folderAError := os.Stat(f.FolderA)
	_, folderBError := os.Stat(f.FolderB)

	// If error is not a `ErrNotExist`
	folderAExists := !os.IsNotExist(folderAError)
	folderBExists := !os.IsNotExist(folderBError)

	if !folderAExists { // If `Folder A` is a `ErrNotExist`
		output = 1
	} else if !folderBExists { // If `Folder B` is a `ErrNotExist`
		output = 2
	} else if !folderAExists && !folderBExists { // If both `Folder A` and `Folder B` is a `ErrNotExist`
		output = 3
	}

	// If neither folders have an error, or if neither of the errors are a `ErrNotExist` error
	return output
}

// Returns the `Modification Time` for the given file
//
// If the `link` specifies that the mod time is rounded, applies the round to the file
func (f *FolderLink) roundedModTime(input os.FileInfo) (output time.Time) {
	// Grabs the `modification time` from the file
	inputModTime := input.ModTime()

	if f.FileTimeRound == 0 { // If the `link` specifies no rounding to be done, return time as is
		return inputModTime
	} else { // If the `link` specifies a rounding time
		// Round the time before returning it
		return inputModTime.Round(time.Second * time.Duration(f.FileTimeRound))
	}
}

// Void return function that `time.Sleep` for the `link's` `FileModDelay` time
//
// This function does nothing other than sleep for the designated time
func (f *FolderLink) fileModDelay() {
	if f.FileModDelay != 0 { // If the `link` specifies a `FileModDelay`
		// Sleep for the amount of time
		time.Sleep(time.Second * time.Duration(f.FileModDelay))
	}
}

/*
The goal of this function is to walk through a directory, and return a list of files found in the directory.

The function takes in account acceptions/exclusions listed in the related `FolderLink`.

These found files are the files to be synced/targetted within the `FolderLink`

The function takes three parameters:

  - `inverse` = Specifies whether it is `Folder A` (false) or `Folder B` (true)

  - `checkedFiles` = A list of files (represented by their `relative path`) that were checked

  - `walkItems` = A list of files (represented as a `WalkItem`) that were checked

    These `WalkItem` contain:

    1) The `relative path` to the file

    2) The accompanying/related `link` to the file

    (What the hell was I thinking having `checkedFiles` when this exists)
    (Ok maybe I had it as a way to quickly check if a file has been already checked? idk man)

The function only returns an `nil` or `error`, new values are directly applied to the parameter variables
*/
func (f *FolderLink) walkCheck(inverse bool, checkedFiles *[]string, walkItems *[]WalkItem) (output error) {
	// Signifies the directory to walk through
	var source string

	// The source directory is based upon what the `inverse` value is
	if !inverse { // No inverse means source = `Folder A`
		source = f.FolderA
	} else { // Inverse means source = `Folder B`
		source = f.FolderB
	}

	// Initiates a walk on the source directory
	filepath.Walk(source, func(filePath string, file os.FileInfo, err error) error {
		// If there is an error, abort the walk
		if err != nil {
			output = err
			return filepath.SkipAll
		}

		// Grabs the name of the file (eg. `file.txt`)
		fileName := file.Name()

		// Checks if the file is a directory/folder
		fileIsDir := file.IsDir()

		// If the file is a directory
		if fileIsDir {

			// Check if the file is root
			// NOTE: THIS IS BROKEN, DOES NOT ACCOUNT FOR WHEN A CHILD FOLDER HAS THE SAME NAME AS ROOT
			// The fix is checking the full file path, rather than the name of the folder
			fileIsRoot := (filepath.Base(filePath) == filepath.Base(source))

			// fmt.Println("\nFile Path:\nBase: ", filepath.Base(filePath), "\nFull: ", filePath)
			// fmt.Println("\nSource/Root:\nBase: ", filepath.Base(source), "\nFull: ", source)
			// fmt.Println("\nIsRoot?", fileIsRoot)
			// fmt.Println("====")

			// If the folder is not root
			if !fileIsRoot {
				// Check if it fails the `folderFilterCheck` OR if `FolderRecursion` is not enabled
				// If either-or is true, skip the directory (ie. don't enter and continue the walk within it)
				if (!folderFilterCheck(f, fileName)) || (!f.FolderRecursion) {
					return filepath.SkipDir
				}
			}

		} else { // If the file is not a directory/folder (ie. its a normal file)
			// Grabs the file name and the file extension as seperate variables
			fileNameShort, fileNameExt := FileNameSplitter(fileName)

			// Checks if it fails the `fileFilterCheck`
			if !fileFilterCheck(f, inverse, fileNameShort, fileNameExt) {
				return nil // If it fails the check, skip this file
			}

			// Checks if `SyncType` is `specific-exts`
			// Since there are only two options for `SyncType`, we just check if it's not `normal`
			extensionlessRelativePath := false
			if f.SyncType != "normal" {
				extensionlessRelativePath = true
			}

			// Find the relative path (from source/root) of the file
			relativePath := relativePathFromRoot(source, filePath, extensionlessRelativePath)

			// If this same relative path does not exist within `checkedFiles`, add it to both `checkedFiles` and `walkItems`
			// This same path would only exist if `walkCheck` is on a second run on the sibling folder
			// 	- eg. `walkCheck` ran on `Folder A`, relative path was added to `checkedFiles`, `walkCheck` is ran again but on `Folder B`, file exists in both folders, so relative path is the same
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

	// If there was an error, return its value
	return output
}

// Function to quickly initiate a `walkCheck` based on the `FolderLinks` `SyncDirection` parameter
// Returns:
//   - `checkedFiles` = a `string` array of `relative paths`
//   - `walkItem` = an object with `relative path` and accompanying `FolderLink`
func (f *FolderLink) quickWalk() (checkedFiles []string, walkItems []WalkItem) {
	// Checks if the `link` is still alive
	if f.stillAlive() != 0 {
		return nil, nil // If the link is dead, return all `nil`
	}

	// Holds any error obtained from a `walkCheck`
	var walkError error

	// Based on the `SyncDirection`, the walk path is different
	switch f.SyncDirection {
	case "bi":
		{
			// For bi-directional, we need to walk both folders

			// Walk `Folder A`, return all `nil` if an error occurs
			walkError = f.walkCheck(false, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}

			// Walk `Folder B`, return all `nil` if an error occurs
			walkError = f.walkCheck(true, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}

		}
	case "ab":
		{
			// For 'ab', we need to walk only `Folder A`

			// Walk `Folder A`, return all `nil` if an error occurs
			walkError = f.walkCheck(false, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}
		}
	case "ba":
		{
			// For 'ba', we need to walk only `Folder B`

			// Walk `Folder B`, return all `nil` if an error occurs
			walkError = f.walkCheck(true, &checkedFiles, &walkItems)
			if walkError != nil {
				return nil, nil
			}
		}
	}

	// Return the final `checkedFiles`,
	return checkedFiles, walkItems
}

// Function to quickly initate a sync of a `link`
//
// This is 'quick' since it requires no parameters, and only uses the properties of the `FolderLink` object
//
// Returns a value if the sync was successful or not
func (f *FolderLink) quickSync() (success bool) {
	// Checks if the `link` is still alive
	if f.stillAlive() != 0 {
		return false
	}

	// Generate the list of files to sync
	checkedFiles, walkItems := f.quickWalk()

	// If both `checkedFiles` and `walkItems` is `nil`, return false
	if checkedFiles == nil && walkItems == nil {
		return false
	}

	// For each of the files, sync them between `Folder A` and `Folder B`
	for _, currentItem := range walkItems {
		currentItem.quickSync()
	}
	return true
}

// This function houses the routine that runs indefinitely, continously watching and syncing the `FolderLink`
//
// This is run when the `FolderLink` has `continuousSync` enabled
func (f *FolderLink) pollingRoutine() {
	// When it's all said and done, decrease `WaitGroup` value by 1
	defer wg.Done()

	fmt.Println("\n=======", f.Name, "|", f.UUID, "=======")
	fmt.Println("Folder A: ", f.FolderA, "\nFolder B:", f.FolderB)

	// Signifies if an error was hit
	var errorHit bool

	// Indefinite loop
	for {
		// Checks if `link` is still alive
		stillAlive := f.stillAlive()

		if stillAlive == 0 { // If the `link` is alive and well

			// If there was a previous error, there clearly isn't one now
			// ie. the `link` has reconnected
			if errorHit {
				fmt.Println("Reconnected to", f.Name, f.UUID)
			}
			// Reset `error` state back to `false`
			errorHit = false

			// Generate list of files in `link`
			checkedFiles, walkItems := f.quickWalk()

			// If both the lists are `nil`, go to next loop iteration
			if checkedFiles == nil && walkItems == nil {
				continue
			}

			// Variable to define if a polling error has occured
			pollingError := false

			// For each of the files in the `link`
			for _, currentWalkItem := range walkItems {
				// Run a safe sync, designed for use within `pollingRoutine`
				pollingError = currentWalkItem.pollingCheck()
				// If any errors occur, break free/stop further checks
				if pollingError {
					break
				}
			}

			// If no errors occur
			if !pollingError {
				// Update the folder history with the current list of files within the `link`
				f.folderHistory = &checkedFiles

			}

		} else { // If the `link` is NOT alive and well

			// We only want this print statement to print once
			// ie. When the error is first hit, print the statement, and keep silently looping for the error to resolve
			if stillAlive == 1 && !errorHit {
				fmt.Println("Folder A does not exist")
			} else if stillAlive == 2 && !errorHit {
				fmt.Println("Folder B does not exist")
			} else if stillAlive == 3 && !errorHit {
				fmt.Println("Both folders do not exist")
			}

			// Set error as hit
			errorHit = true

			// Clear folder history
			f.folderHistory = nil

		}
	}

}
