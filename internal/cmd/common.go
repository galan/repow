package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"repo/internal/hoster"
	"repo/internal/model"
	"repo/internal/say"
	"repo/internal/util"

	"github.com/spf13/cobra"
)

func handleFatalError(err error) {
	if err != nil {
		say.Error("%s", err)
		os.Exit(1)
	}
}

func validateConditions(conds ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, cond := range conds {
			err := cond(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func validateArgGitDir(argIndex int, repoParent bool, repoRoot bool) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		//pre-construct error-message
		var msg string
		if repoParent {
			msg = "repository directory"
		}
		if repoRoot {
			if len(msg) > 0 {
				msg = msg + " or "
			}
			msg = msg + "root-directory with repositories"
		}
		msg = "Missing required argument to " + msg

		if argIndex >= len(args) {
			return errors.New(msg)
		}
		repoPath := args[argIndex]
		if !util.ExistsDir(repoPath) {
			return errors.New(msg)
		}
		condParent := repoParent && util.ExistsDir(path.Join(repoPath, ".git"))
		condRoot := repoRoot

		if !condParent && !condRoot {
			return errors.New(msg)
		}
		return nil
	}
}

// check if .git in dirRoot
// if yes add to array
// if no go thru every directory within dir
// collect all directories that contain .git
func collectGitDirs(dir string, provider hoster.Hoster) (result []model.RepoDir, err error) {
	dirAbs, _ := filepath.Abs(dir)
	if util.ExistsDir(path.Join(dirAbs, ".git")) { // check if given path is git-repository
		repo, err := model.MakeRepoDir(dirAbs, provider.Host())
		if err != nil {
			say.Verbose("Failed determine repository directory: %s", err)
			return result, err
		}
		result = append(result, *repo)
	} else { // else search for all sub-dir git-repositories
		dirs, err := ioutil.ReadDir(dirAbs)
		if err != nil {
			return result, err
		}
		for _, d := range dirs {
			if util.ExistsDir(path.Join(dirAbs, d.Name(), ".git")) {
				repo, err := model.MakeRepoDir(path.Join(dir, d.Name()), provider.Host())
				if err != nil {
					say.Verbose("Not adding dir %s: %s", d.Name(), err)
					return result, err
				}
				result = append(result, *repo)
			}
		}
	}
	return result, nil
}

// convenience function, would exit automatically on error or empty result
func collectGitDirsHandled(dir string, provider hoster.Hoster) []model.RepoDir {
	gitDirs, err := collectGitDirs(dir, provider)
	handleFatalError(err)

	if len(gitDirs) == 0 {
		handleFatalError(errors.New("Argument must point to single git-directory or parent-directory that contains git-directories."))
	}
	return gitDirs
}
