package main

type fLink struct {
	pathA     string
	pathB     string
	linkSumA  string
	linkSumB  string
	groupUUID string
	status    int
	uuid      int
	//0 = no changes/idle state, 1 = changes detected, 2 = approved sync, 3 = cancel changes (folder B has changes, folder A reject changes and makes B in sync with A)

}

type fGroup struct {
	name string
	uuid string
}

type linkSum struct {
	linkSumPath string
}
