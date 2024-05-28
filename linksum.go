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

// linkSums is what holds all directory info (name+lastModified)

func initalizeLinkSum() {
	fmt.Println("---INITALIZE LINK SUM---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		absFolderA, _ := filepath.Abs(currentFolderLink.FolderA)
		absFolderB, _ := filepath.Abs(currentFolderLink.FolderB)

		if linkSums[absFolderA] == nil {
			linkSums[absFolderA] = make(map[string]time.Time)
			filepath.Walk(absFolderA, func(path string, f os.FileInfo, err error) error {
				if err == nil {
					// fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
					linkSums[absFolderA][f.Name()] = f.ModTime()
					//we dont want to enter folders on the directory we watching, so we stop ourselves from doing that
					if f.IsDir() && filepath.Base(path) != filepath.Base(absFolderA) {
						return filepath.SkipDir
					}

				} else { //error means that len(linkSums[folder]) = 0!!!!
					fmt.Println("Error loading", absFolderA, len(linkSums[absFolderA]))
				}
				return nil
			})
		}

		if linkSums[absFolderB] == nil {
			linkSums[absFolderB] = make(map[string]time.Time)
			filepath.Walk(absFolderB, func(path string, f os.FileInfo, err error) error {
				if err == nil {
					// fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
					linkSums[absFolderB][f.Name()] = f.ModTime()
					if f.IsDir() && filepath.Base(path) != filepath.Base(absFolderB) {
						return filepath.SkipDir
					}
				} else {
					fmt.Println("Error loading", absFolderB, len(linkSums[absFolderB]))
				}
				return nil
			})
		}

	}
	// fmt.Println(linkSums)
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

func normalCopySync(source string, destination string, newTime time.Time) {
	copy.Copy(source, destination)
	fmt.Println("Normal copy sync")

	err := os.Chtimes(destination, newTime, newTime)
	if err != nil {
		fmt.Println(err)
	}

}

func convertCopySync(source string, destination string, newFormat string, newTime time.Time) {
	_, sourceFile := filepath.Split(source)
	sourceFileSplit := strings.Split(sourceFile, ".")
	destinationDir, _ := filepath.Split(destination)

	destinationNewFile := destinationDir + sourceFileSplit[0] + "." + newFormat

	fmt.Println("BETTER COPY SYNC", destinationNewFile)

	copy.Copy(source, destinationNewFile)

	err := os.Chtimes(destinationNewFile, newTime, newTime)
	if err != nil {
		fmt.Println(err)
	}

}

func relativePath(source string, destination string) string {

	output := ""
	sourceClean := filepath.ToSlash(source)
	destinationClean := filepath.ToSlash(destination)
	// fmt.Println("relativePath", sourceClean, destinationClean)

	sourceSplit := strings.Split(sourceClean, "/")
	if len(sourceSplit) == 1 {
		sourceSplit = strings.Split(sourceClean, "\\")

	}

	destinationSplit := strings.Split(destinationClean, "/")
	if len(destinationSplit) == 1 {
		destinationSplit = strings.Split(destinationClean, "\\")
	}

	for i := len(sourceSplit) - 1; i >= 0; i-- {
		if sourceSplit[i] == destinationSplit[i] {
			output = "/" + sourceSplit[i] + output
		} else {
			break
		}
	}

	return output

}

func normalWalkCheck(source string, destination string, checkedFiles []string) []string {

	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err == nil { //make sure there is no error from the walk
			//just grabbing these to reference later
			currentFileName := f.Name()
			currentFileIsDir := f.IsDir()
			currentFileModTime := f.ModTime()
			currentFilePath := path //just to make naming the same

			//Building the equivalent path at destination
			pathArr := []string{destination, currentFileName} //temp array used for pathConnector()
			destinationFilePath := pathConnector(pathArr)

			//When file is a directory (and not the root folder), we want to enter it and walk it
			if currentFileIsDir && (filepath.Base(currentFilePath) != filepath.Base(source)) {
				checkedFiles = normalWalkCheck(currentFilePath, destinationFilePath, checkedFiles)
				return filepath.SkipDir
			} else if currentFileIsDir && (filepath.Base(currentFilePath) == filepath.Base(source)) { //skips copying root directory, (eg. /a --copies to--> /b/a)
				return nil
			} else {

				currentRelativePath := relativePath(currentFilePath, destinationFilePath)

				if !(slices.Contains(checkedFiles, currentRelativePath)) { //If the file has not been checked yet
					// fmt.Println("\n"+currentFilePath, " | ", currentFileModTime) //Source File
					fmt.Println("\n"+currentRelativePath, " | ", currentFileModTime) //relativePath + source modification time

					//we start opening it and checking it
					destinationFile, destError := os.Open(destinationFilePath)
					if destError == nil {
						destinationFile, destError := destinationFile.Stat()
						destinationFileModTime := destinationFile.ModTime() //saving that mod time to compare with source
						if destError == nil {
							// fmt.Println("("+destinationFilePath, " | ", destinationFileModTime, ")") //Destination File

							if currentFileModTime.After(destinationFileModTime) {
								fmt.Println("File is more recent in source")
								normalCopySync(currentFilePath, destinationFilePath, currentFileModTime)
							} else if currentFileModTime.Before(destinationFileModTime) {
								fmt.Println("File is more recent at destination (", destinationFileModTime, ")")
								normalCopySync(destinationFilePath, currentFilePath, destinationFileModTime)
							} else if currentFileModTime.Equal(destinationFileModTime) {
								fmt.Println("Files are in sync")
							} else {
								fmt.Println("Yo what", currentFileModTime, destinationFileModTime)
							}
							fmt.Print("\n")

						} else {
							fmt.Println("Error loading file at destionation ", destinationFilePath)
							// normalCopySync(currentFilePath, destinationFilePath, currentFileModTime)
						}
					} else {
						fmt.Println("File does not exist at destination ", destinationFilePath)
						normalCopySync(currentFilePath, destinationFilePath, currentFileModTime)
					}

					checkedFiles = append(checkedFiles, currentRelativePath)
				} else {
					fmt.Println("Checked file", currentRelativePath)
				}
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil
	})

	return checkedFiles
}

func firstBootSync() {
	fmt.Println("---BETTER FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		switch currentFolderLink.SyncType {
		case "normal":
			normalFBS(currentFolderLink)
			// fmt.Println("skipping normal")
		case "convert":
			fmt.Println("skipping convert")
			// convertFBS(currentFolderLink)
		}
	}
}

