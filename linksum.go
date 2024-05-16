package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// linkSums is what holds all directory info (name+lastModified)

func initalizeLinkSum() {
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

func firstBootSync() {
	for _, currentFolderLink := range folderLinks.FolderLinks {
		absFolderA, _ := filepath.Abs(currentFolderLink.FolderA)
		absFolderB, _ := filepath.Abs(currentFolderLink.FolderB)

		folderA := linkSums[absFolderA]
		folderB := linkSums[absFolderB]

		checkedFiles := []string{}

		if folderA[filepath.Base(absFolderA)] != folderB[filepath.Base(absFolderB)] {
			fmt.Println("These folders are not in sync!", absFolderA, " & ", absFolderB)

			if len(folderA) != 0 && len(folderB) != 0 {

				steps := 0
				fmt.Println("WALKING FOLDER A")
				filepath.Walk(absFolderA, func(path string, f os.FileInfo, err error) error {
					if err == nil {
						if steps >= 1 {
							currentFileName := f.Name()
							fmt.Print("FBS ", currentFileName, " ")

							if !(slices.Contains(checkedFiles, currentFileName)) {
								checkedFiles = append(checkedFiles, f.Name())
								if folderB[currentFileName].IsZero() {
									fmt.Print("File does not exist in folder B \n")
								} else {
									fmt.Print("File exists in folder B \n")
								}

							} else {
								fmt.Print("File already checked \n")
							}

							//we dont want to enter folders on the directory we watching, so we stop ourselves from doing that
							if f.IsDir() && filepath.Base(path) != filepath.Base(absFolderA) {
								return filepath.SkipDir
							}

						} else { //The first file in pathwalk is just the root folder, we skip it to avoid adding a whole copy of A into B
							fmt.Println("Skipping first file")
							steps++
						}

					} else {
						fmt.Println("Error loading", absFolderA)
					}
					return nil
				})
				fmt.Println("WALKING FOLDER B")

				steps = 0
				filepath.Walk(absFolderB, func(path string, f os.FileInfo, err error) error {
					if err == nil {
						if steps >= 1 {
							currentFileName := f.Name()
							fmt.Print("FBS ", currentFileName, " ")

							if !(slices.Contains(checkedFiles, currentFileName)) {
								checkedFiles = append(checkedFiles, f.Name())
								if folderA[currentFileName].IsZero() {
									fmt.Print("File does not exist in folder A \n")
								} else {
									fmt.Print("File exists in folder A \n")
								}

							} else {
								fmt.Print("File already checked \n")
							}

							//we dont want to enter folders on the directory we watching, so we stop ourselves from doing that
							if f.IsDir() && filepath.Base(path) != filepath.Base(absFolderB) {
								return filepath.SkipDir
							}

						} else { //The first file in pathwalk is just the root folder, we skip it to avoid adding a whole copy of A into B
							fmt.Println("Skipping first file")
							steps++
						}

					} else {
						fmt.Println("Error loading", absFolderB)
						steps++
					}
					return nil
				})

			} else {
				fmt.Println("Error loading folder link", currentFolderLink.FolderA, currentFolderLink.FolderB)
			}
		}

	}

}
