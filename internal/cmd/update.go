package cmd

import (
	"errors"
	"fmt"
	"math"
	"repo/internal/config"
	"repo/internal/gitclient"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
)

var updateQuiet bool
var updateParallelism int

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&updateQuiet, "quiet", "q", false, "Output only affected repositories")
	updateCmd.Flags().IntVarP(&updateParallelism, "parallelism", "p", 32, "How many process should run in parallel, 1 would be no parallelism.")
}

var updateCmd = &cobra.Command{
	Use:   "update [mode] [dir]",
	Short: "Checks/fetches/pulls updates for the given repository/repositories",
	Long: `Checks/fetches/pulls updates for the given repository or repositories below the given directory.

Mode can be one of:
  check - Outputs the current state of the local repositories
  fetch - Fetches remote changes and outputs the changes
  pull  - Fetches remote changes, merges them (if fast-forward is possible) and outputs the changes`,
	Args: validateConditions(cobra.ExactArgs(2), validateArgGitDir(1, false, true)),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		defer say.Timer(time.Now())
		modesAvailable := []string{"check", "fetch", "pull"}

		mode := args[0]
		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		if !slices.Contains(modesAvailable, mode) {
			handleFatalError(errors.New(fmt.Sprintf("mode has to be one of: %s", modesAvailable)))
		}

		dirReposRoot := getAbsoluteRepoRoot(args[1])
		gitDirs := collectGitDirsHandled(dirReposRoot, hoster)

		if mode == "fetch" || mode == "pull" {
			gitclient.PrepareSsh(hoster.Host())
		}

		tasks := make(chan *StateContext)
		var wg sync.WaitGroup
		for i := 0; i < getParallelism(config.Values.Options.Parallelism); i++ {
			wg.Add(1)
			go processRepository(mode, tasks, &wg)
		}
		counter := int32(0)
		for _, gd := range gitDirs {
			//TODO clarify why intermediate var is required?
			var rdIntermediate model.RepoDir
			rdIntermediate = gd
			dirRelative := getRelativRepoDir(gd.Path, dirReposRoot)
			tasks <- &StateContext{
				total:       len(gitDirs),
				counter:     &counter,
				repo:        &rdIntermediate,
				dirRelative: dirRelative,
			}
		}

		close(tasks)
		wg.Wait()
	},
}

type State int

const (
	clean State = iota
	dirty
	failed
)

type StateContext struct {
	total       int
	counter     *int32
	mutex       sync.Mutex // avoid mixed outputs
	repo        *model.RepoDir
	dirRelative string
	state       State
	ref         string
	behind      int
	message     string
}

func processRepository(mode string, tasks chan *StateContext, wg *sync.WaitGroup) {
	defer wg.Done()
	for ctx := range tasks {
		ctx.ref = gitclient.GetCurrentBranch(ctx.repo.Path)
		switch mode {
		case "check":
			updateCheck(ctx)
		case "fetch":
			updateFetch(ctx)
		case "pull":
			updateFetch(ctx)
			if ctx.state != failed && ctx.state != clean {
				updatePull(ctx)
			}
		}
		printContext(ctx)
	}
}

func updateCheck(ctx *StateContext) {
	if gitclient.IsDirty(ctx.repo.Path) {
		ctx.message = gitclient.GetLocalChanges(ctx.repo.Path)
		ctx.state = dirty
	} else {
		ctx.state = clean
	}
}

func updateFetch(ctx *StateContext) {
	fetched := gitclient.Fetch(ctx.repo.Path)
	if !fetched {
		ctx.state = failed
		ctx.message = "Could not be fetched"
		return
	}
	if gitclient.IsEmpty(ctx.repo.Path) {
		ctx.state = failed
		ctx.message = "Empty git repository"
		return
	}
	ctx.behind = gitclient.GetBehindCount(ctx.repo.Path, ctx.ref)
	if ctx.behind == 0 {
		ctx.state = clean
		return
	}
	remotes := gitclient.IsRemoteExisting(ctx.repo.Path, ctx.ref)
	if !remotes {
		ctx.state = clean
		ctx.message = "No remote for the current branch"
		return
	}
	ctx.state = dirty
	ctx.message = gitclient.GetChanges(ctx.repo.Path, ctx.behind)
	return
}

func updatePull(ctx *StateContext) {
	success := gitclient.MergeFF(ctx.repo.Path)
	if !success {
		ctx.message = "Can not be merged, conflicting changes"
		ctx.state = failed
		return
	}
}

func printContext(ctx *StateContext) {
	ctx.mutex.Lock()
	var outState string
	switch ctx.state {
	case clean:
		if updateQuiet {
			return
		}
		outState = aurora.Green("✔").Bold().String()
	case dirty:
		outState = aurora.Yellow("●").Bold().String()
	case failed:
		outState = aurora.Red("✖").Bold().String()
	default:
		outState = "?"
	}

	// 80 chars for the separator, minus name of repo, minus spaces

	outSep := strings.Repeat("_", int(math.Max(0, (float64)(80-len(ctx.dirRelative)-1))))
	outBranch := aurora.Magenta(ctx.ref).String()
	outBehind := ""
	if ctx.behind > 0 {
		outBehind = "↓" + strconv.Itoa(ctx.behind)
	}
	say.ProgressGeneric(ctx.counter, ctx.total, outState, ctx.dirRelative, "%s (%s%s)", outSep, outBranch, outBehind)

	msg := strings.TrimSpace(ctx.message)
	if len(msg) > 0 {
		if ctx.state == failed {
			say.Raw(aurora.Red(msg).String() + "\n")
		} else {
			say.Raw(msg + "\n")
		}
	}
	ctx.mutex.Unlock()
}
