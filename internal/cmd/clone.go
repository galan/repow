package cmd

import (
	"os"
	"path"
	"repo/internal/gitclient"
	"repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/say"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var cloneTopics []string
var cloneExcludePatterns []string
var cloneIncludePatterns []string

var cloneParallel bool
var cloneStarred bool

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringSliceVarP(&cloneTopics, "topic", "t", nil, "Topics (aka tags/labels) to be filtered. Multiple topics are possible (and).")
	cloneCmd.Flags().StringSliceVarP(&cloneExcludePatterns, "exclude", "e", nil, "Regex-pattern not to be matched for the path. Multiple patterns are possible (and).")
	cloneCmd.Flags().StringSliceVarP(&cloneIncludePatterns, "include", "i", nil, "Regex-pattern that needs to be matched for the path. Multiple patterns are possible (and).")
	cloneCmd.Flags().BoolVarP(&cloneParallel, "parallel", "p", true, "Process operations parallel")
	cloneCmd.Flags().BoolVarP(&cloneStarred, "starred", "s", false, "Filter for starred projects")
}

var cloneCmd = &cobra.Command{
	Use:   "clone [root-dir]",
	Short: "Clones selected repositories to the passed location. Adds new ones on reoccurring calls.",
	Long:  `Clones selected repositories to the passed location. Adds new ones on reoccurring calls.`,
	Args:  validateConditions(cobra.ExactArgs(1), validateArgGitDir(0, false, true)),
	Run: func(cmd *cobra.Command, args []string) {
		defer func(start time.Time) { say.InfoLn("🦊 Finished, took %s", time.Since(start)) }(time.Now())
		dirReposRoot := args[0]

		provider, err := gitlab.MakeProvider()
		handleFatalError(err)

		gitclient.PrepareSsh(provider.Host()) // TODO parse url?
		repos := provider.Repositories(hoster.RequestOptions{
			Topics:          cloneTopics,
			Starred:         cloneStarred,
			ExcludePatterns: cloneExcludePatterns,
			IncludePatterns: cloneIncludePatterns,
		})
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Name < repos[j].Name
		})

		repos = filterExisting(dirReposRoot, repos)
		cloneAll(dirReposRoot, repos)
	},
}

func filterExisting(dirReposRoot string, repos []hoster.ProviderRepository) (result []hoster.ProviderRepository) {
	for _, r := range repos {
		dirName := determineDirectoryName(r.SshUrl)
		dirRepository := path.Join(dirReposRoot, dirName)

		_, err := os.Stat(dirRepository)
		if os.IsNotExist(err) {
			result = append(result, r)
		} else {
			say.Verbose("Repository exists already: %s ", dirName)
		}
	}
	return
}

func determineDirectoryName(sshUrl string) string {
	last := sshUrl[strings.LastIndex(sshUrl, "/")+1:]
	if strings.HasSuffix(last, ".git") {
		last = last[:len(last)-4]
	}
	return last
}

func cloneAll(dirReposRoot string, repos []hoster.ProviderRepository) {
	counter := int32(0)
	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		if cloneParallel {
			go clone(dirReposRoot, repo.SshUrl, &counter, len(repos), false, &wg)
		} else {
			clone(dirReposRoot, repo.SshUrl, &counter, len(repos), true, &wg)
		}
	}
	wg.Wait()
}

func clone(dirReposRoot string, sshUrl string, counter *int32, total int, verbose bool, wg *sync.WaitGroup) error {
	defer wg.Done()
	dirName := determineDirectoryName(sshUrl)
	err := gitclient.Clone(dirReposRoot, dirName, sshUrl, verbose)
	if err != nil {
		say.ProgressError(counter, total, err, dirName, "- Unable to clone")
	} else {
		say.ProgressSuccess(counter, total, dirName, "")
	}
	return err
}
