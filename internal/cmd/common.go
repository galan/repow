package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"repo/internal/config"
	h "repo/internal/hoster"
	"repo/internal/model"
	"repo/internal/say"
	"repo/internal/util"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	REPOW_STYLE string = "REPOW_STYLE"
)

const (
	dirArchived string = "_archived"
	dirRemoved  string = "_removed"

	styleFlat      string = "flat"
	styleRecursive string = "recursive"
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
		msg = "Missing required argument for " + msg

		if argIndex >= len(args) {
			return errors.New(msg)
		}
		repoPath := args[argIndex]
		if !util.ExistsDir(repoPath) {
			return errors.New(msg + " (directory does not exist)")
		}
		condParent := repoParent && util.ExistsDir(path.Join(repoPath, ".git"))
		condRoot := repoRoot

		if !condParent && !condRoot {
			return errors.New(msg)
		}
		return nil
	}
}

var Style string

func validateFlags(cmd *cobra.Command, args []string) error {
	stylesAvailable := []string{styleFlat, styleRecursive}

	styleSelected := config.UsedConfig.Options.Style
	if !slices.Contains(stylesAvailable, styleSelected) {
		return fmt.Errorf("invalid value for style: %q", styleSelected)
	}
	Style = styleSelected
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
			return fs.SkipDir
		}

		if util.ExistsDir(path.Join(dir, ".git")) { // check if given path is git-repository
			repo, err := model.MakeRepoDir(dir, hoster.Host())
			if err != nil {
				say.Verbose("Failed determine repository directory: %s", e)
				return e
			}
			result = append(result, *repo)
			return fs.SkipDir
		}
		return nil
	}
	err = filepath.WalkDir(root, walk)
	return result, err
}

// convenience function, would exit automatically on error or empty result
func collectGitDirsHandled(dir string, hoster h.Hoster) []model.RepoDir {
	gitDirs, err := collectGitDirs(dir, hoster)
	handleFatalError(err)

	if len(gitDirs) == 0 {
		handleFatalError(errors.New("argument must point to single git-directory or parent-directory that contains git-directories"))
	}
	return gitDirs
}

func getParallelism(given int) int {
	return int(math.Max(1, float64(given)))
}

func getAbsoluteRepoRoot(repositoryArg string) string {
	abs, err := filepath.Abs(repositoryArg)
	if err != nil {
		say.Error("Unable to determine absolute root-dir for %s", err)
		os.Exit(1)
	}
	say.Verbose("Absolute repository root: %s", abs)
	return abs
}

func getRelativRepoDir(dirAbsRepoRoot string, dirAbsRepo string) string {
	return strings.TrimPrefix(strings.TrimPrefix(dirAbsRepoRoot, dirAbsRepo), "/")
}