func normalFBS(folderLink FolderLink) {
	fmt.Println("NORMAL FBS")
	absFolderA, _ := filepath.Abs(folderLink.FolderA)
	absFolderB, _ := filepath.Abs(folderLink.FolderB)

	_, folderAError := os.Open(absFolderA)
	_, folderBError := os.Open(absFolderB)

	if folderAError == nil && folderBError == nil {
		checkedFiles := []string{}

		fmt.Println("WALKING FOLDER ", absFolderA)
		checkedFiles = normalWalkCheck(absFolderA, absFolderB, checkedFiles)

		fmt.Println("WALKING FOLDER ", absFolderB)
		normalWalkCheck(absFolderB, absFolderA, checkedFiles)

		fmt.Println(" ")
	} else {
		fmt.Println("Error loading link A:", folderAError, "|| B:", folderBError)
	}
}

func convertFBS(folderLink FolderLink) {
	fmt.Println("CONVERT FBS")
	absFolderA, _ := filepath.Abs(folderLink.FolderA)
	absFolderB, _ := filepath.Abs(folderLink.FolderB)

	_, folderAError := os.Open(absFolderA)
	_, folderBError := os.Open(absFolderB)

	if folderAError == nil && folderBError == nil {
		// checkedFiles := []string{}

		fmt.Println("WALKING FOLDER ", absFolderA)
		// checkedFiles = convertWalkCheck(absFolderA, absFolderB, folderLink.FolderAExt, folderLink.FolderBExt, checkedFiles)

		fmt.Println("WALKING FOLDER ", absFolderB)
		// convertWalkCheck(absFolderB, absFolderA, folderLink.FolderBExt, folderLink.FolderAExt, checkedFiles)

		fmt.Println(" ")
	} else {
		fmt.Println("Error loading link A:", folderAError, "|| B:", folderBError)
	}
}
