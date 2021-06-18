package cmd

import (
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"time"

	"github.com/spf13/cobra"
)

var validateQuiet bool

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Output only affected repositories")
}

var validateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Validates the repo.yaml manifest file",
	Long:  `Validates the repo.yaml manifest file for the given repository or repositories below the given directory.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		gitDirs := collectGitDirsHandled(args[0], hoster)
		validateProcess(hoster, gitDirs)
	},
}

func validateProcess(hoster h.Hoster, gitDirs []model.RepoDir) {
	defer say.Timer(time.Now())
	counter := int32(0)

	for _, gd := range gitDirs {
		say.Verbose("Validating %s", gd.Name)
		errValidate := hoster.Validate(gd.RepoMeta)
		if errValidate != nil {
			say.ProgressErrorArray(&counter, len(gitDirs), errValidate, gd.Name, "")
		} else {
			if !validateQuiet {
				say.ProgressSuccess(&counter, len(gitDirs), gd.Name, "")
			}
		}
	}
}
