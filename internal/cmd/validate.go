package cmd

import (
	"repo/internal/hoster/gitlab"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Validates the repo.yaml manifest file",
	Long:  `Validates the repo.yaml manifest file for the given repository or repositories below the given directory.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		provider, err := gitlab.MakeProvider()
		handleFatalError(err)

		gitDirs := collectGitDirsHandled(args[0], provider)
		counter := int32(0)

		for _, gd := range gitDirs {
			say.Verbose("Validating %s", gd.Name)
			errValidate := provider.Validate(gd.RepoMeta)
			if errValidate != nil {
				say.ProgressErrorArray(&counter, len(gitDirs), errValidate, gd.Name, "")
			} else {
				say.ProgressSuccess(&counter, len(gitDirs), gd.Name, "")
			}
		}
	},
}
