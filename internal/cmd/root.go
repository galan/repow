package cmd

import (
	"repo/internal/config"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repow",
	Short: "repository managment",
	Long:  "repow " + say.Repow() + " convenient and fast repository management with self-containing meta-data.\n\nhttps://github.com/galan/repow",
}

var VersionPassed string

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&say.VerboseEnabled, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is $HOME/.repow.yaml)")
	cobra.OnInitialize(config.InitConfig)
	err := rootCmd.Execute()
	handleFatalError(err)
}
