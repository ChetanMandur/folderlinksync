package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	copy "github.com/otiai10/copy"
)

// Returns false if file does not pass file filter tests
func fileFilterCheck(folderLink *FolderLink, sourceFileNameShort string, sourceFileNameExt string) bool {
	if folderLink.FileExclude != nil {
		if slices.Contains(folderLink.FileExclude, sourceFileNameShort) {
			return false
		}
	}
	if folderLink.FileInclude != nil {
		if !(slices.Contains(folderLink.FileInclude, sourceFileNameShort)) && len(folderLink.FileInclude) != 0 {
			return false
		}
	}
	if folderLink.ExtExclude != nil {
		if slices.Contains(folderLink.ExtExclude, sourceFileNameExt) {
			return false
		}

	}
	if folderLink.ExtInclude != nil {
		if !(slices.Contains(folderLink.ExtInclude, sourceFileNameExt)) && len(folderLink.ExtInclude) != 0 {

			return false
		}

	}
	return true
}

func folderFilterCheck(folderLink *FolderLink, sourceFileName string) bool {
	if folderLink.FolderExclude != nil {
		if slices.Contains(folderLink.FolderExclude, sourceFileName) {
			return false
		}
	}

	if folderLink.FolderInclude != nil {
		if !(slices.Contains(folderLink.FolderInclude, sourceFileName)) && len(folderLink.FolderInclude) != 0 {
			return false
		}
	}

	return true
}

func walkCheck(folderLink *FolderLink, inverse bool, source string, destination string, checkedFiles []string, walkItems []WalkItem) ([]WalkItem, []string) {

	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err == nil { //make sure there is no error from the walk

			//// Grabbing file props as variables
			sourceFileName := f.Name()
			sourceFileIsDir := f.IsDir()
			// sourceFileModTime := f.ModTime()
			sourceFilePath := path //just to make naming the same

			sourceFileNameSplit := strings.Split(sourceFileName, ".")
			var sourceFileNameShort string
			var sourceFileNameExt string
			if len(sourceFileNameSplit) == 2 && !sourceFileIsDir {
				sourceFileNameShort = sourceFileNameSplit[0]
				sourceFileNameExt = sourceFileNameSplit[1]
			}

			//// Building the equivalent path at destination
			destinationFilePath := pathConnector([]string{destination, sourceFileName})

			//// When file is a directory, and not the root folder
			if sourceFileIsDir && (filepath.Base(sourceFilePath) != filepath.Base(source)) {
				if !(folderFilterCheck(folderLink, sourceFileName)) {
					return filepath.SkipDir
				}

				if folderLink.FolderRecursion {

					walkItems, checkedFiles = walkCheck(folderLink, inverse, sourceFilePath, destinationFilePath, checkedFiles, walkItems)
				}
				return filepath.SkipDir

				//// When file is root (always the inital file read)
			} else if sourceFileIsDir && (filepath.Base(sourceFilePath) == filepath.Base(source)) {

				return nil

				//// When file is not a directory
			} else {
				if !(fileFilterCheck(folderLink, sourceFileNameShort, sourceFileNameExt)) {
					return nil
				}

				//// Relative path is the path of the file without directory info beyond root,
				//// For example, the files C://folderA/folder1/test.txt and C://folderB/folder1/test.txt have the relative path of /folder1/test.txt
				//// This is useful to make note of each file, including their position in deep root
				//// In our case, we take the file @ source and file @ destionation (always assume there is one, as thats the goal) and generate the relative path
				var fileRelativePath string

				//// For a convert sync type, we need an extension-less 'currentRelativePath' and specific extensions for 'currentFilePath' and 'destinationFilePath'
				if folderLink.SyncType == "specific-exts" {
					//// Grabbing folder props for Specific Exts as variables
					var SpecificExts_SourceFolderExt string
					// var SpecificExts_DestinationFolderExt string
					if inverse {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderB
						// SpecificExts_DestinationFolderExt = folderLink.SpecificExts_FolderA
					} else {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderA
						// SpecificExts_DestinationFolderExt = folderLink.SpecificExts_FolderB
					}

					//// Check if extensions are case sensitive or not
					var extensionMatch = false
					if folderLink.SpecificExts_CaseSensitive {
						extensionMatch = (SpecificExts_SourceFolderExt == sourceFileNameExt)

					} else {
						extensionMatch = (strings.EqualFold(SpecificExts_SourceFolderExt, sourceFileNameExt))
					}

					//// If file matches extension conditions, begin generating new 'currentRelativePath', 'sourceFilePath', and 'destinationFilePath'
					if extensionMatch {
						//Creating extensionless sourceFile and destinationFile paths (leads to extensionless currentRelativePath)
						sourceFilePath = pathConnector([]string{source, sourceFileNameShort})
						// destinationFilePath = pathConnector([]string{destination, sourceFileNameShort})

						fileRelativePath = relativePathFromFolderLinkRoot(folderLink, inverse, sourceFilePath)

						//Updating sourceFile and destinationFilePath with specific file extension
						// sourceFilePath = sourceFilePath + "." + SpecificExts_SourceFolderExt
						// destinationFilePath = destinationFilePath + "." + SpecificExts_DestinationFolderExt

						//If they don't match, ignore file
					} else {
						return nil
					}

					//For a normal sync, the file is 1:1, with the only difference being the root folder
				} else if folderLink.SyncType == "normal" {

					fileRelativePath = relativePathFromFolderLinkRoot(folderLink, inverse, sourceFilePath)

				} else {

					return nil

				}

				//// Ignore files that we have already checked (saves some compute time and helpful when walking on both A and B)
				if !(slices.Contains(checkedFiles, fileRelativePath)) {

					// fmt.Println("Opened!", fileRelativePath)
					walkItems = append(walkItems, WalkItem{FolderLink: folderLink, Inverse: inverse, RelativePath: fileRelativePath})
					checkedFiles = append(checkedFiles, fileRelativePath)
				} else {
					// fmt.Println("Checked!", fileRelativePath)
					return nil //When a file has been previously checked

				}
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil //catch all end
	})

	return walkItems, checkedFiles
}

