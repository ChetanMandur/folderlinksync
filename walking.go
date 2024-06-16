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
	yes := true

	if yes {
		copy.Copy(source, destination)
		err := os.Chtimes(destination, newTime, newTime)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Copy complete!", source, "--->", destination)

	} else {
		fmt.Println(source, "--->", destination)
	}

}

// Given a path (assumed to be within one of the folder link paths), find its relative path from root
// Relative path is the path of the file without directory info beyond root,
// For example, given the folder root (either FolderA or FolderB) and the path of a file within its directory, return relative path
// FolderRoot = C:/FolderRoot/, FilePath = C:/FolderRoot/depth/test.txt, therefore RelativePath= /depth/test.txt
// root is the path of the root folder, file is the path of the file in question (with file assumed to be within root)
func relativePathFromRoot(root string, file string, extensionless bool) string {
	//Preping the strings
	output := ""
	file = filepath.ToSlash(file) //cleans the string

	//Splitting  root
	rootSplit := strings.Split(root, "/")

	//use backslash if forward slash didnt work
	if len(rootSplit) == 1 {
		rootSplit = strings.Split(root, "\\")
	}

	//Splitting file path
	fileSplit := strings.Split(file, "/")
	temp_slash := "/" //We want to continue using the same type of slashes for the relative path output

	//use backslash if forward slash didnt work (and update temp_slash)
	if len(fileSplit) == 1 {
		fileSplit = strings.Split(file, "\\")
		temp_slash = "\\"
	}

	//Chop off root from the source
	fileReduced := fileSplit[len(rootSplit):]

	//Chopping off the extension if we extensionless
	if extensionless {
		// //               split by '.'  (the last element in path (ie. file))
		// fileNameSplit := strings.Split(fileReduced[len(fileReduced)-1], ".")
		// fileReduced[len(fileReduced)-1] = fileNameSplit[0] // 0 = file name, 1= ext
		name, _ := FileNameSplitter(fileReduced[len(fileReduced)-1])
		fileReduced[len(fileReduced)-1] = name
	}

	//make it all into a string and output
	output = temp_slash + pathConnector(fileReduced)

	return output

}

// Returns false if file does not pass file filter tests
func fileFilterCheck(folderLink *FolderLink, inverse bool, sourceFileNameShort string, sourceFileNameExt string) bool {
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

	if folderLink.SyncType == "specific-exts" {
		//// Grabbing folder props for Specific Exts as variables
		var sourceFolderExt string

		if inverse {
			sourceFolderExt = folderLink.SpecificExts_FolderB

		} else {
			sourceFolderExt = folderLink.SpecificExts_FolderA

		}

		//// Check if extensions are case sensitive or not
		var extensionMatch = false
		if folderLink.SpecificExts_CaseSensitive {
			extensionMatch = (sourceFolderExt == sourceFileNameExt)

		} else {
			extensionMatch = (strings.EqualFold(sourceFolderExt, sourceFileNameExt))
		}

		if !extensionMatch {
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

func FileNameSplitter(originalName string) (finalName string, finalExt string) {
	originalNameSplit := strings.Split(originalName, ".")

	finalExt = originalNameSplit[len(originalNameSplit)-1]

	finalNameSplit := originalNameSplit[0 : len(originalNameSplit)-1]

	for index, value := range finalNameSplit {
		if index != 0 {
			finalName = finalName + "."
		}
		finalName = finalName + value
	}

	return finalName, finalExt

}

/*
The goal of this function is to walk through a directory, and return a list of files found in the directory.
The function takes in account acceptions/exclusions listed in the related FolderLink.
The list is of type WalkItem, which consists of a file's relative path, related FolderLink, whether it is 'inverse' or not (aka whether it is folder B or A in the Link)
*/
func WalkCheck(folderLink *FolderLink, inverse bool, checkedFiles *[]string, walkItems *[]WalkItem) (output error) {
	var source string
	// var destination string

	if !inverse {
		source = folderLink.FolderA
	} else {
		source = folderLink.FolderB
	}

	filepath.Walk(source, func(filePath string, file os.FileInfo, err error) error {

		if err == nil {
			fileName := file.Name()
			fileIsDir := file.IsDir()

			if fileIsDir {
				fileIsRoot := (filepath.Base(filePath) == filepath.Base(source))

				if !fileIsRoot {
					if (!folderFilterCheck(folderLink, fileName)) || (!folderLink.FolderRecursion) {
						return filepath.SkipDir
					}
				}

			} else {
				fileNameShort, fileNameExt := FileNameSplitter(fileName)

				if !fileFilterCheck(folderLink, inverse, fileNameShort, fileNameExt) {
					return nil
				}

				extensionlessRelativePath := false
				if folderLink.SyncType != "normal" {
					extensionlessRelativePath = true
				}

				relativePath := relativePathFromRoot(source, filePath, extensionlessRelativePath)

				if !slices.Contains(*checkedFiles, relativePath) {
					*walkItems = append(*walkItems, WalkItem{FolderLink: folderLink, Inverse: inverse, RelativePath: relativePath})
					*checkedFiles = append(*checkedFiles, relativePath)
					// fmt.Println("Added", relativePath)

				} /* else {
					fmt.Println("Already checked!", relativePath)
				} */

			}

			return nil
		} else {
			output = err
			return filepath.SkipAll
		}

	})

	return output
}
