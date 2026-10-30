package main

// Helper function to determine if the status given was an error
//
// Returns `true` if the status was an error, and `false` if not
func errorCheck(status string) bool {
	if len(status) > 6 {
		return status[0:7] == "ERROR::"
	}
	return false
}
