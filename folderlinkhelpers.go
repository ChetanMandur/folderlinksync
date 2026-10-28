package main

import (
	"path/filepath"
	"slices"
	"strings"
)

// Function takes an array of strings, and concats it into a single string, seperated by forward slash
//
// The array of strings represent folders, with locations[0] being a folder at root, while locations[len(locations)-1] is the last ((great-)grand-)child element
//
// The final string represents a tradtional path from root to folder
func pathConnector(locations []string) string {
	// Holds final path string
	output := ""

	// Iteriate through all strings
	for i := 0; i < len(locations); i++ {
		if i == len(locations)-1 { // If the current iteration is the last
			// Do not need trailing forward slash
			output += locations[i]
		} else { // If the current iteration is not the last
			// Need a trailing forward slash
			output += locations[i] + "/"
		}
	}

	// Normalizes slashes to forward slash?
	// Does this also clean the string? What was I thinking? This is probs a holdover from some old implementation that I didn't remove
	// Or, I meant to do `.fromSlash()` and I didn't realize since I only test on windows
	output = filepath.ToSlash(output)
	return output
}

/*
Function returns the 'relative path' of the 'file' from its 'root' folder

`relative path` is the path of the `file` without the directory info of `root` (In other words: it is the path to the `file` that contains path data past `root`)

In this case, `root` is not the same as the 'system root' (which varies between OSs), `root` is the source parent folder of `file` and is defined as a path from 'system root'

Parameters include:
  - `root` = The targeted parent directory for the file; defined as path from 'system root'
  - `file` = The targeted file, which is a child of `root`; defined as path from 'system root'
  - `extensionless` = Whether the final `relativePath` should include the '.ext' on the file

@example - `root` = C:/FolderRoot/, `file` = C:/FolderRoot/depth/test.txt, therefore `relativePath`= /depth/test.txt
*/
func relativePathFromRoot(root string, file string, extensionless bool) (relativePath string) {
	// Preping the 'file' path string
	file = filepath.ToSlash(file) //cleans the string

	// Splitting the `root` path into an array
	rootSplit := strings.Split(root, "/")

	// Use backslash if forward slash didnt work
	// (Why would forward slash not work if I already use `.ToSlash()`????)
	if len(rootSplit) == 1 {
		rootSplit = strings.Split(root, "\\")
	}

	// Splitting the `file` path into an array
	fileSplit := strings.Split(file, "/")
	temp_slash := "/" //We want to continue using the same type of slashes for the relative path output

	// Use backslash if forward slash didnt work (and update temp_slash)
	// (Again, why would forward slash not work if I already use `.ToSlash()`????)
	if len(fileSplit) == 1 {
		fileSplit = strings.Split(file, "\\")
		temp_slash = "\\"
	}

	// Chop off the root path from the file path
	relativePathSplit := fileSplit[len(rootSplit):]

	// Chopping off the extension from the relative path if we are extensionless
	// This is done, because the file will have a different extension depending on its location
	// Example: `FolderA` has `.jpg`, while `FolderB` has `.jpeg` ==> Data is 1:1 but the extension is different
	if extensionless {
		// The file itself is just the final entry in `relativePathSplit`
		// Grabs just the file name, ignores the ext
		name, _ := FileNameSplitter(relativePathSplit[len(relativePathSplit)-1])

		// Sets final entry (ie. the file) as just it's name
		relativePathSplit[len(relativePathSplit)-1] = name
	}

	// Turn `relativePathSplit` array back into an array and output it
	relativePath = temp_slash + pathConnector(relativePathSplit)

	return relativePath

}

