package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
					fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
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
					fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
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
	fmt.Println(linkSums)
}

func copySync(name string, source string, destination string) {
	fmt.Println("CopySync", name, source, destination)

	folderSource := linkSums[source]
	folderDestination := linkSums[destination]

	copyError := copy.Copy(filepath.ToSlash(source+"/"+filepath.Base(name)), filepath.ToSlash(destination+"/"+filepath.Base(name)))
	if copyError != nil {
		fmt.Println("Error:", copyError)
	} else {
		folderDestination[name] = folderSource[name]
		modifyDateError := os.Chtimes(filepath.ToSlash(destination+"/"+filepath.Base(name)), folderDestination[name], folderDestination[name])
		if modifyDateError != nil {
			fmt.Println(modifyDateError)
		}
	}
}

func walkCheck(source string, destination string, checkedFiles []string) []string {
	folderSource := linkSums[source]
	folderDestination := linkSums[destination]

	steps := 0

	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err == nil {
			if steps >= 1 {
				currentFileName := f.Name()
				fmt.Print("FBS ", currentFileName, " ")

				if !(slices.Contains(checkedFiles, currentFileName)) {
					checkedFiles = append(checkedFiles, currentFileName)
					if folderDestination[currentFileName].IsZero() {
						fmt.Print("File does not exist in destination, copying \n")
						copySync(currentFileName, source, destination)

					} else {
						fmt.Print("File exists in destination -- ")
						if folderSource[currentFileName].After(folderDestination[currentFileName]) {
							fmt.Print("File is more recent in source, copying to destination \n")
							copySync(currentFileName, source, destination)

						} else {
							fmt.Print("File is more recent in destination, copying to source \n")
							copySync(currentFileName, destination, source)
						}
					}

				} else {
					fmt.Print("File already checked \n")
				}

				//we dont want to enter folders on the directory we watching, so we stop ourselves from doing that
				if f.IsDir() && filepath.Base(path) != filepath.Base(source) {
					return filepath.SkipDir
				}

			} else { //The first file in pathwalk is just the root folder, we skip it to avoid adding a whole copy of A into B
				// fmt.Println("Skipping first file")
				steps++
			}

		} else {
			fmt.Println("Error loading", source)
		}
		return nil
	})
	return checkedFiles
}

func firstBootSync() {
	fmt.Println("---FIRST BOOT SYNC---")
	for _, currentFolderLink := range folderLinks.FolderLinks {
		absFolderA, _ := filepath.Abs(currentFolderLink.FolderA)
		absFolderB, _ := filepath.Abs(currentFolderLink.FolderB)

		folderA := linkSums[absFolderA]
		folderB := linkSums[absFolderB]

		checkedFiles := []string{}

		if len(folderA) != 0 && len(folderB) != 0 {

			fmt.Println("WALKING FOLDER ", absFolderA)
			checkedFiles = walkCheck(absFolderA, absFolderB, checkedFiles)

			fmt.Println("WALKING FOLDER ", absFolderB)
			walkCheck(absFolderB, absFolderA, checkedFiles)

		} else {
			fmt.Println("Error loading folder link", currentFolderLink.FolderA, currentFolderLink.FolderB)
		}
		fmt.Println(" ")
	}

}
