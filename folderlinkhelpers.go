package main

import (
	"path/filepath"
	"slices"
	"strings"
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

// Given a path (assumed to be within one of the folder link paths), find its relative path from root
// Relative path is the path of the file without directory info beyond root,
// For example, given the folder root (either FolderA or FolderB) and the path of a file within its directory, return relative path
// FolderRoot = C:/FolderRoot/, FilePath = C:/FolderRoot/depth/test.txt, therefore RelativePath= /depth/test.txt
// root is the path of the root folder, file is the path of the file in question (with file assumed to be within root)
func relativePathFromRoot(root string, file string, extensionless bool) (relativePath string) {
	//Preping the 'file' path string
	file = filepath.ToSlash(file) //cleans the string

	//Splitting root
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
	relativePathSplit := fileSplit[len(rootSplit):]

	//Chopping off the extension if we extensionless
	if extensionless {
		//Splits the file name (from relative path), ignores the ext
		name, _ := FileNameSplitter(relativePathSplit[len(relativePathSplit)-1])
		//Sets file name (from relative path) to just name, no ext
		relativePathSplit[len(relativePathSplit)-1] = name
	}

	//make it all into a string and output
	relativePath = temp_slash + pathConnector(relativePathSplit)

	return relativePath

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

// Given a name in style "file.ext", can return both "file" and "ext" separately
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
