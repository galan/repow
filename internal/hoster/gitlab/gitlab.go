package gitlab

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"

	"repo/internal/hoster"
	"repo/internal/model"
	"repo/internal/say"
	"repo/internal/util"

	gg "github.com/xanzy/go-gitlab"
)

const REPOW_GITLAB_API_TOKEN = "REPOW_GITLAB_API_TOKEN"
const GITLAB_API_TOKEN = "GITLAB_API_TOKEN"

func MakeHoster() (*Gitlab, error) {
	result := &Gitlab{}
	value := util.GetEnv(REPOW_GITLAB_API_TOKEN, util.GetEnv(GITLAB_API_TOKEN, ""))
	if value == "" {
		return result, errors.New("The " + REPOW_GITLAB_API_TOKEN + " or " + GITLAB_API_TOKEN + " environment-variable has to be set")
	}

	var errClient error
	result.client, errClient = gg.NewClient(value)
	if errClient != nil {
		return nil, errClient
	}
	return result, nil
}

type Gitlab struct {
	client *gg.Client
}

func (g Gitlab) Host() string {
	return "gitlab.com"
}

func (g Gitlab) Repositories(options hoster.RequestOptions) []hoster.HosterRepository {
	say.Info("Retrieving gitlab projects")
	projectOptions := &gg.ListProjectsOptions{
		ListOptions: gg.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Archived:   gg.Bool(false),
		Membership: gg.Bool(true),
		Starred:    &options.Starred,
		//TODO include topics here to reduce query time
	}

	var total int
	var repos []hoster.HosterRepository
	var lastResponse *gg.Response
	for ok := true; ok; ok = lastResponse.NextPage != 0 { // Loop through all pages and get list of projects
		say.Info(".")
		projectsPage, responsePage, err := g.client.Projects.ListProjects(projectOptions)
		lastResponse = responsePage
		if err != nil {
			say.Error("Failed retrieving response: %s", err)
			os.Exit(21) // unknown error behaviour, fail-fast
		}
		say.Verbose("\nPage: %d, Projects: %d, Statuscode: %d", projectOptions.Page, len(projectsPage), responsePage.StatusCode)
		for _, project := range projectsPage {
			total++

			if matches(options, project.PathWithNamespace, project.TagList) {
				repos = append(repos, hoster.HosterRepository{
					Id:     project.ID,
					Name:   project.Name,
					Topics: project.TagList,
					SshUrl: project.SSHURLToRepo})
			}
		}
		projectOptions.Page++
	}
	say.InfoLn(" %d retrieved (%d filtered)\n", len(repos), total-len(repos))
	return repos
}

func matchesPattern(value string, patterns []string, inverse bool) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, value)
		if err != nil {
			say.Error("Pattern matching failed unexpected for '%s' with %s", value, err)
			os.Exit(22) // fail-fast
		}
		if !inverse && !matched {
			return false
		}
	}
	return !inverse && true
}

func matches(options hoster.RequestOptions, path string, tags []string) bool {
	if !matchesPattern(path, options.IncludePatterns, false) {
		return false
	}
	if matchesPattern(path, options.ExcludePatterns, true) {
		return false
	}

	for _, topic := range options.Topics {
		var contains bool
		for _, remote := range tags {
			if topic == remote {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}

	return true
}

func (g Gitlab) ProjectState(projectPath string) (hoster.CleanupState, error) {
	say.Verbose("Retrieving gitlab project %s", projectPath)
	projectOptions := &gg.GetProjectOptions{}

	proj, response, err := g.client.Projects.GetProject(projectPath, projectOptions)
	if err != nil {
		if response == nil {
			return hoster.Unknown, errors.New("No response from gitlab")
		}
		if response.StatusCode == 404 {
			return hoster.Removed, nil
		}
		return hoster.Unknown, err
	}
	if response.StatusCode == 404 {
		return hoster.Removed, nil
	}
	if proj.Archived {
		return hoster.Archived, nil
	}
	return hoster.Ok, nil
}

func (g Gitlab) Validate(repo model.RepoMeta) []error {
	var errs []error
	// repo.yaml itself
	if repo.RepoYaml == nil {
		errs = append(errs, errors.New("No repo.yaml file exists"))
		return errs
	}
	if !repo.RepoYamlValid {
		errs = append(errs, errors.New("Invalid repo.yaml file"))
		return errs
	}

	pattern := `^[a-z][a-z0-9-]{0,99}$`

	// name
	if repo.Name != repo.RepoYaml.Name {
		errs = append(errs, errors.New("names do not match ("+repo.Name+" vs. "+repo.RepoYaml.Name+")"))
	}
	// language
	for _, lang := range repo.RepoYaml.Languages {
		errs = *validatePattern(lang, "Language", pattern, &errs)
	}
	// topics
	for _, topic := range repo.RepoYaml.Topics {
		errs = *validatePattern(topic, "Topic", pattern, &errs)
	}
	// orgs
	for orgUnit, orgName := range repo.RepoYaml.Org {
		errs = *validatePattern(orgUnit, "Organization unit", pattern, &errs)
		errs = *validatePattern(orgName, "Organization name", pattern, &errs)
	}
	// contacts
	if len(repo.RepoYaml.Contacts) == 0 {
		errs = append(errs, errors.New("No contacts provided"))
	} else {
		for _, contact := range repo.RepoYaml.Contacts {
			if !g.contactExists(repo.RemotePath, contact) {
				errs = append(errs, errors.New(fmt.Sprintf("User %s does not exists", contact)))
			}
		}
	}
	// gitlab features
	gl := repo.RepoYaml.Gitlab
	allowed := []string{"", "enabled", "private", "disabled"}
	if gl.WikiAccessLevel != nil && !util.IsInSlice(*gl.WikiAccessLevel, allowed...) {
		errs = append(errs, errors.New(fmt.Sprintf("WikiAccessLevel must be one of: %s", allowed)))
	}
	//TODO check more gitlab values
	//TODO check hardcoded arg passed values
	return errs
}

func validatePattern(value, descriptiveName, pattern string, errs *[]error) *[]error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		newErrs := append(*errs, errors.New(fmt.Sprintf("%s '%s' does not match pattern '%s'", descriptiveName, value, pattern)))
		return &newErrs
	}
	return errs
}

