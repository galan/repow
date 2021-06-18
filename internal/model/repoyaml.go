package model

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const RepoYamlFilename string = "repo.yaml"

/*
// Unicorning blup
type Unicorning struct {
	Unit     string `yaml:"unit"`
	Vertical string `yaml:"vertical"`
}

type Org struct {
	Team     string `yaml:"team"`
	Squad    string `yaml:"squad"`
	Chapter  string `yaml:"chapter"`
	Unicorn  string `yaml:"unicorn"`
	Vertical string `yaml:"vertical"`
}
*/

type Gitlab struct {
	WikiAccessLevel    *string `yaml:"wiki_access_level"`
	IssuesAccessLevel  *string `yaml:"issues_access_level"`
	ForkingAccessLevel *string `yaml:"forking_access_level"`
	BuildTimeOut       *int    `yaml:"build_timeout"`

	OnlyAllowMergeIfPipelineSucceeds          *bool `yaml:"only_allow_merge_if_pipeline_succeeds"`
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool `yaml:"only_allow_merge_if_all_discussions_are_resolved"`
	RemoveSourceBranchAfterMerge              *bool `yaml:"remove_source_branch_after_merge"`
	SharedRunnersEnabled                      *bool `yaml:"shared_runners_enabled"`
}

// Repo bla
type RepoYaml struct {
	Name        string            `yaml:"name"`
	Description *string           `yaml:"description"`
	Type        string            `yaml:"type"`
	Languages   []string          `yaml:"languages"`
	Topics      []string          `yaml:"topics"`
	Org         map[string]string `yaml:"org"`
	Annotations map[string]string `yaml:"annotations"`
	Contacts    []string          `yaml:"contacts"`
	Gitlab      Gitlab            `yaml:"gitlab"`
	//Github
}

func (r *RepoYaml) ReadFromString(content string) error {
	return r.ReadFromByteArray([]byte(content))
}

func (r *RepoYaml) ReadFromByteArray(content []byte) error {
	return yaml.Unmarshal(content, r)
}

func (r *RepoYaml) ReadFromFile(filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	//defer yamlFile.Close() //TODO required?

	return yaml.Unmarshal(yamlFile, r)
}
