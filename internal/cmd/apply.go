package cmd

import (
	"repo/internal/hoster/gitlab"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply [dir]",
	Short: "Applies the repo.yaml configuration to the providers repository settings",
	Long:  `Applies the repositories repo.yaml manifest file configuration to the providers repository settings`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		provider, err := gitlab.MakeProvider()
		handleFatalError(err)

		gitDirs := collectGitDirsHandled(args[0], provider)

		for _, gd := range gitDirs {
			// validate
			errs := provider.Validate(gd.RepoMeta)
			if errs != nil {
				say.InfoLn("Skipping invalid %s", gd.Name)
				continue
			}
			// apply
			provider.Apply(gd.RepoMeta)
		}
	},
}
