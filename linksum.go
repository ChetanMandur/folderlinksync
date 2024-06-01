package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// linkSums is what holds all directory info (name+lastModified)

func initalizeLinkSum() {
	fmt.Println("---INITALIZE LINK SUM---")
	for _, currentFolderLink := range folderLinks.FolderLinks {

		if linkSums[currentFolderLink.FolderA] == nil {
			linkSums[currentFolderLink.FolderA] = make(map[string]time.Time)
			filepath.Walk(currentFolderLink.FolderA, func(path string, f os.FileInfo, err error) error {
				if err == nil {
					// fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
					linkSums[currentFolderLink.FolderA][f.Name()] = f.ModTime()
					//we dont want to enter folders on the directory we watching, so we stop ourselves from doing that
					if f.IsDir() && filepath.Base(path) != filepath.Base(currentFolderLink.FolderA) {
						return filepath.SkipDir
					}

				} else { //error means that len(linkSums[folder]) = 0!!!!
					fmt.Println("Error loading", currentFolderLink.FolderA, len(linkSums[currentFolderLink.FolderA]))
				}
				return nil
			})
		}

		if linkSums[currentFolderLink.FolderB] == nil {
			linkSums[currentFolderLink.FolderB] = make(map[string]time.Time)
			filepath.Walk(currentFolderLink.FolderB, func(path string, f os.FileInfo, err error) error {
				if err == nil {
					// fmt.Println(f.Name(), f.ModTime(), filepath.Base(path))
					linkSums[currentFolderLink.FolderB][f.Name()] = f.ModTime()
					if f.IsDir() && filepath.Base(path) != filepath.Base(currentFolderLink.FolderB) {
						return filepath.SkipDir
					}
				} else {
					fmt.Println("Error loading", currentFolderLink.FolderB, len(linkSums[currentFolderLink.FolderB]))
				}
				return nil
			})
		}

	}
	// fmt.Println(linkSums)
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
