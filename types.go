package main

import (
	"github.com/fsnotify/fsnotify"
)

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	Name    string `json:"name"`
	FolderA string `json:"folderA"`
	FolderB string `json:"folderB"`
	//0 = no changes/idle state, 1 = changes detected, 2 = approved sync, 3 = cancel changes (folder B has changes, folder A reject changes and makes B in sync with A)
	status          int
	UUID            int    `json:"uuid"`
	FolderRecursion bool   `json:"folderRecursion"` // whether or not we want to enter/copy folders
	SyncType        string `json:"syncType"`        //'normal' = bidirectional, 'specific-exts' = only looks for files in specified format, creates 1:1 copy with different file ext
	////Additional properties for 'specific-exts' type
	//File format used for each folder, respectively
	SpecificExts_FolderA       string `json:"specificExts_FolderA"`
	SpecificExts_FolderB       string `json:"specificExts_FolderB"`
	SpecificExts_CaseSensitive bool   `json:"specificExts_CaseSensitive"`

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
type Event struct {
	fileName    string
	path        string
	isDirectory bool
	op          fsnotify.Op
}
