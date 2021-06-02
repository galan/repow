package cmd

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"repo/internal/hoster"
	"repo/internal/hoster/gitlab"
	"repo/internal/model"
	"repo/internal/say"

	"github.com/spf13/cobra"
)

const envServerPort string = "REPOW_SERVER_PORT"
const defaultServerPort = "8081"

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
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/webhook/gitlab", func(w http.ResponseWriter, r *http.Request) {
		provider, webhook, err := gitlab.HandleWebhookGitlab(w, r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		if provider == nil || webhook == nil {
			say.Verbose("provider or webhook empty")
			return
		}
		// check default branch
		if "refs/heads/"+webhook.Project.DefaultBranch != webhook.Ref {
			w.Write([]byte(fmt.Sprintf("Skipping non-default branch %s for %s", webhook.Project.DefaultBranch, webhook.Project.Name)))
			return
		}

		processWebhook(w, r, provider, webhook.Project.Name, webhook.Project.PathWithNamespace, webhook.Ref)
	})

	beforeServer()

	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

func initServer() {
	say.InfoLn("Starting server...")
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

func processWebhook(w http.ResponseWriter, r *http.Request, provider hoster.Hoster, name string, remotePath string, ref string) {
	// fetch repo.yaml
	repoYaml, validYaml, err := provider.DownloadRepoyaml(remotePath, ref)
	if err != nil {
		say.Error("%s", err)
		msg := fmt.Sprintf("error: %v", err)
		w.Write([]byte(msg))
		return
	}

	repoRemote := model.MakeRepoRemote(remotePath, repoYaml, validYaml)

	// validate
	errs := provider.Validate(repoRemote.RepoMeta)
	if errs != nil {
		say.Error("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		msg := fmt.Sprintf("Repository manifest for %s is not valid: %s", repoRemote.RemotePath, errs)
		w.Write([]byte(msg))
		return
	}

	say.InfoLn("Repoyaml: %s", repoYaml)

	provider.Apply(repoRemote.RepoMeta)
}
