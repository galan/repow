package gitclient

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"repo/internal/say"
	"repo/internal/util"
	"strconv"
	"strings"
)

func PrepareSsh(host string) {
	// not sure how to improve this.
	// maybe load ssh config if available and determine identity for host?
	_, e, code := util.RunCommandDir(nil, "ssh", "git@"+host)
	if code != 0 {
		say.Error("Failed loading ssh key: %s", e)
		os.Exit(3)
	}
}

func Clone(rootDir string, repoDir string, sshUrl string) error {
	dirRepository := path.Join(rootDir, repoDir)
	cmdGo := exec.Command("git", "clone", sshUrl, dirRepository)
	if say.VerboseEnabled {
		cmdGo.Stdout = os.Stdout
		cmdGo.Stderr = os.Stderr
	} else {
		cmdGo.Stdout = new(bytes.Buffer)
		cmdGo.Stderr = new(bytes.Buffer)
	}
	err := cmdGo.Run()
	if err != nil {
		say.Verbose("git clone failed with %s", err)
		return err
	}
	return nil
}

func IsDirty(repoDir string) bool {
	o, _, _ := util.RunCommandDir(&repoDir, "git", "status", "--porcelain")
	return len(o) > 0
}

// get changes in short form
func GetLocalChanges(repoDir string) string {
	o, _, _ := util.RunCommandDir(&repoDir, "git", "-c", "color.ui=always", "status", "-s")
	return strings.TrimSpace(o)
}

func IsEmpty(repoDir string) bool {
	fis, err := ioutil.ReadDir(path.Join(repoDir, ".git/objects"))
	if err != nil {
		return true
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			return true
		}
	}
	return false
	//o, _, _ := util.RunCommandDir(&repoDir, "find", ".git/objects", "-type", "f")
	//return len(o) == 0
}

func GetCurrentBranch(repoDir string) string {
	if IsEmpty(repoDir) {
		return "-"
	}
	o, _, _ := util.RunCommandDir(&repoDir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimSpace(o)
}

func GetBehindCount(repoDir string, branch string) int {
	o, _, _ := util.RunCommandDir(&repoDir, "git", "rev-list", "HEAD...origin/"+branch, "--count")
	count, _ := strconv.Atoi(strings.TrimSpace(o))
	return count
}

func GetChanges(repoDir string, behind int) string {
	o, _, _ := util.RunCommandDir(&repoDir, "git", "-c", "color.ui=always", "--no-pager", "log", "--format=%C(yellow)%h%Creset %C(blue)%ar%Creset%C(green)%d%Creset %s %C(dim normal)(%an)%Creset", "-n", strconv.Itoa(behind))
	return strings.TrimSpace(o)
}

func IsRemoteExisting(repoDir string, ref string) bool {
	o, _, _ := util.RunCommandDir(&repoDir, "git", "ls-remote", ".", "refs/remotes/origin/"+ref)
	return len(o) >= 0
}

func Fetch(repoDir string) bool {
	_, _, code := util.RunCommandDir(&repoDir, "git", "fetch", "-q")
	return code == 0
}

func MergeFF(repoDir string) bool {
	_, _, code := util.RunCommandDir(&repoDir, "git", "merge", "FETCH_HEAD", "--ff")
	return code == 0
}
