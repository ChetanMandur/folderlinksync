package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	Name      string `json:"name"`
	FolderA   string `json:"folderA"`
	FolderB   string `json:"folderB"`
	GroupUUID int    `json:"groupUUID"`
	status    int
	UUID      int `json:"uuid"`
	//0 = no changes/idle state, 1 = changes detected, 2 = approved sync, 3 = cancel changes (folder B has changes, folder A reject changes and makes B in sync with A)

}

type FolderGroups struct {
	FolderGroups []FolderGroup `json:"groups"`
}
type FolderGroup struct {
	Name string `json:"name"`
	UUID int    `json:"uuid"`
}

type linkSumItem struct {
	name         string
	lastModified time.Time
}

type Event struct {
	fileName    string
	path        string
	isDirectory bool
	op          fsnotify.Op
}
