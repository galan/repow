package cmd

import (
	"fmt"
	"log"
	"net/http"
	"repo/internal/config"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/notification"
	"repo/internal/say"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the webserver for processing webhooks",
	Long:  `Starts the webserver for processing webhooks`,
	Args:  validateConditions(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cmd.Flags())
		startServer()
	},
}

func startServer() {
	initServer()

	// Can act as healthcheck/readiness
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
		fmt.Fprintf(w, "pong")
	})

	http.HandleFunc("/webhook/gitlab", func(w http.ResponseWriter, r *http.Request) {
		hoster, webhook, err := gitlab.HandleWebhookGitlab(w, r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if hoster == nil || webhook == nil {
			say.Verbose("hoster or webhook empty")
			return
		}
		// check default branch
		if "refs/heads/"+webhook.Project.DefaultBranch != webhook.Ref {
			w.Write([]byte(fmt.Sprintf("Skipping non-default branch %s for %s", webhook.Project.DefaultBranch, webhook.Project.Name)))
			return
		}

		go processWebhook(w, r, hoster, webhook.Project.Name, webhook.Project.PathWithNamespace, webhook.Ref)
	})

	beforeServer()

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Values.Server.Port), nil))
}

func initServer() {
	say.InfoLn("Starting repow %s server...", say.Repow())
}

func beforeServer() {
	say.InfoLn("Started server, listening on port %d", config.Values.Server.Port)
}

type WebHookRequest struct {
	RepoYaml model.RepoYaml
}

func isContactsOptional(r *http.Request) bool {
	urlValue := r.URL.Query().Get("optionalContacts")
	if urlValue == "true" {
		return true
	} else if urlValue == "false" {
		return false
	}
	return config.Values.Options.OptionalContacts
}

func processWebhook(w http.ResponseWriter, r *http.Request, hoster h.Hoster, name string, remotePath string, ref string) {
	say.InfoLn("Processing webhook for %s", remotePath)

	// fetch repo.yaml
	repoYaml, validYaml, err := hoster.DownloadRepoyaml(remotePath, ref)
	if err != nil {
		notification.NotifyInvalidRepository(remotePath, err.Error())
		say.Error("%s", err)
		msg := fmt.Sprintf("error: %v", err)
		w.Write([]byte(msg))
		return
	}

	repoRemote := model.MakeRepoRemote(remotePath, repoYaml, validYaml)

	// validate
	errs := hoster.Validate(repoRemote.RepoMeta, isContactsOptional(r))
	if errs != nil {
		notification.NotifyInvalidRepository(remotePath, fmt.Sprintf("%v", errs))
		say.Error("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		msg := fmt.Sprintf("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		w.Write([]byte(msg))
		return
	}

	say.Verbose("Repoyaml: %v", repoYaml)

	hoster.Apply(repoRemote.RepoMeta)
}
