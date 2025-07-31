package cmd

import (
	"repo/internal/config"
	"repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"time"

	"github.com/spf13/cobra"
)

var applyOptionalContacts bool

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().BoolVarP(&applyOptionalContacts, "optionalContacts", "e", false, "Allow empty contacts (existing contacts still will be validated)")
}

var applyCmd = &cobra.Command{
	Use:   "apply [dir]",
	Short: "Applies the repo.yaml configuration to the hosters repository settings",
	Long:  `Applies the repositories repo.yaml manifest file configuration to the hosters repository settings`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, true, true)),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		dirReposRoot := getAbsoluteRepoRoot(args[0])
		gitDirs := collectGitDirsHandled(dirReposRoot, hoster)
		applyProcess(hoster, gitDirs)
	},
}

func applyProcess(hoster hoster.Hoster, gitDirs []model.RepoDir) {
	defer say.Timer(time.Now())
	for _, gd := range gitDirs {
		// validate
		errs := hoster.Validate(gd.RepoMeta, config.Values.Options.OptionalContacts)
		if errs != nil {
			say.InfoLn("Skipping invalid %s", gd.Name)
			continue
		}
		// apply
		hoster.Apply(gd.RepoMeta)
	}
}
