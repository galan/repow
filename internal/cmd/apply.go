package cmd

import (
	"repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply [dir]",
	Short: "Applies the repo.yaml configuration to the hosters repository settings",
	Long:  `Applies the repositories repo.yaml manifest file configuration to the hosters repository settings`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		gitDirs := collectGitDirsHandled(args[0], hoster)
		applyProcess(hoster, gitDirs)
	},
}

func applyProcess(hoster hoster.Hoster, gitDirs []model.RepoDir) {
	defer say.Timer()
	for _, gd := range gitDirs {
		// validate
		errs := hoster.Validate(gd.RepoMeta)
		if errs != nil {
			say.InfoLn("Skipping invalid %s", gd.Name)
			continue
		}
		// apply
		hoster.Apply(gd.RepoMeta)
	}
}
