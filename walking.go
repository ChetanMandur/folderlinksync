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

func walkCheck(folderLink *FolderLink, inverse bool, source string, destination string, checkedFiles []string) []string {

	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err == nil { //make sure there is no error from the walk

			//// Grabbing file props as variables
			sourceFileName := f.Name()
			sourceFileIsDir := f.IsDir()
			sourceFileModTime := f.ModTime()
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
					checkedFiles = walkCheck(folderLink, inverse, sourceFilePath, destinationFilePath, checkedFiles)
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
					var SpecificExts_DestinationFolderExt string
					if inverse {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderB
						SpecificExts_DestinationFolderExt = folderLink.SpecificExts_FolderA
					} else {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderA
						SpecificExts_DestinationFolderExt = folderLink.SpecificExts_FolderB
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
						destinationFilePath = pathConnector([]string{destination, sourceFileNameShort})

						fileRelativePath = relativePathFromFolderLinkRoot(folderLink, inverse, sourceFilePath)

						//Updating sourceFile and destinationFilePath with specific file extension
						sourceFilePath = sourceFilePath + "." + SpecificExts_SourceFolderExt
						destinationFilePath = destinationFilePath + "." + SpecificExts_DestinationFolderExt

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
					// fmt.Println(currentRelativePath)

					//// Start opening the file and check it out
					destinationFile, destinationError := os.Open(destinationFilePath)
					if destinationError == nil {
						destinationFile, destinationError := destinationFile.Stat() // Grabs more readable data
						destinationFileModTime := destinationFile.ModTime()         // Saving that mod time to compare with source
						if destinationError == nil {

							if sourceFileModTime.After(destinationFileModTime) {

								fmt.Println("\n" + fileRelativePath)
								fmt.Println("File is more recent in source (", sourceFileModTime, ")")
								copySync(folderLink, inverse, fileRelativePath, sourceFileModTime)

							} else if sourceFileModTime.Before(destinationFileModTime) {

								fmt.Println("\n" + fileRelativePath) //relativePath
								fmt.Println("File is more recent at destination (", destinationFileModTime, ")")
								copySync(folderLink, inverse, fileRelativePath, destinationFileModTime)

							}

						} else {
							fmt.Println("Error loading file at destionation ", destinationFilePath)
						}
					} else {
						fmt.Println("\n" + fileRelativePath) //relativePath
						fmt.Println("File does not exist at destination. Creating a copy at destination!")
						copySync(folderLink, inverse, fileRelativePath, sourceFileModTime)
					}

					checkedFiles = append(checkedFiles, fileRelativePath)
				} else {
					// fmt.Println("Checked!", currentRelativePath)
					return nil //When a file has been previously checked

				}
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil //catch all end
	})

	return checkedFiles
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

func copySync(folderLink *FolderLink, inverse bool, relativePath string, newTime time.Time) {
	yes := false
	var source string
	var destination string

	if !inverse {
		source = folderLink.FolderA + relativePath
		destination = folderLink.FolderB + relativePath
	} else {
		source = folderLink.FolderB + relativePath
		destination = folderLink.FolderA + relativePath
	}
	source = filepath.Clean(source)
	destination = filepath.Clean(destination)
	if yes {
		copy.Copy(source, destination)
		err := os.Chtimes(destination, newTime, newTime)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Copy complete!")

	} else {
		fmt.Println(source, destination)
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
