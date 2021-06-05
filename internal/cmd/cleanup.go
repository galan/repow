package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"sync"
	"sync/atomic"
	"time"

	color "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

const dirArchived string = "_archived"
const dirRemoved string = "_removed"

var cleanupQuiet bool
var cleanupParallelism int

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().BoolVarP(&cleanupQuiet, "quiet", "q", false, "Output only affected repositories")
	cleanupCmd.Flags().IntVarP(&cleanupParallelism, "parallelism", "p", 64, "How many process should run in parallel, 1 would be no parallelism.")
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup [root-dir]",
	Short: "Pushes archived and deleted repositories from the checkout-directory aside",
	Long:  `Archived or deleted repositories at the hoster are moved aside from the checkout-directory. They are collected non-destructive into separate directories.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, false, true)),
	Run: func(cmd *cobra.Command, args []string) {
		dirReposRoot := args[0]
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		dirs, err := ioutil.ReadDir(dirReposRoot)
		if err != nil {
			handleFatalError(errors.New("Unable to read repository root directory (" + err.Error() + ")"))
		}

		checkRepositories(dirReposRoot, dirs, hoster)
	},
}

func checkRepositories(dirReposRoot string, dirs []os.FileInfo, hoster h.Hoster) {
	counter := int32(0)
	counterOk := int32(0)
	counterSkipped := int32(0)
	counterArchived := int32(0)
	counterRemoved := int32(0)

	defer func(start time.Time) {
		say.Plain("%s Finished, took %s (%d Ok, %d Skipped, %d Archived, %d Removed)",
			say.Repow(), time.Since(start), color.Green(counterOk).Bold(), color.Yellow(counterSkipped).Bold(), color.Blue(counterArchived).Bold(), color.Cyan(counterRemoved).Bold())
	}(time.Now())

	dirsFiltered := []os.FileInfo{}
	for _, dirRepository := range dirs {
		if dirRepository.Name() != dirArchived && dirRepository.Name() != dirRemoved {
			dirsFiltered = append(dirsFiltered, dirRepository)
		}
	}

	tasks := make(chan os.FileInfo)
	var wg sync.WaitGroup
	for i := 0; i < getParallelism(cleanupParallelism); i++ {
		wg.Add(1)
		go processDir(dirReposRoot, hoster, &counter, len(dirsFiltered), &counterOk, &counterSkipped, &counterArchived, &counterRemoved, tasks, &wg)
	}

	for _, dirRepository := range dirsFiltered {
		tasks <- dirRepository
	}
	close(tasks)
	wg.Wait()
}

func processDir(dirReposRoot string, hoster h.Hoster, counter *int32, total int, counterOk *int32, counterSkipped *int32, counterArchived *int32, counterRemoved *int32, tasks chan os.FileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for dirRepository := range tasks {
		dirRepositoryName := dirRepository.Name()

		if !dirRepository.IsDir() {
			if !cleanupQuiet {
				say.ProgressWarn(counter, total, nil, dirRepositoryName, "Not a directory (skipping)")
			}
			atomic.AddInt32(counterSkipped, 1)
			return
		}

		dirRepositoryAbsolute := path.Join(dirReposRoot, dirRepositoryName, ".git")
		if _, err := os.Stat(dirRepositoryAbsolute); os.IsNotExist(err) {
			if !cleanupQuiet {
				say.ProgressWarn(counter, total, nil, dirRepositoryName, "- Does not contain a git repository (skipping)")
			}
			atomic.AddInt32(counterSkipped, 1)
			return
		}

		remotePath := model.DetermineRemotePath(dirRepositoryAbsolute, hoster.Host())

		if remotePath == "" {
			say.ProgressWarn(counter, total, nil, dirRepositoryName, "- Unable to determine git remote name (skipping)")
			atomic.AddInt32(counterSkipped, 1)
			return
		}

		say.Verbose("RemotePath: %s: %s", dirRepositoryName, remotePath)
		state, err := hoster.ProjectState(remotePath)
		if err != nil {
			say.ProgressWarn(counter, total, err, dirRepositoryName, "- Unable to determine git remote state (skipping)")
			atomic.AddInt32(counterSkipped, 1)
			return
		}
		say.Verbose("State for %s: %v", dirRepositoryName, state)

		var errorMove error = nil
		var code string = ""
		switch state {
		case h.Ok:
			say.Verbose("Repository %s ok", dirRepositoryName)
			code = color.Green("✔").Bold().String()
			atomic.AddInt32(counterOk, 1)
		case h.Archived:
			errorMove = move(dirReposRoot, dirRepositoryName, dirArchived)
			code = color.Blue("A").Bold().String() // 📦
			atomic.AddInt32(counterArchived, 1)
		case h.Removed:
			errorMove = move(dirReposRoot, dirRepositoryName, dirRemoved)
			code = color.Cyan("R").Bold().String() // 🗑
			atomic.AddInt32(counterRemoved, 1)
		default:
			say.ProgressError(counter, total, err, dirRepositoryName, "- State for repository is unknown (skipping)")
			code = color.White("?").Bold().String()
			atomic.AddInt32(counterSkipped, 1)
			return
		}

		if errorMove != nil {
			say.ProgressError(counter, total, errorMove, dirRepositoryName, "- Unable to move")
		} else {
			if !cleanupQuiet || state != h.Ok {
				say.ProgressGeneric(counter, total, code, dirRepositoryName, "")
			}
		}
	}
}

func move(dirReposRoot string, dirRepository string, dirTarget string) error {
	dirAbsSource := path.Join(dirReposRoot, dirRepository)
	dirAbsTarget := path.Join(dirReposRoot, dirTarget)
	dirAbsTargetRepository := path.Join(dirReposRoot, dirTarget, dirRepository)

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
		return errors.New("Directory in target location is not a directory (skipping)")
	}
	// check if target-repository dir already exists
	if _, err := os.Stat(dirAbsTargetRepository); err == nil {
		return errors.New("Directory in target location already exists (skipping)")
	}

	//move
	err := os.Rename(dirAbsSource, dirAbsTargetRepository)
	if err != nil {
		return err
	}

	say.Verbose("moved %s %s", dirAbsSource, dirAbsTargetRepository)
	return nil
}
