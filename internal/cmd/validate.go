package cmd

import (
	"repo/internal/config"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"time"

	"github.com/spf13/cobra"
)

var validateQuiet bool
var validateOptionalContacts bool

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVarP(&validateQuiet, "quiet", "q", false, "Output only affected repositories")
	validateCmd.Flags().BoolVarP(&validateOptionalContacts, "optionalContacts", "e", false, "Allow empty contacts (existing contacts still will be validated)")
}

var validateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Validates the repo.yaml manifest file",
	Long:  `Validates the repo.yaml manifest file for the given repository or repositories below the given directory.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		dirReposRoot := getAbsoluteRepoRoot(args[0])
		gitDirs := collectGitDirsHandled(dirReposRoot, hoster)
		validateProcess(hoster, gitDirs, dirReposRoot)
	},
}

func validateProcess(hoster h.Hoster, gitDirs []model.RepoDir, dirReposRoot string) {
	defer say.Timer(time.Now())
	counter := int32(0)

	for _, gd := range gitDirs {
		dirRepoRelative := getRelativRepoDir(gd.Path, dirReposRoot)

		say.Verbose("Validating %s", dirRepoRelative)
		errValidate := hoster.Validate(gd.RepoMeta, config.Values.Options.OptionalContacts)
		if errValidate != nil {
			say.ProgressErrorArray(&counter, len(gitDirs), errValidate, dirRepoRelative, "")
		} else {
			if !validateQuiet {
				say.ProgressSuccess(&counter, len(gitDirs), dirRepoRelative, "")
			}
		}
	}
}
