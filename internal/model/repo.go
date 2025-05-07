package model

import (
	"bytes"
	"os"
	"os/exec"
	"path"
	"regexp"
	"repo/internal/say"
	"repo/internal/util"
	"strings"
)

type RepoMeta struct {
	RepoYaml      *RepoYaml // The parsed repo.yaml model, nil if not exists or not parseable
	RepoYamlValid bool      // The repo.yaml couldn't be parsed correctly
	RemotePath    string    // The parsed remote directory of the repository
	Name          string    // The name of the repository
	//Path     string   // The absolute path to the repository
	//RepoYamlFile string   // Absolute path to repo.yaml
}

type RepoDir struct {
	RepoMeta
	Path string // The absolute path to the repository
}

type RepoRemote struct {
	RepoMeta
}

func (rd RepoDir) RepoYamlFilename() string {
	return path.Join(rd.Path, RepoYamlFilename)
}

func (rd RepoDir) PathDirName() string {
	//p, _ := filepath.Abs(rd.Path) // resolve potential relative path to absolute
	fiRepo, _ := os.Stat(rd.Path)
	return fiRepo.Name()
}

func MakeRepoDir(pathRepository string, hosterHost string) (*RepoDir, error) {
	result := &RepoDir{Path: pathRepository}
	result.RemotePath = DetermineRemotePath(pathRepository, hosterHost)
	result.Name = result.PathDirName()

	if util.ExistsFile(result.RepoYamlFilename()) {
		result.RepoYaml = &RepoYaml{}
		err := result.RepoYaml.ReadFromFile(result.RepoYamlFilename())
		result.RepoYamlValid = err == nil
		if err != nil {
			say.Verbose("Failed parsing file %s: %s", result.RepoYamlFilename(), err)
		}
	}
	return result, nil // TODO errors required?
}

func MakeRepoRemote(remotePath string, repoYaml *RepoYaml, validYaml bool) *RepoRemote {
	result := &RepoRemote{}
	result.RemotePath = remotePath
	result.RepoYaml = repoYaml
	result.RepoYamlValid = validYaml
	result.Name = determineName(remotePath)
	return result
}

func determineName(remotePath string) string {
	split := strings.Split(remotePath, "/")
	return split[len(split)-1]
}

func DetermineRemotePath(pathRepository string, hosterHost string) string {
	bufferOut := new(bytes.Buffer)
	bufferErr := new(bytes.Buffer)
	cmdGo := exec.Command("git", "remote", "-v")
	cmdGo.Dir = pathRepository
	cmdGo.Stdout = bufferOut
	cmdGo.Stderr = bufferErr
	err := cmdGo.Run()
	if err != nil {
		say.Error("git remote failed for %s: %s", pathRepository, err)
		say.Error("%s", bufferErr.String())
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(bufferOut.String(), "\r\n", "\n"), "\n")

	var result string
	for _, line := range lines {
		result = ParseRemotePath(line, hosterHost)
		if len(result) > 0 {
			break
		}
	}
	return result
}

// TODO distinguish remote url notations, improve this approach
func ParseRemotePath(path string, hosterHost string) string {
	var result string
	re, _ := regexp.Compile(`^origin[\t ]+((https|ssh):\/\/.*@?|git@)` + hosterHost + `[\/:]([a-zA-Z0-9_\/.-]+)[\t ]+.fetch.$`)
	matches := re.MatchString(path)
	say.Verbose("Checking remote: %s, matches: %v", path, matches)
	if matches {
		result = re.FindStringSubmatch(path)[3]
		result = strings.TrimSuffix(result, ".git") // greedy regex has issues in golang, workaround
	}
	return result
}
