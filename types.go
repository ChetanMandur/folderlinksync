package main

type FolderLinks struct {
	FolderLinks []FolderLink `json:"links"`
}

type FolderLink struct {
	Name      string `json:"name"`
	FolderA   string `json:"folderA"`
	FolderB   string `json:"folderB"`
	linkSumA  string
	linkSumB  string
	GroupUUID int `json:"groupUUID"`
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

type linkSum struct {
	linkSumPath string
}