func pathConnector(locations []string) string {
	output := ""
	for i := 0; i < len(locations); i++ {
		if i == len(locations)-1 {
			output += locations[i]

		} else {
			output += locations[i] + "/"
		}

	}

	output = filepath.ToSlash(output)
	return output
}

// Copies file to destination, and updates the file's mod time to the given mod time
func copySync(source string, destination string, newTime time.Time) {
	//True= actually copies False = Just print the paths it was gonna use (to -> from)
	yes := false

	if yes {
		copy.Copy(source, destination)
		err := os.Chtimes(destination, newTime, newTime)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Copy complete!")

	} else {
		fmt.Println(source, "--->", destination)
	}

}

func walkItemCompareCopySync(walkItem *WalkItem) {

	//Splits the struct into its own variables
	folderLink := walkItem.FolderLink
	inverse := walkItem.Inverse
	relativePath := walkItem.RelativePath

	//Paths for file @ each location
	var source string
	var destination string

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
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)

	//Opening the files (YOLO for the errors on source, im hoping no errors cuz it was read w/o error during its walk find)
	sourceFile, _ := os.Open(source)
	sourceFileStat, _ := sourceFile.Stat()
	sourceFileModTime := sourceFileStat.ModTime()

	//Opening destination AND checking for error, cuz error means file does not exist (aka something important)
	destinationFile, destinationError := os.Open(destination)
	fmt.Println("\n" + relativePath)
	if destinationError == nil {
		destinationFileStat, destinationError := destinationFile.Stat()

		if destinationError == nil {
			// No errors means we are in! We grab the mod time for the destination file and begin our work of finding the most up-to date version
			destinationFileModTime := destinationFileStat.ModTime()

			if sourceFileModTime.After(destinationFileModTime) {

				fmt.Println("File is more recent in source (", source, sourceFileModTime, ")")
				copySync(source, destination, sourceFileModTime)

			} else if sourceFileModTime.Before(destinationFileModTime) {

				fmt.Println("File is more recent at destination (", destination, destinationFileModTime, ")")
				copySync(destination, source, destinationFileModTime)

			} else {
				fmt.Println("Files are in sync!")
			}

		} else {
			fmt.Println("Error loading file at destionation ", destination)
		}
	} else {
		fmt.Println("File does not exist at destination.(", destination, ") Copying and Creating!")
		copySync(source, destination, sourceFileModTime)
	}

}

// Given a path (assumed to be within one of the folder link paths), find its relative path from root
func relativePathFromFolderLinkRoot(folderLink *FolderLink, inverse bool, source string) string {
	//Preping the strings
	output := ""
	sourceClean := filepath.ToSlash(source)

	//Finding the root path
	var sourceRoot string
	if inverse {
		sourceRoot = folderLink.FolderB

	} else {
		sourceRoot = folderLink.FolderA

	}

	//Splitting source root
	sourceRootSplit := strings.Split(sourceRoot, "/")

	//use backslash if forward slash didnt work
	if len(sourceRootSplit) == 1 {
		sourceRootSplit = strings.Split(sourceRoot, "\\")
	}

	//Splitting source
	sourceCleanSplit := strings.Split(sourceClean, "/")
	temp_slash := "/" //We want to continue using the same type of slashes for the relative path output

	//use backslash if forward slash didnt work (and update temp_slash)
	if len(sourceCleanSplit) == 1 {
		sourceCleanSplit = strings.Split(sourceClean, "\\")
		temp_slash = "\\"
	}

	//Chop off source root from source clean
	sourceReduced := sourceCleanSplit[len(sourceRootSplit):]

	//make it all into a string and output
	output = temp_slash + pathConnector(sourceReduced)
	return output

}

func firstBootSync() {
	fmt.Println("---FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		fmt.Println("\n=======", currentFolderLink.Name, "|", currentFolderLink.UUID, "=======")

		_, folderAError := os.Open(currentFolderLink.FolderA)
		_, folderBError := os.Open(currentFolderLink.FolderB)

		if folderAError == nil && folderBError == nil {
			checkedFiles := []string{}
			walkItems := []WalkItem{}
			fmt.Println("Folder A: ", currentFolderLink.FolderA, "\nFolder B:", currentFolderLink.FolderB)

			// fmt.Println("\nWALKING FOLDER ", currentFolderLink.FolderA, "-> ("+currentFolderLink.FolderB+")")
			walkItems, checkedFiles = walkCheck(&currentFolderLink, false, currentFolderLink.FolderA, currentFolderLink.FolderB, checkedFiles, walkItems)

			// fmt.Println("\nWALKING FOLDER ", currentFolderLink.FolderB, "-> ("+currentFolderLink.FolderA+")")
			walkItems, _ = walkCheck(&currentFolderLink, true, currentFolderLink.FolderB, currentFolderLink.FolderA, checkedFiles, walkItems)

			// fmt.Println("HELLO WORLD\n", len(walkItems), len(checkedFiles))
			for _, currentWalkItem := range walkItems {
				walkItemCompareCopySync(&currentWalkItem)
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
