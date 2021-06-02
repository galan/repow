package util

import "os"

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func ExistsFile(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// checks if the directory exists and is not a file
func ExistsDir(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
