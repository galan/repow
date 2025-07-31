package cmd

import (
	"repo/internal/config"
	"repo/internal/say"

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repow",
	Short: "repository managment",
	Long:  "repow " + say.Repow() + " convenient and fast repository management with self-containing meta-data.\n\n" + aurora.Hyperlink("https://github.com/galan/repow", "https://github.com/galan/repow").String(),
}

var VersionPassed string

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&say.VerboseEnabled, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&config.ConfigFile, "configfile", "c", "", "custom config-file location (default "+config.DefaultConfigFile()+")")
	err := rootCmd.Execute()
	handleFatalError(err)
}
