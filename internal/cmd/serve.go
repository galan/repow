package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	h "repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

const envServerPort string = "REPOW_SERVER_PORT"
const defaultServerPort = "8080"

var serverPort string = defaultServerPort

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the webserver for processing webhooks",
	Long:  `Starts the webserver for processing webhooks`,
	Args:  validateConditions(cobra.NoArgs),
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func startServer() {
	initServer()

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

	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

func initServer() {
	say.InfoLn("Starting repow %s server...", say.Repow())
	port := os.Getenv(envServerPort)
	if port != "" {
		serverPort = port
	}
}

func beforeServer() {
	say.InfoLn("Started server, listening on port %s", serverPort)
}

type WebHookRequest struct {
	RepoYaml model.RepoYaml
}

func processWebhook(w http.ResponseWriter, r *http.Request, hoster h.Hoster, name string, remotePath string, ref string) {
	say.InfoLn("Processing webhook for %s", remotePath)
	// fetch repo.yaml
	repoYaml, validYaml, err := hoster.DownloadRepoyaml(remotePath, ref)
	if err != nil {
		say.Error("%s", err)
		msg := fmt.Sprintf("error: %v", err)
		w.Write([]byte(msg))
		return
	}

	repoRemote := model.MakeRepoRemote(remotePath, repoYaml, validYaml)

	// validate
	errs := hoster.Validate(repoRemote.RepoMeta)
	if errs != nil {
		say.Error("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		msg := fmt.Sprintf("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		w.Write([]byte(msg))
		return
	}

	say.Verbose("Repoyaml: %s", repoYaml)

	hoster.Apply(repoRemote.RepoMeta)
}
