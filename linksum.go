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

func betterCopySync(source string, destination string, newTime time.Time) {
	copy.Copy(source, destination)
	err := os.Chtimes(destination, newTime, newTime)
	if err != nil {
		fmt.Println(err)
	}

}

func relativePath(source string, destination string) string {

	output := ""
	sourceClean := filepath.ToSlash(source)
	destinationClean := filepath.ToSlash(destination)

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

func betterWalkCheck(source string, destination string, checkedFiles []string) []string {
	steps := 0 //gotta count our steps

	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err == nil { //make sure there is no error from the walk

			//just grabbing these to reference later
			currentFileName := f.Name()
			currentFileIsDir := f.IsDir()
			currentFileModTime := f.ModTime()

			if steps >= 1 { //Skip first file (aka 'self'), now check each file
				//Building the equivalent path at destination
				pathArr := []string{destination, currentFileName}
				destinationPath := pathConnector(pathArr)

				currentRelativePath := relativePath(path, destinationPath)

				if !(slices.Contains(checkedFiles, currentRelativePath)) { //If the file has not been checked yet
					if !currentFileIsDir {
						fmt.Print(path, " | ", currentFileModTime, "\n") //Source File
					} else {
						fmt.Print(path, "\n") //Source File
					}

					//When file is a directory (and not the source folder), we want to enter it and walk it
					if currentFileIsDir && filepath.Base(path) != filepath.Base(source) {
						checkedFiles = betterWalkCheck(path, destinationPath, checkedFiles)
						return filepath.SkipDir
					}

					//if we get here, that means file is NOT a directory, so we start opening it and checking it
					destinationFile, destError := os.Open(destinationPath)
					if destError == nil {
						destinationFile, destError := destinationFile.Stat()
						destinationFileModTime := destinationFile.ModTime() //saving that mod time to compare with source
						if destError == nil {
							fmt.Print("(", destinationPath, " | ", destinationFileModTime, ")\n") //Destination File

							if currentFileModTime.After(destinationFileModTime) {
								fmt.Print("File is more recent in source\n")
								betterCopySync(path, destinationPath, currentFileModTime)
							} else if currentFileModTime.Before(destinationFileModTime) {
								fmt.Print("File is more recent at destination\n")
								betterCopySync(destinationPath, path, destinationFileModTime)
							} else if currentFileModTime.Equal(destinationFileModTime) {
								fmt.Print("Files are in sync\n")
							} else {
								fmt.Print("Yo what", currentFileModTime, destinationFileModTime)
							}
							fmt.Println("")

						} else {
							fmt.Print("File does not exist at destination\n")
							betterCopySync(path, destinationPath, currentFileModTime)
						}
					} else {
						fmt.Print("File does not exist at destination\n")
						betterCopySync(path, destinationPath, currentFileModTime)
					}

					checkedFiles = append(checkedFiles, currentRelativePath)
				} else {
					fmt.Println("Checked file", currentRelativePath)
				}

			} else { //The first file in pathwalk is just the root folder, we skip it to avoid adding a whole copy of A into B (this is always ran first, and never again)
				steps++
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil
	})

	return checkedFiles
}

func betterFirstBootSync() {
	fmt.Println("---BETTER FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		fmt.Println("FBS NEW WALK", currentFolderLink.UUID)
		absFolderA, _ := filepath.Abs(currentFolderLink.FolderA)
		absFolderB, _ := filepath.Abs(currentFolderLink.FolderB)

		_, folderAError := os.Open(absFolderA)
		_, folderBError := os.Open(absFolderB)

		if folderAError == nil && folderBError == nil {
			checkedFiles := []string{}

			fmt.Println("WALKING FOLDER ", absFolderA)
			checkedFiles = betterWalkCheck(absFolderA, absFolderB, checkedFiles)

			fmt.Println("WALKING FOLDER ", absFolderB)
			betterWalkCheck(absFolderB, absFolderA, checkedFiles)

			fmt.Println(" ")
		} else {
			fmt.Println("Error loading link A:", folderAError, "|| B:", folderBError)
		}
	}
}
