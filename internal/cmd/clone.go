package cmd

import (
	"os"
	"path"
	"repo/internal/gitclient"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/say"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var cloneStyle string
var cloneTopics []string
var cloneExcludePatterns []string
var cloneIncludePatterns []string

var cloneParallelism int
var cloneStarred bool

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVarP(&cloneStyle, "style", "y", "flat", "Either repositories are cloned 'flat' into the root-dir, or 'recursive' using the groups as directories.")
	cloneCmd.Flags().StringSliceVarP(&cloneTopics, "topic", "t", nil, "Topics (aka tags/labels) to be filtered. Multiple topics are possible (and).")
	cloneCmd.Flags().StringSliceVarP(&cloneExcludePatterns, "exclude", "e", nil, "Regex-pattern not to be matched for the path. Multiple patterns are possible (and).")
	cloneCmd.Flags().StringSliceVarP(&cloneIncludePatterns, "include", "i", nil, "Regex-pattern that needs to be matched for the path. Multiple patterns are possible (and).")
	cloneCmd.Flags().IntVarP(&cloneParallelism, "parallelism", "p", 32, "How many process should run in parallel, 1 would be no parallelism.")
	cloneCmd.Flags().BoolVarP(&cloneStarred, "starred", "s", false, "Filter for starred projects")
}

var cloneCmd = &cobra.Command{
	Use:   "clone [root-dir]",
	Short: "Clones selected repositories to the passed location. Adds new ones on reoccurring calls.",
	Long:  `Clones selected repositories to the passed location. Adds new ones on reoccurring calls.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, false, true)),
	Run: func(cmd *cobra.Command, args []string) {
		defer say.Timer(time.Now())
		dirReposRoot := getAbsoluteRepoRoot(args[0])

		hoster, err := gitlab.MakeHoster()
		handleFatalError(err)

		gitclient.PrepareSsh(hoster.Host()) // TODO parse url?
		repos := hoster.Repositories(h.RequestOptions{
			Topics:          cloneTopics,
			Starred:         cloneStarred,
			ExcludePatterns: cloneExcludePatterns,
			IncludePatterns: cloneIncludePatterns,
		})
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].PathWithNamespace < repos[j].PathWithNamespace
		})

		repos = filterExisting(dirReposRoot, repos)
		cloneAll(dirReposRoot, repos)
	},
	PreRunE: validateFlags,
}

func filterExisting(dirReposRoot string, repos []h.HosterRepository) (result []h.HosterRepository) {
	for _, r := range repos {

		var dirRepository string
		if Style == styleFlat {
			dirRepository = path.Join(dirReposRoot, r.Path)
		} else if Style == styleRecursive {
			dirRepository = path.Join(dirReposRoot, r.PathWithNamespace)
		}

		_, err := os.Stat(dirRepository)
		if os.IsNotExist(err) {
			result = append(result, r)
		} else {
			say.Verbose("Repository exists already: %s ", r.PathWithNamespace)
		}
	}
	return
}

func cloneAll(dirReposRoot string, repos []h.HosterRepository) {
	tasks := make(chan h.HosterRepository)
	var wg sync.WaitGroup
	counter := int32(0)
	for i := 0; i < getParallelism(cloneParallelism); i++ {
		wg.Add(1)
		go clone(dirReposRoot, &counter, len(repos), tasks, &wg)
	}

	for _, repo := range repos {
		tasks <- repo
	}

	close(tasks)
	wg.Wait()
}

func clone(dirReposRoot string, counter *int32, total int, tasks chan h.HosterRepository, wg *sync.WaitGroup) {
	defer wg.Done()
	for repo := range tasks {

		var dirTarget string
		if Style == styleFlat {
			dirTarget = repo.Path
		} else if Style == styleRecursive {
			dirTarget = repo.PathWithNamespace
		}

		err := gitclient.Clone(dirReposRoot, dirTarget, repo.SshUrl)
		if err != nil {
			say.ProgressError(counter, total, err, repo.PathWithNamespace, "- Unable to clone")
		} else {
			say.ProgressSuccess(counter, total, dirTarget, "")
		}
	}
}
