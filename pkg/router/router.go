package router

import (
	"context"
	"log"
	"net/http"

	paulclient "github.com/Spazzy757/paul/pkg/client"
	"github.com/Spazzy757/paul/pkg/github"
	"github.com/gorilla/mux"
)

//GetRouter .
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", GithubWebHookHandler)
	return r
}

//GithubWebHookHandler .
func GithubWebHookHandler(w http.ResponseWriter, r *http.Request) {
	installationIDHeader := r.Header.Get("X-GitHub-Hook-Installation-Target-ID")
	gClient, err := paulclient.GetClient(installationIDHeader)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	githubClient := &paulclient.GithubClient{
		Ctx:                context.Background(),
		GitService:         gClient.Git,
		RepoService:        gClient.Repositories,
		PullRequestService: gClient.PullRequests,
		IssueService:       gClient.Issues,
	}
	err = github.IncomingWebhook(r, githubClient)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
