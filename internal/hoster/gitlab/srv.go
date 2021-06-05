package gitlab

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"repo/internal/hoster"

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

const GITLAB_SECRET_TOKEN = "GITLAB_SECRET_TOKEN"

func HandleWebhookGitlab(w http.ResponseWriter, r *http.Request) (hoster.Hoster, *WebHookPush, error) {
	eventType := r.Header.Get("X-Gitlab-Event")
	if eventType != "Push Hook" {
		w.Write([]byte("ignored"))
		return nil, nil, nil
	}

	if !matchesSecurityToken(r) {
		return nil, nil, errors.New("security-token does not match")
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
	if secretToken == "" {
		return true
	}
	value, exists := os.LookupEnv(GITLAB_SECRET_TOKEN)
	return exists && value == secretToken
}
