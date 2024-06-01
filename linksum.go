package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// linkSums is what holds all directory info (name+lastModified)

func initalizeLinkSum() {
	fmt.Println("---INITALIZE LINK SUM---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		_, folderAError := os.Open(currentFolderLink.FolderA)
		_, folderBError := os.Open(currentFolderLink.FolderB)

		if folderAError == nil && folderBError == nil {
			linkSums[currentFolderLink.UUID] = make(map[string]time.Time)
			checkedFiles := linkSumWalk(&currentFolderLink, false, currentFolderLink.FolderA, []string{})
			linkSumWalk(&currentFolderLink, true, currentFolderLink.FolderB, checkedFiles)

		} else {
			if folderAError != nil {
				fmt.Println("Error loading links A:", folderAError)
			} else if folderBError != nil {
				fmt.Println("Error loading links B:", folderBError)
			}

		}

	}
	fmt.Println(linkSums)
}
func linkSumWalk(folderLink *FolderLink, inverse bool, source string, checkedFiles []string) []string {

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

			//// When file is a directory, and not the root folder
			if sourceFileIsDir && (filepath.Base(sourceFilePath) != filepath.Base(source)) {
				if !(folderFilterCheck(folderLink, sourceFileName)) {
					return filepath.SkipDir
				}

				if folderLink.FolderRecursion {
					checkedFiles = linkSumWalk(folderLink, inverse, sourceFilePath, checkedFiles)
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
				var currentRelativePath string

				//// For a convert sync type, we need an extension-less 'currentRelativePath' and specific extensions for 'currentFilePath' and 'destinationFilePath'
				if folderLink.SyncType == "specific-exts" {
					//// Grabbing folder props for Specific Exts as variables
					var SpecificExts_SourceFolderExt string
					if inverse {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderB
					} else {
						SpecificExts_SourceFolderExt = folderLink.SpecificExts_FolderA
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
						currentRelativePath = relativePathFromFolderLinkRoot(folderLink, inverse, sourceFilePath)
						//If they don't match, ignore file
					} else {
						return nil
					}
					//For a normal sync, the file is 1:1, with the only difference being the root folder
				} else if folderLink.SyncType == "normal" {
					currentRelativePath = relativePathFromFolderLinkRoot(folderLink, inverse, sourceFilePath)
				} else {
					return nil
				}
				//// Ignore files that we have already checked (saves some compute time and helpful when walking on both A and B)
				if !(slices.Contains(checkedFiles, currentRelativePath)) {
					linkSums[folderLink.UUID][currentRelativePath] = sourceFileModTime
					checkedFiles = append(checkedFiles, currentRelativePath)
				}
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil //catch all end
	})

	return checkedFiles
}

func firstBootSync() {
	fmt.Println("---FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		fmt.Println("\n=======", currentFolderLink.Name, "|", currentFolderLink.UUID, "=======")

		_, folderAError := os.Open(currentFolderLink.FolderA)
		_, folderBError := os.Open(currentFolderLink.FolderB)

		if folderAError == nil && folderBError == nil {
			checkedFiles := []string{}

			fmt.Println("WALKING FOLDER ", currentFolderLink.FolderA, "-> ("+currentFolderLink.FolderB+")")
			checkedFiles = walkCheck(&currentFolderLink, false, currentFolderLink.FolderA, currentFolderLink.FolderB, checkedFiles)

			fmt.Println("WALKING FOLDER ", currentFolderLink.FolderB, "-> ("+currentFolderLink.FolderA+")")
			walkCheck(&currentFolderLink, true, currentFolderLink.FolderB, currentFolderLink.FolderA, checkedFiles)

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