/*
Function to test a file against the `FolderLink` inclusion and exclusion filters

Parameters:
- `folderLink` = The link whose filters are being checked
- `inverse` = Whether or not the file is coming from `Folder A` (`true`) or `Folder B` (`false`)
- `sourceFileNameShort` = The name of the file (without the extension)
- `sourceFileNameExt` = The extension of the file (without the name and the period)

Returns `true` the file passes the filters
*/
func fileFilterCheck(folderLink *FolderLink, inverse bool, sourceFileNameShort string, sourceFileNameExt string) bool {
	// If the `FolderLink` has a list of specific file names to exclude (blacklist)
	if folderLink.FileExclude != nil {
		// Check if the name name exists within the list
		if slices.Contains(folderLink.FileExclude, sourceFileNameShort) {
			// If the name is found, the check fails
			return false
		}
	}

	// If the `FolderLink` has a list of specific file names that it can only include (whitelist)
	if folderLink.FileInclude != nil {
		// Check if the name does NOT exist within the list, and that the list has items within it
		if !(slices.Contains(folderLink.FileInclude, sourceFileNameShort)) && len(folderLink.FileInclude) != 0 {
			// If the name is not found, the check fails
			return false
		}
	}

	// If the `FolderLink` has a list of specific file extensions to exclude (blacklist)
	if folderLink.ExtExclude != nil {
		// Checks if the extension is within the list
		if slices.Contains(folderLink.ExtExclude, sourceFileNameExt) {
			// If the extension is found, the check fails
			return false
		}

	}

	// If the `FolderLink` has a list of specific file extensions that it can only include (whitelist)
	if folderLink.ExtInclude != nil {
		// Check if the extension does NOT exist withn the list, and that the list has items within it
		if !(slices.Contains(folderLink.ExtInclude, sourceFileNameExt)) && len(folderLink.ExtInclude) != 0 {
			// If the extension is not found, the check fails
			return false
		}

	}

	// If the `FolderLink` has a `SyncType` of `specific-exts`
	// We need to check if the file extension passes the filter
	// The filter being: if the file extension matches the specific extension
	if folderLink.SyncType == "specific-exts" {
		// The `specific extension` required for the specific folder of the `FolderLink`
		var sourceFolderExt string

		if inverse { // If inverse is `true`, grab the `specific-ext` for `Folder B`
			sourceFolderExt = folderLink.SpecificExts_FolderB
		} else { // If inverse is `true`, grab the `specific-ext` for `Folder A`
			sourceFolderExt = folderLink.SpecificExts_FolderA
		}

		// Used to signify if the extension names match
		var extensionMatch = false
		if folderLink.SpecificExts_CaseSensitive { // If the extension name requires case sensitivity
			// Simple check if the strings are equal
			extensionMatch = (sourceFolderExt == sourceFileNameExt)

		} else { // If the extension name does not require case sensitivity
			// Check without care for case sensitivity
			extensionMatch = (strings.EqualFold(sourceFolderExt, sourceFileNameExt))
		}

		// If the specific extension does not match the file extension, the check fails
		if !extensionMatch {
			return false
		}
	}

	// All checks pass, therefore file passes the filter check
	return true
}

func folderFilterCheck(folderLink *FolderLink, sourceFileName string) bool {
	// If the `FolderLink` has a list of specific folder names to exclude (blacklist)
	if folderLink.FolderExclude != nil {
		// Check if the name name exists within the list
		if slices.Contains(folderLink.FolderExclude, sourceFileName) {
			// If the name is found, the check fails
			return false
		}
	}

	// If the `FolderLink` has a list of specific folder names that it can only include (whitelist)
	if folderLink.FolderInclude != nil {
		// Check if the name does NOT exist within the list, and that the list has items within it
		if !(slices.Contains(folderLink.FolderInclude, sourceFileName)) && len(folderLink.FolderInclude) != 0 {
			// If the name is not found, the check fails
			return false
		}
	}

	// All checks pass, therefore file passes the filter check
	return true
}

// Given a name in style "file.ext", can return both "file" and "ext" separately
func FileNameSplitter(originalName string) (finalName string, finalExt string) {
	// Splits the name by '.'
	originalNameSplit := strings.Split(originalName, ".")

	// NOTE: the following logic only works for files with extensions, extensionless files do not work
	// ik how to fix it and im gonna fix it....later

	// Extension will always be the last item in the split
	finalExt = originalNameSplit[len(originalNameSplit)-1]

	// A file can naturally have "." throughout it's name
	// so it is possible to have more than 2 substrings with a seperator of "."
	// We want to grab all but the last substring, (since we know, the last substring is the extension)
	finalNameSplit := originalNameSplit[0 : len(originalNameSplit)-1]

	// Reconstruct the file name
	for index, value := range finalNameSplit {
		if index != 0 {
			finalName = finalName + "."
		}
		finalName = finalName + value
	}

	// Return the file name and extension
	return finalName, finalExt

}
