package cmd

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"repo/internal/config"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"sync"
	"sync/atomic"
	"time"

	color "github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
)

var cleanupQuiet bool
var cleanupParallelism int

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().BoolVarP(&cleanupQuiet, "quiet", "q", false, "Output only affected repositories")
	cleanupCmd.Flags().IntVarP(&cleanupParallelism, "parallelism", "p", 32, "How many process should run in parallel, 1 would be no parallelism.")
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [root-dir]",
	Short: "Pushes archived and deleted repositories from the checkout-directory aside",
	Long:  `Archived or deleted repositories at the hoster are moved aside from the checkout-directory. They are collected non-destructive into separate directories.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, false, true)),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		dirReposRoot := getAbsoluteRepoRoot(args[0])

		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		gitDirs := collectGitDirsHandled(dirReposRoot, hoster)

		checkRepositories(dirReposRoot, gitDirs, hoster)
	},
}

func checkRepositories(dirReposRoot string, dirs []model.RepoDir, hoster h.Hoster) {
	counter := int32(0)
	counterOk := int32(0)
	counterSkipped := int32(0)
	counterArchived := int32(0)
	counterRemoved := int32(0)

	defer func(start time.Time) {
		say.Plain("%s Finished, took %s (%d Ok, %d Skipped, %d Archived, %d Removed)",
			say.Repow(), time.Since(start), color.Green(counterOk).Bold(), color.Yellow(counterSkipped).Bold(), color.Blue(counterArchived).Bold(), color.Cyan(counterRemoved).Bold())
	}(time.Now())

	tasks := make(chan model.RepoDir)
	var wg sync.WaitGroup
	for i := 0; i < getParallelism(config.Values.Options.Parallelism); i++ {
		wg.Add(1)
		go processDir(dirReposRoot, hoster, &counter, len(dirs), &counterOk, &counterSkipped, &counterArchived, &counterRemoved, tasks, &wg)
	}

	for _, dirRepository := range dirs {
		tasks <- dirRepository
	}
	close(tasks)
	wg.Wait()
}

func processDir(dirReposRoot string, hoster h.Hoster, counter *int32, total int, counterOk *int32, counterSkipped *int32, counterArchived *int32, counterRemoved *int32, tasks chan model.RepoDir, wg *sync.WaitGroup) {
	defer wg.Done()
	for dirRepository := range tasks {

		dirRepoRelative := getRelativRepoDir(dirRepository.Path, dirReposRoot)

		remotePath := model.DetermineRemotePath(dirRepository.Path, hoster.Host())

		if remotePath == "" {
			say.ProgressWarn(counter, total, nil, dirRepoRelative, "- Unable to determine git remote name (skipping)")
			atomic.AddInt32(counterSkipped, 1)
			continue
		}

		say.Verbose("RemotePath: %s: %s", dirRepoRelative, remotePath)
		state, err := hoster.ProjectState(remotePath)
		if err != nil {
			say.ProgressWarn(counter, total, err, dirRepoRelative, "- Unable to determine git remote state (skipping)")
			atomic.AddInt32(counterSkipped, 1)
			continue
		}
		say.Verbose("State for %s: %v", dirRepoRelative, state)

		var errorMove error = nil
		var code string = ""
		switch state {
		case h.Ok:
			say.Verbose("Repository %s ok", dirRepoRelative)
			code = color.Green("✔").Bold().String()
			atomic.AddInt32(counterOk, 1)
		case h.Archived:
			errorMove = move(dirReposRoot, dirRepository.Path, dirArchived)
			code = color.Blue("A").Bold().String() // 📦
			atomic.AddInt32(counterArchived, 1)
		case h.Removed:
			errorMove = move(dirReposRoot, dirRepository.Path, dirRemoved)
			code = color.Cyan("R").Bold().String() // 🗑
			atomic.AddInt32(counterRemoved, 1)
		default:
			say.ProgressError(counter, total, err, dirRepoRelative, "- State for repository is unknown (skipping)")
			code = color.White("?").Bold().String()
			atomic.AddInt32(counterSkipped, 1)
			continue
		}

		if errorMove != nil {
			say.ProgressError(counter, total, errorMove, dirRepoRelative, "- Unable to move")
		} else {
			if !cleanupQuiet || state != h.Ok {
				say.ProgressGeneric(counter, total, code, dirRepoRelative, "")
			}
		}
	}
}

func move(dirReposRoot string, dirRepository string, dirTarget string) error {
	dirRepoRelative := getRelativRepoDir(dirRepository, dirReposRoot)
	dirAbsSource := dirRepository
	dirAbsTarget := path.Join(dirReposRoot, dirTarget)
	dirAbsTargetRepository := path.Join(dirReposRoot, dirTarget, dirRepoRelative)

	// check & prepare target
	// create if not exists
	if _, err := os.Stat(dirAbsTarget); os.IsNotExist(err) {
		errMk := os.Mkdir(dirAbsTarget, 0775)
		if errMk != nil {
			return errMk
		}
	}
	// check if target is directory
	fiTarget, _ := os.Stat(dirAbsTarget)
	if !fiTarget.IsDir() {
		return errors.New("directory in target location is not a directory (skipping)")
	}
	// check if target-repository dir already exists
	if _, err := os.Stat(dirAbsTargetRepository); err == nil {
		return errors.New("directory in target location already exists (skipping)")
	}

	//move
	say.Verbose("Moving %s to %s\n", dirRepoRelative, dirAbsTargetRepository)
	// mkdirs except last in target directory
	dirAbsTargetRepositoryParent := filepath.Dir(dirAbsTargetRepository)
	errMkdir := os.MkdirAll(dirAbsTargetRepositoryParent, 0755)
	if errMkdir != nil {
		return errMkdir
	}

	err := os.Rename(dirAbsSource, dirAbsTargetRepository)
	if err != nil {
		return err
	}

	say.Verbose("moved %s %s", dirAbsSource, dirAbsTargetRepository)
	return nil
}
