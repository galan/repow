package hoster

import (
	"fmt"
	"repo/internal/model"
)

type RequestOptions struct {
	Topics          []string
	Starred         bool
	ExcludePatterns []string
	IncludePatterns []string
}

type Hoster interface {
	Repositories(options RequestOptions) []HosterRepository
	ProjectState(projectPath string) (CleanupState, error)
	Host() string
	Validate(repo model.RepoMeta) []error
	DownloadRepoyaml(remotePath string, ref string) (*model.RepoYaml, bool, error)
	Apply(repo model.RepoMeta) error
}

type HosterRepository struct {
	Id     int
	Name   string
	Topics []string
	SshUrl string
}

type CleanupState int

const (
	Unknown CleanupState = iota
	Ok
	Removed
	Archived
)

func (state CleanupState) String() string {
	switch state {
	case Unknown:
		return "Unknown"
	case Ok:
		return "Ok"
	case Removed:
		return "Removed"
	case Archived:
		return "Archived"
	default:
		return fmt.Sprintf("%d", int(state))
	}
}