func (g Gitlab) contactExists(remotePath, contact string) bool {
	opt := &gg.ListProjectUserOptions{
		Search: &contact,
	}
	users, resp, err := g.client.Projects.ListProjectsUsers(remotePath, opt)
	if err != nil {
		say.Error("Unable to determine user: %s", err)
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	for _, user := range users {
		if user.Username == contact && user.State == "active" {
			return true
		}
	}
	return false
}

func (g Gitlab) DownloadRepoyaml(remotePath string, branch string) (*model.RepoYaml, bool, error) {
	say.Verbose("Downloading repo.yaml for project %s", remotePath)

	gfo := &gg.GetFileOptions{
		Ref: gg.String(branch),
	}
	file, response, err := g.client.RepositoryFiles.GetFile(remotePath, model.RepoYamlFilename, gfo)

	if response == nil {
		return nil, false, errors.New("No gitlab server response")
	}
	if response.StatusCode == 404 {
		return nil, false, nil
	}
	if response.StatusCode != 200 {
		say.Verbose("Unable to download repository manifest: %s", response.StatusCode)
		return nil, false, errors.New(fmt.Sprintf("Statuscode %d", response.StatusCode))
	}
	if err != nil {
		return nil, false, err
	}
	contentDecoded, errDecode := base64.StdEncoding.DecodeString(file.Content)
	if errDecode != nil {
		return nil, false, errors.New(fmt.Sprintf("Invalid base64: %s", file.Content))
	}

	result := &model.RepoYaml{}
	err = result.ReadFromByteArray(contentDecoded)
	if err != nil {
		say.Verbose("Invalid content (yaml): %s", result)
		return nil, false, nil
	}

	return result, true, nil
}

func (g Gitlab) Apply(repo model.RepoMeta) error {
	say.InfoLn("Apply %s", repo.Name)

	// topics
	var topics []string
	for _, topic := range repo.RepoYaml.Topics {
		topics = append(topics, topic)
	}
	for _, lang := range repo.RepoYaml.Languages {
		topics = append(topics, "lang_"+lang)
	}
	if repo.RepoYaml.Type != "" {
		topics = append(topics, "type_"+repo.RepoYaml.Type)
	}
	for k, v := range repo.RepoYaml.Org {
		topics = append(topics, "org_"+k+"_"+v)

	}

	var desc *string
	if repo.RepoYaml.Description != nil {
		desc = repo.RepoYaml.Description
	}

	epo := &gg.EditProjectOptions{
		TagList:     &topics,
		Description: desc,

		BuildTimeout:                              repo.RepoYaml.Gitlab.BuildTimeOut,
		OnlyAllowMergeIfPipelineSucceeds:          repo.RepoYaml.Gitlab.OnlyAllowMergeIfPipelineSucceeds,
		OnlyAllowMergeIfAllDiscussionsAreResolved: repo.RepoYaml.Gitlab.OnlyAllowMergeIfAllDiscussionsAreResolved,
		RemoveSourceBranchAfterMerge:              repo.RepoYaml.Gitlab.RemoveSourceBranchAfterMerge,
		SharedRunnersEnabled:                      repo.RepoYaml.Gitlab.SharedRunnersEnabled,
	}

	// gitlab features
	if repo.RepoYaml.Gitlab.WikiAccessLevel != nil {
		wal := gg.AccessControlValue(*repo.RepoYaml.Gitlab.WikiAccessLevel)
		epo.WikiAccessLevel = &wal
	}
	if repo.RepoYaml.Gitlab.IssuesAccessLevel != nil {
		ial := gg.AccessControlValue(*repo.RepoYaml.Gitlab.IssuesAccessLevel)
		epo.IssuesAccessLevel = &ial
	}

	project, response, err := g.client.Projects.EditProject(repo.RemotePath, epo)
	if err != nil {
		say.Error("%s", err)
	}
	say.InfoLn("%s %s %s", project, response, err)
	return nil
}
