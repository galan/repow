package cmd

import (
	"repo/internal/say"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the application version",
	Long:  `Prints the application version`,
	Run: func(cmd *cobra.Command, args []string) {
		say.InfoLn("%s", VersionPassed)
	},
}
