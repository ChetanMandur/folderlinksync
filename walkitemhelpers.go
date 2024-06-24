package main

func errorCheck(status string) bool {
	if len(status) > 6 {
		return status[0:7] == "ERROR::"
	}
	return false
}
