package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	h "repo/internal/hoster"
	"repo/internal/model"
	"repo/internal/say"
	"repo/internal/util"
	"slices"

	"github.com/spf13/cobra"
)

const dirArchived string = "_archived"
const dirRemoved string = "_removed"

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

func validateFlags(cmd *cobra.Command, args []string) error {
	var stylesAvailable = []string{"flat", "recursive"}
	var style = cmd.Flag("style").Value.String()
	if !slices.Contains(stylesAvailable, style) {
		return fmt.Errorf("invalid value for style: %q", style)
	}
	return nil
}

// check all directories recursivly for .git directory
// collect them and return the array
func collectGitDirs(root string, hoster h.Hoster) (result []model.RepoDir, err error) {

	ignored := []string{path.Join(root, dirArchived), path.Join(root, dirRemoved)}

	walk := func(dir string, d fs.DirEntry, e error) error {
		if !d.IsDir() {
			return nil
		}

		if slices.Contains(ignored, dir) {
			return nil
		}

		if util.ExistsDir(path.Join(dir, ".git")) { // check if given path is git-repository
			repo, err := model.MakeRepoDir(dir, hoster.Host())
			if err != nil {
				say.Verbose("Failed determine repository directory: %s", e)
				return e
			}
			result = append(result, *repo)
			//say.Info("Appended %s\n", dir)
			return fs.SkipDir
		}

		//fmt.Println(dir, d.Name(), "directory?", d.IsDir())
		return nil
	}
	err = filepath.WalkDir(root, walk)
	return result, err

	/*
		dirAbs, _ := filepath.Abs(dir)
		if util.ExistsDir(path.Join(dirAbs, ".git")) { // check if given path is git-repository
			repo, err := model.MakeRepoDir(dirAbs, hoster.Host())
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
					repo, err := model.MakeRepoDir(path.Join(dir, d.Name()), hoster.Host())
					if err != nil {
						say.Verbose("Not adding dir %s: %s", d.Name(), err)
						return result, err
					}
					result = append(result, *repo)
				}
			}
		}
	*/
	say.Info("found paths: %v", result)
	return result, nil
}

// convenience function, would exit automatically on error or empty result
func collectGitDirsHandled(dir string, hoster h.Hoster) []model.RepoDir {
	gitDirs, err := collectGitDirs(dir, hoster)
	handleFatalError(err)

	if len(gitDirs) == 0 {
		handleFatalError(errors.New("Argument must point to single git-directory or parent-directory that contains git-directories."))
	}
	return gitDirs
}

func getParallelism(given int) int {
	return int(math.Max(1, float64(given)))
}
