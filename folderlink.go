package main

import (
	"os"
	"time"
)

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	//// General info (Bare minimum)
	Name    string `json:"name"`
	FolderA string `json:"folderA"`
	FolderB string `json:"folderB"`
	UUID    int    `json:"uuid"`

	//// General syncing settings
	FolderRecursion bool   `json:"folderRecursion"` // whether or not we want to enter/copy folders
	SyncOnStartup   bool   `json:"syncOnStartup"`
	ContinuousSync  bool   `json:"continuousSync"`
	SyncType        string `json:"syncType"` //'normal' = bidirectional, 'specific-exts' = only looks for files in specified format, creates 1:1 copy with different file ext

	//// Sync comparison settings
	FileTimeRound int `json:"fileTimeRound"` // (time.Seconds) File modification times are rounded by these many seconds (can reduce constant updates, will increase chance of missed data)
	FileModDelay  int `json:"fileModDelay"`  // (time.Seconds) The amount of time the program waits before modying destination copy properties

	////Additional properties for 'specific-exts' type
	//File format used for each folder, respectively
	SpecificExts_FolderA       string `json:"specificExts_FolderA"`
	SpecificExts_FolderB       string `json:"specificExts_FolderB"`
	SpecificExts_CaseSensitive bool   `json:"specificExts_CaseSensitive"`

	//// General (Basic) Filtering rules
	//In/Ex-clude Files
	FileInclude []string `json:"fileInclude"`
	FileExclude []string `json:"fileExclude"`
	//In/Ex-clude Folders
	FolderInclude []string `json:"folderInclude"`
	FolderExclude []string `json:"folderExclude"`
	//In/Ex-clude Extensions
	ExtInclude []string `json:"extInclude"`
	ExtExclude []string `json:"extExclude"`
}

func (f *FolderLink) roundedModTime(input os.FileInfo) (output time.Time) {
	inputModTime := input.ModTime()

	if f.FileTimeRound == 0 {
		return inputModTime
	} else {
		return inputModTime.Round(time.Second * time.Duration(f.FileTimeRound))
	}
}

func (f *FolderLink) fileModDelay() {
	if f.FileModDelay != 0 {
		time.Sleep(time.Second * time.Duration(f.FileModDelay))
	}

}
