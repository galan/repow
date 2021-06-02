package main

import (
	"repo/internal/cmd"
)

// passed via ldflags during build, must be empty
var VersionLdFlag string

func main() {
	cmd.VersionPassed = VersionLdFlag
	cmd.Execute()
}
