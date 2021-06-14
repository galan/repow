package gitlab

import (
	"errors"
	"io/ioutil"
	"net/http"
	"repo/internal/hoster"
	"repo/internal/util"

	"gopkg.in/yaml.v2"
)

type WebHookPush struct {
	Ref     string `yaml:"ref"`
	Project struct {
		Name              string `yaml:"name"`
		PathWithNamespace string `yaml:"path_with_namespace"`
		DefaultBranch     string `yaml:"default_branch"`
	} `yaml:"project"`
}

const REPOW_GITLAB_SECRET_TOKEN = "REPOW_GITLAB_SECRET_TOKEN"
const GITLAB_SECRET_TOKEN = "GITLAB_SECRET_TOKEN"

func HandleWebhookGitlab(w http.ResponseWriter, r *http.Request) (hoster.Hoster, *WebHookPush, error) {
	if !matchesSecurityToken(r) {
		return nil, nil, errors.New("security-token does not match")
	}

	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType != "Push Hook" {
		w.Write([]byte("ignored"))
		return nil, nil, nil
	}

	hoster, err := MakeHoster()
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	//create some struct from webhook
	push := WebHookPush{}
	yaml.Unmarshal(body, &push)

	return hoster, &push, nil

}

func matchesSecurityToken(r *http.Request) bool {
	// docs: https://docs.gitlab.com/ce/user/project/integrations/webhooks.html#secret-token
	secretToken := r.Header.Get("X-Gitlab-Token") // if configured
	value := util.GetEnv(REPOW_GITLAB_SECRET_TOKEN, util.GetEnv(GITLAB_SECRET_TOKEN, ""))
	return value == secretToken
}
