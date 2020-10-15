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
	ctx := context.Background()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = github.IncomingWebhook(ctx, r, gClient)
	if err != nil {
		log.Printf("Error: %v", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
